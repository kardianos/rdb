// Copyright 2014 Daniel Theophanes.
// Use of this source code is governed by a zlib-style
// license that can be found in the rdb LICENSE file.

package pg

import (
	"testing"

	"bitbucket.org/kardianos/rdb"
	"bitbucket.org/kardianos/rdb/must"
)

var connectionString = "pg2://postgres:letmein@localhost:5432?db=postgres"

func TestBasicQuery(t *testing.T) {
	conf := must.Config(rdb.ParseConfigURL(connectionString))
	db, err := rdb.Open(conf)
	if err != nil {
		t.Fatalf("Failed to open DB: %v", err)
	}

	var foo, fii int
	var fox string

	res, err := db.Query(&rdb.Command{
		Sql: `
select 1 as "foo", cast(2 as int8) as "fii", 'Hello' as "fox", cast(E'\\xDEADBEEF' as bytea) as "BN"
union all
select 5, 7, 'World', E'\\xDEAD0123'
; 
		`,
	})
	if err != nil {
		t.Fatalf("Failed to run query: %v", err)
	}

	schema := res.Schema()

	for res.Next() {
		err = res.Scan(&foo, &fii, &fox)
		if err != nil {
			t.Fatalf("Failed to run query: %v", err)
		}
	}
	res.Close()

	if foo != 5 && fii != 7 && fox != "World" {
		t.Logf("foo: %d, fii: %d, fox: %s", foo, fii, fox)
		t.Errorf("Failed to get correct values.")
	}

	if len(schema) < 3 {
		t.Fatalf("Not enough schema columns.")
	}

	if schema[0].Type != rdb.TypeInt32 {
		t.Errorf("Failed to get correct type for 'foo', got %v.", schema[0].Type)
	}
	if schema[2].Type != rdb.TypeText {
		t.Errorf("Failed to get correct type for 'fox', got %v.", schema[2].Type)
	}
}
