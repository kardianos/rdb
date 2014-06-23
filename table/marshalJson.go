// Copyright 2014 Daniel Theophanes.
// Use of this source code is governed by a zlib-style
// license that can be found in the LICENSE file.

package table

import (
	"bytes"
	"encoding/json"
	"io"

	"bitbucket.org/kardianos/rdb"
)

// type Converter func(value *rdb.Nullable) error

// Serialize table buffer as an array of JSON objects.
type JsonRowObject struct {
	*Buffer
	FlushAt int
}

func (coder *JsonRowObject) WriteTo(writer io.Writer) (n int64, err error) {
	flushAt := coder.FlushAt
	if flushAt == 0 {
		flushAt = 16 * 1024
	}

	var bb []byte
	buf := &bytes.Buffer{}
	var bbLen int
	buf.WriteRune('[')
	for i, row := range coder.Buffer.Row {
		if i != 0 {
			buf.WriteRune(',')
		}
		buf.WriteRune('{')
		for j, field := range row.Field {
			if j != 0 {
				buf.WriteRune(',')
			}
			col := coder.Buffer.schema[j]
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
type JsonRowArray struct {
	*Buffer
	FlushAt int

	ColumnHeadersName string
	DataRowsName      string
}

func (coder *JsonRowArray) WriteTo(writer io.Writer) (n int64, err error) {
	flushAt := coder.FlushAt
	if flushAt == 0 {
		flushAt = 16 * 1024
	}
	names := "Names"
	data := "Data"
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
	bb, err = json.Marshal(names)
	if err != nil {
		return
	}
	buf.Write(bb)
	buf.WriteRune(':')
	// Write headers array.
	buf.WriteRune('[')
	schema := coder.Buffer.Schema()
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
	for i, row := range coder.Buffer.Row {
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
