// Copyright 2014 Daniel Theophanes.
// Use of this source code is governed by a zlib-style
// license that can be found in the LICENSE file.

package ms

import (
	"testing"

	"bitbucket.org/kardianos/rdb"
	"bitbucket.org/kardianos/rdb/must"
)

func TestMultiResult(t *testing.T) {
	// Handle multiple result sets.
	t.Skip()
	defer func() {
		if re := recover(); re != nil {
			if localError, is := re.(must.Error); is {
				t.Errorf("SQL Error: %v", localError)
				return
			}
			panic(re)
		}
	}()

	var myFav string
	db.Query(&rdb.Command{
		Sql: `
			select @animal as 'MyAnimal';
			-- New query.
			select 3 as 'Pants';
		`,
		Arity:         rdb.Any,
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
