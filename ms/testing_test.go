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
	"time"

	"github.com/kardianos/rdb"
	"github.com/kardianos/rdb/must"
)

const parallel = false

var testConnectionString = os.Getenv("APP_DSN") // "ms://username:password@localhost?db=master&dial_timeout=3s"

var config *rdb.Config
var db must.ConnPool

func TestMain(m *testing.M) {
	if db.Valid() {
		return
	}
	if len(testConnectionString) == 0 {
		os.Exit(m.Run())
	}
	config = must.Config(rdb.ParseConfigURL(testConnectionString))
	config.PoolInitCapacity = runtime.NumCPU()
	config.DialTimeout = time.Millisecond * 100
	db = must.Open(config)
	err := db.Normal().Ping(context.Background())
	if err != nil {
		db = must.ConnPool{}
	}

	os.Exit(m.Run())
}

func checkSkip(t *testing.T) {
	if !db.Valid() {
		t.Skip("DB connection not configured, check APP_DSN")
	}
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
