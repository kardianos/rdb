// Copyright (c) 2014, The pg Authors. All Rights Reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package pg

import (
	"bitbucket.org/kardianos/rdb"
	"testing"
)

var connectionString = "pg://postgres:AgainMoreToday@localhost:5432?db=photosite"

func TestBasicQuery(t *testing.T) {
	conf := rdb.ParseConfigMust(connectionString)
	db := rdb.OpenMust(conf)

	var foo, fii int

	db.Query(&rdb.Command{
		Sql: `
			select 1 as "foo", 2 as "fii"; 
		`,
	}).Scan(&foo, &fii).Close()

	if foo != 1 && fii != 2 {
		t.Logf("foo: %d, fii: %d", foo, fii)
		t.Errorf("Failed to get correct values.")
	}
}
