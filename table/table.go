// Copyright 2014 Daniel Theophanes.
// Use of this source code is governed by a zlib-style
// license that can be found in the LICENSE file.

// table give a logical in-memory buffer row a database table.
/*

*/
package table

import (
	"errors"

	"bitbucket.org/kardianos/rdb"
)

type Row struct {
	buffer *Buffer
	Field  []rdb.Nullable
}

func (row Row) Get(name string) (interface{}, error) {
	index, found := row.buffer.nameIndexLookup[name]
	if !found {
		return rdb.Nullable{}, rdb.ErrorColumnNotFound{At: "Get", Name: name}
	}
	return row.Field[index].Value, nil
}
func (row Row) GetN(name string) (rdb.Nullable, error) {
	index, found := row.buffer.nameIndexLookup[name]
	if !found {
		return rdb.Nullable{}, rdb.ErrorColumnNotFound{At: "GetN", Name: name}
	}
	return row.Field[index], nil
}

type Buffer struct {
	Row             []Row
	schema          []*rdb.Column
	nameIndexLookup map[string]int
}

func (b *Buffer) Len() int {
	return len(b.Row)
}

func Fill(res *rdb.Result) (*Buffer, error) {
	tb := &Buffer{}

	err := tb.SetSchema(res.Schema())
	if err != nil {
		return nil, err
	}
	for res.Next() {
		err = res.Scan()
		if err != nil {
			return nil, err
		}
		tb.Row = append(tb.Row, Row{
			buffer: tb,
			Field:  res.GetRowN(),
		})
	}
	return tb, nil
}
func FillCommand(cp *rdb.ConnPool, cmd *rdb.Command, params ...rdb.Param) (*Buffer, error) {
	res, err := cp.Query(cmd, params...)
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
