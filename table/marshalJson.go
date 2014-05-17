// Copyright 2014 Daniel Theophanes.
// Use of this source code is governed by a zlib-style
// license that can be found in the LICENSE file.

package table

import (
	"bytes"
	"encoding/json"
	"io"
)

type JsonObjectArray struct {
	*Buffer
	FlushAt int
}

func (coder *JsonObjectArray) WriteTo(writer io.Writer) (n int64, err error) {
	if coder.FlushAt == 0 {
		coder.FlushAt = 16 * 1024
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
				bb, err = json.Marshal(field.V)
				if err != nil {
					return
				}
				buf.Write(bb)
			}
		}
		buf.WriteRune('}')
		if buf.Len() > coder.FlushAt {
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
