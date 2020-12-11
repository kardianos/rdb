// Copyright 2014 Daniel Theophanes.
// Use of this source code is governed by a zlib-style
// license that can be found in the LICENSE file.

package batch

import (
	"fmt"
	"strings"
	"testing"
)

func TestBatchSplit(t *testing.T) {
	type testItem struct {
		Sql    string
		Expect []string
	}

	list := []testItem{
		{
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
		{
			Sql: `go
use DB go
`,
			Expect: []string{`
use DB go
`,
			},
		},
		{
			Sql: `select 'It''s go time'
go
select top 1 1`,
			Expect: []string{`select 'It''s go time'
`, `
select top 1 1`,
			},
		},
		{
			Sql: `select 1 /* go */
go
select top 1 1`,
			Expect: []string{`select 1 /* go */
`, `
select top 1 1`,
			},
		},
		{
			Sql: `select 1 -- go
go
select top 1 1`,
			Expect: []string{`select 1 -- go
`, `
select top 1 1`,
			},
		},
		{
			Sql: `select 1;
go
select 2;
Go
select 3;
gO
select 4;
GO
select 5;`,
			Expect: []string{
				`select 1;`,
				`select 2;`,
				`select 3;`,
				`select 4;`,
				`select 5;`,
			},
		},
		{
			Sql: `
create table X (
	Google bigint
);
			`,
			Expect: []string{
				`
create table X (
	Google bigint
);
				`,
			},
		},
	}

	for i := range list {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			ss := BatchSplitSql(list[i].Sql, "go")
			if len(ss) != len(list[i].Expect) {
				t.Fatalf("Test Item index %d; expect %d items, got %d.", i, len(list[i].Expect), len(ss))
			}
			for j := 0; j < len(ss); j++ {
				if strings.TrimSpace(ss[j]) != strings.TrimSpace(list[i].Expect[j]) {
					t.Errorf("Test Item index %d, batch index %d; expect <%s>, got <%s>.", i, j, list[i].Expect[j], ss[j])
				}
			}
		})
	}
}
