// Copyright 2014 Daniel Theophanes.
// Use of this source code is governed by a zlib-style
// license that can be found in the LICENSE file.

package rdb

import (
	"fmt"
	"io"
	"math/big"
	"time"
)

type Valuer interface {
	Columns(cc []*SqlColumn) error
	WriteField(c *SqlColumn, value *DriverValue) error
	SqlMessage(err *SqlMessage)
	RowScanned()
	Done() error
}

type valuer struct {
	errors SqlErrors
	fields []*Field
	eof    bool
	arity  Arity

	columns      []*SqlColumn
	columnLookup map[string]*SqlColumn
	buffer       []*DriverValue
	prep         []interface{}

	initFields []*Field

	rowCount uint64
}

func (v *valuer) clearBuffer() {
	for i := range v.buffer {
		v.buffer[i] = nil
	}
}

func (v *valuer) Columns(cc []*SqlColumn) error {
	v.columns = cc
	v.columnLookup = make(map[string]*SqlColumn, len(cc))
	for _, col := range cc {
		v.columnLookup[col.Name] = col
	}
	v.buffer = make([]*DriverValue, len(cc))
	v.prep = make([]interface{}, len(cc))

	v.fields = make([]*Field, len(cc))
	for i, field := range v.initFields {
		if len(field.N) == 0 {
			if i < len(v.columns) {
				return ErrorColumnNotFound{Index: i}
			}
			v.fields[i] = field
		} else {
			col, found := v.columnLookup[field.N]
			if !found {
				return ErrorColumnNotFound{Name: field.N}
			}
			v.fields[col.Index] = field
		}
	}
	v.initFields = nil
	return nil
}
func (v *valuer) SqlMessage(msg *SqlMessage) {
	switch msg.Type {
	case SqlInfo:
		// TODO: Where should info messages go?

	case SqlError:
		v.errors = append(v.errors, msg)
	}
}
func (v *valuer) RowScanned() {
	v.rowCount += 1
	for i := range v.prep {
		v.prep[i] = nil
	}
	/*
		TODO: Fix arity check.
		if v.arity&One != 0 {
			v.EOF = true
			if v.rowCount == 1 {
				return v.conn.Scan()
			}
			if v.arity&ArityMust != 0 && v.rowCount > 1 {
				return arityError
			}
		}
	*/
	return
}

func (v *valuer) Done() error {
	v.eof = true
	for i := range v.prep {
		v.prep[i] = nil
	}
	if len(v.errors) != 0 {
		return v.errors
	}
	return nil
}

func (v *valuer) WriteField(c *SqlColumn, value *DriverValue) error {
	prep := v.prep[c.Index]
	f := v.fields[c.Index]
	if value.Null {
		value.Value = f.NullValue
	}
	if prep == nil {
		if value.Chunked {
			bf := v.buffer[c.Index]
			if bf == nil {
				v.buffer[c.Index] = value
				return nil
			}
			switch in := value.Value.(type) {
			case []byte:
				bf.Value = append(bf.Value.([]byte), in...)
			}
		}
		v.buffer[c.Index] = value
		return nil
	}
	if value.Value == nil {
		// May happen with null values.
		// Nothing to convert or write.
		return nil
	}
	var err error
	switch in := value.Value.(type) {
	case string:
		switch out := prep.(type) {
		case io.Writer:
			_, err = out.Write([]byte(in))
		case *string:
			*out = in
		case *[]byte:
			*out = []byte(in)
		default:
			return errorTypeNotSupported(in, out, c)
		}
	case []byte:
		switch out := prep.(type) {
		case io.Writer:
			_, err = out.Write(in)
		case *string:
			*out = string(in)
		case *[]byte:
			*out = in
		default:
			return errorTypeNotSupported(in, out, c)
		}
	case bool:
		switch out := prep.(type) {
		case *bool:
			*out = bool(in)
		default:
			return errorTypeNotSupported(in, out, c)
		}
	case uint8:
		switch out := prep.(type) {
		case *uint8:
			*out = uint8(in)
		case *int8:
			*out = int8(in)
		case *uint16:
			*out = uint16(in)
		case *int16:
			*out = int16(in)
		case *uint32:
			*out = uint32(in)
		case *int32:
			*out = int32(in)
		case *uint64:
			*out = uint64(in)
		case *int64:
			*out = int64(in)
		default:
			return errorTypeNotSupported(in, out, c)
		}
	case int8:
		switch out := prep.(type) {
		case *uint8:
			*out = uint8(in)
		case *int8:
			*out = int8(in)
		case *uint16:
			*out = uint16(in)
		case *int16:
			*out = int16(in)
		case *uint32:
			*out = uint32(in)
		case *int32:
			*out = int32(in)
		case *uint64:
			*out = uint64(in)
		case *int64:
			*out = int64(in)
		default:
			return errorTypeNotSupported(in, out, c)
		}
	case uint16:
		switch out := prep.(type) {
		case *uint16:
			*out = uint16(in)
		case *int16:
			*out = int16(in)
		case *uint32:
			*out = uint32(in)
		case *int32:
			*out = int32(in)
		case *uint64:
			*out = uint64(in)
		case *int64:
			*out = int64(in)
		default:
			return errorTypeNotSupported(in, out, c)
		}
	case int16:
		switch out := prep.(type) {
		case *uint16:
			*out = uint16(in)
		case *int16:
			*out = int16(in)
		case *uint32:
			*out = uint32(in)
		case *int32:
			*out = int32(in)
		case *uint64:
			*out = uint64(in)
		case *int64:
			*out = int64(in)
		default:
			return errorTypeNotSupported(in, out, c)
		}
	case uint32:
		switch out := prep.(type) {
		case *uint32:
			*out = uint32(in)
		case *int32:
			*out = int32(in)
		case *uint64:
			*out = uint64(in)
		case *int64:
			*out = int64(in)
		case *uint:
			*out = uint(in)
		case *int:
			*out = int(in)
		default:
			return errorTypeNotSupported(in, out, c)
		}
	case int32:
		switch out := prep.(type) {
		case *uint32:
			*out = uint32(in)
		case *int32:
			*out = int32(in)
		case *uint64:
			*out = uint64(in)
		case *int64:
			*out = int64(in)
		case *uint:
			*out = uint(in)
		case *int:
			*out = int(in)
		default:
			return errorTypeNotSupported(in, out, c)
		}
	case uint64:
		switch out := prep.(type) {
		case *uint64:
			*out = uint64(in)
		case *int64:
			*out = int64(in)
		case *uint:
			*out = uint(in)
		case *int:
			*out = int(in)
		default:
			return errorTypeNotSupported(in, out, c)
		}
	case int64:
		switch out := prep.(type) {
		case *uint64:
			*out = uint64(in)
		case *int64:
			*out = int64(in)
		case *uint:
			*out = uint(in)
		case *int:
			*out = int(in)
		default:
			return errorTypeNotSupported(in, out, c)
		}
	case float32:
		switch out := prep.(type) {
		case **big.Rat:
			out.SetFloat64(float64(in))
		case *big.Rat:
			out.SetFloat64(float64(in))
		case *float64:
			*out = float64(in)
		case *float32:
			*out = float32(in)
		default:
			return errorTypeNotSupported(in, out, c)
		}
	case float64:
		switch out := prep.(type) {
		case **big.Rat:
			out.SetFloat64(float64(in))
		case *big.Rat:
			out.SetFloat64(float64(in))
		case *float64:
			*out = float64(in)
		case *float32:
			*out = float32(in)
		default:
			return errorTypeNotSupported(in, out, c)
		}
	case *big.Rat:
		switch out := prep.(type) {
		case **big.Rat:
			*out = in
		case *big.Rat:
			out.Set(in)
		case *float64:
			fl, _ := in.Float64()
			*out = float64(fl)
		case *float32:
			fl, _ := in.Float64()
			*out = float32(fl)
		default:
			return errorTypeNotSupported(in, out, c)
		}
	case time.Time:
		switch out := prep.(type) {
		case *time.Time:
			*out = in
		default:
			return errorTypeNotSupported(in, out, c)
		}
	case time.Duration:
		switch out := prep.(type) {
		case *time.Duration:
			*out = in
		default:
			return errorTypeNotSupported(in, out, c)
		}
	default:
		return errorTypeNotSupported(nil, nil, c)
	}
	return err
}

func errorTypeNotSupported(in, out interface{}, c *SqlColumn) error {
	if out == nil && in == nil {
		return fmt.Errorf("Unsupported column type: %s", c.Name)
	}
	return fmt.Errorf("Prep type (%T) cannot fit data type (%T) in %s", out, in, c.Name)
}
