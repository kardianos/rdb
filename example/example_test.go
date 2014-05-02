// Copyright 2014 Daniel Theophanes.
// Use of this source code is governed by a zlib-style
// license that can be found in the LICENSE file.

package example

import (
	"testing"

	"bitbucket.org/kardianos/rdb"
	_ "bitbucket.org/kardianos/tds"
)

func TestSimpleQuery(t *testing.T) {
	defer func() {
		if re := recover(); re != nil {
			if localError, is := re.(rdb.MustError); is {
				t.Errorf("SQL Error: %v", localError)
				return
			}
			panic(re)
		}
	}()
	config := rdb.ParseConfigMust("tds://TESTU@localhost/SqlExpress?db=master")

	cmd := &rdb.Command{
		Sql: `
			select
				cast('fox' as varchar(7)) as dock,
				box = cast(@animal as nvarchar(max))
			;
		`,
		Arity: rdb.OneMust,
		Input: []rdb.Param{
			rdb.Param{
				N: "animal",
				T: rdb.TypeString,
			},
		},
	}

	db := rdb.OpenMust(config)
	defer db.Close()

	var dock string

	res := db.Query(cmd, rdb.Value{V: "Fish"})
	defer res.Close()

	// Prep all or some of the values.
	res.PrepAll(&dock)
	res.Scan()
	// The other values in the row are buffered until the next call to Scan().
	box := string(res.Get("box").([]byte))

	_ = box

}
