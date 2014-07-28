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

// TODO: Rename to DriverAssigner and document better.
// Assigner can be used by the driver to put special values directly into prepped
// value pointer.
type Assigner func(input, output interface{}) (handled bool, err error)

type DriverValuer interface {
	Columns([]*Column) error
	Done() error
	RowScanned()
	Message(*Message)
	WriteField(c *Column, reportRow bool, value *DriverValue, assign Assigner) error
	RowsAffected(count uint64)
}

type valuer struct {
	cmd *Command

	errorList Errors
	infoList  []*Message
	fields    []*Field
	eof       bool

	columns      []*Column
	columnLookup map[string]*Column
	buffer       []Nullable
	prep         []interface{}

	convert []ColumnConverter

	rowCount     uint64
	rowsAffected uint64
}

func (v *valuer) clearBuffer() {
	for i := range v.buffer {
		v.buffer[i] = Nullable{
			Null: true,
		}

	}
}
func (v *valuer) clearPrep() {
	for i := range v.prep {
		v.prep[i] = nil
	}
}

func (v *valuer) Columns(cc []*Column) error {
	v.columns = cc
	v.columnLookup = make(map[string]*Column, len(cc))
	for _, col := range cc {
		v.columnLookup[col.Name] = col
	}
	v.buffer = make([]Nullable, len(cc))
	v.prep = make([]interface{}, len(cc))

	// Prepare fields.
	v.fields = make([]*Field, len(cc))
	for i, field := range v.cmd.Fields {
		if len(field.Name) == 0 {
			if i >= len(v.columns) {
				// Don't error. Some queries may return
				// different number of columns.
				continue
			}
			v.fields[i] = &field
		} else {
			col, found := v.columnLookup[field.Name]
			if !found {
				// Don't error. Some queries may return
				// different number of columns.
				continue
			}
			v.fields[col.Index] = &field
		}
	}

	if v.cmd.Converter != nil {
		v.convert = make([]ColumnConverter, len(cc))
		for i, col := range cc {
			v.convert[i] = v.cmd.Converter.ColumnConverter(col)
		}
	}

	return nil
}
func (v *valuer) Message(msg *Message) {
	switch msg.Type {
	case SqlInfo:
		v.infoList = append(v.infoList, msg)

	case SqlError:
		v.errorList = append(v.errorList, msg)
	}
}
func (v *valuer) RowScanned() {
	v.rowCount += 1
	return
}

func (v *valuer) Done() error {
	v.eof = true
	for i := range v.prep {
		v.prep[i] = nil
	}
	if len(v.errorList) != 0 {
		return v.errorList
	}
	return nil
}

func (v *valuer) RowsAffected(count uint64) {
	v.rowsAffected = count
}

/*
	if (value is null) && (has default value) {
		set value to default value
	}
	if no prepped value {
		if chunked {
			append to any existing value.
		}
	}
	if prepped value {
		if prepped value is Nullable
	}
	If there is a default value for the field, use the default value.
	If there is no prepped value, put it in a buffer.
	If using a buffer, append any value
*/
func (v *valuer) WriteField(c *Column, reportRow bool, value *DriverValue, assign Assigner) error {
	// TODO: Respect value.MustCopy.
	if !reportRow {
		return nil
	}

	var convert ColumnConverter
	if v.convert != nil {
		convert = v.convert[c.Index]
	}

	prep := v.prep[c.Index]
	f := v.fields[c.Index]
	if value.Null && f != nil && f.Null != nil {
		value.Null = false
		value.Value = f.Null
	}
	if prep == nil {
		if value.Chunked {
			bf := v.buffer[c.Index]
			if bf.Value == nil {
				outValue := Nullable{
					Null:  value.Null,
					Value: value.Value,
				}
				if !value.More && convert != nil {
					convert(c, &outValue)
				}
				v.buffer[c.Index] = outValue
				return nil
			}
			switch in := value.Value.(type) {
			case []byte:
				bf.Value = append(bf.Value.([]byte), in...)
			default:
				return fmt.Errorf("Type not supported for chunked read: %T", in)
			}
			if !value.More && convert != nil {
				convert(c, &v.buffer[c.Index])
			}
			return nil
		}
		outValue := Nullable{
			Null:  value.Null,
			Value: value.Value,
		}
		if convert != nil {
			convert(c, &outValue)
		}
		v.buffer[c.Index] = outValue
		return nil
	}
	outValue := Nullable{
		Null:  value.Null,
		Value: value.Value,
	}
	if convert != nil {
		convert(c, &outValue)
	}
	if nullable, is := prep.(*Nullable); is {
		*nullable = outValue
		return nil
	}
	if outValue.Null || outValue.Value == nil {
		// Can only scan a null value into a nullable type.
		return ScanNullError
	}
	var err error
	var handled = false
	if assign != nil {
		handled, err = assign(outValue.Value, prep)
		if handled {
			return err
		}
	}

	switch in := outValue.Value.(type) {
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
		case *uint:
			*out = uint(in)
		case *int:
			*out = int(in)
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
		case *uint:
			*out = uint(in)
		case *int:
			*out = int(in)
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
		case *uint:
			*out = uint(in)
		case *int:
			*out = int(in)
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
		case *uint:
			*out = uint(in)
		case *int:
			*out = int(in)
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

func errorTypeNotSupported(in, out interface{}, c *Column) error {
	if out == nil && in == nil {
		return fmt.Errorf("Unsupported column type: %s", c.Name)
	}
	return fmt.Errorf("Prep type (%T) cannot fit data type (%T) in %s", out, in, c.Name)
}
