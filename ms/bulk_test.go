package ms

import (
	"context"
	"encoding/csv"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/kardianos/rdb"
)

func TestBulkInsert(t *testing.T) {
	const CSVText = `col_a,col_b,col_c
A,1,LP
B,2,LD
C,3,LC`

	cr := csv.NewReader(strings.NewReader(CSVText))
	all, err := cr.ReadAll()
	if err != nil {
		t.Fatal(err)
	}
	cols := all[0]
	data := all[1:]

	_ = cols
	list := []struct {
		Name string
		With bool
	}{
		{Name: "no_with_hints"},
		{Name: "with_hints", With: true},
	}

	const rowMax = 100000

	for _, item := range list {
		var rowIndex = 0
		t.Run(item.Name, func(t *testing.T) {
			b := &Bulk{
				TableName: "tab_01",
				Columns: []rdb.Param{
					{Name: "col_a", Type: rdb.Text, Length: 100},
					{Name: "col_b", Type: rdb.Text, Length: 100},
					{Name: "col_c", Type: rdb.Text, Length: 100},
				},
				Row: func(row []rdb.Param) error {
					if rowIndex >= rowMax {
						return io.EOF
					}
					extIndex := rowIndex % len(data)
					r := data[extIndex]
					for i := range row {
						row[i].Value = r[i]
					}
					rowIndex++
					return nil
				},
			}
			if item.With {
				b.TabLock = true
				b.RowsPerBatch = 1000
				b.KBPerBatch = 1024 * 10
			}
			ctx, cancel := context.WithTimeout(context.Background(), time.Second*20)
			defer cancel()

			db.Query(ctx, &rdb.Command{
				Arity: rdb.ZeroMust,
				SQL: `
drop table if exists tab_01;
create table tab_01 (
	col_a nvarchar(100),
	col_b nvarchar(100),
	col_c nvarchar(100)
);
`,
			})
			db.Query(ctx, &rdb.Command{
				Arity: rdb.ZeroMust,
				Bulk:  b,
			})
			var ct int
			db.Query(ctx, &rdb.Command{
				Arity: rdb.OneMust,
				SQL: `
select CT = count(*) from tab_01;
drop table if exists tab_01;
`,
			}).Prep("CT", &ct).Scan()

			if ct != rowMax {
				t.Fatalf("expected %d rows, got %d rows", rowMax, ct)
			}
		})
	}
}
