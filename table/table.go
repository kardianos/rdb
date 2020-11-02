// Copyright 2014 Daniel Theophanes.
// Use of this source code is governed by a zlib-style
// license that can be found in the LICENSE file.

// table give a logical in-memory buffer row a database table.
/*

 */
package table

import (
	"errors"

	"github.com/kardianos/rdb"
)

type Row struct {
	buffer *Buffer
	Field  []rdb.Nullable
}

func (row Row) Index(name string) int {
	index, found := row.buffer.nameIndexLookup[name]
	if !found {
		panic(rdb.ErrorColumnNotFound{At: "Index", Name: name})
	}
	return index
}
func (row Row) Get(name string) interface{} {
	index, found := row.buffer.nameIndexLookup[name]
	if !found {
		panic(rdb.ErrorColumnNotFound{At: "Get", Name: name})
	}
	return row.Field[index].Value
}
func (row Row) GetN(name string) rdb.Nullable {
	index, found := row.buffer.nameIndexLookup[name]
	if !found {
		panic(rdb.ErrorColumnNotFound{At: "GetN", Name: name})
	}
	return row.Field[index]
}
func (row Row) HasColumn(name string) bool {
	_, found := row.buffer.nameIndexLookup[name]
	return found
}

type Buffer struct {
	Name            string
	Row             []Row
	schema          []*rdb.Column
	nameIndexLookup map[string]int

	// Truncated may be manually set to true if the returned row set has been truncated.
	Truncated bool

	// Result set, which should include current buffer if not nil.
	Set []*Buffer
}

func (b *Buffer) Len() int {
	return len(b.Row)
}
func Fill(res *rdb.Result) (*Buffer, error) {
	set, err := FillSet(res)
	if err != nil {
		return nil, err
	}
	if len(set) == 0 {
		return nil, nil
	}
	return set[0], nil
}

func FillSet(res *rdb.Result) ([]*Buffer, error) {
	defer res.Close()
	set := make([]*Buffer, 0, 1)
	for {
		tb := &Buffer{}

		err := tb.SetSchema(res.Schema())
		if err != nil {
			return nil, err
		}
		hasRow := false
		for res.Next() {
			hasRow = true
			err = res.Scan()
			if err != nil {
				return nil, err
			}
			tb.Row = append(tb.Row, Row{
				buffer: tb,
				Field:  res.GetRowN(),
			})
		}
		if len(tb.schema) != 0 || hasRow {
			set = append(set, tb)
		}

		nextRes, err := res.NextResult()
		if err != nil {
			return nil, err
		}
		if !nextRes {
			break
		}
	}
	for _, tb := range set {
		tb.Set = set
	}
	return set, nil
}
func FillCommand(q rdb.Queryer, cmd *rdb.Command, params ...rdb.Param) (*Buffer, error) {
	res, err := q.Query(cmd, params...)
	if err != nil {
		return nil, err
	}
	defer res.Close()
	return Fill(res)
}

var errSetSchema = errors.New("Can only set the schema when no rows exist.")

func (b *Buffer) SetSchema(schema []*rdb.Column) error {
	if len(b.Row) != 0 {
		return errSetSchema
	}
	b.schema = schema
	b.nameIndexLookup = make(map[string]int, len(b.schema))
	for i, col := range b.schema {
		b.nameIndexLookup[col.Name] = i
	}
	return nil
}

func (b *Buffer) Schema() []*rdb.Column {
	return b.schema
}

// ColumnIndex returns the index of the named column.
// If the name is not present it returns -1.
func (b *Buffer) ColumnIndex(name string) int {
	index, found := b.nameIndexLookup[name]
	if !found {
		return -1
	}
	return index
}

func (b *Buffer) AddRow(v ...interface{}) *Row {
	b.Row = append(b.Row, Row{buffer: b, Field: make([]rdb.Nullable, len(b.schema))})
	r := &b.Row[len(b.Row)-1]
	ct := len(v)
	if len(r.Field) < ct {
		ct = len(r.Field)
	}
	for i := 0; i < ct; i++ {
		r.Field[i].Null = v[i] == nil
		r.Field[i].Value = v[i]
	}
	return r
}

func (b *Buffer) AddBufferRow(row *Row) {
	x := *row
	x.buffer = b
	b.Row = append(b.Row, x)
}
