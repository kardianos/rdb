// Copyright 2014 Daniel Theophanes.
// Use of this source code is governed by a zlib-style
// license that can be found in the LICENSE file.

package ms

import (
	"math/big"
	"testing"

	"bitbucket.org/kardianos/rdb"
)

func TestOutputParam(t *testing.T) {
	if parallel {
		t.Parallel()
	}
	defer recoverTest(t)

	createProcDrop := &rdb.Command{
		Sql: `
if object_id('AddTen') is not null drop proc AddTen
		`,
		Arity: rdb.ZeroMust,
	}
	createProc := &rdb.Command{
		Sql: `
create proc dbo.AddTen (
	@p1 int,
	@p2 int output
)
as
begin
	select @p2 = @p1 + 10
end
		`,
		Arity: rdb.ZeroMust,
	}

	callProc := &rdb.Command{
		Sql:   `AddTen`,
		Arity: rdb.ZeroMust,
	}

	db.Query(createProcDrop)
	db.Query(createProc)

	var val int
	db.Query(callProc,
		rdb.Param{Name: "p1", Value: 5, Type: rdb.TypeInt32},
		rdb.Param{Name: "p2", Out: true, Value: &val, Type: rdb.TypeInt32},
	)
	if val != 15 {
		t.Fatalf("Incorrect value. Want 15 was %v", val)
	}
	assertFreeConns(t)
}

func TestOutputParamTypes(t *testing.T) {
	if parallel {
		t.Parallel()
	}
	defer recoverTest(t)

	createProcDrop := &rdb.Command{
		Sql: `
if object_id('DataTypes') is not null drop proc DataTypes
		`,
		Arity: rdb.ZeroMust,
	}
	createProc := &rdb.Command{
		Sql: `
create proc dbo.DataTypes (
	@p1 nvarchar(max) output,
	@p2 int output,
	@p3 decimal(38,7) output
)
as
begin
	select @p1 = 'Hello', @p2 = 42, @p3 = 45.678
end
		`,
		Arity: rdb.ZeroMust,
	}

	callProc := &rdb.Command{
		Sql:   `DataTypes`,
		Arity: rdb.ZeroMust,
	}

	db.Query(createProcDrop)
	db.Query(createProc)

	var val1 string
	var val2 int
	var val3 = big.NewRat(5, 1)

	db.Query(callProc,
		rdb.Param{Name: "p1", Out: true, Value: &val1, Type: rdb.Text},
		rdb.Param{Name: "p2", Out: true, Value: &val2, Type: rdb.Integer},
		rdb.Param{Name: "p3", Out: true, Value: &val3, Precision: 38, Scale: 7, Type: rdb.Decimal},
	)
	if val1 != "Hello" {
		t.Fatalf("Incorrect value. Want 'Hello' was %v", val1)
	}
	if val2 != 42 {
		t.Fatalf("Incorrect value. Want 42 was %v", val2)
	}
	if val3.FloatString(3) != "45.678" {
		t.Fatalf("Incorrect value. Want 45.678 was %v", val3)
	}
	assertFreeConns(t)
}
