// Copyright 2014 Daniel Theophanes.
// Use of this source code is governed by a zlib-style
// license that can be found in the LICENSE file.

package ms

import (
	"context"
	"fmt"
	"io"
	"net"
	"os"
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
var dbInvalidHost must.ConnPool

func TestMain(m *testing.M) {
	if db.Valid() {
		return
	}
	if len(testConnectionString) == 0 {
		os.Exit(m.Run())
	}
	config = must.Config(rdb.ParseConfigURL(testConnectionString))
	if false {
		config.PoolInitCapacity = 100 // runtime.NumCPU()
	} else {
		// Force all test on to a single connection to find connection re-use errors.
		config.PoolInitCapacity = 1
		config.PoolMaxCapacity = 1
	}
	config.DialTimeout = time.Millisecond * 100
	config.RollbackTimeout = time.Millisecond * 500
	db = must.Open(config)
	err := db.Normal().Ping(context.Background())
	if err != nil {
		fmt.Printf("DB PING error (tests will skip): %v\n", err)
		db = must.ConnPool{}
	}

	host, port, err := func() (string, int, error) {
		ln, err := net.Listen("tcp", "127.0.209.80:0")
		if err != nil {
			return "", 0, err
		}

		addr := ln.Addr().(*net.TCPAddr)
		host := addr.IP.String()
		port := addr.Port

		go func() {
			defer ln.Close()

			for {
				conn, err := ln.Accept()
				if err != nil {
					if err == io.EOF {
						return
					}
					continue
				}
				go func(c net.Conn) {
					defer c.Close()
					select {}
				}(conn)
			}
		}()
		return host, port, nil
	}()
	if err != nil {
		panic(err)
	}

	dbInvalidHost = must.Open(&rdb.Config{DriverName: "ms", Hostname: host, Port: port})

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
