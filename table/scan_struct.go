// Copyright 2014 Daniel Theophanes.
// Use of this source code is governed by a zlib-style
// license that can be found in the LICENSE file.

package table

import (
	"context"
	"encoding/json"
	"fmt"
	"iter"
	"reflect"
	"strings"

	"github.com/kardianos/rdb"
)

// Query runs cmd on q and scans every row of the first result set into a new []T.
// T must be a struct or a pointer to a struct. Column mapping uses the "db" tag
// (or the field name), same rules as UnmarshalStruct.
//
// Non-nullable scalar fields are populated via Result.Prep / Scan so the driver
// can DirectAssign. SQL NULL into a non-pointer field leaves the zero value;
// NULL into a pointer field leaves nil.
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
		if err := plan.scan(res, &out[len(out)-1]); err != nil {
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
			if err := plan.scan(res, &row); err != nil {
				yield(zero, err)
				return
			}
			if !yield(row, nil) {
				return
			}
		}
	}
}

type fieldBind struct {
	colIdx   int
	fieldIdx int
	isJSON   bool
	// viaNull scans into rdb.Nullable then copies into the field.
	// Used for JSON (after Scan), SQL-nullable non-pointer fields, and pointer fields.
	viaNull  bool
	jsonElem reflect.Type
}

type structPlan struct {
	tType       reflect.Type // underlying struct type
	asPtr       bool         // T is *Struct
	fields      []fieldBind
	nullScratch []rdb.Nullable // one per viaNull field, reused each row
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

	p := &structPlan{tType: tType, asPtr: asPtr}
	for i := 0; i < tType.NumField(); i++ {
		f := tType.Field(i)
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
		colIdx, ok := nameIndex[columnName]
		if !ok {
			return nil, fmt.Errorf("table: field %s (column %q) not found in result schema", f.Name, columnName)
		}
		b := fieldBind{colIdx: colIdx, fieldIdx: i, isJSON: isJSON}
		if isJSON {
			if f.Type.Kind() != reflect.Slice && f.Type.Kind() != reflect.Array {
				return nil, fmt.Errorf("table: field %s with json tag must be a slice or array", f.Name)
			}
			b.jsonElem = f.Type.Elem()
			// JSON is read after Scan via GetxN (may be chunked / large).
		} else if f.Type.Kind() == reflect.Ptr || colNullable[colIdx] {
			// Pointer fields and nullable columns: accept NULL without ErrScanNull.
			b.viaNull = true
			p.nullScratch = append(p.nullScratch, rdb.Nullable{})
		}
		p.fields = append(p.fields, b)
	}
	return p, nil
}

// scan populates dest (a *T) from the current result row.
func (p *structPlan) scan(res *rdb.Result, dest any) error {
	rv := reflect.ValueOf(dest).Elem() // T
	var stru reflect.Value
	if p.asPtr {
		if rv.IsNil() {
			rv.Set(reflect.New(p.tType))
		}
		stru = rv.Elem()
	} else {
		stru = rv
	}

	si := 0
	for i := range p.fields {
		b := &p.fields[i]
		if b.isJSON {
			continue
		}
		f := stru.Field(b.fieldIdx)
		if !f.CanSet() {
			continue
		}
		if b.viaNull {
			p.nullScratch[si] = rdb.Nullable{}
			res.Prepx(b.colIdx, &p.nullScratch[si])
			si++
			continue
		}
		// Direct Prep into the field for DirectAssign.
		res.Prepx(b.colIdx, f.Addr().Interface())
	}

	if err := res.Scan(); err != nil {
		return err
	}

	si = 0
	for i := range p.fields {
		b := &p.fields[i]
		f := stru.Field(b.fieldIdx)
		if !f.CanSet() {
			continue
		}
		if b.isJSON {
			val := res.GetxN(b.colIdx)
			if val.Null {
				f.Set(reflect.Zero(f.Type()))
				continue
			}
			var jsonData []byte
			switch v := val.Value.(type) {
			case []byte:
				jsonData = v
			case string:
				jsonData = []byte(v)
			default:
				return fmt.Errorf("table: column %d: expected []byte or string for JSON field, got %T", b.colIdx, val.Value)
			}
			slicePtr := reflect.New(f.Type())
			if err := json.Unmarshal(jsonData, slicePtr.Interface()); err != nil {
				return fmt.Errorf("table: column %d: JSON unmarshal: %w", b.colIdx, err)
			}
			f.Set(slicePtr.Elem())
			continue
		}
		if !b.viaNull {
			continue
		}
		n := p.nullScratch[si]
		si++
		if n.Null || n.Value == nil {
			f.Set(reflect.Zero(f.Type()))
			continue
		}
		if f.Kind() == reflect.Ptr {
			elemType := f.Type().Elem()
			fv := reflect.ValueOf(n.Value)
			if !fv.Type().ConvertibleTo(elemType) {
				return fmt.Errorf("table: column %d: cannot convert %T to %s", b.colIdx, n.Value, elemType)
			}
			elem := reflect.New(elemType)
			elem.Elem().Set(fv.Convert(elemType))
			f.Set(elem)
			continue
		}
		fv := reflect.ValueOf(n.Value)
		if !fv.Type().ConvertibleTo(f.Type()) {
			return fmt.Errorf("table: column %d: cannot convert %T to %s", b.colIdx, n.Value, f.Type())
		}
		f.Set(fv.Convert(f.Type()))
	}
	return nil
}
