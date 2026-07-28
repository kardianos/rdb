// Copyright 2014 Daniel Theophanes.
// Use of this source code is governed by a zlib-style
// license that can be found in the LICENSE file.

package ms

import (
	"context"
	"fmt"
	"math/big"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/kardianos/rdb"
	"github.com/kardianos/rdb/must"
)

// TestDecodeDateTimePayloadOversized ensures unexpected BYTELEN (e.g. 11 with
// datetime2 scale 7) does not panic after the payload is already consumed.
// Production: "proto error date/time: consumed 8 of 11 payload bytes (dt=3 scale=7)".
func TestDecodeDateTimePayloadOversized(t *testing.T) {
	// datetime2 = dtTime|dtDate = 1|2 = 3; scale 7 → normal payload 8 bytes.
	// Build a valid 8-byte value and pad to 11.
	scale := 7
	timeSz := timeBytesForScale(scale)
	payload := make([]byte, timeSz+3+3) // 5+3+3 = 11
	// time ticks: 1 second at scale 7 = 10^7 ticks
	ticks := int64(10_000_000)
	for i := 0; i < timeSz; i++ {
		payload[i] = byte(ticks >> (8 * i))
	}
	// date: day 0 = 0001-01-01
	// trailing 3 bytes left zero

	v := decodeDateTimePayload(dtTime|dtDate, scale, scale, payload)
	tm, ok := v.(time.Time)
	if !ok {
		t.Fatalf("got %T, want time.Time", v)
	}
	if tm.Year() != 1 || tm.Month() != 1 || tm.Day() != 1 {
		t.Errorf("date: got %v", tm)
	}
	// Must not panic; value should still decode the leading valid portion.
	if tm.Hour() != 0 || tm.Minute() != 0 || tm.Second() != 1 {
		t.Errorf("time-of-day: got %v", tm)
	}

	// Undersized: still no panic.
	_ = decodeDateTimePayload(dtTime|dtDate, scale, scale, []byte{1, 2})
	_ = decodeDateTimePayload(dtDate, 0, 0, []byte{1})
	_ = decodeDateTimePayload(dtTime, 0, 0, []byte{1, 2})
}

// TestTimeThenIntNoDesync is the production regression for:
//   proto error IntN, unknown data len 3
// Old time/datetime2 decode used dataLen as the time size, so for scale 0–2
// (time payload 3 bytes) it read 4–5 bytes and stole the following IntN length
// byte. The next "length" was then often 0x03 (first byte of a small int).
func TestTimeThenIntNoDesync(t *testing.T) {
	checkSkip(t)
	defer recoverTest(t)

	for scale := 0; scale <= 7; scale++ {
		scale := scale
		t.Run(fmt.Sprintf("scale_%d", scale), func(t *testing.T) {
			defer recoverTest(t)
			cmd := &rdb.Command{
				SQL: fmt.Sprintf(`
					SELECT
						t   = CAST('10:20:19.1234567' AS time(%d)),
						dt  = CAST('2026-07-28 10:20:19.1234567' AS datetime2(%d)),
						d   = CAST('2026-07-28' AS date),
						i1  = CAST(3 AS int),
						i2  = CAST(878 AS int),
						i3  = CAST(42 AS bigint),
						dec = CAST(1234.567891 AS decimal(38,20));
				`, scale, scale),
				Arity: rdb.OneMust,
			}
			res := db.Query(context.Background(), cmd)
			defer res.Close()
			if !res.Next() {
				t.Fatal("expected row")
			}
			res.Scan()
			// Must successfully decode ints after time/date (would panic with dataLen=3 before fix).
			if v := res.Get("i1"); v == nil {
				t.Fatal("i1 nil")
			}
			if v := res.Get("i2"); v == nil {
				t.Fatal("i2 nil")
			}
			// Follow-up: connection still usable.
			res.Close()
			res2 := db.Query(context.Background(), &rdb.Command{SQL: `SELECT CAST(1 AS int) AS x`, Arity: rdb.OneMust})
			defer res2.Close()
			if !res2.Next() {
				t.Fatal("follow-up failed — connection desynced")
			}
			res2.Scan()
		})
	}
}

// TestSampleLikeMixedRow stresses a Sample-shaped row: many ints after
// date/time/datetime2 at low scales plus high-scale decimal.
func TestSampleLikeMixedRow(t *testing.T) {
	checkSkip(t)
	defer recoverTest(t)

	cmd := &rdb.Command{
		SQL: `
			SELECT TOP 100
				ID = v.i,
				DateOrdered = CAST(DATEADD(day, -v.i % 10, GETUTCDATE()) AS date),
				TimeOfOrdered = CAST('08:15:30' AS time(0)),
				TimeWindow = CAST('09:00:00.123' AS time(3)),
				DateStart = CAST(GETUTCDATE() AS date),
				DateDone = CAST(SYSUTCDATETIME() AS datetime2(0)),
				TimeCreated = CAST(SYSUTCDATETIME() AS datetime2(3)),
				StatusLink = CAST(v.i % 5 AS int),
				Account = CAST(878 AS int),
				Qty = CAST(3 AS int),
				Amt = CAST(1234.567891 AS decimal(38,20)),
				MoneyAmt = CAST(100.5 AS decimal(38,6))
			FROM (VALUES
				(1),(2),(3),(4),(5),(6),(7),(8),(9),(10),
				(11),(12),(13),(14),(15),(16),(17),(18),(19),(20),
				(21),(22),(23),(24),(25),(26),(27),(28),(29),(30),
				(31),(32),(33),(34),(35),(36),(37),(38),(39),(40),
				(41),(42),(43),(44),(45),(46),(47),(48),(49),(50),
				(51),(52),(53),(54),(55),(56),(57),(58),(59),(60),
				(61),(62),(63),(64),(65),(66),(67),(68),(69),(70),
				(71),(72),(73),(74),(75),(76),(77),(78),(79),(80),
				(81),(82),(83),(84),(85),(86),(87),(88),(89),(90),
				(91),(92),(93),(94),(95),(96),(97),(98),(99),(100)
			) AS v(i)
			WHERE v.i >= @MinID
			ORDER BY v.i;
		`,
		Arity: rdb.Any,
	}

	// Repeat to catch desync on connection reuse.
	for iter := 0; iter < 30; iter++ {
		res := db.Query(context.Background(), cmd,
			rdb.Param{Name: "MinID", Type: rdb.Integer, Value: int64(1)},
			rdb.Param{Name: "AppAccount", Type: rdb.Integer, Value: int64(878)},
		)
		rows := 0
		for res.Next() {
			res.Scan()
			for i := range res.Schema() {
				_ = res.Getx(i)
			}
			rows++
		}
		res.Close()
		if rows == 0 {
			t.Fatalf("iter %d: expected rows", iter)
		}
	}
}

// TestTimeBytesForScale documents MS-TDS time component sizes.
func TestTimeBytesForScale(t *testing.T) {
	cases := []struct {
		scale int
		want  int
	}{
		{0, 3}, {1, 3}, {2, 3},
		{3, 4}, {4, 4},
		{5, 5}, {6, 5}, {7, 5},
	}
	for _, tc := range cases {
		if got := timeBytesForScale(tc.scale); got != tc.want {
			t.Errorf("timeBytesForScale(%d)=%d want %d", tc.scale, got, tc.want)
		}
	}

	// datetime2: dataLen = timeSize + 3
	for scale, timeSz := range map[int]int{0: 3, 2: 3, 3: 4, 4: 4, 5: 5, 7: 5} {
		dataLen := timeSz + 3
		got, err := timeBytesFromDataLen(dtDate|dtTime, dataLen)
		if err != nil || got != timeSz {
			t.Errorf("datetime2 scale=%d dataLen=%d: got %d err %v want %d", scale, dataLen, got, err, timeSz)
		}
	}
	// time only
	for _, dataLen := range []int{3, 4, 5} {
		got, err := timeBytesFromDataLen(dtTime, dataLen)
		if err != nil || got != dataLen {
			t.Errorf("time dataLen=%d: got %d err %v", dataLen, got, err)
		}
	}
	// datetimeoffset: dataLen = timeSize + 5
	for timeSz, dataLen := range map[int]int{3: 8, 4: 9, 5: 10} {
		got, err := timeBytesFromDataLen(dtDate|dtTime|dtZone, dataLen)
		if err != nil || got != timeSz {
			t.Errorf("dto dataLen=%d: got %d err %v want %d", dataLen, got, err, timeSz)
		}
	}
}

// TestDateTime2ScaleDecode stresses datetime2/time at every legal scale.
// Regression: wrong time-component size for scale<=4 desynced the TDS stream
// (unknown token peek: tdsToken(0)).
func TestDateTime2ScaleDecode(t *testing.T) {
	checkSkip(t)
	defer recoverTest(t)

	for scale := 0; scale <= 7; scale++ {
		scale := scale
		t.Run(fmt.Sprintf("scale_%d", scale), func(t *testing.T) {
			defer recoverTest(t)
			cmd := &rdb.Command{
				SQL: fmt.Sprintf(`
					SELECT
						d  = CAST('2026-07-28' AS date),
						t  = CAST('10:20:19.1234567' AS time(%d)),
						dt = CAST('2026-07-28 10:20:19.1234567' AS datetime2(%d)),
						dto = CAST('2026-07-28 10:20:19.1234567 +00:00' AS datetimeoffset(%d)),
						dec = CAST(1234.567891 AS decimal(38,20)),
						i   = CAST(878 AS int);
				`, scale, scale, scale),
				Arity: rdb.OneMust,
			}
			res := db.Query(context.Background(), cmd)
			defer res.Close()
			if !res.Next() {
				t.Fatal("expected row")
			}
			res.Scan()
			// Touch every column so decode runs fully.
			schema := res.Schema()
			for i := range schema {
				_ = res.Getx(i)
			}
			// Ensure stream is clean for a follow-up query on the same connection.
			res.Close()

			res2 := db.Query(context.Background(), &rdb.Command{
				SQL:   `SELECT 1 AS x`,
				Arity: rdb.OneMust,
			})
			defer res2.Close()
			if !res2.Next() {
				t.Fatal("follow-up query failed — connection likely desynced")
			}
			res2.Scan()
		})
	}
}

// TestWorkOrderLikeStress mimics the failing production query shape:
// many typed columns (date/time/datetime2 at mixed scales, decimal high scale),
// parameters, ORDER BY, and repeated concurrent execution on a shared pool.
// Uses an inline CTE (no session temp tables) so pool RESETCONNECTION is fine.
func TestWorkOrderLikeStress(t *testing.T) {
	checkSkip(t)
	defer recoverTest(t)

	ctx := context.Background()

	// Inline row source: mixed time scales (0, 3, 7), datetime2(0/3/7), decimal(38,20).
	// Use a small numbers table (not sys cross-join) so each iteration stays fast.
	const sql = `
;WITH n AS (
	SELECT v.i
	FROM (VALUES
		(1),(2),(3),(4),(5),(6),(7),(8),(9),(10),
		(11),(12),(13),(14),(15),(16),(17),(18),(19),(20),
		(21),(22),(23),(24),(25),(26),(27),(28),(29),(30),
		(31),(32),(33),(34),(35),(36),(37),(38),(39),(40),
		(41),(42),(43),(44),(45),(46),(47),(48),(49),(50)
	) AS v(i)
),
arity AS (
	SELECT
		i AS ID,
		i % 50 AS Sample,
		i % 10 AS Lab,
		DATEADD(day, -i % 30, CAST(GETUTCDATE() AS date)) AS DateOrdered,
		CAST(DATEADD(second, i, CAST('08:00:00' AS time)) AS time(7)) AS TimeOfOrdered,
		i % 5 AS WorkOrderStatus,
		DATEADD(day, i % 14, CAST(GETUTCDATE() AS date)) AS DateStart,
		CAST((i % 100) * 0.25 AS decimal(18,4)) AS DaySpanDue,
		CAST(1234.567891 + i * 0.000001 AS decimal(38,20)) AS HourDurationWindow,
		CAST('09:15:30.123' AS time(3)) AS TimeOfStart,
		CAST('17:00:00' AS time(0)) AS TimeOfEnd,
		DATEADD(day, i % 14 + 2, CAST(GETUTCDATE() AS date)) AS DateDue,
		CAST(DATEADD(hour, i % 24, SYSUTCDATETIME()) AS datetime2(7)) AS DateDone,
		CAST(DATEADD(minute, i, CAST('12:00:00' AS time)) AS time(7)) AS TimeOfDone,
		i % 20 AS LocationWard,
		CASE WHEN i % 2 = 0 THEN CAST(1 AS bit) ELSE CAST(0 AS bit) END AS Fasting,
		CONCAT(N'comment-', i) AS Comment,
		CASE WHEN i % 17 = 0 THEN CAST(1 AS bit) ELSE CAST(0 AS bit) END AS Deleted,
		CAST(DATEADD(millisecond, i, SYSUTCDATETIME()) AS datetime2(3)) AS TimeCreated,
		CAST(SYSUTCDATETIME() AS datetime2(0)) AS TimeUpdated,
		CAST(100.50 + i AS decimal(38,6)) AS MoneyAmt
	FROM n
)
SELECT TOP 500
	arity.ID, arity.Sample, arity.Lab, arity.DateOrdered, arity.TimeOfOrdered,
	arity.WorkOrderStatus, arity.DateStart, arity.DaySpanDue, arity.HourDurationWindow,
	arity.TimeOfStart, arity.TimeOfEnd, arity.DateDue, arity.DateDone, arity.TimeOfDone,
	arity.LocationWard, arity.Fasting, arity.Comment, arity.Deleted,
	arity.TimeCreated, arity.TimeUpdated, arity.MoneyAmt
FROM arity
WHERE
	(arity.Deleted IS NULL OR arity.Deleted = 0)
	AND arity.DateStart <= @DateStartEnd
	AND (@WorkOrderStatusLink = 0 OR arity.WorkOrderStatus = @WorkOrderStatusLink)
ORDER BY arity.DateDue ASC, arity.ID ASC;
`

	runOnce := func(pool must.ConnPool, id int) (err error) {
		defer func() {
			if recovered := recover(); recovered != nil {
				err = fmt.Errorf("panic: %v", recovered)
			}
		}()
		now := time.Now().UTC()
		cmd := &rdb.Command{SQL: sql, Arity: rdb.Any}
		res := pool.Query(ctx, cmd,
			rdb.Param{Name: "AppAccount", Type: rdb.Integer, Value: int64(878)},
			rdb.Param{Name: "AppNow", Type: rdb.TypeTimestamp, Value: now},
			rdb.Param{Name: "DateStartEnd", Type: rdb.TypeDate, Value: now.Add(48 * time.Hour)},
			rdb.Param{Name: "WorkOrderStatusLink", Type: rdb.Integer, Value: int64(id % 5)},
		)
		defer res.Close()

		rows := 0
		for res.Next() {
			// Scan advances the row; Next only reports whether one is available.
			// must.Result.Scan panics on error (caught by defer recover above).
			res.Scan()
			schema := res.Schema()
			for i := range schema {
				v := res.Getx(i)
				switch schema[i].Name {
				case "HourDurationWindow", "MoneyAmt", "DaySpanDue":
					if v != nil {
						if _, ok := v.(*big.Rat); !ok {
							return fmt.Errorf("col %s type %T", schema[i].Name, v)
						}
					}
				}
			}
			rows++
		}
		if rows == 0 {
			return fmt.Errorf("expected rows")
		}
		return nil
	}

	// Serial stress on the shared test pool (capacity 1): many back-to-back queries.
	const serialIters = 80
	for i := 0; i < serialIters; i++ {
		if err := runOnce(db, i); err != nil {
			t.Fatalf("serial %d: %v", i, err)
		}
	}

	// Concurrent stress on a dedicated multi-conn pool (test db is capacity 1).
	if config == nil {
		t.Logf("serial stress ok (%d iters); skip concurrent (no config)", serialIters)
		return
	}
	ccfg := *config
	ccfg.PoolInitCapacity = 4
	ccfg.PoolMaxCapacity = 8
	cpool := must.Open(&ccfg)
	defer cpool.Close()

	const workers = 8
	const iters = 25
	var failures atomic.Int64
	var wg sync.WaitGroup
	wg.Add(workers)
	for w := 0; w < workers; w++ {
		w := w
		go func() {
			defer wg.Done()
			for i := 0; i < iters; i++ {
				if err := runOnce(cpool, w*iters+i); err != nil {
					failures.Add(1)
					t.Errorf("worker %d iter %d: %v", w, i, err)
					return
				}
			}
		}()
	}
	wg.Wait()
	if failures.Load() != 0 {
		t.Fatalf("%d concurrent failures", failures.Load())
	}

	// Follow-up on shared pool: ensures it is not desynced.
	res := db.Query(ctx, &rdb.Command{SQL: `SELECT 1 AS x`, Arity: rdb.OneMust})
	defer res.Close()
	if !res.Next() {
		t.Fatal("follow-up query failed — connection likely desynced")
	}
	t.Logf("stress ok; serial=%d concurrent=%d×%d", serialIters, workers, iters)
}
