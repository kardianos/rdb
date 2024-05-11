// Copyright 2014 Daniel Theophanes.
// Use of this source code is governed by a zlib-style
// license that can be found in the LICENSE file.

package ms

import (
	"context"
	"testing"

	"github.com/kardianos/rdb"
)

func TestErrorQuery(t *testing.T) {
	if parallel {
		t.Parallel()
	}
	defer recoverTest(t)
	res, err := db.Normal().Query(context.Background(), &rdb.Command{
		SQL: `
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
	if parallel {
		t.Parallel()
	}
	defer recoverTest(t)
	var myFav string
	db.Query(context.Background(), &rdb.Command{
		SQL: `
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
	if parallel {
		t.Parallel()
	}
	defer assertFreeConns(t)
	defer recoverTest(t)
	var myFav string
	res := db.Query(context.Background(), &rdb.Command{
		SQL: `
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
		// t.Logf("got: %s\n", myFav)

		if myFav != check[i] {
			t.Errorf("Got <%s>, want <%s>", myFav, check[i])
		}
		i++
	}

	if res.RowsAffected() != 3 {
		t.Errorf("Invalid number of rows affected. Want 3 got %d.", res.RowsAffected())
	}

}
func TestRowsQueryNull(t *testing.T) {
	if parallel {
		t.Parallel()
	}
	defer recoverTest(t)
	var colA string
	cmd := &rdb.Command{
		SQL: `
			select N'Sleep well' as 'ColA'
			union all
			select NULL
		;`,
	}
	res := db.Query(context.Background(), cmd)
	defer res.Close()
	i := 0
	norm := res.Normal()
	for norm.Next() {
		colA = ""
		err := norm.Scan(&colA)
		if i == 1 && err != rdb.ErrScanNull {
			t.Errorf("Scanning a null value without a *rdb.Nullable should be an error. Got: %q", colA)
		}
		i++
	}

	assertFreeConns(t)
}
func TestLargerQuery(t *testing.T) {
	if parallel {
		t.Parallel()
	}
	defer recoverTest(t)
	cmd := &rdb.Command{
		SQL: `
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

	res := db.Query(context.Background(), cmd, []rdb.Param{
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
