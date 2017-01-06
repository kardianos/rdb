// Copyright 2014 Daniel Theophanes.
// Use of this source code is governed by a zlib-style
// license that can be found in the LICENSE file.

package ms

import (
	"testing"

	"bitbucket.org/kardianos/rdb"
	"bitbucket.org/kardianos/rdb/table"
)

func TestMultiResultSimple(t *testing.T) {
	if parallel {
		t.Parallel()
	}
	defer assertFreeConns(t)

	// Handle multiple result sets.
	defer recoverTest(t)

	set, err := table.FillCommand(db.Normal(), &rdb.Command{
		Sql: `
			select @animal as 'MyAnimal';
			-- New query.
			select 3 as 'Pants', cast(1 as bit) as 'Shirt';
		`,
		Arity: rdb.Any,
	}, rdb.Param{
		Name:  "animal",
		Type:  rdb.Text,
		Value: "DogIsFriend",
	})
	if err != nil {
		t.Fatalf("failed to fill set %v", err)
	}
	if len(set.Set) != 2 {
		t.Fatalf("expected 2 result sets, got %d", set.Len())
	}


	var myFav string
	res := db.Query(&rdb.Command{
		Sql: `
			select @animal as 'MyAnimal';
			-- New query.
			select 3 as 'Pants', cast(1 as bit) as 'Shirt';
		`,
		Arity: rdb.Any,
	}, []rdb.Param{
		{
			Name:  "animal",
			Type:  rdb.Text,
			Value: "DogIsFriend",
		},
	}...)
	defer res.Close()
res.Next()
	res.Prep("MyAnimal", &myFav).Scan()
	t.Logf("My Animal: %s\n", myFav)
	if res.Next() {
		t.Fatal("expected no more rows")
	}
	moreRes := res.NextResult()
	if !moreRes {
		t.Fatal("expected more result sets")
	}
res.Next()
	var pants int
	var shirt bool
	res.Prep("Pants", &pants).Prep("Shirt", &shirt).Scan()
	t.Logf("Pants: %v, Shirt: %v\n", pants, shirt)
}

func TestMultiResultHalt(t *testing.T) {
	if parallel {
		t.Parallel()
	}
	// Handle multiple result sets.
	defer recoverTest(t)

	var myFav string
	res := db.Query(&rdb.Command{
		Sql: `
			select @animal as 'MyAnimal';
			-- New query.
			select 3 as 'Pants', cast(1 as bit) as 'Shirt';
		`,
		Arity: rdb.Any,
	}, []rdb.Param{
		{
			Name:  "animal",
			Type:  rdb.Text,
			Value: "DogIsFriend",
		},
	}...)

	res.Prep("MyAnimal", &myFav).Scan()
	t.Logf("My Animal: %s\n", myFav)

	// Only fetch the first result.

	res.Close()

	assertFreeConns(t)
}

func TestMultiResultLoop(t *testing.T) {
	if parallel {
		t.Parallel()
	}
	// Handle multiple result sets.
	defer recoverTest(t)

	var myFav string
	res := db.Query(&rdb.Command{
		Sql: `
			select @animal as 'MyAnimal';
			-- New query.
			select 3 as 'Pants', cast(1 as bit) as 'Shirt';
		`,
		Arity: rdb.Any,
	}, []rdb.Param{
		{
			Name:  "animal",
			Type:  rdb.Text,
			Value: "DogIsFriend",
		},
	}...)

	defer res.Close()

	i := 0
	for {
		switch i {
		case 0:
			for res.Next() {
				res.Prep("MyAnimal", &myFav).Scan()
				t.Logf("My Animal: %s\n", myFav)
			}
		case 1:
			for res.Next() {
				var pants int
				var shirt bool
				res.Prep("Pants", &pants).Prep("Shirt", &shirt).Scan()
				t.Logf("Pants: %v, Shirt: %v\n", pants, shirt)
			}
		}
		if res.NextResult() == false {
			break
		}
		i++
	}

	// Only fetch the first result.

	assertFreeConns(t)
}

func TestMultiResultEmpty(t *testing.T) {
	if parallel {
		t.Parallel()
	}
	// Handle multiple result sets.
	defer recoverTest(t)

	res := db.Query(&rdb.Command{
		Sql: `
select
	*
from
	sys.columns
where
	1=1
	and @TestingLink=1
;

select
	*
from
	sys.tables
where
	1=1
	and @TestingLink=1
		`,
		Arity: rdb.Any,
	}, []rdb.Param{
		{
			Name:  "TestingLink",
			Type:  rdb.Bool,
			Value: false,
		},
	}...)

	defer res.Close()

	results := 1

	for {
		if res.Next() {
			t.Fatal("No next rows")
		}
		if !res.NextResult() {
			break
		}
		results++
	}
	if results != 2 {
		t.Fatal("wanted 2 sets, got ", results)
	}

	assertFreeConns(t)
}
