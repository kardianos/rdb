// Copyright 2014 Daniel Theophanes.
// Use of this source code is governed by a zlib-style
// license that can be found in the LICENSE file.

package ms

import (
	"testing"
	"time"

	"bitbucket.org/kardianos/rdb"
)

func TestTimeoutDie(t *testing.T) {
	// t.Skip()
	// Handle multiple result sets.
	defer recoverTest(t)

	res, err := db.Normal().Query(&rdb.Command{
		Sql: `
			waitfor delay '00:00:02';
			select 1 as 'ID';
		`,
		Arity:        rdb.Any,
		QueryTimeout: time.Second * 1,
	})
	defer assertFreeConns(t)
	defer res.Close()

	if err == nil {
		t.Errorf("Failed to timeout: %v", err)
	}

}

func TestTimeoutLive(t *testing.T) {
	// Handle multiple result sets.
	defer recoverTest(t)

	res, err := db.Normal().Query(&rdb.Command{
		Sql: `
			waitfor delay '00:00:01';
			select 1 as 'ID';
		`,
		Arity:        rdb.Any,
		QueryTimeout: time.Second * 2,
	})

	defer assertFreeConns(t)
	defer res.Close()

	if err != nil {
		t.Errorf("Error with query: %v", err)
	}

}
