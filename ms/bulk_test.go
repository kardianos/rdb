package ms

import (
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/kardianos/rdb"
)

func TestBulkInsert(t *testing.T) {
	checkSkip(t)

	type TABLE struct {
		CSV        string
		Columns    []rdb.Param
		SQLColumns string

		name string
		data [][]string
	}
	tableLookup := map[string]*TABLE{
		"tab_01": {
			CSV: `col_a,col_b,col_c
A,1,LPAGHJKSDF
B,2,LDBSDFKJHSD
C,3,LCESDF`,
			Columns: []rdb.Param{
				{Name: "col_a", Type: rdb.Text, Length: 100},
				{Name: "col_b", Type: rdb.Text, Length: 100},
				{Name: "col_c", Type: rdb.Text, Length: 100},
			},
			SQLColumns: `
col_a nvarchar(100),
col_b nvarchar(100),
col_c nvarchar(100)
`,
		},
		"tab_02": {
			CSV: `col_a,col_b
1,` + strings.Repeat("LPAGHJKSDF", 100) + `
2,` + strings.Repeat("LDBSDFKJHSD", 100) + `
3,` + strings.Repeat("LCESDF", 100) + ``,
			Columns: []rdb.Param{
				{Name: "col_a", Type: rdb.Integer},
				{Name: "col_b", Type: rdb.Text},
			},
			SQLColumns: `
col_a bigint,
col_b nvarchar(max)
`,
		},
	}

	for name, tb := range tableLookup {
		cr := csv.NewReader(strings.NewReader(tb.CSV))
		all, err := cr.ReadAll()
		if err != nil {
			t.Fatal(err)
		}
		tb.name = name
		tb.data = all[1:]
	}

	list := []struct {
		Name      string
		Table     string
		With      bool
		RowMax    int
		BatchSize int
		Long      bool
	}{
		{Name: "zero01", Table: "tab_01", With: false, RowMax: 0, BatchSize: 1000},
		{Name: "zero02", Table: "tab_02", With: false, RowMax: 0, BatchSize: 1000},
		{Name: "short", Table: "tab_01", With: false, RowMax: 100, BatchSize: 1000},
		{Name: "no_with_hints", Table: "tab_01", With: false, RowMax: 100_000, BatchSize: 10_000},
		{Name: "with_hints", Table: "tab_01", With: true, RowMax: 100_000, BatchSize: 10_000},
		{Name: "long", Table: "tab_01", With: true, RowMax: 1_000_000, BatchSize: 100_000, Long: true},
		{Name: "wide", Table: "tab_02", With: true, RowMax: 10_000, BatchSize: 1_000, Long: true},
	}

	for _, item := range list {
		var rowIndex = 0
		t.Run(item.Name, func(t *testing.T) {
			if testing.Short() && item.Long {
				t.Skip("skip long test")
			}
			table := tableLookup[item.Table]
			b := &Bulk{
				TableName: item.Table,
				Columns:   table.Columns,
				Row: func(row []rdb.Param) error {
					if rowIndex >= item.RowMax {
						return io.EOF
					}
					extIndex := rowIndex % len(table.data)
					r := table.data[extIndex]
					for i := range row {
						row[i].Value = r[i]
					}
					rowIndex++
					return nil
				},
			}
			if item.With {
				b.TabLock = true
				b.RowsPerBatch = item.BatchSize
			}
			ctx, cancel := context.WithTimeout(context.Background(), time.Second*40)
			defer cancel()

			conn, err := db.Normal().Connection()
			if err != nil {
				t.Fatal(err)
			}
			defer conn.Close()

			_, err = conn.Query(ctx, &rdb.Command{
				Arity: rdb.ZeroMust,
				Log: func(msg *rdb.Message) {
					t.Log(msg.String())
				},
				SQL: fmt.Sprintf(`
drop table if exists %[1]s;
create table %[1]s (
	%[2]s
);
`, table.name, table.SQLColumns),
			})
			if err != nil {
				t.Fatal(err)
			}

			resp, err := conn.Query(ctx, &rdb.Command{
				Arity: rdb.ZeroMust,
				Log: func(msg *rdb.Message) {
					t.Log(msg.String())
				},
				Bulk: b,
			})
			if err != nil {
				t.Fatal(err)
			}
			affected := resp.RowsAffected()
			if affected != uint64(item.RowMax) {
				t.Errorf("expected %d rows affected, got %d rows", item.RowMax, affected)
			}

			var ct int
			resp, err = conn.Query(ctx, &rdb.Command{
				Arity: rdb.OneMust,
				Log: func(msg *rdb.Message) {
					t.Log(msg.String())
				},
				SQL: fmt.Sprintf(`
select CT = count(*) from %[1]s;
drop table if exists %[1]s;
`, table.name),
			})
			if err != nil {
				t.Fatal(err)
			}
			resp.Prep("CT", &ct).Scan()

			if ct != item.RowMax {
				t.Fatalf("expected %d rows, got %d rows", item.RowMax, ct)
			}
		})
	}
}
