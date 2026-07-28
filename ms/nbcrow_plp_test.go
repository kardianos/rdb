package ms

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/kardianos/rdb"
	"github.com/kardianos/rdb/must"
)

// TestNBCRowPLPNullBitmap ensures the NBCROW null bitmap survives multi-packet
// PLP reads. Regression: nulls was a view into msgBuf; fill() during long
// nvarchar(max) overwrote it so later null date/time columns were decoded as
// non-null and desynced the stream (tdsToken(0), IntN bad dataLen, EOF).
func TestNBCRowPLPNullBitmap(t *testing.T) {
	checkSkip(t)
	defer recoverTest(t)

	ctx := context.Background()
	// Multi-chunk PLP (packet body ~4k) + trailing NULL date/time forces NBCROW
	// with a bitmap that must remain valid across fill().
	for _, nchars := range []int{100, 2026, 2861, 4000, 8000} {
		nchars := nchars
		t.Run(fmt.Sprintf("chars_%d", nchars), func(t *testing.T) {
			defer recoverTest(t)
			sql := fmt.Sprintf(`
select top 5
  id = n,
  c = cast(case when n = 3 then replicate(N'Z', %d) else N'' end as nvarchar(max)),
  d = cast(null as date),
  tm = cast(null as time(7))
from (select top 5 n = row_number() over (order by (select 1)) from sys.all_columns) x`, nchars)
			res := db.Query(ctx, &rdb.Command{SQL: sql, Arity: rdb.Any})
			defer res.Close()
			got := 0
			for res.Next() {
				res.Scan()
				// date and time must be null
				if v := res.Getx(2); v != nil {
					t.Fatalf("row %d date want nil got %v", got, v)
				}
				if v := res.Getx(3); v != nil {
					t.Fatalf("row %d time want nil got %v", got, v)
				}
				got++
			}
			if got != 5 {
				t.Fatalf("got %d rows want 5", got)
			}
			// Stream still usable
			res2 := db.Query(ctx, &rdb.Command{SQL: `select cast(1 as int) as x`, Arity: rdb.OneMust})
			defer res2.Close()
			if !res2.Next() {
				t.Fatal("desync probe: no row")
			}
			res2.Scan()
		})
	}
}

// TestSampleQueryLimsF03 is an integration check against the lims Sample shape
// that originally failed with unknown token peek: tdsToken(0).
func TestSampleQueryLimsF03(t *testing.T) {
	dsn := "ms://sa:Limit48%5ECheese@ssdb.internal.coredata.biz:1490?db=kashi_lims_f03&cert=/home/d/cert/le-chain.pem&insecure_skip_verify=true"
	cfg, err := rdb.ParseConfigURL(dsn)
	if err != nil {
		t.Fatal(err)
	}
	cfg.PoolInitCapacity = 1
	cfg.PoolMaxCapacity = 1
	cfg.DialTimeout = 5 * time.Second
	pool := must.Open(cfg)
	defer pool.Close()
	if err := pool.Normal().Ping(context.Background()); err != nil {
		t.Skipf("f03 unavailable: %v", err)
	}
	db := pool.Normal()
	ctx := context.Background()

	sql := `select top 5000 arity.ID,arity.SampleGroup,arity.SampleStudy,arity.SampleType,arity.TestedOutOfLab,arity.SampleIdentifier,arity.ExternalSampleIdentifier,arity.ViewID,arity.DateReceived,arity.TimeOfReceived,arity.Item,arity.ParentItem,arity.OrganizationLOB,arity.ReadOrganization,arity.FromPerson,arity.ForPerson,arity.Relationship,arity.Age,arity.ConfirmatoryTest,arity.LabComment,arity.ReportComment,arity.PortalComment,arity.OrderedBy,arity.OrderedByPhysician,arity.OrderedByPractice,arity.LocationWard,arity.DateCollected,arity.TimeOfCollected,arity.Extracted,arity.DuplicateSampleReason,arity.Surface,arity.ReferralSource,arity.DateBiopsy,arity.DateTransplant,arity.TransplantType,arity.BiopsyIndication,arity.PrimaryDisease,arity.NumberOfPieces,arity.RINDec,arity.Deleted,arity.TimeCreated,arity.AccountCreated,arity.TimeUpdated,arity.AccountUpdated,arity.TimeDeleted,arity.AccountDeleted
from data.Sample arity
where (arity.Deleted is null or arity.Deleted = 0)
and (
exists(select top 1 1 from data.OrganizationLOB ol where ol.ID = arity.OrganizationLOB and ol.LOB = @LOBLink)
or exists (select top 1 1 from data.SamplePanel sp join data.Panel p on p.ID = sp.Panel join data.PanelGroup pg on pg.ID = p.PanelGroup where sp.Sample = arity.ID and pg.LOB = @LOBLink)
or exists (select top 1 1 from data.SampleOrder so join data.ReportDefinition rd on rd.ID = so.ReportDefinition where so.Sample = arity.ID and rd.LOB = @LOBLink and so.Deleted = 0)
)
order by arity.SampleIdentifier asc, arity.TimeUpdated desc`

	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("panic: %v", r)
		}
	}()
	res, err := db.Query(ctx, &rdb.Command{SQL: sql, Arity: rdb.Any},
		rdb.Param{Name: "LOBLink", Type: rdb.TypeInt64, Value: int64(1)},
	)
	if err != nil {
		t.Fatal(err)
	}
	defer res.Close()
	n := 0
	for res.Next() {
		if err := res.Scan(); err != nil {
			t.Fatalf("scan row %d: %v", n, err)
		}
		for i := range res.Schema() {
			_ = res.Getx(i)
		}
		n++
	}
	t.Logf("Sample full query: %d rows OK", n)
	if n == 0 {
		t.Fatal("expected rows")
	}
}
