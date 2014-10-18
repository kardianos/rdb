// Copyright 2014 Daniel Theophanes.
// Use of this source code is governed by a zlib-style
// license that can be found in the LICENSE file.

package ms

import (
	"testing"

	"bitbucket.org/kardianos/rdb"
)

func TestErrorQuery(t *testing.T) {
	defer recoverTest(t)
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

	assertFreeConns(t)
}

func TestSimpleQuery(t *testing.T) {
	defer recoverTest(t)
	var myFav string
	db.Query(&rdb.Command{
		Sql: `
			select @animal as 'MyAnimal';
		`,
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

	assertFreeConns(t)
}

func TestRowsQuerySimple(t *testing.T) {
	defer assertFreeConns(t)
	defer recoverTest(t)
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
	},
		rdb.Param{
			Name:  "animal",
			Type:  rdb.Text,
			Value: "Dreaming boats.",
		},
	)
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

	if res.RowsAffected() != 3 {
		t.Errorf("Invalid number of rows affected. Want 3 got %d.", res.RowsAffected())
	}

}
func TestRowsQueryNull(t *testing.T) {
	defer recoverTest(t)
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

	assertFreeConns(t)
}
func TestLargerQuery(t *testing.T) {
	defer recoverTest(t)
	cmd := &rdb.Command{
		Sql: `
			select
				432 as ID,
				987.654 as Val,
				cast('fox' as varchar(7)) as dock,
				box = cast(@animal as nvarchar(max))
			order by ID
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

	assertFreeConns(t)
}
