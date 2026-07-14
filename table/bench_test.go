// Copyright 2014 Daniel Theophanes.
// Use of this source code is governed by a zlib-style
// license that can be found in the LICENSE file.

package table

import (
	"reflect"
	"testing"
	"unsafe"

	"github.com/kardianos/rdb"
)

// benchRow mirrors a typical app DTO: non-null ints (Prep/DirectAssign)
// plus nullable text (Nullable path).
type benchRow struct {
	ID     int32  `db:"id"`
	Code   int32  `db:"code"`
	Flags  int32  `db:"flags"`
	Score  int32  `db:"score"`
	Name   string `db:"name"`
	Region string `db:"region"`
}

const (
	benchRows   = 200
	benchName   = "hello world row text"
	benchRegion = "US-WEST"
)

func benchSchema() []*rdb.Column {
	return []*rdb.Column{
		{Name: "id", Index: 0, Nullable: false, Type: rdb.TypeInt32},
		{Name: "code", Index: 1, Nullable: false, Type: rdb.TypeInt32},
		{Name: "flags", Index: 2, Nullable: false, Type: rdb.TypeInt32},
		{Name: "score", Index: 3, Nullable: false, Type: rdb.TypeInt32},
		{Name: "name", Index: 4, Nullable: true, Type: rdb.TypeVarChar},
		{Name: "region", Index: 5, Nullable: true, Type: rdb.TypeVarChar},
	}
}

func benchBuffer(rows int) *Buffer {
	schema := benchSchema()
	buf := &Buffer{}
	if err := buf.SetSchema(schema); err != nil {
		panic(err)
	}
	for i := 0; i < rows; i++ {
		buf.AddRow(
			int32(i),
			int32(i*10),
			int32(i%7),
			int32(1000+i),
			benchName,
			benchRegion,
		)
	}
	return buf
}

type rowData struct {
	id, code, flags, score int32
	name, region           string
}

func makeRowData(rows int) []rowData {
	out := make([]rowData, rows)
	for i := range out {
		out[i] = rowData{
			id: int32(i), code: int32(i * 10), flags: int32(i % 7), score: int32(1000 + i),
			name: benchName, region: benchRegion,
		}
	}
	return out
}

// BenchmarkUnmarshalStruct is the classic path: already-filled Buffer → []T
// via reflect per field (no Prep). Buffer construction is outside the timer
// so this isolates UnmarshalStruct only.
func BenchmarkUnmarshalStruct(b *testing.B) {
	buf := benchBuffer(benchRows)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		out, err := UnmarshalStruct[benchRow](buf)
		if err != nil {
			b.Fatal(err)
		}
		if len(out) != benchRows {
			b.Fatalf("len=%d", len(out))
		}
	}
}

// BenchmarkPlanScanDirect is the current Query/Stream mapping loop: one plan,
// then per row offset prep + typed apply (no Result/network). Simulates driver
// DirectAssign by writing through prep pointers.
func BenchmarkPlanScanDirect(b *testing.B) {
	schema := benchSchema()
	data := makeRowData(benchRows)
	plan, err := newStructPlan[benchRow](schema, "db")
	if err != nil {
		b.Fatal(err)
	}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		out := make([]benchRow, 0, benchRows)
		for r := 0; r < benchRows; r++ {
			out = append(out, benchRow{})
			base := unsafe.Pointer(&out[len(out)-1])
			d := &data[r]
			for fi := range plan.fields {
				f := &plan.fields[fi]
				switch f.mode {
				case modeDirect:
					switch f.colIdx {
					case 0:
						*f.prep(base).(*int32) = d.id
					case 1:
						*f.prep(base).(*int32) = d.code
					case 2:
						*f.prep(base).(*int32) = d.flags
					case 3:
						*f.prep(base).(*int32) = d.score
					default:
						b.Fatalf("unexpected direct col %d", f.colIdx)
					}
				case modeNull:
					var n rdb.Nullable
					switch f.colIdx {
					case 4:
						n.Value = d.name
					case 5:
						n.Value = d.region
					}
					if err := f.applyNull(base, n); err != nil {
						b.Fatal(err)
					}
				}
			}
		}
		if out[benchRows-1].ID != int32(benchRows-1) {
			b.Fatal(out[benchRows-1])
		}
	}
}

// BenchmarkPlanScanReflectLoop approximates the first Query/Stream design:
// plan once (col→field map), but each row uses reflect.Value.Field/Set.
func BenchmarkPlanScanReflectLoop(b *testing.B) {
	schema := benchSchema()
	data := makeRowData(benchRows)
	tType := reflect.TypeOf(benchRow{})
	fieldByCol := make([]int, len(schema))
	for i := range fieldByCol {
		fieldByCol[i] = -1
	}
	for i := 0; i < tType.NumField(); i++ {
		f := tType.Field(i)
		tag := f.Tag.Get("db")
		if tag == "" || tag == "-" {
			continue
		}
		for ci, col := range schema {
			if col.Name == tag {
				fieldByCol[ci] = i
			}
		}
	}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		out := make([]benchRow, 0, benchRows)
		for r := 0; r < benchRows; r++ {
			out = append(out, benchRow{})
			rv := reflect.ValueOf(&out[len(out)-1]).Elem()
			d := &data[r]
			for col, fi := range fieldByCol {
				if fi < 0 {
					continue
				}
				f := rv.Field(fi)
				switch col {
				case 0:
					f.SetInt(int64(d.id))
				case 1:
					f.SetInt(int64(d.code))
				case 2:
					f.SetInt(int64(d.flags))
				case 3:
					f.SetInt(int64(d.score))
				case 4:
					f.SetString(d.name)
				case 5:
					f.SetString(d.region)
				}
			}
		}
		if out[benchRows-1].ID != int32(benchRows-1) {
			b.Fatal(out[benchRows-1])
		}
	}
}

// BenchmarkFillBufferOnly measures building the intermediate Buffer
// (Fill materialization without UnmarshalStruct).
func BenchmarkFillBufferOnly(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf := benchBuffer(benchRows)
		if buf.Len() != benchRows {
			b.Fatal(buf.Len())
		}
	}
}

// BenchmarkFillThenUnmarshal is the full classic pipeline: build Buffer then
// UnmarshalStruct (what apps effectively pay for Fill + UnmarshalStruct).
func BenchmarkFillThenUnmarshal(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf := benchBuffer(benchRows)
		out, err := UnmarshalStruct[benchRow](buf)
		if err != nil {
			b.Fatal(err)
		}
		if len(out) != benchRows {
			b.Fatal(len(out))
		}
	}
}
