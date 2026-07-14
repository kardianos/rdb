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

// tdsPrepValuer Preps into the growing []T last element each row (Query style).
type tdsPrepValuer struct {
	out  *[]tdsStructRow
	prep [5]interface{}
	row  *tdsStructRow
}

func (v *tdsPrepValuer) Columns(cc []*rdb.Column) error {
	return nil
}
func (v *tdsPrepValuer) Done() error          { return nil }
func (v *tdsPrepValuer) Message(*rdb.Message) {}
func (v *tdsPrepValuer) RowsAffected(uint64)  {}
func (v *tdsPrepValuer) RowScanned() {
	// row already filled via Prep; append was done before Scan equivalent
}
func (v *tdsPrepValuer) PrepAt(i int) interface{} {
	if i < 0 || i >= len(v.prep) {
		return nil
	}
	return v.prep[i]
}
func (v *tdsPrepValuer) HasConverter(int) bool     { return false }
func (v *tdsPrepValuer) FieldNull(int) interface{} { return nil }
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

// BenchmarkTDS_PrepIntoSlice: wire → Prep/DirectAssign into []T (Query style).
func BenchmarkTDS_PrepIntoSlice(b *testing.B) {
	stream := syntheticIntRowStream(structBenchRows, 4, structBenchText)
	ctx := context.Background()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var out []tdsStructRow
		val := &tdsPrepValuer{out: &out}
		// beginRow before each row token — hook via wrapping
		wrapped := &prepRowStarter{inner: val}
		if err := runTokenStream(ctx, stream, wrapped, nil); err != nil {
			b.Fatal(err)
		}
		if len(out) != structBenchRows {
			b.Fatalf("len=%d", len(out))
		}
	}
}

// prepRowStarter starts a new struct row when columns are known and before each row.
type prepRowStarter struct {
	inner   *tdsPrepValuer
	started bool
}

func (p *prepRowStarter) Columns(cc []*rdb.Column) error {
	p.started = true
	return p.inner.Columns(cc)
}
func (p *prepRowStarter) Done() error          { return p.inner.Done() }
func (p *prepRowStarter) Message(m *rdb.Message) { p.inner.Message(m) }
func (p *prepRowStarter) RowsAffected(n uint64)  { p.inner.RowsAffected(n) }
func (p *prepRowStarter) RowScanned()             { p.inner.RowScanned() }
func (p *prepRowStarter) PrepAt(i int) interface{} {
	// DirectAssign uses PrepAt before WriteField; start a new row on column 0.
	if i == 0 {
		p.inner.beginRow()
	}
	return p.inner.PrepAt(i)
}
func (p *prepRowStarter) HasConverter(i int) bool { return p.inner.HasConverter(i) }
func (p *prepRowStarter) FieldNull(i int) interface{} {
	return p.inner.FieldNull(i)
}
func (p *prepRowStarter) WriteField(c *rdb.Column, value *rdb.DriverValue, assign rdb.Assigner) error {
	return p.inner.WriteField(c, value, assign)
}
