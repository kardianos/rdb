// Copyright 2014 Daniel Theophanes.
// Use of this source code is governed by a zlib-style
// license that can be found in the LICENSE file.

package table

import (
	"bytes"
	"io"
	"testing"

	"github.com/kardianos/rdb"
)

func getTable(single bool) *Buffer {
	table := &Buffer{
		Name: "res1",
	}
	table.SetSchema([]*rdb.Column{
		&rdb.Column{Name: "ColA"},
		&rdb.Column{Name: "ColB"},
	})
	table.Row = []Row{
		{
			Field: []rdb.Nullable{
				{Value: "Hello"},
				{Value: 123.524},
			},
		},
		{
			Field: []rdb.Nullable{
				{Value: "Hi"},
				{Null: true},
			},
		},
	}
	if single {
		return table
	}
	tableNext := &Buffer{
		Name: "res2",
	}
	tableNext.SetSchema([]*rdb.Column{
		&rdb.Column{Name: "Col1"},
	})
	tableNext.Row = []Row{
		{
			Field: []rdb.Nullable{
				{Value: "ABC"},
			},
		},
		{
			Field: []rdb.Nullable{
				{Value: "XYZ"},
			},
		},
	}
	table.Set = []*Buffer{
		table,
		tableNext,
	}
	tableNext.Set = table.Set
	return table
}

func TestJsonMarshal(t *testing.T) {
	type jsonTest struct {
		name   string
		result string
		writer io.WriterTo
	}
	testTable := []jsonTest{
		jsonTest{
			name:   "Single Result JSON Row Object",
			result: `[{"ColA":"Hello","ColB":123.524},{"ColA":"Hi","ColB":null}]`,
			writer: &JsonRowObject{Buffer: getTable(true)},
		},
		jsonTest{
			name:   "Multiple Result JSON Row Object",
			result: `[[{"ColA":"Hello","ColB":123.524},{"ColA":"Hi","ColB":null}],[{"Col1":"ABC"},{"Col1":"XYZ"}]]`,
			writer: &JsonRowObject{Buffer: getTable(false)},
		},
		jsonTest{
			name:   "Single Result JSON Row Array",
			result: `{"T1":"HI","Name":"res1","Column":["ColA","ColB"],"Data":[["Hello",123.524],["Hi",null]]}`,
			writer: &JsonRowArray{
				Buffer: getTable(true),
				Meta: map[string]interface{}{
					"Column": "Ignored",
					"Name":   "Ignored Also",
					"T1":     "HI",
				},
			},
		},
		jsonTest{
			name:   "Multiple Result JSON Row Array",
			result: `[{"T1":"HI","Name":"res1","Column":["ColA","ColB"],"Data":[["Hello",123.524],["Hi",null]]},{"T1":"HI","Name":"res2","Column":["Col1"],"Data":[["ABC"],["XYZ"]]}]`,
			writer: &JsonRowArray{
				Buffer: getTable(false),
				Meta: map[string]interface{}{
					"Column": "Ignored",
					"Name":   "Ignored Also",
					"T1":     "HI",
				},
			},
		},
	}

	var err error
	buf := &bytes.Buffer{}
	for _, jt := range testTable {
		_, err = jt.writer.WriteTo(buf)
		if err != nil {
			t.Error(err)
		}
		if buf.String() != jt.result {
			t.Errorf("Test failed: %s\nDoesn't match:\nwant\n\t%s\ngot\n\t%s", jt.name, jt.result, buf.String())
		}
		buf.Reset()
	}
}
