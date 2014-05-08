// Copyright 2014 Daniel Theophanes.
// Use of this source code is governed by a zlib-style
// license that can be found in the LICENSE file.

package example

import (
	"testing"

	"bitbucket.org/kardianos/rdb"
	_ "bitbucket.org/kardianos/rdb/ms"
)

func TestPing(t *testing.T) {
	config := rdb.ParseConfigMust(testConnectionString)

	db, err := rdb.Open(config)
	if err != nil {
		t.Errorf("Failed to open DB: %v", err)
	}
	defer db.Close()

	err = db.Ping()
	if err != nil {
		t.Errorf("Ping error: %v", err)
	}
}

func TestVersion(t *testing.T) {
	config := rdb.ParseConfigMust(testConnectionString)

	db, err := rdb.Open(config)
	if err != nil {
		t.Errorf("Failed to open DB: %v", err)
	}
	defer db.Close()

	connInfo, err := db.ConnectionInfo()
	if err != nil {
		t.Errorf("ConnectionInfo error: %v", err)
	}
	t.Logf("Server: %v\n", connInfo.Server)
	t.Logf("Protocol: %v\n", connInfo.Protocol)
}
