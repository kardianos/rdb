// Copyright 2014 Daniel Theophanes.
// Use of this source code is governed by a zlib-style
// license that can be found in the LICENSE file.

package table

import (
	"bitbucket.org/kardianos/rdb"
	"bytes"
	"testing"
)

func TestJsonMarshal(t *testing.T) {
	checkRowObject := `[{"ColA":"Hello","ColB":123.524},{"ColA":"Hi","ColB":null}]`
	checkRowArray := `{"Names":["ColA","ColB"],"Data":[["Hello",123.524],["Hi",null]]}`
	table := &Buffer{}
	table.SetSchema([]*rdb.SqlColumn{
		&rdb.SqlColumn{Name: "ColA"},
		&rdb.SqlColumn{Name: "ColB"},
	})
	table.Row = []Row{
		{
			Field: []rdb.Nullable{
				{V: "Hello"},
				{V: 123.524},
			},
		},
		{
			Field: []rdb.Nullable{
				{V: "Hi"},
				{Null: true},
			},
		},
	}

	var err error
	buf := &bytes.Buffer{}

	coderObj := JsonRowObject{Buffer: table}
	_, err = coderObj.WriteTo(buf)
	if err != nil {
		t.Error(err)
	}
	if buf.String() != checkRowObject {
		t.Errorf("Doesn't match: want <%s> got <%s>", checkRowObject, buf.String())
	}
	buf.Reset()

	coderArray := JsonRowArray{Buffer: table}
	_, err = coderArray.WriteTo(buf)
	if err != nil {
		t.Error(err)
	}
	if buf.String() != checkRowArray {
		t.Errorf("Doesn't match:\nwant\n\t%s\ngot\n\t%s", checkRowArray, buf.String())
	}
}
