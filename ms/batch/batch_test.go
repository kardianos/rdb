// Copyright 2014 Daniel Theophanes.
// Use of this source code is governed by a zlib-style
// license that can be found in the LICENSE file.

package batch

import "testing"

func TestBatchSplit(t *testing.T) {
	type testItem struct {
		Sql    string
		Expect []string
	}

	list := []testItem{
		testItem{
			Sql: `use DB
go
select 1
go
select 2
`,
			Expect: []string{`use DB
`, `
select 1
`, `
select 2
`,
			},
		},
		testItem{
			Sql: `go
use DB go
`,
			Expect: []string{`
use DB go
`,
			},
		},
		testItem{
			Sql: `select 'It''s go time'
go
select top 1 1`,
			Expect: []string{`select 'It''s go time'
`, `
select top 1 1`,
			},
		},
		testItem{
			Sql: `select 1 /* go */
go
select top 1 1`,
			Expect: []string{`select 1 /* go */
`, `
select top 1 1`,
			},
		},
		testItem{
			Sql: `select 1 -- go
go
select top 1 1`,
			Expect: []string{`select 1 -- go
`, `
select top 1 1`,
			},
		},
	}

	for i := range list {
		ss := BatchSplitSql(list[i].Sql, "go")
		if len(ss) != len(list[i].Expect) {
			t.Errorf("Test Item index %d; expect %d items, got %d.", i, len(list[i].Expect), len(ss))
			continue
		}
		for j := 0; j < len(ss); j++ {
			if ss[j] != list[i].Expect[j] {
				t.Errorf("Test Item index %d, batch index %d; expect <%s>, got <%s>.", i, j, list[i].Expect[j], ss[j])
			}
		}
	}
}
