// Copyright 2026 Daniel Theophanes.
// Use of this source code is governed by a zlib-style
// license that can be found in the LICENSE file.

package ms

import (
	"context"
	"errors"
	"testing"

	"github.com/kardianos/rdb"
	"github.com/kardianos/rdb/must"
)

// TestServerErrorKeepsConnection checks that an error the server reports (a
// raiserror, a missing table) leaves the connection open. A Connection or
// Transaction holds session state, its temp tables, locks and transaction, so
// closing the connection under it loses that state and fails every later
// query; a pooled connection goes back to the pool for reuse.
func TestServerErrorKeepsConnection(t *testing.T) {
	checkSkip(t)
	config := must.Config(rdb.ParseConfigURL(testConnectionString))
	config.PoolInitCapacity = 1
	config.PoolMaxCapacity = 1
	pool, err := rdb.Open(config)
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()
	ctx := context.Background()

	scalar := func(t *testing.T, q rdb.Queryer, sql string) int64 {
		t.Helper()
		res, err := q.Query(ctx, &rdb.Command{SQL: sql})
		if err != nil {
			t.Fatalf("%s: %v", sql, err)
		}
		defer res.Close()
		var v int64
		if !res.Next() {
			t.Fatalf("%s: no row", sql)
		}
		if err := res.Scan(&v); err != nil {
			t.Fatalf("%s: %v", sql, err)
		}
		return v
	}
	exec := func(t *testing.T, q rdb.Queryer, sql string) {
		t.Helper()
		_, err := q.Query(ctx, &rdb.Command{SQL: sql, Arity: rdb.Zero})
		if err != nil {
			t.Fatalf("%s: %v", sql, err)
		}
	}
	// raise runs a query that fails on the server, on the query or, after a
	// result set, on Close, and checks the error is the server's.
	raise := func(t *testing.T, q rdb.Queryer, cmd *rdb.Command, nextResult bool, params ...rdb.Param) {
		t.Helper()
		res, err := q.Query(ctx, cmd, params...)
		if err == nil {
			// Next only reports a pending row; Scan reads it. With nextResult
			// the query moves on to its second result set and closes with its
			// rows still to read.
			for res.Next() {
				if err = res.Scan(); err != nil {
					break
				}
			}
			if err == nil && nextResult {
				_, err = res.NextResult()
			}
			if cerr := res.Close(); err == nil {
				err = cerr
			}
		}
		if err == nil {
			t.Fatalf("%s: no error", cmd.SQL)
		}
		var list rdb.Errors
		if !errors.As(err, &list) {
			t.Fatalf("%s: error %v (%T), want the server's rdb.Errors", cmd.SQL, err, err)
		}
	}

	// sessionSQL names the physical connection: SQL Server hands a freed
	// session ID to the next connection, so @@spid alone does not tell a
	// reopened connection from the same one.
	const sessionSQL = `
select
	convert(bigint, convert(bigint, @@spid) * convert(bigint, 1000000000000) + datediff_big(millisecond, '2020-01-01', c.connect_time) % convert(bigint, 1000000000000))
from
	sys.dm_exec_connections c
where
	c.session_id = @@spid
;`

	list := []struct {
		Name       string
		Bad        rdb.Command
		NextResult bool // Move to the second result set, then close.
	}{
		{Name: "raiserror, zero arity", Bad: rdb.Command{SQL: `raiserror(N'boom', 16, 1);`, Arity: rdb.Zero}},
		{Name: "raiserror, rows", Bad: rdb.Command{SQL: `raiserror(N'boom', 16, 1);`}},
		{Name: "missing table", Bad: rdb.Command{SQL: `select * from dbo.NoSuchTable;`}},
		{Name: "error after a statement", Bad: rdb.Command{SQL: `declare @x int = 1; raiserror(N'boom', 16, 1);`}},
		{Name: "error with parameters", Bad: rdb.Command{SQL: `raiserror(@Message, 16, 1);`}},
		{Name: "error after a result set", Bad: rdb.Command{SQL: `select v = 1; raiserror(N'boom', 16, 1);`}},
		{Name: "error between result sets", Bad: rdb.Command{SQL: `select v = 1; raiserror(N'boom', 16, 1); select v = 2 union all select 3;`}, NextResult: true},
	}
	params := func(cmd rdb.Command) []rdb.Param {
		if cmd.SQL == `raiserror(@Message, 16, 1);` {
			return []rdb.Param{{Name: "Message", Type: rdb.Text, Value: "boom"}}
		}
		return nil
	}
	for _, tc := range list {
		t.Run(tc.Name+"/connection", func(t *testing.T) {
			conn, err := pool.Connection(ctx)
			if err != nil {
				t.Fatal(err)
			}
			defer conn.Close()
			spid := scalar(t, conn, sessionSQL)
			exec(t, conn, `create table #keep (ID int); insert into #keep (ID) values (1);`)

			bad := tc.Bad
			raise(t, conn, &bad, tc.NextResult, params(tc.Bad)...)

			if got := scalar(t, conn, sessionSQL); got != spid {
				t.Fatalf("session %d after the error, want %d", got, spid)
			}
			if got := scalar(t, conn, `select convert(bigint, count(*)) from #keep;`); got != 1 {
				t.Fatalf("temp table rows %d, want 1", got)
			}
		})
		t.Run(tc.Name+"/transaction", func(t *testing.T) {
			tran, err := pool.Begin(ctx)
			if err != nil {
				t.Fatal(err)
			}
			// A failed check must still give the pool its one connection back.
			defer func() {
				if tran.Active() {
					tran.Rollback()
				}
			}()
			exec(t, tran, `create table #tran (ID int); insert into #tran (ID) values (1);`)

			bad := tc.Bad
			raise(t, tran, &bad, tc.NextResult, params(tc.Bad)...)

			if got := scalar(t, tran, `select convert(bigint, count(*)) from #tran;`); got != 1 {
				t.Fatalf("temp table rows %d, want 1", got)
			}
			if err := tran.Rollback(); err != nil {
				t.Fatalf("rollback: %v", err)
			}
		})
		t.Run(tc.Name+"/pool", func(t *testing.T) {
			// One connection in the pool: the next query gets the same session
			// only if the error did not close it.
			spid := scalar(t, pool, sessionSQL)
			bad := tc.Bad
			raise(t, pool, &bad, tc.NextResult, params(tc.Bad)...)
			if got := scalar(t, pool, sessionSQL); got != spid {
				t.Fatalf("session %d after the error, want the same pooled session %d", got, spid)
			}
			// The reused connection must not report the old error again.
			tran, err := pool.Begin(ctx)
			if err != nil {
				t.Fatalf("begin after the error: %v", err)
			}
			if err := tran.Commit(); err != nil {
				t.Fatalf("commit after the error: %v", err)
			}
		})
	}
}
