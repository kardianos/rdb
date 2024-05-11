// Copyright 2014 Daniel Theophanes.
// Use of this source code is governed by a zlib-style
// license that can be found in the LICENSE file.

package ms

import (
	"context"
	"testing"
)

func TestPing(t *testing.T) {
	checkSkip(t)
	if parallel {
		t.Parallel()
	}
	err := db.Normal().Ping(context.Background())
	if err != nil {
		t.Fatalf("Ping error: %v", err)
	}
}

func TestVersion(t *testing.T) {
	checkSkip(t)
	if parallel {
		t.Parallel()
	}
	connInfo, err := db.Normal().ConnectionInfo(context.Background())
	if err != nil {
		t.Fatalf("ConnectionInfo error: %v", err)
	}
	t.Logf("Server: %v\n", connInfo.Server)
	t.Logf("Protocol: %v\n", connInfo.Protocol)
}
