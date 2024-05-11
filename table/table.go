// Copyright 2014 Daniel Theophanes.
// Use of this source code is governed by a zlib-style
// license that can be found in the LICENSE file.

// Package table give a logical in-memory buffer row a database table.
/*

 */
package table

import (
	"context"
	"errors"

	"github.com/kardianos/rdb"
)

// Row represents a single buffer row of a table.
type Row struct {
	buffer *Buffer
	Field  []rdb.Nullable
}

// Index returns the index of column name.
func (row Row) Index(name string) int {
	index, found := row.buffer.nameIndexLookup[name]
	if !found {
		panic(rdb.ErrorColumnNotFound{At: "Index", Name: name})
	}
	return index
}

// Get returns the column name as a value, nil if null. Panics if name is not a valid column.
func (row Row) Get(name string) interface{} {
	index, found := row.buffer.nameIndexLookup[name]
	if !found {
		panic(rdb.ErrorColumnNotFound{At: "Get", Name: name})
	}
	return row.Field[index].Value
}

// GetN returns a Nullable under the column name. If name does not exist, panics.
func (row Row) GetN(name string) rdb.Nullable {
	index, found := row.buffer.nameIndexLookup[name]
	if !found {
		panic(rdb.ErrorColumnNotFound{At: "GetN", Name: name})
	}
	return row.Field[index]
}

// Set the value v on the row in column name. If name does not exist, panics.
func (row Row) Set(name string, v interface{}) {
	index, found := row.buffer.nameIndexLookup[name]
	if !found {
		panic(rdb.ErrorColumnNotFound{At: "Set", Name: name})
	}
	n := rdb.Nullable{Value: v}
	if v == nil {
		n.Null = true
	}
	row.Field[index] = n
}

// SetN sets the nullable value on column name. If name does not exist, panics.
func (row Row) SetN(name string, v rdb.Nullable) {
	index, found := row.buffer.nameIndexLookup[name]
	if !found {
		panic(rdb.ErrorColumnNotFound{At: "SetN", Name: name})
	}
	row.Field[index] = v
}

// HasColumn returns true if the row contains the column name.
func (row Row) HasColumn(name string) bool {
	_, found := row.buffer.nameIndexLookup[name]
	return found
}

// Buffer represents a table buffer that holds one or more query.
type Buffer struct {
	// Name of the table in the buffer. May be manually set for further encoding.
	Name            string
	Row             []Row // Row data.
	schema          []*rdb.Column
	nameIndexLookup map[string]int

	// Truncated may be manually set to true if the returned row set has been truncated.
	Truncated bool

	// Result set, which should include current buffer if not nil.
	Set []*Buffer
}

// Len is the number of rows in the table buffer.
func (b *Buffer) Len() int {
	if b == nil {
		return 0
	}
	return len(b.Row)
}

// Fill the query result and return the buffer.
func Fill(ctx context.Context, res *rdb.Result) (*Buffer, error) {
	set, err := FillSet(res)
	if err != nil {
		return nil, err
	}
	if len(set) == 0 {
		return nil, nil
	}
	return set[0], nil
}

// FillSet populates a query result and returns a table set.
func FillSet(res *rdb.Result) ([]*Buffer, error) {
	defer res.Close()
	set := make([]*Buffer, 0, 1)
	for {
		tb := &Buffer{}

		err := tb.SetSchema(res.Schema())
		if err != nil {
			return set, err
		}
		hasRow := false
		for res.Next() {
			hasRow = true
			err = res.Scan()
			if err != nil {
				return set, err
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
			return set, err
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

// FillCommand runs a query, fills the result, and closes the query result.
func FillCommand(ctx context.Context, q rdb.Queryer, cmd *rdb.Command, params ...rdb.Param) (*Buffer, error) {
	res, err := q.Query(ctx, cmd, params...)
	if err != nil {
		return nil, err
	}
	defer res.Close()
	return Fill(ctx, res)
}

var errSetSchema = errors.New("can only set the schema when no rows exist")

// SetSchema sets the table buffer schema manually.
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

// Schema returns the table column schema
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

// AddRow adds a slice of values.
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

// AddBufferRow adds a new row to the buffer manually.
func (b *Buffer) AddBufferRow(row Row) {
	row.buffer = b
	b.Row = append(b.Row, row)
}
