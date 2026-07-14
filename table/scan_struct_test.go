// Copyright 2014 Daniel Theophanes.
// Use of this source code is governed by a zlib-style
// license that can be found in the LICENSE file.

package table

import (
	"bytes"
	"context"
	"errors"
	"io"
	"strings"
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

func TestPlanOptAndNullFlag(t *testing.T) {
	type Row struct {
		ID       int32           `db:"id"`
		Name     rdb.Opt[string] `db:"name"`
		Region   string          `db:"region"`
		RegNull  bool            `null:"region"`
		Blob     io.Writer       `db:"blob"`
	}
	schema := []*rdb.Column{
		{Name: "id", Index: 0, Nullable: false},
		{Name: "name", Index: 1, Nullable: true},
		{Name: "region", Index: 2, Nullable: true},
		{Name: "blob", Index: 3, Nullable: false},
	}
	plan, err := newStructPlan[Row](schema, "db")
	if err != nil {
		t.Fatal(err)
	}
	if len(plan.fields) != 4 {
		t.Fatalf("fields=%d want 4 (flag field not a column bind)", len(plan.fields))
	}
	byCol := map[int]fieldMode{}
	for _, f := range plan.fields {
		byCol[f.colIdx] = f.mode
	}
	if byCol[0] != modeDirect {
		t.Errorf("id mode=%v", byCol[0])
	}
	if byCol[1] != modeDirect {
		t.Errorf("Opt name mode=%v want direct", byCol[1])
	}
	if byCol[2] != modeFlag {
		t.Errorf("region mode=%v want flag", byCol[2])
	}
	if byCol[3] != modeDirect {
		t.Errorf("blob mode=%v", byCol[3])
	}

	var row Row
	var buf bytes.Buffer
	row.Blob = &buf
	base := unsafe.Pointer(&row)

	// id
	for _, f := range plan.fields {
		switch f.colIdx {
		case 0:
			p := f.prep(base).(*int32)
			*p = 1
		case 1:
			p := f.prep(base).(*rdb.Opt[string])
			handled, err := rdb.DirectAssignBytes(p, []byte("alice"), false, true, nil)
			if !handled || err != nil {
				t.Fatalf("opt assign: %v %v", handled, err)
			}
		case 2:
			sink := f.flagPrep(base).(*rdb.NullFlagPrep)
			handled, err := rdb.DirectAssignBytes(sink, []byte("US"), false, true, nil)
			if !handled || err != nil {
				t.Fatalf("flag assign: %v %v", handled, err)
			}
			if row.Region != "US" || row.RegNull {
				t.Fatalf("region=%q null=%v", row.Region, row.RegNull)
			}
			handled, err = rdb.DirectAssignBytes(sink, nil, true, true, nil)
			if !handled || err != nil {
				t.Fatalf("flag null: %v %v", handled, err)
			}
			if row.Region != "" || !row.RegNull {
				t.Fatalf("after null region=%q null=%v", row.Region, row.RegNull)
			}
		case 3:
			w := f.prep(base)
			handled, err := rdb.DirectAssignBytes(w, []byte("xyz"), false, false, nil)
			if !handled || err != nil {
				t.Fatalf("writer: %v %v", handled, err)
			}
			if buf.String() != "xyz" {
				t.Fatalf("buf=%q", buf.String())
			}
		}
	}
	if row.ID != 1 || !row.Name.Valid || row.Name.V != "alice" {
		t.Fatalf("row=%+v", row)
	}
}

func TestNullTagErrors(t *testing.T) {
	type badSelf struct {
		NameNull bool `null:"NameNull"`
	}
	_, err := newStructPlan[badSelf](nil, "db")
	if err == nil {
		t.Fatal("expected error for self-referential null tag")
	}

	type badMissing struct {
		NameNull bool `null:"missing"`
	}
	_, err = newStructPlan[badMissing](nil, "db")
	if err == nil {
		t.Fatal("expected error for missing null target")
	}

	type badType struct {
		Name     string `db:"name"`
		NameNull int    `null:"name"`
	}
	_, err = newStructPlan[badType]([]*rdb.Column{{Name: "name", Index: 0}}, "db")
	if err == nil {
		t.Fatal("expected error for non-bool null flag")
	}
}

func TestRejectPointerOpt(t *testing.T) {
	type Row struct {
		Name *rdb.Opt[string] `db:"name"`
	}
	_, err := newStructPlan[Row]([]*rdb.Column{{Name: "name", Index: 0, Nullable: true}}, "db")
	if err == nil {
		t.Fatal("expected error for *Opt[T] field")
	}
	if !strings.Contains(err.Error(), "*Opt") {
		t.Fatalf("err=%v", err)
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
	if err := plan.fields[0].applyNull(base, rdb.Nullable{Value: "bob"}); err != nil {
		t.Fatal(err)
	}
	if row.Name != "bob" {
		t.Fatalf("name=%q", row.Name)
	}
	if err := plan.fields[0].applyNull(base, rdb.Nullable{Null: true}); err != nil {
		t.Fatal(err)
	}
	if row.Name != "" {
		t.Fatalf("name after null=%q", row.Name)
	}
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
