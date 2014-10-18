// Copyright 2014 Daniel Theophanes.
// Use of this source code is governed by a zlib-style
// license that can be found in the LICENSE file.

package ms

import (
	"fmt"
	"runtime"
	"testing"

	"bitbucket.org/kardianos/rdb"
	"bitbucket.org/kardianos/rdb/must"
)

const parallel = true

var testConnectionString = "ms://TESTU:letmein@localhost/SqlExpress?db=master&dial_timeout=3s&init_cap=%d"

var config *rdb.Config
var db must.ConnPool

func init() {
	if db.Normal() != nil {
		return
	}
	testConnectionString = fmt.Sprintf(testConnectionString, runtime.NumCPU())
	config = must.Config(rdb.ParseConfigURL(testConnectionString))
	db = must.Open(config)
}

func assertFreeConns(t *testing.T) {
	capacity, available := db.Normal().PoolAvailable()
	t.Logf("Pool capacity: %v, available: %v.", capacity, available)
	if capacity != available {
		t.Errorf("Not all connections returned to pool.")
	}
}

func recoverTest(t *testing.T) {
	if re := recover(); re != nil {
		if localError, is := re.(must.Error); is {
			t.Errorf("SQL Error: %v", localError)
			return
		}
		panic(re)
	}
}
