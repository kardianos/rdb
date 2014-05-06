// Copyright 2014 Daniel Theophanes.
// Use of this source code is governed by a zlib-style
// license that can be found in the LICENSE file.

package ms

import (
	"bitbucket.org/kardianos/rdb"
	"fmt"
	"math/big"
	"strings"
	"time"
)

type Result struct {
	Errors  rdb.SqlErrors
	Columns []*SqlColumn
	Fields  []*rdb.Field
	EOF     bool
	arity   rdb.Arity

	ColumnLookup map[string]*SqlColumn
	buffer       []*FieldValue
	prep         []interface{}

	initFields []*rdb.Field

	tds *Connection
	mr  *MessageReader

	rowCount uint64
}

func (r *Result) WriteField(c *SqlColumn, f *rdb.Field, value *FieldValue) {
	prep := r.prep[c.Index]
	if value.Null {
		value.Value = f.NullValue
	}
	if prep == nil {
		r.buffer[c.Index] = value
		return
	}
	if value.Value == nil {
		// May happen with null values.
		// Nothing to convert or write.
		return
	}

	switch {
	case c.info.Bytes:
		switch out := prep.(type) {
		case *string:
			switch in := value.Value.(type) {
			case string:
				*out = in
			case []byte:
				*out = string(in)
			default:
				errorTypeNotSupported(in, out, c)
			}
		case *[]byte:
			switch in := value.Value.(type) {
			case string:
				*out = []byte(in)
			case []byte:
				*out = in
			default:
				errorTypeNotSupported(in, out, c)
			}
		default:
			errorTypeNotSupported(nil, out, c)
		}
	case c.code == typeIntN || c.info.Fixed:
		switch out := prep.(type) {
		case *uint8:
			switch in := value.Value.(type) {
			case uint8:
				*out = in
			default:
				errorTypeNotSupported(in, out, c)
			}
		case *int8:
			switch in := value.Value.(type) {
			case uint8:
				*out = int8(in)
			default:
				errorTypeNotSupported(in, out, c)
			}
		case *uint16:
			switch in := value.Value.(type) {
			case uint8:
				*out = uint16(in)
			case uint16:
				*out = uint16(in)
			default:
				errorTypeNotSupported(in, out, c)
			}
		case *int16:
			switch in := value.Value.(type) {
			case uint8:
				*out = int16(in)
			case uint16:
				*out = int16(in)
			default:
				errorTypeNotSupported(in, out, c)
			}
		case *uint32:
			switch in := value.Value.(type) {
			case uint8:
				*out = uint32(in)
			case uint16:
				*out = uint32(in)
			case uint32:
				*out = uint32(in)
			default:
				errorTypeNotSupported(in, out, c)
			}
		case *int32:
			switch in := value.Value.(type) {
			case uint8:
				*out = int32(in)
			case uint16:
				*out = int32(in)
			case uint32:
				*out = int32(in)
			default:
				errorTypeNotSupported(in, out, c)
			}
		case *uint64:
			switch in := value.Value.(type) {
			case uint8:
				*out = uint64(in)
			case uint16:
				*out = uint64(in)
			case uint32:
				*out = uint64(in)
			case uint64:
				*out = uint64(in)
			default:
				errorTypeNotSupported(in, out, c)
			}
		case *int64:
			switch in := value.Value.(type) {
			case uint8:
				*out = int64(in)
			case uint16:
				*out = int64(in)
			case uint32:
				*out = int64(in)
			case uint64:
				*out = int64(in)
			default:
				errorTypeNotSupported(in, out, c)
			}
		case *uint:
			switch in := value.Value.(type) {
			case uint8:
				*out = uint(in)
			case uint16:
				*out = uint(in)
			case uint32:
				*out = uint(in)
			case uint64:
				*out = uint(in)
			default:
				errorTypeNotSupported(in, out, c)
			}
		case *int:
			switch in := value.Value.(type) {
			case uint8:
				*out = int(in)
			case uint16:
				*out = int(in)
			case uint32:
				*out = int(in)
			case uint64:
				*out = int(in)
			default:
				errorTypeNotSupported(in, out, c)
			}
		default:
			errorTypeNotSupported(nil, out, c)
		}
	case c.code == typeBitN:
		switch out := prep.(type) {
		case *bool:
			switch in := value.Value.(type) {
			case bool:
				*out = in
			default:
				errorTypeNotSupported(in, out, c)
			}
		default:
			errorTypeNotSupported(nil, out, c)
		}
	case c.info.IsPrSc:
		switch in := value.Value.(type) {
		case *big.Rat:
			switch out := prep.(type) {
			case **big.Rat:
				*out = in
			case *float64:
				*out, _ = in.Float64()
			case *float32:
				v, _ := in.Float64()
				*out = float32(v)
			default:
				errorTypeNotSupported(in, out, c)
			}
		default:
			errorTypeNotSupported(in, nil, c)
		}
	case c.code == typeFloatN:
		switch out := prep.(type) {
		case *float32:
			switch in := value.Value.(type) {
			case float32:
				*out = in
			case float64:
				*out = float32(in)
			default:
				errorTypeNotSupported(in, out, c)
			}
		case *float64:
			switch in := value.Value.(type) {
			case float32:
				*out = float64(in)
			case float64:
				*out = float64(in)
			default:
				errorTypeNotSupported(in, out, c)
			}
		case **big.Rat:
			switch in := value.Value.(type) {
			case float32:
				out.SetFloat64(float64(in))
			case float64:
				out.SetFloat64(in)
			default:
				errorTypeNotSupported(in, out, c)
			}
		default:
			errorTypeNotSupported(nil, out, c)
		}
	case c.code == typeTimeN:
		fallthrough
	case c.code == typeDateN:
		fallthrough
	case c.code == typeDateTime2N:
		fallthrough
	case c.code == typeDateTimeOffsetN:
		fallthrough
	case c.code == typeDateTimeN:
		switch out := prep.(type) {
		case *time.Time:
			switch in := value.Value.(type) {
			case time.Time:
				*out = in
			default:
				errorTypeNotSupported(in, out, c)
			}
		case *time.Duration:
			switch in := value.Value.(type) {
			case time.Duration:
				*out = in
			default:
				errorTypeNotSupported(in, out, c)
			}
		default:
			errorTypeNotSupported(nil, out, c)
		}
	default:
		panic(fmt.Errorf("Unsupported column type: %s - %v", c.Name, value.Value))
	}
	return
}

func errorTypeNotSupported(in, out interface{}, c *SqlColumn) error {
	if in == nil {
		panic(recoverError{err: fmt.Errorf("Prep type (%T) not valid in %s", out, c.Name)})
	}
	if out == nil {
		panic(recoverError{err: fmt.Errorf("DB type (%T) not valid in %s", in, c.Name)})
	}
	panic(recoverError{err: fmt.Errorf("Prep type (%T) cannot fit data type (%T) in %s", out, in, c.Name)})
}

func (result *Result) Process(clearRowBuffer bool) error {
	m := result.mr
	// Make sure the existing row buffer is purged before filling it again.
	if clearRowBuffer {
		for i := range result.buffer {
			result.buffer[i] = nil
		}
	}
	for {
		res, err := result.tds.getSingleResponse(m, result)
		if err != nil {
			result.EOF = true
			return err
		}
		switch v := res.(type) {
		case *rdb.SqlError:
			result.Errors = append(result.Errors, v)
		case []*SqlColumn:
			result.Columns = v
			result.ColumnLookup = make(map[string]*SqlColumn, len(v))
			for _, col := range v {
				result.ColumnLookup[col.Name] = col
			}
			result.buffer = make([]*FieldValue, len(v))
			result.prep = make([]interface{}, len(v))

			result.Fields = make([]*rdb.Field, len(v))
			for i, field := range result.initFields {
				if len(field.N) == 0 {
					if i < len(result.Columns) {
						return rdb.ErrorColumnNotFound{Index: i}
					}
					result.Fields[i] = field
				} else {
					col, found := result.ColumnLookup[field.N]
					if !found {
						return rdb.ErrorColumnNotFound{Name: field.N}
					}
					result.Fields[col.Index] = field
				}
			}
			result.initFields = nil
			return nil
		case *SqlRow:
			// Sent after the row is scanned.
			// Prep values must be cleared after the initial fill.
			// The prior prep values are no longer valid as they are filled
			// during the row scan.
			result.rowCount += 1
			for i := range result.prep {
				result.prep[i] = nil
			}

			if result.arity&rdb.One != 0 {
				result.EOF = true
				if result.rowCount == 1 {
					return result.Process(false)
				}
				if result.arity&rdb.ArityMust != 0 && result.rowCount > 1 {
					return arityError
				}
			}

			return nil
		case SqlRpcResult:
		case *SqlDone:
			result.EOF = true
			return nil
		}
	}
}

func (r *Result) Close() error {
	err := r.mr.Close()
	r.tds.inUse = false
	return err
}

type SqlDone struct {
	StatusCode uint16
	CurrentCmd uint16
	Rows       uint64
}

func (done *SqlDone) Status() string {
	if done.StatusCode == 0 {
		return "Final"
	}
	codes := []string{}

	if 0x01&done.StatusCode != 0 {
		codes = append(codes, "More")
	}
	if 0x02&done.StatusCode != 0 {
		codes = append(codes, "Error")
	}
	if 0x04&done.StatusCode != 0 {
		codes = append(codes, "Transaction in progress")
	}
	if 0x10&done.StatusCode != 0 {
		codes = append(codes, fmt.Sprintf("Rows: %d", done.Rows))
	}
	if 0x20&done.StatusCode != 0 {
		codes = append(codes, "Attention")
	}
	if 0x100&done.StatusCode != 0 {
		codes = append(codes, "Server Error. Discard results.")
	}
	if len(codes) == 0 {
		panic(fmt.Sprintf("Unknown code: %d", done.StatusCode))
	}
	return strings.Join(codes, " & ")
}
func (done *SqlDone) String() string {
	return fmt.Sprintf("Done Cmd=%d Status=%s", done.CurrentCmd, done.Status())
}
func (done *SqlDone) Error() string {
	return done.Status()
}

type SqlColumn struct {
	rdb.SqlColumn

	Collation [5]byte

	code driverType
	info typeInfo
}

type chunkStatus byte

const (
	chunkNone chunkStatus = iota
	chunked
	chunkDone
)

type FieldValue struct {
	Chunked  chunkStatus
	Value    interface{}
	Null     bool
	MustCopy bool
}

type SqlRow struct{}

type SqlRpcResult int32

type recoverError struct {
	err error
}
