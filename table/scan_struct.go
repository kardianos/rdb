// Copyright 2014 Daniel Theophanes.
// Use of this source code is governed by a zlib-style
// license that can be found in the LICENSE file.

package table

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"iter"
	"math/big"
	"reflect"
	"strings"
	"time"
	"unsafe"

	"github.com/kardianos/rdb"
)

// Query runs cmd on q and scans every row of the first result set into a new []T.
// T must be a struct or a pointer to a struct. Column mapping uses the "db" tag
// (or the field name), same rules as UnmarshalStruct.
//
// Non-nullable scalar fields are populated via Result.Prep / Scan so the driver
// can DirectAssign without interface{} boxing.
//
// Nullable columns without boxing:
//   - rdb.Opt[T] fields (Valid=false for NULL)
//   - plain field + bool with null:"colname" (flag true for NULL)
//   - pointer fields (*T): nil for NULL (still uses a Nullable scratch today)
//   - plain field only on a nullable column: zero value for NULL (Nullable scratch)
//
// io.Writer fields receive bytes via Write without the API retaining a []byte.
//
// Field offsets and typed prep/apply funcs are built once per result set; the
// row loop does not walk the struct with reflect.Value.
//
// Use Stream if you do not need to retain every row.
func Query[T any](ctx context.Context, q rdb.Queryer, cmd *rdb.Command, params ...rdb.Param) ([]T, error) {
	return QueryTag[T](ctx, q, cmd, "db", params...)
}

// QueryTag is Query with an explicit struct tag name (empty means "db").
func QueryTag[T any](ctx context.Context, q rdb.Queryer, cmd *rdb.Command, tagName string, params ...rdb.Param) ([]T, error) {
	res, err := q.Query(ctx, cmd, params...)
	if err != nil {
		return nil, err
	}
	defer res.Close()

	plan, err := newStructPlan[T](res.Schema(), tagName)
	if err != nil {
		return nil, err
	}

	var out []T
	for res.Next() {
		out = append(out, *new(T))
		if err := plan.scan(res, unsafe.Pointer(&out[len(out)-1])); err != nil {
			return out, err
		}
	}
	return out, nil
}

// Stream runs cmd on q and yields each row of the first result set as T.
// The query starts when iteration begins; the result is closed when iteration
// ends (including early break). On failure, yield is invoked once with the
// zero T and the error.
//
//	for row, err := range table.Stream[MyRow](ctx, db, cmd) {
//	    if err != nil {
//	        return err
//	    }
//	    // use row
//	}
func Stream[T any](ctx context.Context, q rdb.Queryer, cmd *rdb.Command, params ...rdb.Param) iter.Seq2[T, error] {
	return StreamTag[T](ctx, q, cmd, "db", params...)
}

// StreamTag is Stream with an explicit struct tag name (empty means "db").
func StreamTag[T any](ctx context.Context, q rdb.Queryer, cmd *rdb.Command, tagName string, params ...rdb.Param) iter.Seq2[T, error] {
	return func(yield func(T, error) bool) {
		var zero T
		res, err := q.Query(ctx, cmd, params...)
		if err != nil {
			yield(zero, err)
			return
		}
		defer res.Close()

		plan, err := newStructPlan[T](res.Schema(), tagName)
		if err != nil {
			yield(zero, err)
			return
		}

		for res.Next() {
			var row T
			if err := plan.scan(res, unsafe.Pointer(&row)); err != nil {
				yield(zero, err)
				return
			}
			if !yield(row, nil) {
				return
			}
		}
	}
}

type fieldMode uint8

const (
	modeDirect fieldMode = iota // Prep into field or Opt[T]; DirectAssign
	modeFlag                    // Prep via NullFlagPrep (payload + null:"…" bool)
	modeNull                    // Prep into Nullable scratch; apply after Scan
	modeJSON                    // GetxN after Scan; json.Unmarshal
)

type fieldBind struct {
	colIdx  int
	offset  uintptr
	mode    fieldMode
	nullIdx int // index into nullScratch when modeNull

	// Built once in newStructPlan; used every row without reflect.Value field walks.
	prep      func(base unsafe.Pointer) any                   // modeDirect
	flagSink  *rdb.NullFlagPrep                               // modeFlag; rebinding each row
	flagPrep  func(base unsafe.Pointer) any                   // modeFlag; returns flagSink after rebind
	applyNull func(base unsafe.Pointer, n rdb.Nullable) error // modeNull
	applyJSON func(base unsafe.Pointer, n rdb.Nullable) error // modeJSON
}

type structPlan struct {
	tType       reflect.Type // underlying struct type
	asPtr       bool         // T is *Struct
	fields      []fieldBind
	nullScratch []rdb.Nullable
}

func isOptType(ft reflect.Type) bool {
	// Instantiated generics report Name as "Opt[string]", not "Opt".
	if ft.Kind() != reflect.Struct || ft.NumField() != 2 {
		return false
	}
	if ft.Field(0).Name != "V" || ft.Field(1).Name != "Valid" || ft.Field(1).Type.Kind() != reflect.Bool {
		return false
	}
	name := ft.Name()
	if name != "Opt" && !strings.HasPrefix(name, "Opt[") {
		return false
	}
	if ft.PkgPath() == "github.com/kardianos/rdb" {
		return true
	}
	return strings.Contains(ft.String(), "rdb.Opt[")
}

func newStructPlan[T any](schema []*rdb.Column, tagName string) (*structPlan, error) {
	if len(tagName) == 0 {
		tagName = "db"
	}
	tType := reflect.TypeOf((*T)(nil)).Elem()
	asPtr := false
	if tType.Kind() == reflect.Ptr {
		asPtr = true
		tType = tType.Elem()
	}
	if tType.Kind() != reflect.Struct {
		return nil, fmt.Errorf("table: type %s must be a struct or *struct", reflect.TypeOf((*T)(nil)).Elem())
	}

	nameIndex := make(map[string]int, len(schema))
	colNullable := make([]bool, len(schema))
	for i, col := range schema {
		if col == nil {
			continue
		}
		nameIndex[col.Name] = i
		colNullable[i] = col.Nullable
	}

	// field name / column name → field index for null:"…" pairing
	fieldByName := make(map[string]int, tType.NumField())
	fieldByColName := make(map[string]int, tType.NumField())
	// nullTag target (column or field name) → bool field index
	nullFlags := make(map[string]int) // key: target column name or field name

	for i := 0; i < tType.NumField(); i++ {
		f := tType.Field(i)
		if !f.IsExported() {
			continue
		}
		fieldByName[f.Name] = i
		if nt := f.Tag.Get("null"); nt != "" {
			if f.Type.Kind() != reflect.Bool {
				return nil, fmt.Errorf("table: field %s has null:%q but is not bool", f.Name, nt)
			}
			if _, dup := nullFlags[nt]; dup {
				return nil, fmt.Errorf("table: duplicate null:%q", nt)
			}
			nullFlags[nt] = i
		}
		// Record db column name for this field when present.
		colName := f.Name
		if dbTag := f.Tag.Get(tagName); dbTag != "" {
			parts := strings.Split(dbTag, ",")
			if parts[0] != "" && parts[0] != "-" {
				colName = parts[0]
			}
			if parts[0] == "-" {
				continue
			}
		}
		fieldByColName[colName] = i
	}

	// Resolve null flag targets to payload field index + flag field index.
	type flagPair struct {
		payloadIdx int
		flagIdx    int
	}
	// colIdx → flag pair (for payload fields that have a null bool)
	flagByPayload := make(map[int]flagPair)

	for target, flagIdx := range nullFlags {
		payloadIdx, ok := fieldByColName[target]
		if !ok {
			payloadIdx, ok = fieldByName[target]
		}
		if !ok {
			return nil, fmt.Errorf("table: null:%q does not match any field or column", target)
		}
		if payloadIdx == flagIdx {
			return nil, fmt.Errorf("table: null:%q cannot refer to itself", target)
		}
		pf := tType.Field(payloadIdx)
		if isOptType(pf.Type) {
			return nil, fmt.Errorf("table: null:%q target field %s is Opt[T]; use one null mechanism", target, pf.Name)
		}
		if pf.Tag.Get("null") != "" {
			return nil, fmt.Errorf("table: null:%q target %s is itself a null flag", target, pf.Name)
		}
		flagByPayload[payloadIdx] = flagPair{payloadIdx: payloadIdx, flagIdx: flagIdx}
	}

	p := &structPlan{tType: tType, asPtr: asPtr}
	// Skip bool fields that are only null flags (not columns).
	nullFlagFields := make(map[int]bool)
	for _, fp := range flagByPayload {
		nullFlagFields[fp.flagIdx] = true
	}

	for i := 0; i < tType.NumField(); i++ {
		f := tType.Field(i)
		if !f.IsExported() {
			continue
		}
		if nullFlagFields[i] {
			continue // consumed as side channel of payload field
		}
		// Pure null-tagged field without a resolved payload already errored above
		// only if target missing; a null tag always pairs. Fields with only null
		// tag and no db mapping are the flag fields we skip.

		columnName := f.Name
		var isJSON bool
		if dbTag := f.Tag.Get(tagName); dbTag != "" {
			parts := strings.Split(dbTag, ",")
			if parts[0] != "" {
				columnName = parts[0]
			}
			for _, part := range parts[1:] {
				if part == "json" {
					isJSON = true
					break
				}
			}
		}
		if columnName == "-" {
			continue
		}
		// Skip fields that are only null flags (have null tag, no independent column).
		if f.Tag.Get("null") != "" {
			continue
		}

		colIdx, ok := nameIndex[columnName]
		if !ok {
			return nil, fmt.Errorf("table: field %s (column %q) not found in result schema", f.Name, columnName)
		}

		b := fieldBind{colIdx: colIdx, offset: f.Offset}
		switch {
		case isJSON:
			if f.Type.Kind() != reflect.Slice && f.Type.Kind() != reflect.Array {
				return nil, fmt.Errorf("table: field %s with json tag must be a slice or array", f.Name)
			}
			b.mode = modeJSON
			fn, err := makeJSONApply(f.Type, f.Offset)
			if err != nil {
				return nil, err
			}
			b.applyJSON = fn
		case isOptType(f.Type):
			// Opt[T]: Prep *Opt[T] directly (DirectAssign sets Valid).
			b.mode = modeDirect
			fn, err := makeDirectPrep(f.Type, f.Offset)
			if err != nil {
				return nil, fmt.Errorf("table: field %s: %w", f.Name, err)
			}
			b.prep = fn
		default:
			if pair, hasFlag := flagByPayload[i]; hasFlag {
				b.mode = modeFlag
				flagOff := tType.Field(pair.flagIdx).Offset
				sink := &rdb.NullFlagPrep{}
				b.flagSink = sink
				valPrep, err := makeDirectPrep(f.Type, f.Offset)
				if err != nil {
					return nil, fmt.Errorf("table: field %s: %w", f.Name, err)
				}
				b.flagPrep = func(base unsafe.Pointer) any {
					sink.Value = valPrep(base)
					sink.Null = (*bool)(unsafe.Add(base, flagOff))
					return sink
				}
				break
			}
			if f.Type.Kind() == reflect.Ptr || colNullable[colIdx] {
				// Legacy path: Nullable scratch (may box non-null values).
				b.mode = modeNull
				b.nullIdx = len(p.nullScratch)
				p.nullScratch = append(p.nullScratch, rdb.Nullable{})
				fn, err := makeNullApply(f.Type, f.Offset)
				if err != nil {
					return nil, fmt.Errorf("table: field %s: %w", f.Name, err)
				}
				b.applyNull = fn
			} else {
				b.mode = modeDirect
				fn, err := makeDirectPrep(f.Type, f.Offset)
				if err != nil {
					return nil, fmt.Errorf("table: field %s: %w", f.Name, err)
				}
				b.prep = fn
			}
		}
		p.fields = append(p.fields, b)
	}
	return p, nil
}

// rowBase returns a pointer to the struct value for this row.
// dest points at T (either Struct or *Struct).
func (p *structPlan) rowBase(dest unsafe.Pointer) unsafe.Pointer {
	if !p.asPtr {
		return dest
	}
	slot := (*unsafe.Pointer)(dest)
	if *slot == nil {
		*slot = reflect.New(p.tType).UnsafePointer()
	}
	return *slot
}

// scan populates the row at dest (pointer to T) from the current result row.
func (p *structPlan) scan(res *rdb.Result, dest unsafe.Pointer) error {
	base := p.rowBase(dest)

	for i := range p.fields {
		b := &p.fields[i]
		switch b.mode {
		case modeDirect:
			res.Prepx(b.colIdx, b.prep(base))
		case modeFlag:
			res.Prepx(b.colIdx, b.flagPrep(base))
		case modeNull:
			p.nullScratch[b.nullIdx] = rdb.Nullable{}
			res.Prepx(b.colIdx, &p.nullScratch[b.nullIdx])
		case modeJSON:
			// filled after Scan via GetxN
		}
	}

	if err := res.Scan(); err != nil {
		return err
	}

	for i := range p.fields {
		b := &p.fields[i]
		switch b.mode {
		case modeDirect, modeFlag:
			// already written by DirectAssign
		case modeNull:
			if err := b.applyNull(base, p.nullScratch[b.nullIdx]); err != nil {
				return fmt.Errorf("table: column %d: %w", b.colIdx, err)
			}
		case modeJSON:
			if err := b.applyJSON(base, res.GetxN(b.colIdx)); err != nil {
				return fmt.Errorf("table: column %d: %w", b.colIdx, err)
			}
		}
	}
	return nil
}

var ioWriterType = reflect.TypeOf((*io.Writer)(nil)).Elem()

func makeDirectPrep(ft reflect.Type, off uintptr) (func(unsafe.Pointer) any, error) {
	// Opt[T]: return *Opt[T] for DirectAssign.
	if isOptType(ft) {
		return func(base unsafe.Pointer) any {
			return reflect.NewAt(ft, unsafe.Add(base, off)).Interface()
		}, nil
	}
	// io.Writer interface field: Prep the concrete writer value (not *io.Writer).
	if ft.Kind() == reflect.Interface && ft.Implements(ioWriterType) {
		return func(base unsafe.Pointer) any {
			return reflect.NewAt(ft, unsafe.Add(base, off)).Elem().Interface()
		}, nil
	}
	// Named types that implement io.Writer (e.g. *bytes.Buffer stored as concrete field).
	if ft.Implements(ioWriterType) {
		return func(base unsafe.Pointer) any {
			return reflect.NewAt(ft, unsafe.Add(base, off)).Interface()
		}, nil
	}

	switch ft.Kind() {
	case reflect.Bool:
		return func(base unsafe.Pointer) any { return (*bool)(unsafe.Add(base, off)) }, nil
	case reflect.Int:
		return func(base unsafe.Pointer) any { return (*int)(unsafe.Add(base, off)) }, nil
	case reflect.Int8:
		return func(base unsafe.Pointer) any { return (*int8)(unsafe.Add(base, off)) }, nil
	case reflect.Int16:
		return func(base unsafe.Pointer) any { return (*int16)(unsafe.Add(base, off)) }, nil
	case reflect.Int32:
		return func(base unsafe.Pointer) any { return (*int32)(unsafe.Add(base, off)) }, nil
	case reflect.Int64:
		return func(base unsafe.Pointer) any { return (*int64)(unsafe.Add(base, off)) }, nil
	case reflect.Uint:
		return func(base unsafe.Pointer) any { return (*uint)(unsafe.Add(base, off)) }, nil
	case reflect.Uint8:
		return func(base unsafe.Pointer) any { return (*uint8)(unsafe.Add(base, off)) }, nil
	case reflect.Uint16:
		return func(base unsafe.Pointer) any { return (*uint16)(unsafe.Add(base, off)) }, nil
	case reflect.Uint32:
		return func(base unsafe.Pointer) any { return (*uint32)(unsafe.Add(base, off)) }, nil
	case reflect.Uint64:
		return func(base unsafe.Pointer) any { return (*uint64)(unsafe.Add(base, off)) }, nil
	case reflect.Float32:
		return func(base unsafe.Pointer) any { return (*float32)(unsafe.Add(base, off)) }, nil
	case reflect.Float64:
		return func(base unsafe.Pointer) any { return (*float64)(unsafe.Add(base, off)) }, nil
	case reflect.String:
		return func(base unsafe.Pointer) any { return (*string)(unsafe.Add(base, off)) }, nil
	case reflect.Slice:
		if ft.Elem().Kind() == reflect.Uint8 {
			return func(base unsafe.Pointer) any { return (*[]byte)(unsafe.Add(base, off)) }, nil
		}
	case reflect.Struct:
		switch ft {
		case reflect.TypeOf(time.Time{}):
			return func(base unsafe.Pointer) any { return (*time.Time)(unsafe.Add(base, off)) }, nil
		case reflect.TypeOf(big.Rat{}):
			return func(base unsafe.Pointer) any { return (*big.Rat)(unsafe.Add(base, off)) }, nil
		}
	}
	// Fallback: reflect.NewAt still avoids Field-by-name walks.
	return func(base unsafe.Pointer) any {
		return reflect.NewAt(ft, unsafe.Add(base, off)).Interface()
	}, nil
}

func makeNullApply(ft reflect.Type, off uintptr) (func(unsafe.Pointer, rdb.Nullable) error, error) {
	if ft.Kind() == reflect.Ptr {
		elem := ft.Elem()
		elemApply, err := makeNullApply(elem, 0)
		if err != nil {
			return nil, err
		}
		// For pointer fields we allocate a new element and point at it.
		return func(base unsafe.Pointer, n rdb.Nullable) error {
			slot := (*unsafe.Pointer)(unsafe.Add(base, off))
			if n.Null || n.Value == nil {
				*slot = nil
				return nil
			}
			// Allocate elem, apply value into it, store pointer.
			ptr := reflect.New(elem)
			if err := elemApply(ptr.UnsafePointer(), n); err != nil {
				return err
			}
			*slot = ptr.UnsafePointer()
			return nil
		}, nil
	}

	switch ft.Kind() {
	case reflect.Bool:
		return func(base unsafe.Pointer, n rdb.Nullable) error {
			p := (*bool)(unsafe.Add(base, off))
			if n.Null || n.Value == nil {
				*p = false
				return nil
			}
			v, err := asBool(n.Value)
			if err != nil {
				return err
			}
			*p = v
			return nil
		}, nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return makeIntNullApply(ft.Kind(), off), nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return makeUintNullApply(ft.Kind(), off), nil
	case reflect.Float32:
		return func(base unsafe.Pointer, n rdb.Nullable) error {
			p := (*float32)(unsafe.Add(base, off))
			if n.Null || n.Value == nil {
				*p = 0
				return nil
			}
			v, err := asFloat64(n.Value)
			if err != nil {
				return err
			}
			*p = float32(v)
			return nil
		}, nil
	case reflect.Float64:
		return func(base unsafe.Pointer, n rdb.Nullable) error {
			p := (*float64)(unsafe.Add(base, off))
			if n.Null || n.Value == nil {
				*p = 0
				return nil
			}
			v, err := asFloat64(n.Value)
			if err != nil {
				return err
			}
			*p = v
			return nil
		}, nil
	case reflect.String:
		return func(base unsafe.Pointer, n rdb.Nullable) error {
			p := (*string)(unsafe.Add(base, off))
			if n.Null || n.Value == nil {
				*p = ""
				return nil
			}
			s, err := asString(n.Value)
			if err != nil {
				return err
			}
			*p = s
			return nil
		}, nil
	case reflect.Slice:
		if ft.Elem().Kind() == reflect.Uint8 {
			return func(base unsafe.Pointer, n rdb.Nullable) error {
				p := (*[]byte)(unsafe.Add(base, off))
				if n.Null || n.Value == nil {
					*p = nil
					return nil
				}
				b, err := asBytes(n.Value)
				if err != nil {
					return err
				}
				*p = b
				return nil
			}, nil
		}
	case reflect.Struct:
		if ft == reflect.TypeOf(time.Time{}) {
			return func(base unsafe.Pointer, n rdb.Nullable) error {
				p := (*time.Time)(unsafe.Add(base, off))
				if n.Null || n.Value == nil {
					*p = time.Time{}
					return nil
				}
				t, err := asTime(n.Value)
				if err != nil {
					return err
				}
				*p = t
				return nil
			}, nil
		}
	}

	// Reflect fallback for uncommon types.
	return func(base unsafe.Pointer, n rdb.Nullable) error {
		rv := reflect.NewAt(ft, unsafe.Add(base, off)).Elem()
		if n.Null || n.Value == nil {
			rv.Set(reflect.Zero(ft))
			return nil
		}
		fv := reflect.ValueOf(n.Value)
		if !fv.Type().ConvertibleTo(ft) {
			return fmt.Errorf("cannot convert %T to %s", n.Value, ft)
		}
		rv.Set(fv.Convert(ft))
		return nil
	}, nil
}

func makeIntNullApply(kind reflect.Kind, off uintptr) func(unsafe.Pointer, rdb.Nullable) error {
	return func(base unsafe.Pointer, n rdb.Nullable) error {
		if n.Null || n.Value == nil {
			switch kind {
			case reflect.Int:
				*(*int)(unsafe.Add(base, off)) = 0
			case reflect.Int8:
				*(*int8)(unsafe.Add(base, off)) = 0
			case reflect.Int16:
				*(*int16)(unsafe.Add(base, off)) = 0
			case reflect.Int32:
				*(*int32)(unsafe.Add(base, off)) = 0
			case reflect.Int64:
				*(*int64)(unsafe.Add(base, off)) = 0
			}
			return nil
		}
		v, err := asInt64(n.Value)
		if err != nil {
			return err
		}
		switch kind {
		case reflect.Int:
			*(*int)(unsafe.Add(base, off)) = int(v)
		case reflect.Int8:
			*(*int8)(unsafe.Add(base, off)) = int8(v)
		case reflect.Int16:
			*(*int16)(unsafe.Add(base, off)) = int16(v)
		case reflect.Int32:
			*(*int32)(unsafe.Add(base, off)) = int32(v)
		case reflect.Int64:
			*(*int64)(unsafe.Add(base, off)) = v
		}
		return nil
	}
}

func makeUintNullApply(kind reflect.Kind, off uintptr) func(unsafe.Pointer, rdb.Nullable) error {
	return func(base unsafe.Pointer, n rdb.Nullable) error {
		if n.Null || n.Value == nil {
			switch kind {
			case reflect.Uint:
				*(*uint)(unsafe.Add(base, off)) = 0
			case reflect.Uint8:
				*(*uint8)(unsafe.Add(base, off)) = 0
			case reflect.Uint16:
				*(*uint16)(unsafe.Add(base, off)) = 0
			case reflect.Uint32:
				*(*uint32)(unsafe.Add(base, off)) = 0
			case reflect.Uint64:
				*(*uint64)(unsafe.Add(base, off)) = 0
			}
			return nil
		}
		v, err := asUint64(n.Value)
		if err != nil {
			return err
		}
		switch kind {
		case reflect.Uint:
			*(*uint)(unsafe.Add(base, off)) = uint(v)
		case reflect.Uint8:
			*(*uint8)(unsafe.Add(base, off)) = uint8(v)
		case reflect.Uint16:
			*(*uint16)(unsafe.Add(base, off)) = uint16(v)
		case reflect.Uint32:
			*(*uint32)(unsafe.Add(base, off)) = uint32(v)
		case reflect.Uint64:
			*(*uint64)(unsafe.Add(base, off)) = v
		}
		return nil
	}
}

func makeJSONApply(ft reflect.Type, off uintptr) (func(unsafe.Pointer, rdb.Nullable) error, error) {
	return func(base unsafe.Pointer, n rdb.Nullable) error {
		ptr := reflect.NewAt(ft, unsafe.Add(base, off)).Interface()
		if n.Null || n.Value == nil {
			reflect.ValueOf(ptr).Elem().Set(reflect.Zero(ft))
			return nil
		}
		var jsonData []byte
		switch v := n.Value.(type) {
		case []byte:
			jsonData = v
		case string:
			jsonData = []byte(v)
		default:
			return fmt.Errorf("expected []byte or string for JSON field, got %T", n.Value)
		}
		if err := json.Unmarshal(jsonData, ptr); err != nil {
			return fmt.Errorf("JSON unmarshal: %w", err)
		}
		return nil
	}, nil
}

func asInt64(v any) (int64, error) {
	switch x := v.(type) {
	case int:
		return int64(x), nil
	case int8:
		return int64(x), nil
	case int16:
		return int64(x), nil
	case int32:
		return int64(x), nil
	case int64:
		return x, nil
	case uint:
		return int64(x), nil
	case uint8:
		return int64(x), nil
	case uint16:
		return int64(x), nil
	case uint32:
		return int64(x), nil
	case uint64:
		return int64(x), nil
	case float32:
		return int64(x), nil
	case float64:
		return int64(x), nil
	default:
		return 0, fmt.Errorf("cannot convert %T to integer", v)
	}
}

func asUint64(v any) (uint64, error) {
	switch x := v.(type) {
	case uint:
		return uint64(x), nil
	case uint8:
		return uint64(x), nil
	case uint16:
		return uint64(x), nil
	case uint32:
		return uint64(x), nil
	case uint64:
		return x, nil
	case int:
		return uint64(x), nil
	case int8:
		return uint64(x), nil
	case int16:
		return uint64(x), nil
	case int32:
		return uint64(x), nil
	case int64:
		return uint64(x), nil
	default:
		return 0, fmt.Errorf("cannot convert %T to unsigned integer", v)
	}
}

func asFloat64(v any) (float64, error) {
	switch x := v.(type) {
	case float32:
		return float64(x), nil
	case float64:
		return x, nil
	case int:
		return float64(x), nil
	case int32:
		return float64(x), nil
	case int64:
		return float64(x), nil
	default:
		return 0, fmt.Errorf("cannot convert %T to float", v)
	}
}

func asBool(v any) (bool, error) {
	switch x := v.(type) {
	case bool:
		return x, nil
	default:
		return false, fmt.Errorf("cannot convert %T to bool", v)
	}
}

func asString(v any) (string, error) {
	switch x := v.(type) {
	case string:
		return x, nil
	case []byte:
		return string(x), nil
	default:
		return "", fmt.Errorf("cannot convert %T to string", v)
	}
}

func asBytes(v any) ([]byte, error) {
	switch x := v.(type) {
	case []byte:
		return x, nil
	case string:
		return []byte(x), nil
	default:
		return nil, fmt.Errorf("cannot convert %T to []byte", v)
	}
}

func asTime(v any) (time.Time, error) {
	switch x := v.(type) {
	case time.Time:
		return x, nil
	default:
		return time.Time{}, fmt.Errorf("cannot convert %T to time.Time", v)
	}
}
