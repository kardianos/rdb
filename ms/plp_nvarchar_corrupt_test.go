package ms

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/kardianos/rdb"
	"github.com/kardianos/rdb/must"
)

func TestPLPNvarcharLargeRoundTrip(t *testing.T) {
	checkSkip(t)
	defer recoverTest(t)

	ctx := context.Background()
	payload, err := os.ReadFile("/home/d/work/kashi_project/mmdx_debug/260727046MM.json")
	if err != nil {
		payload = bytes.Repeat([]byte("0123456789abcdef"), 26111/16+1)[:26111]
		t.Logf("using synthetic payload %d bytes", len(payload))
	} else {
		t.Logf("using real JSON payload %d bytes", len(payload))
	}

	// Single dedicated connection (avoid pool / temp-table session issues).
	cfg := must.Config(rdb.ParseConfigURL(testConnectionString))
	cfg.PoolInitCapacity = 1
	cfg.PoolMaxCapacity = 1
	pool := must.Open(cfg)
	defer pool.Close()
	conn, err := pool.Normal().Connection(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()

	q := func(sql string, params ...rdb.Param) *rdb.Result {
		res, err := conn.Query(ctx, &rdb.Command{SQL: sql, Arity: rdb.Any}, params...)
		if err != nil {
			t.Fatalf("query: %v\nSQL: %s", err, sql)
		}
		// drain / close for non-select
		return res
	}

	res := q(`set textsize -1;`)
	res.Close()

	res = q(`
if object_id('tempdb..#rdb_plp_nv') is not null drop table #rdb_plp_nv;
create table #rdb_plp_nv (id int primary key, name nvarchar(100), body nvarchar(max) not null, d date null);
`)
	res.Close()

	res = q(`insert into #rdb_plp_nv (id, name, body, d) values (1, N'row1', @body, null);`,
		rdb.Param{Name: "body", Type: rdb.Text, Value: string(payload)})
	res.Close()

	res = q(`select body from #rdb_plp_nv where id = 1;`)
	if !res.Next() {
		res.Close()
		t.Fatal("no row body-only")
	}
	if err := res.Scan(); err != nil {
		res.Close()
		t.Fatal(err)
	}
	gotB := asBytes(t, res.Get("body"))
	res.Close()
	assertBytes(t, "body-only", payload, gotB)

	res = q(`select id, name, body, d from #rdb_plp_nv where id = 1;`)
	if !res.Next() {
		res.Close()
		t.Fatal("no row wide")
	}
	if err := res.Scan(); err != nil {
		res.Close()
		t.Fatal(err)
	}
	gotB = asBytes(t, res.Get("body"))
	res.Close()
	assertBytes(t, "wide-row", payload, gotB)

	for i := 2; i <= 5; i++ {
		res = q(`insert into #rdb_plp_nv (id, name, body, d) values (@id, N'x', @body, null);`,
			rdb.Param{Name: "id", Type: rdb.TypeInt32, Value: int32(i)},
			rdb.Param{Name: "body", Type: rdb.Text, Value: string(payload)},
		)
		res.Close()
	}
	res = q(`select id, name, body, d from #rdb_plp_nv order by id;`)
	n := 0
	for res.Next() {
		if err := res.Scan(); err != nil {
			res.Close()
			t.Fatalf("scan row %d: %v", n, err)
		}
		gotB = asBytes(t, res.Get("body"))
		assertBytes(t, fmt.Sprintf("row-%d", n), payload, gotB)
		n++
	}
	res.Close()
	if n != 5 {
		t.Fatalf("got %d rows want 5", n)
	}
	t.Logf("all %d rows OK, payload %d bytes", n, len(payload))
}

func asBytes(t *testing.T, v interface{}) []byte {
	t.Helper()
	switch x := v.(type) {
	case []byte:
		return x
	case string:
		return []byte(x)
	case nil:
		t.Fatal("nil body")
	default:
		t.Fatalf("body type %T", v)
	}
	return nil
}

func assertBytes(t *testing.T, label string, want, got []byte) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("%s: len got %d want %d", label, len(got), len(want))
	}
	diffs := 0
	var first []string
	var offs []int
	for i := range want {
		if want[i] != got[i] {
			diffs++
			offs = append(offs, i)
			if len(first) < 15 {
				first = append(first, fmt.Sprintf("%d: want %q 0x%02X got %q 0x%02X", i, want[i], want[i], got[i], got[i]))
			}
		}
	}
	if diffs != 0 {
		gaps := make([]int, 0)
		for i := 1; i < len(offs) && i < 12; i++ {
			gaps = append(gaps, offs[i]-offs[i-1])
		}
		t.Fatalf("%s: %d bytes differ; offs=%v gaps=%v\nfirst: %v", label, diffs, offs[:min(12, len(offs))], gaps, first)
	}
}
