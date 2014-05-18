// Copyright 2014 Daniel Theophanes.
// Use of this source code is governed by a zlib-style
// license that can be found in the LICENSE file.

package ms

import (
	"testing"

	"bitbucket.org/kardianos/rdb"
)

func TestPing(t *testing.T) {
	config := rdb.ParseConfigMust(testConnectionString)

	db, err := rdb.Open(config)
	if err != nil {
		t.Fatalf("Failed to open DB: %v", err)
	}
	defer db.Close()

	err = db.Ping()
	if err != nil {
		t.Fatalf("Ping error: %v", err)
	}
}

func TestVersion(t *testing.T) {
	config := rdb.ParseConfigMust(testConnectionString)

	db, err := rdb.Open(config)
	if err != nil {
		t.Fatalf("Failed to open DB: %v", err)
	}
	defer db.Close()

	connInfo, err := db.ConnectionInfo()
	if err != nil {
		t.Fatalf("ConnectionInfo error: %v", err)
	}
	t.Logf("Server: %v\n", connInfo.Server)
	t.Logf("Protocol: %v\n", connInfo.Protocol)
}
