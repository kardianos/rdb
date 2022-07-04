// Copyright 2015 Daniel Theophanes.
// Use of this source code is governed by a zlib-style
// license that can be found in the LICENSE file.

package ms

import (
	"context"
	"testing"

	"github.com/kardianos/rdb"
	"github.com/kardianos/rdb/must"
)

func TestReset(t *testing.T) {
	config := must.Config(rdb.ParseConfigURL(testConnectionString))
	config.PoolInitCapacity = 1
	config.PoolMaxCapacity = 1
	config.ResetQuery = `set XACT_ABORT on;`

	db := must.Open(config)
	defer db.Close()

	cmd := &rdb.Command{Sql: `select 16384 & @@OPTIONS;`}
	for i := range [100]struct{}{} {
		res := db.Query(context.Background(), cmd)
		v := 0
		res.Scan(&v)
		res.Close()
		if v == 0 {
			t.Fail()
			t.Logf("Run %d: Should always be 1, but value is 0", i+1)
		}
	}
}
