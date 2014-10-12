// Copyright 2014 Daniel Theophanes.
// Use of this source code is governed by a zlib-style
// license that can be found in the LICENSE file.

package ms

import (
	"testing"

	"bitbucket.org/kardianos/rdb"
)

func TestMultiResult(t *testing.T) {
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
	res.NextResult()
	var pants int
	var shirt bool
	res.Prep("Pants", &pants).Prep("Shirt", &shirt).Scan()
	t.Logf("Pants: %v, Shirt: %v\n", pants, shirt)

	res.Close()

	assertFreeConns(t)
}

func TestMultiResultHalt(t *testing.T) {
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
