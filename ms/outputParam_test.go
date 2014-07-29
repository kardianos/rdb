// Copyright 2014 Daniel Theophanes.
// Use of this source code is governed by a zlib-style
// license that can be found in the LICENSE file.

package ms

import (
	"testing"

	"bitbucket.org/kardianos/rdb"
	"bitbucket.org/kardianos/rdb/must"
)

func TestOutputParam(t *testing.T) {
	defer func() {
		if re := recover(); re != nil {
			if localError, is := re.(must.Error); is {
				t.Errorf("SQL Error: %v", localError)
				return
			}
			panic(re)
		}
	}()

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

	openConnPool()
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
