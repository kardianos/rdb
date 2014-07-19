// Copyright 2014 Daniel Theophanes.
// Use of this source code is governed by a zlib-style
// license that can be found in the LICENSE file.

package ms

import (
	"reflect"
	"testing"

	"bitbucket.org/kardianos/rdb"
	"bitbucket.org/kardianos/rdb/must"
)

func TestTransaction(t *testing.T) {
	defer func() {
		if re := recover(); re != nil {
			if localError, is := re.(must.Error); is {
				t.Errorf("SQL Error: %v", localError)
				return
			}
			panic(re)
		}
	}()

	cmd := &rdb.Command{
		Sql: `
			select
				v1 = @v1
		`,
		Arity: rdb.OneMust,
	}

	openConnPool()

	params := []rdb.Param{
		{Name: "v1", Type: rdb.Text, Value: "Hello"},
	}

	tran := db.Begin()

	var v1 string

	res := tran.Query(cmd, params...)
	res.Scan(&v1)
	res.Close()

	savePointName := "PointA"

	tran.SavePoint(savePointName)

	res = tran.Query(cmd, params...)
	res.Scan(&v1)
	res.Close()

	tran.RollbackTo(savePointName)

	res = tran.Query(cmd, params...)
	res.Scan(&v1)
	res.Close()

	tran.Commit()

	tran = db.Begin()

	res = tran.Query(cmd, params...)
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

	capacity, available := db.Normal().PoolAvailable()
	t.Logf("Pool capacity: %v, available: %v.", capacity, available)
	if capacity != available {
		t.Errorf("Not all connections returned to pool.")
	}
}
