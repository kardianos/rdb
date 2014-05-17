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
	check := `[{"ColA":"Hello","ColB":123.524},{"ColA":"Hi","ColB":null}]`
	buf := &Buffer{}
	buf.SetSchema([]*rdb.SqlColumn{
		&rdb.SqlColumn{Name: "ColA"},
		&rdb.SqlColumn{Name: "ColB"},
	})
	buf.Row = []Row{
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

	bb := &bytes.Buffer{}

	coder := JsonObjectArray{Buffer: buf}
	_, err := coder.WriteTo(bb)
	if err != nil {
		t.Error(err)
	}
	if bb.String() != check {
		t.Errorf("Doesn't match: want <%s> got <%s>", check, bb.String())
	}
}
