// Copyright 2014 Daniel Theophanes.
// Use of this source code is governed by a zlib-style
// license that can be found in the LICENSE file.

package ms

import (
	"context"
	"fmt"
	"testing"

	"github.com/kardianos/rdb"
	"github.com/kardianos/rdb/table"
)

func TestMultiResultSimple(t *testing.T) {
	checkSkip(t)
	if parallel {
		t.Parallel()
	}
	defer assertFreeConns(t)

	// Handle multiple result sets.
	defer recoverTest(t)

	set, err := table.FillCommand(context.Background(), db.Normal(), &rdb.Command{
		SQL: `
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
		t.Fatalf("expected 2 result sets, got %d", len(set.Set))
	}

	var myFav string
	res := db.Query(context.Background(), &rdb.Command{
		SQL: `
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
	checkSkip(t)
	if parallel {
		t.Parallel()
	}
	// Handle multiple result sets.
	defer recoverTest(t)

	var myFav string
	res := db.Query(context.Background(), &rdb.Command{
		SQL: `
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
	checkSkip(t)
	if parallel {
		t.Parallel()
	}
	// Handle multiple result sets.
	defer recoverTest(t)

	var myFav string
	res := db.Query(context.Background(), &rdb.Command{
		SQL: `
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

func TestMultiResultEmpty1(t *testing.T) {
	checkSkip(t)
	if parallel {
		t.Parallel()
	}

	defer assertFreeConns(t)

	// Handle multiple result sets.
	defer recoverTest(t)

	res := db.Query(context.Background(), &rdb.Command{
		SQL: `
declare @T table(ID int);
insert into @T
select 1;

select
	1 as set1, *
from
	sys.columns
where
	1=1
	and @TestingLink=1
;

select
	2 as set2, *
from
	sys.tables
where
	1=1
	and @TestingLink=1
;

select
	3 as set3, *
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
		if len(res.Schema()) == 0 {
			t.Fatal("column schema not populating in empty result set")
		}
		gotColName := res.Schema()[0].Name
		expectColName := fmt.Sprintf("set%d", results)
		if gotColName != expectColName {
			t.Fatalf("expected first column to be %q, got %q", expectColName, gotColName)
		}
		t.Logf("column count %d, first column %q", len(res.Schema()), res.Schema()[0].Name)
		if !res.NextResult() {
			break
		}
		results++
	}
	if results != 3 {
		t.Fatal("wanted 3 sets, got ", results)
	}

}

func TestMultiResultEmpty2(t *testing.T) {
	checkSkip(t)
	if parallel {
		t.Parallel()
	}

	defer assertFreeConns(t)

	// Handle multiple result sets.
	defer recoverTest(t)

	res := db.Query(context.Background(), &rdb.Command{
		SQL: `
	declare @SampleSelect table (ID bigint)
	declare @Sample table (ID bigint);
	declare @Locus table (ID bigint);

	insert into @Sample
	select ID
	from
		(select 1 as ID) a
	where
		1=0
	;

	select
		1 as set1
	from
		sys.tables
	where
		1=0

	select
		2 as set2
	from
		sys.tables
	where
		1=0
	order by
		name asc
	;

	select
		3 as set3
	from
		sys.tables
	where
		1=0
	order by
		name asc
	;
		`,
		Arity: rdb.Any,
	})

	defer res.Close()

	results := 1

	for {
		if res.Next() {
			t.Fatal("No next rows")
		}
		if len(res.Schema()) == 0 {
			t.Fatal("column schema not populating in empty result set")
		}
		gotColName := res.Schema()[0].Name
		expectColName := fmt.Sprintf("set%d", results)
		if gotColName != expectColName {
			t.Fatalf("expected first column to be %q, got %q", expectColName, gotColName)
		}
		t.Logf("column count %d, first column %q", len(res.Schema()), res.Schema()[0].Name)
		if !res.NextResult() {
			t.Logf("done with %d results", results)
			break
		}
		results++
	}
	if results != 3 {
		t.Fatal("wanted 3 sets, got ", results)
	}
}

func TestMultiResultEmpty3(t *testing.T) {
	checkSkip(t)
	if parallel {
		t.Parallel()
	}

	defer assertFreeConns(t)

	// Handle multiple result sets.
	defer recoverTest(t)

	res := db.Query(context.Background(), &rdb.Command{
		SQL: `
	declare @SampleSelect table (ID bigint)
	declare @Sample table (ID bigint);
	declare @Locus table (ID bigint);

	insert into @Sample
	select ID
	from
		(select 1 as ID) a
	where
		1=0
	;

	select
		1 as set1
	from
		sys.tables
	where
		1=0

	select
		2 as set2
	from
		sys.tables
	where
		1=0
	;

	select
		3 as set3
	from
		sys.tables
	where
		1=0
	;
		`,
		Arity: rdb.Any,
	})

	defer res.Close()

	results := 1

	for {
		if res.Next() {
			t.Fatal("No next rows")
		}
		if len(res.Schema()) == 0 {
			t.Fatal("column schema not populating in empty result set")
		}
		gotColName := res.Schema()[0].Name
		expectColName := fmt.Sprintf("set%d", results)
		if gotColName != expectColName {
			t.Fatalf("expected first column to be %q, got %q", expectColName, gotColName)
		}
		t.Logf("column count %d, first column %q", len(res.Schema()), res.Schema()[0].Name)
		if !res.NextResult() {
			t.Logf("done with %d results", results)
			break
		}
		results++
	}
	if results != 3 {
		t.Fatal("wanted 3 sets, got ", results)
	}
}
func TestMultiResultNotEmpty1(t *testing.T) {
	checkSkip(t)
	if parallel {
		t.Parallel()
	}

	defer assertFreeConns(t)

	// Handle multiple result sets.
	defer recoverTest(t)

	cmd := &rdb.Command{
		SQL:   `select name from sys.tables order by name asc;`,
		Arity: rdb.Any,
	}

	tb, err := table.FillCommand(context.Background(), db.Normal(), cmd)
	if err != nil {
		t.Fatal(err)
	}
	if tb.Len() == 0 {
		t.Fatalf("got %d rows, expected at last one row", tb.Len())
	}
}

func TestMultiResultAnotherTest(t *testing.T) {
	checkSkip(t)
	if parallel {
		t.Parallel()
	}

	defer assertFreeConns(t)

	// Handle multiple result sets.
	defer recoverTest(t)

	cmd := &rdb.Command{
		SQL: `

	select
		1 as set1
	from
		sys.tables
	where
		1=0

	select
		2 as set2
	from
		sys.tables
	where
		1=0
	;

	select top 1
		3 as set3
	from
		sys.tables
	where
		1=1

		`,
		Arity: rdb.Any,
	}

	tb, err := table.FillCommand(context.Background(), db.Normal(), cmd)
	if err != nil {
		t.Fatal(err)
	}

	set := tb.Set
	list := []int{0, 0, 1}
	if len(set) != len(list) {
		t.Fatalf("expected %d result sets, got %d", len(list), len(set))
	}
	for index, tb := range set {
		if len(tb.Row) != list[index] {
			t.Errorf("in result set index %d, wanted %d rows, got %d", index, list[index], len(tb.Row))
		}
	}

}
