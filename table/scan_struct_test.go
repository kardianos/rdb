// Copyright 2014 Daniel Theophanes.
// Use of this source code is governed by a zlib-style
// license that can be found in the LICENSE file.

package table

import (
	"context"
	"errors"
	"testing"
	"unsafe"

	"github.com/kardianos/rdb"
)

func TestNewStructPlan(t *testing.T) {
	type Row struct {
		ID   int32  `db:"id"`
		Name string `db:"name"`
		Skip string `db:"-"`
		Age  *int32 `db:"age"`
	}
	schema := []*rdb.Column{
		{Name: "id", Index: 0, Nullable: false},
		{Name: "name", Index: 1, Nullable: true},
		{Name: "age", Index: 2, Nullable: true},
	}
	plan, err := newStructPlan[Row](schema, "db")
	if err != nil {
		t.Fatal(err)
	}
	if plan.asPtr {
		t.Fatal("Row should not be pointer plan")
	}
	if len(plan.fields) != 3 {
		t.Fatalf("fields=%d want 3", len(plan.fields))
	}
	byCol := map[int]fieldBind{}
	for _, f := range plan.fields {
		byCol[f.colIdx] = f
	}
	if byCol[0].mode != modeDirect || byCol[0].prep == nil {
		t.Error("id should be modeDirect with prep func")
	}
	if byCol[1].mode != modeNull || byCol[1].applyNull == nil {
		t.Error("name should be modeNull with applyNull")
	}
	if byCol[2].mode != modeNull || byCol[2].applyNull == nil {
		t.Error("age pointer should be modeNull with applyNull")
	}
}

func TestDirectPrepWritesField(t *testing.T) {
	type Row struct {
		ID   int32  `db:"id"`
		Name string `db:"name"`
	}
	schema := []*rdb.Column{
		{Name: "id", Index: 0, Nullable: false},
		{Name: "name", Index: 1, Nullable: false},
	}
	plan, err := newStructPlan[Row](schema, "db")
	if err != nil {
		t.Fatal(err)
	}
	var row Row
	base := unsafe.Pointer(&row)
	for _, f := range plan.fields {
		if f.mode != modeDirect {
			t.Fatalf("col %d mode=%v", f.colIdx, f.mode)
		}
		ptr := f.prep(base)
		switch f.colIdx {
		case 0:
			*ptr.(*int32) = 42
		case 1:
			*ptr.(*string) = "alice"
		}
	}
	if row.ID != 42 || row.Name != "alice" {
		t.Fatalf("row=%+v", row)
	}
}

func TestNullApply(t *testing.T) {
	type Row struct {
		Name string `db:"name"`
		Age  *int32 `db:"age"`
	}
	schema := []*rdb.Column{
		{Name: "name", Index: 0, Nullable: true},
		{Name: "age", Index: 1, Nullable: true},
	}
	plan, err := newStructPlan[Row](schema, "db")
	if err != nil {
		t.Fatal(err)
	}
	var row Row
	base := unsafe.Pointer(&row)
	for _, f := range plan.fields {
		if f.mode != modeNull {
			t.Fatalf("col %d want modeNull", f.colIdx)
		}
	}
	// name non-null
	if err := plan.fields[0].applyNull(base, rdb.Nullable{Value: "bob"}); err != nil {
		t.Fatal(err)
	}
	if row.Name != "bob" {
		t.Fatalf("name=%q", row.Name)
	}
	// name null → zero
	if err := plan.fields[0].applyNull(base, rdb.Nullable{Null: true}); err != nil {
		t.Fatal(err)
	}
	if row.Name != "" {
		t.Fatalf("name after null=%q", row.Name)
	}
	// age pointer
	if err := plan.fields[1].applyNull(base, rdb.Nullable{Value: int32(9)}); err != nil {
		t.Fatal(err)
	}
	if row.Age == nil || *row.Age != 9 {
		t.Fatalf("age=%v", row.Age)
	}
	if err := plan.fields[1].applyNull(base, rdb.Nullable{Null: true}); err != nil {
		t.Fatal(err)
	}
	if row.Age != nil {
		t.Fatalf("age after null=%v", row.Age)
	}
}

func TestNewStructPlanSkipsUnexported(t *testing.T) {
	type Row struct {
		ID     int32 `db:"id"`
		hidden string
	}
	plan, err := newStructPlan[Row]([]*rdb.Column{{Name: "id", Index: 0}}, "db")
	if err != nil {
		t.Fatal(err)
	}
	if len(plan.fields) != 1 {
		t.Fatalf("fields=%d want 1", len(plan.fields))
	}
}

func TestNewStructPlanMissingColumn(t *testing.T) {
	type Row struct {
		ID int `db:"missing"`
	}
	_, err := newStructPlan[Row]([]*rdb.Column{{Name: "id", Index: 0}}, "db")
	if err == nil {
		t.Fatal("expected error for missing column")
	}
}

func TestNewStructPlanNotStruct(t *testing.T) {
	_, err := newStructPlan[int](nil, "db")
	if err == nil {
		t.Fatal("expected error for non-struct T")
	}
}

type fakeQueryer struct {
	err error
}

func (f fakeQueryer) Query(ctx context.Context, cmd *rdb.Command, params ...rdb.Param) (*rdb.Result, error) {
	return nil, f.err
}

func TestQueryPropagatesQueryError(t *testing.T) {
	want := errors.New("boom")
	_, err := Query[struct {
		A int `db:"a"`
	}](context.Background(), fakeQueryer{err: want}, &rdb.Command{SQL: "select 1"})
	if !errors.Is(err, want) {
		t.Fatalf("err=%v want %v", err, want)
	}
}

func TestStreamPropagatesQueryError(t *testing.T) {
	want := errors.New("boom")
	var got error
	for _, err := range Stream[struct {
		A int `db:"a"`
	}](context.Background(), fakeQueryer{err: want}, &rdb.Command{SQL: "select 1"}) {
		got = err
		break
	}
	if !errors.Is(got, want) {
		t.Fatalf("err=%v want %v", got, want)
	}
}
