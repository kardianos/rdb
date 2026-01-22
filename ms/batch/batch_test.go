// Copyright 2014 Daniel Theophanes.
// Use of this source code is governed by a zlib-style
// license that can be found in the LICENSE file.

package batch

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/kardianos/rdb"
	"github.com/kardianos/rdb/must"

	_ "github.com/kardianos/rdb/ms" // register the ms driver
)

var testConnectionString = os.Getenv("APP_DSN")
var db must.ConnPool

func TestMain(m *testing.M) {
	if len(testConnectionString) == 0 {
		os.Exit(m.Run())
	}
	config := must.Config(rdb.ParseConfigURL(testConnectionString))
	config.PoolInitCapacity = 1
	config.PoolMaxCapacity = 1
	config.DialTimeout = time.Millisecond * 100
	db = must.Open(config)
	err := db.Normal().Ping(context.Background())
	if err != nil {
		fmt.Printf("DB PING error (tests will skip): %v\n", err)
		db = must.ConnPool{}
	}
	os.Exit(m.Run())
}

func checkSkip(t *testing.T) {
	if !db.Valid() {
		t.Skip("DB connection not configured, check APP_DSN")
	}
}

func TestBatchSplit(t *testing.T) {
	type testItem struct {
		SQL    string
		Expect []string
	}

	list := []testItem{
		{
			SQL: `use DB
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
			SQL: `go
use DB go
`,
			Expect: []string{`
use DB go
`,
			},
		},
		{
			SQL: `select 'It''s go time'
go
select top 1 1`,
			Expect: []string{`select 'It''s go time'
`, `
select top 1 1`,
			},
		},
		{
			SQL: `select 1 /* go */
go
select top 1 1`,
			Expect: []string{`select 1 /* go */
`, `
select top 1 1`,
			},
		},
		{
			SQL: `select 1 -- go
go
select top 1 1`,
			Expect: []string{`select 1 -- go
`, `
select top 1 1`,
			},
		},
		{
			SQL: `select 1;
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
			SQL: `
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
			ss := SplitSQL(list[i].SQL, "go")
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

// TestSplitCmd tests the SplitCmd function
func TestSplitCmd(t *testing.T) {
	cmd := &rdb.Command{
		SQL:   "select 1\ngo\nselect 2",
		Arity: rdb.OneMust,
		Name:  "test_cmd",
	}

	cmds := SplitCmd(cmd, "go")
	if len(cmds) != 2 {
		t.Fatalf("expected 2 commands, got %d", len(cmds))
	}

	// Verify the SQL is split correctly
	if strings.TrimSpace(cmds[0].SQL) != "select 1" {
		t.Errorf("expected 'select 1', got %q", cmds[0].SQL)
	}
	if strings.TrimSpace(cmds[1].SQL) != "select 2" {
		t.Errorf("expected 'select 2', got %q", cmds[1].SQL)
	}

	// Verify other fields are preserved
	if cmds[0].Arity != rdb.OneMust {
		t.Errorf("expected Arity to be preserved")
	}
	if cmds[0].Name != "test_cmd" {
		t.Errorf("expected Name to be preserved")
	}
}

// TestExecuteSQL tests executing batched SQL statements
func TestExecuteSQL(t *testing.T) {
	checkSkip(t)

	ctx := context.Background()

	// Test simple batch execution
	batchSQL := `
declare @x int = 1
go
declare @y int = 2
go
select 1
`
	err := ExecuteSQL(ctx, db.Normal(), batchSQL, "go")
	if err != nil {
		t.Fatalf("ExecuteSQL failed: %v", err)
	}
}

// TestExecuteSQLWithError tests ExecuteSQL error handling
func TestExecuteSQLWithError(t *testing.T) {
	checkSkip(t)

	ctx := context.Background()

	// Test batch with SQL error - this should fail on the second statement
	batchSQL := `
select 1
go
select * from nonexistent_table_12345
go
select 2
`
	err := ExecuteSQL(ctx, db.Normal(), batchSQL, "go")
	if err == nil {
		t.Fatal("expected error for invalid table, got nil")
	}
	t.Logf("Got expected error: %v", err)

	// The error should contain context information
	errStr := err.Error()
	if !strings.Contains(errStr, "nonexistent_table") {
		t.Errorf("error should contain table name, got: %v", err)
	}
}

// TestSQLErrorWithContext tests the SQLErrorWithContext function
func TestSQLErrorWithContext(t *testing.T) {
	sql := `line 1
line 2
line 3 with error
line 4
line 5`

	// Create a mock rdb.Errors
	msgs := rdb.Errors{
		&rdb.Message{
			Type:       rdb.SqlError,
			Number:     208,
			Message:    "Invalid object name 'foo'",
			LineNumber: 3,
			ServerName: "testserver",
		},
	}

	err := SQLErrorWithContext(sql, msgs, 1)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	errStr := err.Error()
	t.Logf("Error with context:\n%s", errStr)

	// Should contain the error message
	if !strings.Contains(errStr, "Invalid object name") {
		t.Error("error should contain original message")
	}

	// Should contain context lines
	if !strings.Contains(errStr, "line 3 with error") {
		t.Error("error should contain the error line")
	}

	// Should have the arrow indicator
	if !strings.Contains(errStr, "-->") {
		t.Error("error should have --> indicator for error line")
	}
}

// TestSQLErrorWithContextEdgeCases tests edge cases for SQLErrorWithContext
func TestSQLErrorWithContextEdgeCases(t *testing.T) {
	t.Run("negative_context_lines", func(t *testing.T) {
		sql := "line 1\nline 2\nline 3"
		msgs := rdb.Errors{
			&rdb.Message{LineNumber: 2, Message: "error"},
		}
		err := SQLErrorWithContext(sql, msgs, -5) // negative should be treated as 0
		if err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("error_at_start", func(t *testing.T) {
		sql := "line 1\nline 2\nline 3"
		msgs := rdb.Errors{
			&rdb.Message{LineNumber: 1, Message: "error at start"},
		}
		err := SQLErrorWithContext(sql, msgs, 2)
		if err == nil {
			t.Fatal("expected error")
		}
		if !strings.Contains(err.Error(), "line 1") {
			t.Error("should contain first line")
		}
	})

	t.Run("error_at_end", func(t *testing.T) {
		sql := "line 1\nline 2\nline 3"
		msgs := rdb.Errors{
			&rdb.Message{LineNumber: 3, Message: "error at end"},
		}
		err := SQLErrorWithContext(sql, msgs, 2)
		if err == nil {
			t.Fatal("expected error")
		}
		if !strings.Contains(err.Error(), "line 3") {
			t.Error("should contain last line")
		}
	})

	t.Run("multiple_errors", func(t *testing.T) {
		sql := "line 1\nline 2\nline 3\nline 4\nline 5"
		msgs := rdb.Errors{
			&rdb.Message{LineNumber: 2, Message: "first error"},
			&rdb.Message{LineNumber: 4, Message: "second error"},
		}
		err := SQLErrorWithContext(sql, msgs, 1)
		if err == nil {
			t.Fatal("expected error")
		}
		errStr := err.Error()
		if !strings.Contains(errStr, "first error") || !strings.Contains(errStr, "second error") {
			t.Error("should contain both error messages")
		}
	})
}

// TestSplitSQLEdgeCases tests edge cases for SplitSQL
func TestSplitSQLEdgeCases(t *testing.T) {
	t.Run("empty_separator", func(t *testing.T) {
		result := SplitSQL("select 1", "")
		if len(result) != 1 || result[0] != "select 1" {
			t.Error("empty separator should return original")
		}
	})

	t.Run("separator_longer_than_sql", func(t *testing.T) {
		result := SplitSQL("x", "go")
		if len(result) != 1 || result[0] != "x" {
			t.Error("separator longer than SQL should return original")
		}
	})

	t.Run("only_separator", func(t *testing.T) {
		result := SplitSQL("go", "go")
		// Should result in empty batch
		if len(result) != 0 {
			t.Errorf("expected 0 batches, got %d: %v", len(result), result)
		}
	})

	t.Run("consecutive_separators", func(t *testing.T) {
		result := SplitSQL("select 1\ngo\ngo\nselect 2", "go")
		// Empty batch between separators should be filtered
		nonEmpty := 0
		for _, s := range result {
			if strings.TrimSpace(s) != "" {
				nonEmpty++
			}
		}
		if nonEmpty != 2 {
			t.Errorf("expected 2 non-empty batches, got %d", nonEmpty)
		}
	})
}
