// Copyright 2014 Daniel Theophanes.
// Use of this source code is governed by a zlib-style
// license that can be found in the LICENSE file.

package ms

import (
	"bytes"
	"context"
	"testing"

	"github.com/kardianos/rdb"
	"github.com/kardianos/rdb/table"
)

// Full-stack: synthetic TDS token stream → either Buffer+UnmarshalStruct
// or Prep into []struct via a DriverValuer that uses table.Query's plan style.

const structBenchRows = 200
const structBenchText = "hello world row text"

// tdsBufferValuer accumulates rows like Fill (no Prep).
type tdsBufferValuer struct {
	buf   *table.Buffer
	cur   []rdb.Nullable
	ncols int
}

func (v *tdsBufferValuer) Columns(cc []*rdb.Column) error {
	v.buf = &table.Buffer{}
	if err := v.buf.SetSchema(cc); err != nil {
		return err
	}
	v.ncols = len(cc)
	v.cur = make([]rdb.Nullable, v.ncols)
	return nil
}
func (v *tdsBufferValuer) Done() error                 { return nil }
func (v *tdsBufferValuer) Message(*rdb.Message)        {}
func (v *tdsBufferValuer) RowsAffected(uint64)         {}
func (v *tdsBufferValuer) RowScanned()                 {}
func (v *tdsBufferValuer) WriteField(c *rdb.Column, value *rdb.DriverValue, assign rdb.Assigner) error {
	// Match valuer buffer path (MustCopy).
	var val interface{}
	if !value.Null {
		val = value.Value
		if value.MustCopy {
			if bb, ok := val.([]byte); ok {
				cp := make([]byte, len(bb))
				copy(cp, bb)
				val = cp
			}
		}
	}
	v.cur[c.Index] = rdb.Nullable{Null: value.Null, Value: val}
	// Last column finishes the row.
	if c.Index == v.ncols-1 {
		fields := make([]rdb.Nullable, v.ncols)
		copy(fields, v.cur)
		// Use AddRow via values for simplicity.
		vals := make([]interface{}, v.ncols)
		for i, n := range fields {
			if n.Null {
				vals[i] = nil
			} else {
				vals[i] = n.Value
			}
		}
		v.buf.AddRow(vals...)
		for i := range v.cur {
			v.cur[i] = rdb.Nullable{Null: true}
		}
	}
	return nil
}

type tdsStructRow struct {
	C0   int32  `db:"c0"`
	C1   int32  `db:"c1"`
	C2   int32  `db:"c2"`
	C3   int32  `db:"c3"`
	Name string `db:"name"`
}

type tdsStructRowOpt struct {
	C0   int32           `db:"c0"`
	C1   int32           `db:"c1"`
	C2   int32           `db:"c2"`
	C3   int32           `db:"c3"`
	Name rdb.Opt[string] `db:"name"`
}

type tdsStructRowFlag struct {
	C0       int32  `db:"c0"`
	C1       int32  `db:"c1"`
	C2       int32  `db:"c2"`
	C3       int32  `db:"c3"`
	Name     string `db:"name"`
	NameNull bool   `null:"name"`
}

// tdsPrepValuer Preps into the growing []T last element each row (Query style).
type tdsPrepValuer struct {
	out  *[]tdsStructRow
	prep [5]interface{}
	row  *tdsStructRow
}

func (v *tdsPrepValuer) Columns(cc []*rdb.Column) error { return nil }
func (v *tdsPrepValuer) Done() error                    { return nil }
func (v *tdsPrepValuer) Message(*rdb.Message)           {}
func (v *tdsPrepValuer) RowsAffected(uint64)            {}
func (v *tdsPrepValuer) RowScanned()                    {}
func (v *tdsPrepValuer) PrepAt(i int) interface{} {
	if i < 0 || i >= len(v.prep) {
		return nil
	}
	return v.prep[i]
}
func (v *tdsPrepValuer) HasConverter(int) bool          { return false }
func (v *tdsPrepValuer) FieldNull(int) interface{}      { return nil }
func (v *tdsPrepValuer) WriteField(c *rdb.Column, value *rdb.DriverValue, assign rdb.Assigner) error {
	return rdb.AssignValue(c, rdb.Nullable{Null: value.Null, Value: value.Value}, v.PrepAt(c.Index), assign)
}
func (v *tdsPrepValuer) beginRow() {
	*v.out = append(*v.out, tdsStructRow{})
	v.row = &(*v.out)[len(*v.out)-1]
	v.prep[0] = &v.row.C0
	v.prep[1] = &v.row.C1
	v.prep[2] = &v.row.C2
	v.prep[3] = &v.row.C3
	v.prep[4] = &v.row.Name
}

type tdsPrepValuerOpt struct {
	out  *[]tdsStructRowOpt
	prep [5]interface{}
	row  *tdsStructRowOpt
}

func (v *tdsPrepValuerOpt) Columns([]*rdb.Column) error { return nil }
func (v *tdsPrepValuerOpt) Done() error                 { return nil }
func (v *tdsPrepValuerOpt) Message(*rdb.Message)        {}
func (v *tdsPrepValuerOpt) RowsAffected(uint64)         {}
func (v *tdsPrepValuerOpt) RowScanned()                 {}
func (v *tdsPrepValuerOpt) PrepAt(i int) interface{} {
	if i < 0 || i >= len(v.prep) {
		return nil
	}
	return v.prep[i]
}
func (v *tdsPrepValuerOpt) HasConverter(int) bool     { return false }
func (v *tdsPrepValuerOpt) FieldNull(int) interface{} { return nil }
func (v *tdsPrepValuerOpt) WriteField(c *rdb.Column, value *rdb.DriverValue, assign rdb.Assigner) error {
	return rdb.AssignValue(c, rdb.Nullable{Null: value.Null, Value: value.Value}, v.PrepAt(c.Index), assign)
}
func (v *tdsPrepValuerOpt) beginRow() {
	*v.out = append(*v.out, tdsStructRowOpt{})
	v.row = &(*v.out)[len(*v.out)-1]
	v.prep[0] = &v.row.C0
	v.prep[1] = &v.row.C1
	v.prep[2] = &v.row.C2
	v.prep[3] = &v.row.C3
	v.prep[4] = &v.row.Name
}

type tdsPrepValuerFlag struct {
	out      *[]tdsStructRowFlag
	prep     [5]interface{}
	row      *tdsStructRowFlag
	nameSink rdb.NullFlagPrep
}

func (v *tdsPrepValuerFlag) Columns([]*rdb.Column) error { return nil }
func (v *tdsPrepValuerFlag) Done() error                 { return nil }
func (v *tdsPrepValuerFlag) Message(*rdb.Message)        {}
func (v *tdsPrepValuerFlag) RowsAffected(uint64)         {}
func (v *tdsPrepValuerFlag) RowScanned()                 {}
func (v *tdsPrepValuerFlag) PrepAt(i int) interface{} {
	if i < 0 || i >= len(v.prep) {
		return nil
	}
	return v.prep[i]
}
func (v *tdsPrepValuerFlag) HasConverter(int) bool     { return false }
func (v *tdsPrepValuerFlag) FieldNull(int) interface{} { return nil }
func (v *tdsPrepValuerFlag) WriteField(c *rdb.Column, value *rdb.DriverValue, assign rdb.Assigner) error {
	return rdb.AssignValue(c, rdb.Nullable{Null: value.Null, Value: value.Value}, v.PrepAt(c.Index), assign)
}
func (v *tdsPrepValuerFlag) beginRow() {
	*v.out = append(*v.out, tdsStructRowFlag{})
	v.row = &(*v.out)[len(*v.out)-1]
	v.prep[0] = &v.row.C0
	v.prep[1] = &v.row.C1
	v.prep[2] = &v.row.C2
	v.prep[3] = &v.row.C3
	v.nameSink.Value = &v.row.Name
	v.nameSink.Null = &v.row.NameNull
	v.prep[4] = &v.nameSink
}

func runTokenStream(ctx context.Context, stream []byte, val rdb.DriverValuer, onRow func()) error {
	r := bytes.NewReader(stream)
	pr := NewPacketReader(&deadlineNop{r: r})
	mr := pr.BeginMessage(ctx, packetTabularResult)
	defer mr.Close()
	conn := &Connection{pr: pr, mr: mr, val: val}
	for {
		res, err := conn.getSingleResponse(ctx, mr, true)
		if err != nil {
			return err
		}
		switch res.(type) {
		case MsgEom, MsgFinalDone:
			return nil
		case MsgRow:
			if onRow != nil {
				onRow()
			}
			if rv, ok := val.(interface{ RowScanned() }); ok {
				rv.RowScanned()
			}
		}
	}
}

// BenchmarkTDS_FillThenUnmarshal: wire → Buffer (Fill-like) → UnmarshalStruct.
func BenchmarkTDS_FillThenUnmarshal(b *testing.B) {
	// 4 ints + name (nullable text on wire as nvarchar non-null)
	stream := syntheticIntRowStream(structBenchRows, 4, structBenchText)
	ctx := context.Background()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		val := &tdsBufferValuer{}
		if err := runTokenStream(ctx, stream, val, nil); err != nil {
			b.Fatal(err)
		}
		// Column names are c0..c3,name — match tdsStructRow tags.
		out, err := table.UnmarshalStruct[tdsStructRow](val.buf)
		if err != nil {
			b.Fatal(err)
		}
		if len(out) != structBenchRows {
			b.Fatalf("len=%d", len(out))
		}
	}
}

// BenchmarkTDS_PrepIntoSlice: wire → Prep into plain string (nullable col, may box via fallback).
func BenchmarkTDS_PrepIntoSlice(b *testing.B) {
	stream := syntheticIntRowStream(structBenchRows, 4, structBenchText)
	ctx := context.Background()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var out []tdsStructRow
		val := &tdsPrepValuer{out: &out}
		wrapped := &prepRowStarter{begin: val.beginRow, prep: val, write: val}
		if err := runTokenStream(ctx, stream, wrapped, nil); err != nil {
			b.Fatal(err)
		}
		if len(out) != structBenchRows {
			b.Fatalf("len=%d", len(out))
		}
	}
}

// BenchmarkTDS_PrepOpt: wire → Prep into Opt[string] (no interface box for name).
func BenchmarkTDS_PrepOpt(b *testing.B) {
	stream := syntheticIntRowStream(structBenchRows, 4, structBenchText)
	ctx := context.Background()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var out []tdsStructRowOpt
		val := &tdsPrepValuerOpt{out: &out}
		wrapped := &prepRowStarter{begin: val.beginRow, prep: val, write: val}
		if err := runTokenStream(ctx, stream, wrapped, nil); err != nil {
			b.Fatal(err)
		}
		if len(out) != structBenchRows {
			b.Fatalf("len=%d", len(out))
		}
		if !out[structBenchRows-1].Name.Valid {
			b.Fatal("expected name Valid")
		}
	}
}

// BenchmarkTDS_PrepNullFlag: wire → Prep into string + NameNull bool via NullFlagPrep.
func BenchmarkTDS_PrepNullFlag(b *testing.B) {
	stream := syntheticIntRowStream(structBenchRows, 4, structBenchText)
	ctx := context.Background()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var out []tdsStructRowFlag
		val := &tdsPrepValuerFlag{out: &out}
		wrapped := &prepRowStarter{begin: val.beginRow, prep: val, write: val}
		if err := runTokenStream(ctx, stream, wrapped, nil); err != nil {
			b.Fatal(err)
		}
		if len(out) != structBenchRows {
			b.Fatalf("len=%d", len(out))
		}
		if out[structBenchRows-1].NameNull {
			b.Fatal("expected non-null name")
		}
	}
}

// prepRowStarter starts a new struct row on first PrepAt of each row (column 0).
type prepRowStarter struct {
	begin func()
	prep  interface {
		PrepAt(int) interface{}
		HasConverter(int) bool
		FieldNull(int) interface{}
	}
	write interface {
		Columns([]*rdb.Column) error
		Done() error
		Message(*rdb.Message)
		RowsAffected(uint64)
		RowScanned()
		WriteField(*rdb.Column, *rdb.DriverValue, rdb.Assigner) error
	}
}

func (p *prepRowStarter) Columns(cc []*rdb.Column) error { return p.write.Columns(cc) }
func (p *prepRowStarter) Done() error                    { return p.write.Done() }
func (p *prepRowStarter) Message(m *rdb.Message)         { p.write.Message(m) }
func (p *prepRowStarter) RowsAffected(n uint64)          { p.write.RowsAffected(n) }
func (p *prepRowStarter) RowScanned()                    { p.write.RowScanned() }
func (p *prepRowStarter) HasConverter(i int) bool        { return p.prep.HasConverter(i) }
func (p *prepRowStarter) FieldNull(i int) interface{}    { return p.prep.FieldNull(i) }
func (p *prepRowStarter) PrepAt(i int) interface{} {
	if i == 0 {
		p.begin()
	}
	return p.prep.PrepAt(i)
}
func (p *prepRowStarter) WriteField(c *rdb.Column, value *rdb.DriverValue, assign rdb.Assigner) error {
	return p.write.WriteField(c, value, assign)
}
