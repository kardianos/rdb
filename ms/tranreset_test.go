// Copyright 2026 Daniel Theophanes.
// Use of this source code is governed by a zlib-style
// license that can be found in the LICENSE file.

package ms

import (
	"context"
	"testing"

	"github.com/kardianos/rdb"
	"github.com/kardianos/rdb/must"
)

// TestTransactionHolds checks that a Transaction is a transaction on the
// server: @@trancount is 1 inside it and Rollback undoes its writes. A
// connection reset still pending when Begin runs must not reach the server
// after the transaction started, or the reset rolls the transaction back.
func TestTransactionHolds(t *testing.T) {
	checkSkip(t)
	for _, resetQuery := range []string{"", "set nocount on;"} {
		name := "no reset query"
		if resetQuery != "" {
			name = "reset query"
		}
		t.Run(name, func(t *testing.T) {
			config := must.Config(rdb.ParseConfigURL(testConnectionString))
			config.ResetQuery = resetQuery
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
				if _, err := q.Query(ctx, &rdb.Command{SQL: sql, Arity: rdb.Zero}); err != nil {
					t.Fatalf("%s: %v", sql, err)
				}
			}

			exec(t, pool, `if object_id('dbo.TranHolds') is not null drop table dbo.TranHolds; create table dbo.TranHolds (ID int);`)
			defer exec(t, pool, `drop table dbo.TranHolds;`)

			// The first round begins on a connection that was used and
			// released; the second on one just released by a Rollback.
			for round := range 2 {
				// Session state left by the last user must be reset away
				// before the transaction, not skipped.
				exec(t, pool, `create table #stale (ID int);`)
				tran, err := pool.Begin(ctx)
				if err != nil {
					t.Fatal(err)
				}
				if got := scalar(t, tran, `select convert(bigint, case when object_id('tempdb..#stale') is null then 0 else 1 end);`); got != 0 {
					tran.Rollback()
					t.Fatalf("round %d: #stale survived into the transaction, want it reset", round)
				}
				if got := scalar(t, tran, `select convert(bigint, @@trancount);`); got != 1 {
					tran.Rollback()
					t.Fatalf("round %d: @@trancount %d in the transaction, want 1", round, got)
				}
				exec(t, tran, `insert into dbo.TranHolds (ID) values (1);`)
				if err := tran.Rollback(); err != nil {
					t.Fatalf("round %d: rollback: %v", round, err)
				}
				if got := scalar(t, pool, `select convert(bigint, count(*)) from dbo.TranHolds;`); got != 0 {
					t.Fatalf("round %d: %d rows after rollback, want 0", round, got)
				}
			}

			// Commit keeps the write.
			tran, err := pool.Begin(ctx)
			if err != nil {
				t.Fatal(err)
			}
			exec(t, tran, `insert into dbo.TranHolds (ID) values (1);`)
			if err := tran.Commit(); err != nil {
				t.Fatalf("commit: %v", err)
			}
			if got := scalar(t, pool, `select convert(bigint, count(*)) from dbo.TranHolds;`); got != 1 {
				t.Fatalf("%d rows after commit, want 1", got)
			}
		})
	}
}

// TestTransactionHoldsFresh begins a transaction on a connection that has
// run nothing since it was opened.
func TestTransactionHoldsFresh(t *testing.T) {
	checkSkip(t)
	config := must.Config(rdb.ParseConfigURL(testConnectionString))
	config.ResetQuery = ""
	config.PoolInitCapacity = 1
	config.PoolMaxCapacity = 1
	pool, err := rdb.Open(config)
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()
	ctx := context.Background()

	tran, err := pool.Begin(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer tran.Rollback()
	res, err := tran.Query(ctx, &rdb.Command{SQL: `select convert(bigint, @@trancount);`})
	if err != nil {
		t.Fatal(err)
	}
	defer res.Close()
	var n int64
	if !res.Next() {
		t.Fatal("no row")
	}
	if err := res.Scan(&n); err != nil {
		t.Fatal(err)
	}
	if n != 1 {
		t.Fatalf("@@trancount %d in the transaction, want 1", n)
	}
}
