// Copyright 2014 Daniel Theophanes.
// Use of this source code is governed by a zlib-style
// license that can be found in the LICENSE file.

package ms

import (
	"testing"

	"bitbucket.org/kardianos/rdb"
)

const testConnectionString = "ms://TESTU@localhost/SqlExpress?db=master&dial_timeout=3s"

var config *rdb.Config
var db rdb.ConnPoolMust

func openConnPool() {
	if db.Normal() != nil {
		return
	}
	config = rdb.ParseConfigMust(testConnectionString)
	db = rdb.OpenMust(config)
}

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
	openConnPool()

	ErrorQuery(db, t)
	SimpleQuery(db, t)
	RowsQuery(db, t)
	LargerQuery(db, t)
	return nil
}

func ErrorQuery(db rdb.ConnPoolMust, t *testing.T) {
	res, err := db.Normal().Query(&rdb.Command{
		Sql: `
			s3l3c1 @animal as 'MyAnimal';`,
		Arity:         rdb.OneMust,
		TruncLongText: true,
	}, []rdb.Param{
		{
			N: "animal",
			T: rdb.TypeString,
			L: 8,
			V: "DogIsFriend",
		},
	}...)
	if err == nil {
		t.Errorf("Expecting an error.")
	}
	if _, is := err.(rdb.SqlErrors); !is {
		t.Errorf("Expecting SqlErrors type.")
	}
	res.Close()
}

func SimpleQuery(db rdb.ConnPoolMust, t *testing.T) {
	var myFav string
	db.Query(&rdb.Command{
		Sql: `
			select @animal as 'MyAnimal';`,
		Arity:         rdb.OneMust,
		TruncLongText: true,
	}, []rdb.Param{
		{
			N: "animal",
			T: rdb.TypeString,
			L: 8,
			V: "DogIsFriend",
		},
	}...).Prep("MyAnimal", &myFav).Scan()
	t.Logf("Animal_1: %s\n", myFav)
}
func RowsQuery(db rdb.ConnPoolMust, t *testing.T) {
	var myFav string
	res := db.Query(&rdb.Command{
		Sql: `
			select @animal as 'MyAnimal'
			union all
			select N'Hello again!'
		;`,
		Arity:         rdb.Any,
		TruncLongText: true,
	}, []rdb.Param{
		{
			N: "animal",
			T: rdb.TypeString,
			V: "Dreaming boats.",
		},
	}...)
	defer res.Close()
	for {
		res.Prep("MyAnimal", &myFav)
		if !res.Scan().Next() {
			break
		}
		t.Logf("Animal_2: %s\n", myFav)
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
	}

	var dock string
	var id int
	var val float64

	res := db.Query(cmd, []rdb.Param{
		{
			N: "animal",
			T: rdb.TypeString,
			V: "Fish",
		},
	}...)
	defer res.Close()

	res.Scan(&id, &val, &dock)

	// The other values in the row are buffered until the next call to Scan().
	box := string(res.Get("box").([]byte))
	_ = box
	t.Logf("ID: %d\n", id)
	t.Logf("Val: %f\n", val)
}
