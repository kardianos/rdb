// Copyright 2014 Daniel Theophanes.
// Use of this source code is governed by a zlib-style
// license that can be found in the LICENSE file.

package ms

import (
	"testing"

	"bitbucket.org/kardianos/rdb"
	"bitbucket.org/kardianos/rdb/must"
)

const testConnectionString = "ms://TESTU@localhost/SqlExpress?db=master&dial_timeout=3s"

var config *rdb.Config
var db must.ConnPool

func openConnPool() {
	if db.Normal() != nil {
		return
	}
	config = must.Config(rdb.ParseConfigURL(testConnectionString))
	db = must.Open(config)
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
			if localError, is := re.(must.Error); is {
				ferr = localError
				return
			}
			panic(re)
		}
	}()
	openConnPool()

	ErrorQuery(db, t)
	SimpleQuery(db, t)
	RowsQuerySimple(db, t)
	RowsQueryNull(db, t)
	LargerQuery(db, t)

	capacity, available := db.Normal().PoolAvailable()
	t.Logf("Pool capacity: %v, available: %v.", capacity, available)
	if capacity != available {
		t.Errorf("Not all connections returned to pool.")
	}
	return nil
}

func ErrorQuery(db must.ConnPool, t *testing.T) {
	res, err := db.Normal().Query(&rdb.Command{
		Sql: `
			s3l3c1 @animal as 'MyAnimal';`,
		Arity:         rdb.OneMust,
		TruncLongText: true,
	}, []rdb.Param{
		{
			Name:   "animal",
			Type:   rdb.Text,
			Length: 8,
			Value:  "DogIsFriend",
		},
	}...)
	if err == nil {
		t.Errorf("Expecting an error.")
	}
	if _, is := err.(rdb.Errors); !is {
		t.Errorf("Expecting SqlErrors type.")
	}
	res.Close()
}

func SimpleQuery(db must.ConnPool, t *testing.T) {
	var myFav string
	db.Query(&rdb.Command{
		Sql: `
			select @animal as 'MyAnimal';`,
		Arity:         rdb.OneMust,
		TruncLongText: true,
	}, []rdb.Param{
		{
			Name:   "animal",
			Type:   rdb.Text,
			Length: 8,
			Value:  "DogIsFriend",
		},
	}...).Prep("MyAnimal", &myFav).Scan()
	t.Logf("Animal_1: %s\n", myFav)
}
func RowsQuerySimple(db must.ConnPool, t *testing.T) {
	var myFav string
	res := db.Query(&rdb.Command{
		Sql: `
			select @animal as 'MyAnimal'
			union all
			select N'Hello again!'
			union all
			select NULL
		;`,
		Arity: rdb.Any,
		Fields: []rdb.Field{
			{Null: "null-value"},
		},
		TruncLongText: true,
	}, []rdb.Param{
		{
			Name:  "animal",
			Type:  rdb.Text,
			Value: "Dreaming boats.",
		},
	}...)
	check := []string{
		"Dreaming boats.",
		"Hello again!",
		"null-value",
	}
	defer res.Close()
	i := 0
	for res.Next() {
		res.Scan(&myFav)
		if myFav != check[i] {
			t.Errorf("Got <%s>, want <%s>", myFav, check[i])
		}
		i++
		t.Logf("Animal_2: %s\n", myFav)
	}
}
func RowsQueryNull(db must.ConnPool, t *testing.T) {
	var colA string
	cmd := &rdb.Command{
		Sql: `
			select N'Sleep well' as 'ColA'
			union all
			select NULL
		;`,
	}
	res := db.Query(cmd)
	defer res.Close()
	i := 0
	norm := res.Normal()
	for norm.Next() {
		err := norm.Scan(&colA)
		if i == 1 && err != rdb.ScanNullError {
			t.Error("Scanning a null value without a *rdb.Nullable should be an error.")
		}
		i++
	}
}
func LargerQuery(db must.ConnPool, t *testing.T) {
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
			Name:  "animal",
			Type:  rdb.Text,
			Value: "Fish",
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
