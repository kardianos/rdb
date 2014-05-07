// Copyright 2014 Daniel Theophanes.
// Use of this source code is governed by a zlib-style
// license that can be found in the LICENSE file.

package example

import (
	"testing"

	"bitbucket.org/kardianos/rdb"
	_ "bitbucket.org/kardianos/rdb/ms"
)

const testConnectionString = "ms://TESTU@localhost/SqlExpress?db=master"

func TestSimpleQuery(t *testing.T) {
	err := QueryTest(t)
	if err != nil {
		t.Error(err)
	}
}

func QueryTest(t *testing.T) (ferr error) {
	defer func() {
		if re := recover(); re != nil {
			if localError, is := re.(rdb.MustError); is {
				ferr = localError
				return
			}
			panic(re)
		}
	}()
	config := rdb.ParseConfigMust(testConnectionString)

	db := rdb.OpenMust(config)
	defer db.Close()

	SimpleQuery(db, t)
	RowsQuery(db, t)
	LargerQuery(db, t)
	return nil
}

func SimpleQuery(db rdb.ConnPoolMust, t *testing.T) {
	var myFav string
	db.Query(&rdb.Command{
		Sql: `
			select @animal as 'MyAnimal';`,
		Arity: rdb.OneMust,
		Input: []rdb.Param{
			rdb.Param{
				N: "animal",
				T: rdb.TypeString,
				L: 8,
				V: "DogIsFriend",
			},
		},
		TruncLongText: true,
	}).Prep("MyAnimal", &myFav).Scan()
	t.Logf("Animal: %s\n", myFav)
}
func RowsQuery(db rdb.ConnPoolMust, t *testing.T) {
	var myFav string
	res := db.Query(&rdb.Command{
		Sql: `
			select @animal as 'MyAnimal'
			union all
			select N'Hello again!'
		;`,
		Arity: rdb.Any,
		Input: []rdb.Param{
			rdb.Param{
				N: "animal",
				T: rdb.TypeString,
				V: "Dreaming boats.",
			},
		},
		TruncLongText: true,
	})
	defer res.Close()
	for {
		res.Prep("MyAnimal", &myFav)
		if !res.Scan() {
			break
		}
		t.Logf("Animal: %s\n", myFav)
	}
}
func LargerQuery(db rdb.ConnPoolMust, t *testing.T) {
	cmd := &rdb.Command{
		Sql: `
			select
				432 as ID,
				987.654 as Val,
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

	var dock string
	var id int
	var val float64

	res := db.Query(cmd, rdb.Value{V: "Fish"})
	defer res.Close()

	res.PrepAll(&id, &val, &dock)

	res.Scan()

	// The other values in the row are buffered until the next call to Scan().
	box := string(res.Get("box").([]byte))
	_ = box
	t.Logf("ID: %d\n", id)
	t.Logf("Val: %f\n", val)
}
