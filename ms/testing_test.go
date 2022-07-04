// Copyright 2014 Daniel Theophanes.
// Use of this source code is governed by a zlib-style
// license that can be found in the LICENSE file.

package ms

import (
	"context"
	"os"
	"runtime"
	"runtime/debug"
	"testing"

	"github.com/kardianos/rdb"
	"github.com/kardianos/rdb/must"
)

const parallel = false

var testConnectionString = os.Getenv("APP_DSN") // "ms://username:password@localhost?db=master&dial_timeout=3s"

var config *rdb.Config
var db must.ConnPool

func TestMain(m *testing.M) {
	if db.Normal() != nil {
		return
	}
	config = must.Config(rdb.ParseConfigURL(testConnectionString))
	config.PoolInitCapacity = runtime.NumCPU()
	db = must.Open(config)
	db.Ping(context.Background())

	os.Exit(m.Run())
}

func assertFreeConns(t *testing.T) {
	if parallel {
		return
	}
	capacity, available := db.Normal().PoolAvailable()
	t.Logf("Pool capacity: %v, available: %v.", capacity, available)
	if capacity != available {
		t.Errorf("Not all connections returned to pool.")
	}
}

func recoverTest(t *testing.T) {
	if re := recover(); re != nil {
		if localError, is := re.(must.Error); is {
			t.Logf("%s", debug.Stack())
			t.Errorf("SQL Error: %v", localError)
			return
		}
		panic(re)
	}
}
