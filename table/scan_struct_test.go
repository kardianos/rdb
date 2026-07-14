// Copyright 2014 Daniel Theophanes.
// Use of this source code is governed by a zlib-style
// license that can be found in the LICENSE file.

package table

import (
	"context"
	"errors"
	"testing"

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
	// id direct, name viaNull (nullable), age viaNull (pointer)
	byCol := map[int]fieldBind{}
	for _, f := range plan.fields {
		byCol[f.colIdx] = f
	}
	if byCol[0].viaNull {
		t.Error("id should Prep directly")
	}
	if !byCol[1].viaNull {
		t.Error("name should use Nullable path")
	}
	if !byCol[2].viaNull {
		t.Error("age pointer should use Nullable path")
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

// fakeQueryer implements rdb.Queryer for API wiring tests (Query error path).
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
