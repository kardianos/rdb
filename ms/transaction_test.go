// Copyright 2014 Daniel Theophanes.
// Use of this source code is governed by a zlib-style
// license that can be found in the LICENSE file.

package ms

import (
	"context"
	"reflect"
	"testing"

	"github.com/kardianos/rdb"
)

func TestTransaction(t *testing.T) {
	checkSkip(t)
	if parallel {
		t.Parallel()
	}
	defer recoverTest(t)

	ctx := context.Background()
	cmd := &rdb.Command{
		SQL: `
			select
				v1 = @v1
		`,
		Arity: rdb.OneMust,
	}

	params := []rdb.Param{
		{Name: "v1", Type: rdb.Text, Value: "Hello"},
	}
	tran := db.Begin(ctx)

	var v1 string

	res := tran.Query(ctx, cmd, params...)
	res.Scan(&v1)
	res.Close()

	savePointName := "PointA"

	tran.SavePoint(savePointName)

	res = tran.Query(ctx, cmd, params...)
	res.Scan(&v1)
	res.Close()

	tran.RollbackTo(savePointName)

	res = tran.Query(ctx, cmd, params...)
	res.Scan(&v1)
	res.Close()

	tran.Commit()

	tran = db.Begin(ctx)

	res = tran.Query(ctx, cmd, params...)
	res.Scan(&v1)
	res.Close()

	tran.Rollback()

	compare := []interface{}{v1}

	for i := range compare {
		in := params[i]
		if !reflect.DeepEqual(compare[i], in.Value) {
			t.Errorf("Param %s did not round trip: Want (%v) got (%v)", in.Name, in.Value, compare[i])
		}
	}

	assertFreeConns(t)
}

func TestTransactionIsolationLevel(t *testing.T) {
	checkSkip(t)
	defer recoverTest(t)

	const SQL = `
select
	IsoLevel = case es.transaction_isolation_level
		when 0 then 'Unspecified' 
		when 1 then 'ReadUncommitted' 
		when 2 then 'ReadCommitted' 
		when 3 then 'RepeatableRead' 
		when 4 then 'Serializable' 
		when 5 then 'Snapshot'
	end
FROM
	sys.dm_exec_sessions es
where 1=1
	and es.session_id = @@SPID
;
`
	ctx := context.Background()

	list := []rdb.IsolationLevel{
		rdb.LevelReadUncommitted,
		rdb.LevelReadCommitted,
		rdb.LevelRepeatableRead,
		rdb.LevelSerializable,
		rdb.LevelSnapshot,
	}

	for _, item := range list {
		t.Run(item.String(), func(t *testing.T) {
			tran := db.BeginLevel(ctx, item)
			defer tran.Rollback()

			var levelName string
			tran.Query(ctx, &rdb.Command{
				Arity: rdb.OneMust,
				SQL:   SQL,
			}).Prep("IsoLevel", &levelName).Scan()

			if g, w := levelName, item.String(); g != w {
				t.Fatalf("got %s, want %s", g, w)
			}
		})
	}
	assertFreeConns(t)
}
