// Copyright 2014 Daniel Theophanes.
// Use of this source code is governed by a zlib-style
// license that can be found in the LICENSE file.

package table

import (
	"bytes"
	"encoding/json"
	"io"

	"github.com/kardianos/rdb"
)

// type Converter func(value *rdb.Nullable) error

// Serialize table buffer as an array of JSON objects.
// When multiple results are returned, turns into an array of arrays.
type JsonRowObject struct {
	*Buffer
	FlushAt int
}

func (coder *JsonRowObject) MarshalJSON() ([]byte, error) {
	buf := &bytes.Buffer{}
	_, err := coder.WriteTo(buf)
	return buf.Bytes(), err
}

func (coder *JsonRowObject) WriteTo(writer io.Writer) (n int64, err error) {
	flushAt := coder.FlushAt
	if flushAt == 0 {
		flushAt = 16 * 1024
	}

	// 	var bb []byte
	buf := &bytes.Buffer{}
	var bbLen int

	if len(coder.Set) > 1 {
		set := coder.Set

		buf.WriteRune('[')

		for i, tableBuffer := range set {
			if i != 0 {
				buf.WriteRune(',')
			}
			bbLen, err = writer.Write(buf.Bytes())
			buf.Reset()
			n += int64(bbLen)
			if err != nil {
				return
			}

			var nextN int64
			nextN, err = coder.writeToSingle(writer, tableBuffer)
			n += nextN
			if err != nil {
				return
			}
		}

		buf.WriteRune(']')
		bbLen, err = writer.Write(buf.Bytes())
		buf.Reset()
		n += int64(bbLen)
		return
	}

	return coder.writeToSingle(writer, coder.Buffer)
}

func (coder *JsonRowObject) writeToSingle(writer io.Writer, table *Buffer) (n int64, err error) {
	flushAt := coder.FlushAt
	if flushAt == 0 {
		flushAt = 16 * 1024
	}

	var bb []byte
	buf := &bytes.Buffer{}
	var bbLen int

	buf.WriteRune('[')
	for i, row := range table.Row {
		if i != 0 {
			buf.WriteRune(',')
		}
		buf.WriteRune('{')
		for j, field := range row.Field {
			if j != 0 {
				buf.WriteRune(',')
			}
			col := table.schema[j]
			bb, err = json.Marshal(col.Name)
			if err != nil {
				return
			}
			buf.Write(bb)
			buf.WriteRune(':')
			if field.Null {
				buf.WriteString("null")
			} else {
				val := field.Value
				if col.Generic == rdb.Text {
					valBytes, is := val.([]byte)
					if is {
						val = string(valBytes)
					}
				}
				bb, err = json.Marshal(val)
				if err != nil {
					return
				}
				buf.Write(bb)
			}
		}
		buf.WriteRune('}')
		if buf.Len() > flushAt {
			bbLen, err = writer.Write(buf.Bytes())
			buf.Reset()
			n += int64(bbLen)
			if err != nil {
				return
			}
		}
	}
	buf.WriteRune(']')
	bbLen, err = writer.Write(buf.Bytes())
	n += int64(bbLen)
	return
}

// Serialize the table buffer as an object with a column name array and an
// an array of rows. Each row is an array of values.
// Supports many result sets by chaining them together.
type JsonRowArray struct {
	*Buffer
	FlushAt int

	// Additional properties to add to the output.
	Meta map[string]interface{}

	ResultNameName    string // Default field name is "Name".
	ColumnHeadersName string // Default field name is "Column".
	DataRowsName      string // Default field name is "Data".
}

func (coder *JsonRowArray) MarshalJSON() ([]byte, error) {
	buf := &bytes.Buffer{}
	_, err := coder.WriteTo(buf)
	return buf.Bytes(), err
}
func (coder *JsonRowArray) WriteTo(writer io.Writer) (n int64, err error) {
	flushAt := coder.FlushAt
	if flushAt == 0 {
		flushAt = 16 * 1024
	}

	buf := &bytes.Buffer{}
	var bbLen int

	if len(coder.Set) > 1 {
		set := coder.Set

		buf.WriteRune('[')

		for i, tableBuffer := range set {
			if i != 0 {
				buf.WriteRune(',')
			}
			bbLen, err = writer.Write(buf.Bytes())
			buf.Reset()
			n += int64(bbLen)
			if err != nil {
				return
			}

			var nextN int64
			nextN, err = coder.writeToSingle(writer, tableBuffer)
			n += nextN
			if err != nil {
				return
			}
		}

		buf.WriteRune(']')
		bbLen, err = writer.Write(buf.Bytes())
		buf.Reset()
		n += int64(bbLen)
		return
	}

	return coder.writeToSingle(writer, coder.Buffer)
}

func (coder *JsonRowArray) writeToSingle(writer io.Writer, table *Buffer) (n int64, err error) {
	flushAt := coder.FlushAt
	if flushAt == 0 {
		flushAt = 16 * 1024
	}
	resultName := "Name"
	names := "Column"
	data := "Data"
	if len(coder.ResultNameName) != 0 {
		resultName = coder.ResultNameName
	}
	if len(coder.ColumnHeadersName) != 0 {
		names = coder.ColumnHeadersName
	}
	if len(coder.DataRowsName) != 0 {
		data = coder.DataRowsName
	}

	var bb []byte
	buf := &bytes.Buffer{}
	var bbLen int

	// Write header.
	buf.WriteRune('{')
	for propName, prop := range coder.Meta {
		if propName == names || propName == data || propName == resultName {
			continue
		}
		bb, err = json.Marshal(propName)
		if err != nil {
			return
		}
		buf.Write(bb)
		buf.WriteRune(':')
		bb, err = json.Marshal(prop)
		if err != nil {
			return
		}
		buf.Write(bb)
		buf.WriteRune(',')
	}
	_ = resultName
	// Write result name.
	bb, err = json.Marshal(resultName)
	if err != nil {
		return
	}
	buf.Write(bb)
	buf.WriteRune(':')
	bb, err = json.Marshal(table.Name)
	if err != nil {
		return
	}
	buf.Write(bb)
	buf.WriteRune(',')

	bb, err = json.Marshal(names)
	if err != nil {
		return
	}
	buf.Write(bb)
	buf.WriteRune(':')
	// Write headers array.
	buf.WriteRune('[')
	schema := table.Schema()
	for i, col := range schema {
		if i != 0 {
			buf.WriteRune(',')
		}
		bb, err = json.Marshal(col.Name)
		if err != nil {
			return
		}
		buf.Write(bb)
	}
	buf.WriteRune(']')
	buf.WriteRune(',')
	bb, err = json.Marshal(data)
	if err != nil {
		return
	}
	buf.Write(bb)
	buf.WriteRune(':')
	// Write data array.
	buf.WriteRune('[')
	for i, row := range table.Row {
		if i != 0 {
			buf.WriteRune(',')
		}
		buf.WriteRune('[')
		for j, field := range row.Field {
			col := schema[j]
			if j != 0 {
				buf.WriteRune(',')
			}
			if field.Null {
				buf.WriteString("null")
			} else {
				val := field.Value
				if col.Generic == rdb.Text {
					valBytes, is := val.([]byte)
					if is {
						val = string(valBytes)
					}
				}
				bb, err = json.Marshal(val)
				if err != nil {
					return
				}
				buf.Write(bb)
			}
		}
		buf.WriteRune(']')
		if buf.Len() > flushAt {
			bbLen, err = writer.Write(buf.Bytes())
			buf.Reset()
			n += int64(bbLen)
			if err != nil {
				return
			}
		}
	}
	buf.WriteRune(']')
	buf.WriteRune('}')

	bbLen, err = writer.Write(buf.Bytes())
	n += int64(bbLen)
	return
}
