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

	// TODO: Cheat for now, alter when native output param support is added.
	callProc := &rdb.Command{
		Sql: `
declare @val int
exec AddTen @p1=5, @p2 = @val output
select Val = @val
		`,
		Arity: rdb.OneMust,
	}

	openConnPool()
	db.Query(createProcDrop)
	db.Query(createProc)
	res := db.Query(callProc)
	var val int
	if res.Next() == false {
		t.Fatalf("No rows from proc.")
	}
	res.Scan(&val)

	if val != 15 {
		t.Fatalf("Incorrect value. Want 15 was %v", val)
	}
}
