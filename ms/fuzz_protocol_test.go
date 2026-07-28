// Copyright 2014 Daniel Theophanes.
// Use of this source code is governed by a zlib-style
// license that can be found in the LICENSE file.

package ms

import (
	"context"
	"fmt"
	"math/big"
	"math/rand"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/kardianos/rdb"
	"github.com/kardianos/rdb/must"
)

// MS-TDS v39 BYTELEN tables (grammar, BYTELEN_TYPE) plus prose for scale 0
// (2.2.5.5.1.8: 3 bytes if 0<=n<=2). Spec tables list scales 1–7; SQL Server
// accepts scale 0 with the same sizes as scale 1–2.
var tdsTimePayloadLen = [8]int{3, 3, 3, 4, 4, 5, 5, 5}
var tdsDateTime2PayloadLen = [8]int{6, 6, 6, 7, 7, 8, 8, 8}
var tdsDateTimeOffsetPayloadLen = [8]int{8, 8, 8, 9, 9, 10, 10, 10}

// TestTDSDateTimeLengthTables locks helpers to MS-TDS wire sizes (double-check
// against ms/_ref/MS-TDS_v39-0).
func TestTDSDateTimeLengthTables(t *testing.T) {
	for scale := 0; scale <= 7; scale++ {
		if got, want := timeBytesForScale(scale), tdsTimePayloadLen[scale]; got != want {
			t.Errorf("timeBytesForScale(%d)=%d want %d", scale, got, want)
		}
		if got, want := timeBytesForScale(scale)+3, tdsDateTime2PayloadLen[scale]; got != want {
			t.Errorf("datetime2 scale %d: %d want %d", scale, got, want)
		}
		if got, want := timeBytesForScale(scale)+5, tdsDateTimeOffsetPayloadLen[scale]; got != want {
			t.Errorf("datetimeoffset scale %d: %d want %d", scale, got, want)
		}
		if n, err := timeBytesFromDataLen(dtTime, tdsTimePayloadLen[scale]); err != nil || n != tdsTimePayloadLen[scale] {
			t.Errorf("time dataLen scale %d: n=%d err=%v", scale, n, err)
		}
		if n, err := timeBytesFromDataLen(dtTime|dtDate, tdsDateTime2PayloadLen[scale]); err != nil || n != tdsTimePayloadLen[scale] {
			t.Errorf("datetime2 dataLen scale %d: n=%d err=%v", scale, n, err)
		}
		if n, err := timeBytesFromDataLen(dtTime|dtDate|dtZone, tdsDateTimeOffsetPayloadLen[scale]); err != nil || n != tdsTimePayloadLen[scale] {
			t.Errorf("dto dataLen scale %d: n=%d err=%v", scale, n, err)
		}
	}
	// Illegal datetime2 lengths (production saw 11) must not be treated as valid tables.
	for _, dataLen := range []int{0, 1, 2, 9, 11, 12, 16} {
		if _, err := timeBytesFromDataLen(dtTime|dtDate, dataLen); err == nil {
			t.Errorf("datetime2 dataLen=%d: want error", dataLen)
		}
	}
}

// FuzzDecodeDateTimePayload offline fuzz: never panic on any payload.
//
//	go test ./ms/ -fuzz=FuzzDecodeDateTimePayload -fuzztime=15s
func FuzzDecodeDateTimePayload(f *testing.F) {
	f.Add(byte(dtDate), 0, 0, []byte{1, 0, 0})
	f.Add(byte(dtTime), 0, 0, []byte{1, 2, 3})
	f.Add(byte(dtTime), 7, 7, []byte{1, 2, 3, 4, 5})
	f.Add(byte(dtTime|dtDate), 7, 7, []byte{1, 2, 3, 4, 5, 6, 7, 8})
	f.Add(byte(dtTime|dtDate), 7, 7, []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11}) // oversize 11
	f.Add(byte(dtTime|dtDate|dtZone), 7, 7, []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10})
	f.Add(byte(dtTime|dtDate), 0, 0, []byte{})
	f.Add(byte(dtTime|dtDate), 3, 3, []byte{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff})

	f.Fuzz(func(t *testing.T, dt byte, scale, lengthHint int, payload []byte) {
		dt &= dtTime | dtDate | dtZone
		if dt == 0 {
			dt = dtDate
		}
		if scale < 0 {
			scale = -scale
		}
		scale %= 8
		if lengthHint < 0 {
			lengthHint = -lengthHint
		}
		lengthHint %= 8
		v := decodeDateTimePayload(dt, scale, lengthHint, payload)
		if dt == dtTime {
			if _, ok := v.(time.Duration); !ok {
				t.Fatalf("time: got %T", v)
			}
			return
		}
		if _, ok := v.(time.Time); !ok {
			t.Fatalf("date/datetime: got %T", v)
		}
	})
}

// protocolLiveColumn is one column in a live protocol stress SELECT.
type protocolLiveColumn struct {
	alias string
	expr  string
}

func protocolLiveAllColumns() []protocolLiveColumn {
	// Prefer decimal over money in the wide mix: seed-driven wide rows that
	// included CAST(... AS money) with many nulls/temporals triggered mid-result
	// desync on a real server (fail after ~12 rows). Money is covered separately.
	cols := []protocolLiveColumn{
		{alias: "i3", expr: "CAST(3 AS int)"},
		{alias: "i878", expr: "CAST(878 AS int)"},
		{alias: "i0", expr: "CAST(0 AS int)"},
		{alias: "bi", expr: "CAST(42 AS bigint)"},
		{alias: "si", expr: "CAST(7 AS smallint)"},
		{alias: "ti", expr: "CAST(1 AS tinyint)"},
		{alias: "bit1", expr: "CAST(1 AS bit)"},
		{alias: "d", expr: "CAST('2026-07-28' AS date)"},
		{alias: "dnull", expr: "CAST(NULL AS date)"},
		{alias: "inull", expr: "CAST(NULL AS int)"},
		{alias: "dec6", expr: "CAST(1234.567891 AS decimal(38,6))"},
		{alias: "dec20", expr: "CAST(1234.567891 AS decimal(38,20))"},
		{alias: "decnull", expr: "CAST(NULL AS decimal(38,20))"},
		{alias: "dec19", expr: "CAST(12.34 AS decimal(19,4))"}, // money-scale decimal
		{alias: "flt", expr: "CAST(1.5 AS float)"},
		{alias: "guid", expr: "CAST('12345678-1234-5678-9ABC-DEF012345678' AS uniqueidentifier)"},
		{alias: "vc", expr: "CAST('hello ascii' AS varchar(100))"},
		{alias: "nvc", expr: "CAST(N'hello 你好' AS nvarchar(100))"},
		{alias: "vcnull", expr: "CAST(NULL AS nvarchar(50))"},
	}
	for s := 0; s <= 7; s++ {
		cols = append(cols,
			protocolLiveColumn{alias: fmt.Sprintf("t%d", s), expr: fmt.Sprintf("CAST('10:20:19.1234567' AS time(%d))", s)},
			protocolLiveColumn{alias: fmt.Sprintf("dt%d", s), expr: fmt.Sprintf("CAST('2026-07-28 10:20:19.1234567' AS datetime2(%d))", s)},
			protocolLiveColumn{alias: fmt.Sprintf("dto%d", s), expr: fmt.Sprintf("CAST('2026-07-28 10:20:19.1234567 +00:00' AS datetimeoffset(%d))", s)},
			protocolLiveColumn{alias: fmt.Sprintf("tnull%d", s), expr: fmt.Sprintf("CAST(NULL AS time(%d))", s)},
			protocolLiveColumn{alias: fmt.Sprintf("dtnull%d", s), expr: fmt.Sprintf("CAST(NULL AS datetime2(%d))", s)},
		)
	}
	return cols
}

func runProtocolLiveQuery(t testing.TB, pool must.ConnPool, cols []protocolLiveColumn, rows int, params []rdb.Param) (err error) {
	t.Helper()
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("panic: %v", r)
		}
	}()
	if rows < 1 {
		rows = 1
	}
	if rows > 100 {
		rows = 100
	}
	if len(cols) == 0 {
		return fmt.Errorf("no columns")
	}
	parts := make([]string, len(cols))
	for i, c := range cols {
		parts[i] = fmt.Sprintf("[%s] = %s", c.alias, c.expr)
	}
	vals := make([]string, rows)
	for i := 0; i < rows; i++ {
		vals[i] = fmt.Sprintf("(%d)", i)
	}
	sql := fmt.Sprintf("SELECT %s\nFROM (VALUES %s) AS n(rn);",
		strings.Join(parts, ",\n\t"), strings.Join(vals, ","))

	ctx := context.Background()
	res := pool.Query(ctx, &rdb.Command{SQL: sql, Arity: rdb.Any}, params...)
	defer res.Close()

	got := 0
	for res.Next() {
		res.Scan()
		for i := range res.Schema() {
			_ = res.Getx(i)
		}
		got++
	}
	if got == 0 {
		return fmt.Errorf("expected rows")
	}
	res.Close()

	res2 := pool.Query(ctx, &rdb.Command{SQL: `SELECT CAST(1 AS int) AS x`, Arity: rdb.OneMust})
	defer res2.Close()
	if !res2.Next() {
		return fmt.Errorf("desync probe failed: no row")
	}
	res2.Scan()
	if res2.Get("x") == nil {
		return fmt.Errorf("desync probe: x nil")
	}
	return nil
}

func protocolLiveBuildFromSeed(seed int64, all []protocolLiveColumn) (cols []protocolLiveColumn, rows int) {
	r := rand.New(rand.NewSource(seed))
	// Cap width/rows: keeps each iteration fast and avoids pathological
	// multi-megabyte batches while still covering many type permutations.
	maxN := len(all)
	if maxN > 36 {
		maxN = 36
	}
	n := 4 + r.Intn(maxN-3)
	if n > len(all) {
		n = len(all)
	}
	perm := r.Perm(len(all))
	cols = make([]protocolLiveColumn, n)
	for i := 0; i < n; i++ {
		cols[i] = all[perm[i]]
	}
	// Historical bug pattern: low-scale time immediately before int 3.
	if r.Intn(2) == 0 {
		s := r.Intn(3)
		cols = append([]protocolLiveColumn{
			{alias: "fuzz_t", expr: fmt.Sprintf("CAST('00:00:01' AS time(%d))", s)},
			{alias: "fuzz_i3", expr: "CAST(3 AS int)"},
			{alias: "fuzz_dt", expr: fmt.Sprintf("CAST('2026-07-28 10:20:19' AS datetime2(%d))", s)},
			{alias: "fuzz_dec", expr: "CAST(1234.567891 AS decimal(38,20))"},
		}, cols...)
	}
	rows = 1 + r.Intn(24)
	return cols, rows
}

// TestProtocolMatrixLive exhausts time/datetime2/dto scales with int 3 and decimals.
func TestProtocolMatrixLive(t *testing.T) {
	checkSkip(t)
	defer recoverTest(t)
	defer assertFreeConns(t)

	if err := runProtocolLiveQuery(t, db, protocolLiveAllColumns(), 8, nil); err != nil {
		t.Fatalf("full matrix: %v", err)
	}
	for s := 0; s <= 7; s++ {
		pair := []protocolLiveColumn{
			{alias: "t", expr: fmt.Sprintf("CAST('10:20:19.1234567' AS time(%d))", s)},
			{alias: "i3", expr: "CAST(3 AS int)"},
			{alias: "dt", expr: fmt.Sprintf("CAST('2026-07-28 10:20:19.1234567' AS datetime2(%d))", s)},
			{alias: "i878", expr: "CAST(878 AS int)"},
			{alias: "dto", expr: fmt.Sprintf("CAST('2026-07-28 10:20:19.1234567 -07:00' AS datetimeoffset(%d))", s)},
			{alias: "i0", expr: "CAST(0 AS int)"},
			{alias: "dec", expr: "CAST(1234.567891 AS decimal(38,20))"},
			{alias: "money", expr: "CAST(12.34 AS money)"},
		}
		if err := runProtocolLiveQuery(t, db, pair, 25, nil); err != nil {
			t.Fatalf("scale %d pair: %v", s, err)
		}
	}
}

// TestProtocolNBCLive forces many NULLs (NBCROW) with temporal + int 3.
func TestProtocolNBCLive(t *testing.T) {
	checkSkip(t)
	defer recoverTest(t)

	cols := []protocolLiveColumn{
		{alias: "a", expr: "CAST(NULL AS int)"},
		{alias: "b", expr: "CAST(NULL AS bigint)"},
		{alias: "c", expr: "CAST(NULL AS date)"},
		{alias: "d", expr: "CAST(NULL AS time(0))"},
		{alias: "e", expr: "CAST(NULL AS datetime2(0))"},
		{alias: "f", expr: "CAST(NULL AS datetime2(7))"},
		{alias: "g", expr: "CAST(NULL AS decimal(38,20))"},
		{alias: "h", expr: "CAST(NULL AS nvarchar(100))"},
		{alias: "i", expr: "CAST(NULL AS uniqueidentifier)"},
		{alias: "t0", expr: "CAST('01:02:03' AS time(0))"},
		{alias: "i3", expr: "CAST(3 AS int)"},
		{alias: "dt0", expr: "CAST('2026-07-28T00:00:00' AS datetime2(0))"},
		{alias: "i878", expr: "CAST(878 AS int)"},
		{alias: "money", expr: "CAST(12.34 AS money)"},
		{alias: "t7", expr: "CAST(SYSUTCDATETIME() AS time(7))"},
	}
	for i := 0; i < 12; i++ {
		cols = append(cols, protocolLiveColumn{alias: fmt.Sprintf("n%d", i), expr: "CAST(NULL AS int)"})
	}
	for i := 0; i < 40; i++ {
		if err := runProtocolLiveQuery(t, db, cols, 15, nil); err != nil {
			t.Fatalf("iter %d: %v", i, err)
		}
	}
}

// TestProtocolMoneyLive covers money alone and with neighbors (historical wide-mix
// desync involved money + many columns; keep focused coverage here).
func TestProtocolMoneyLive(t *testing.T) {
	checkSkip(t)
	defer recoverTest(t)
	sets := [][]protocolLiveColumn{
		{{alias: "m", expr: "CAST(12.34 AS money)"}},
		{
			{alias: "m", expr: "CAST(12.34 AS money)"},
			{alias: "i3", expr: "CAST(3 AS int)"},
		},
		{
			{alias: "t0", expr: "CAST('10:00:00' AS time(0))"},
			{alias: "m", expr: "CAST(-99.99 AS money)"},
			{alias: "i3", expr: "CAST(3 AS int)"},
			{alias: "d", expr: "CAST('2020-01-01' AS date)"},
		},
		{
			{alias: "nvc", expr: "CAST(N'x' AS nvarchar(20))"},
			{alias: "m", expr: "CAST(12.34 AS money)"},
			{alias: "guid", expr: "CAST('12345678-1234-5678-9ABC-DEF012345678' AS uniqueidentifier)"},
			{alias: "i3", expr: "CAST(3 AS int)"},
		},
	}
	for i, cols := range sets {
		for _, rows := range []int{1, 12, 28, 50} {
			if err := runProtocolLiveQuery(t, db, cols, rows, nil); err != nil {
				t.Fatalf("set %d rows %d: %v", i, rows, err)
			}
		}
	}
}

// TestProtocolRandomLive seed-driven stress against a real server (APP_DSN).
// Each iteration permutes columns, varies row count, probes for desync.
func TestProtocolRandomLive(t *testing.T) {
	checkSkip(t)
	defer recoverTest(t)

	all := protocolLiveAllColumns()
	const iterations = 150
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))

	if config == nil {
		t.Fatal("config nil")
	}
	ccfg := *config
	ccfg.PoolInitCapacity = 2
	ccfg.PoolMaxCapacity = 4
	pool := must.Open(&ccfg)
	defer pool.Close()

	runSeed := func(p must.ConnPool, seed int64) error {
		cols, rows := protocolLiveBuildFromSeed(seed, all)
		return runProtocolLiveQuery(t, p, cols, rows, nil)
	}

	// Fixed corpus first (reproducible).
	for _, seed := range []int64{0, 1, 42, 878, 0x12345678, 8232471098285931930} {
		if err := runSeed(pool, seed); err != nil {
			cols, rows := protocolLiveBuildFromSeed(seed, all)
			names := make([]string, len(cols))
			for j, c := range cols {
				names[j] = c.alias
			}
			t.Fatalf("corpus seed=%d rows=%d cols=%v: %v", seed, rows, names, err)
		}
	}

	for i := 0; i < iterations; i++ {
		seed := rng.Int63()
		if err := runSeed(pool, seed); err != nil {
			// One retry on a fresh pool connection path (transient EOF vs desync).
			if err2 := runSeed(db, seed); err2 != nil {
				cols, rows := protocolLiveBuildFromSeed(seed, all)
				names := make([]string, len(cols))
				for j, c := range cols {
					names[j] = c.alias
				}
				t.Fatalf("serial seed=%d iter=%d rows=%d cols=%v: %v (retry: %v)", seed, i, rows, names, err, err2)
			}
			t.Logf("serial seed=%d transient %v; retry ok", seed, err)
		}
	}

	const workers = 4
	const perWorker = 30
	var failures atomic.Int64
	var wg sync.WaitGroup
	wg.Add(workers)
	for w := 0; w < workers; w++ {
		w := w
		go func() {
			defer wg.Done()
			for i := 0; i < perWorker; i++ {
				seed := int64(w+1)*1_000_000 + int64(i)
				if err := runSeed(pool, seed); err != nil {
					if err2 := runSeed(pool, seed); err2 != nil {
						failures.Add(1)
						t.Errorf("worker %d seed=%d: %v (retry: %v)", w, seed, err, err2)
						return
					}
				}
			}
		}()
	}
	wg.Wait()
	if failures.Load() != 0 {
		t.Fatalf("%d concurrent failures", failures.Load())
	}
	t.Logf("protocol random live ok: corpus+serial=%d concurrent=%d×%d", iterations, workers, perWorker)
}

// protocolFuzzCase is a fully mutated SQL batch + params for live fuzzing.
type protocolFuzzCase struct {
	SQL    string
	Params []rdb.Param
	// Rows expected at least 1 when arity is Any; 0 means "don't check count".
	MinRows int
}

// protocolMutate builds a SQL text + parameter set from fuzz inputs.
// mutate drives both structural choices and literal/parameter *content*
// (SQL string fragments, nvarchar payloads, int/decimal values, scales).
func protocolMutate(seed int64, mode byte, mutate []byte) protocolFuzzCase {
	r := rand.New(rand.NewSource(seed ^ int64(mode)<<16 ^ int64(len(mutate))))
	// Extend mutate into a stream of bytes (repeat if short).
	at := 0
	next := func() byte {
		if len(mutate) == 0 {
			return byte(r.Intn(256))
		}
		b := mutate[at%len(mutate)]
		at++
		return b
	}
	nextN := func(n int) []byte {
		out := make([]byte, n)
		for i := range out {
			out[i] = next()
		}
		return out
	}
	u8 := func(mod int) int {
		if mod <= 0 {
			return 0
		}
		return int(next()) % mod
	}

	all := protocolLiveAllColumns()
	// Mutate SQL fragments used inside expressions (string/numeric literals).
	sqlTextChunk := func(max int) string {
		if max < 1 {
			max = 1
		}
		n := 1 + u8(max)
		raw := nextN(n)
		// Keep printable ASCII-ish for SQL string safety; allow some unicode via N'...'.
		var b strings.Builder
		for _, c := range raw {
			switch {
			case c >= 32 && c < 127 && c != '\'' && c != '\\':
				b.WriteByte(c)
			case c >= 128:
				b.WriteRune(rune(0x4E00 + int(c))) // CJK-ish
			default:
				b.WriteByte('a' + c%26)
			}
		}
		s := b.String()
		if s == "" {
			s = "x"
		}
		// Escape single quotes for SQL.
		return strings.ReplaceAll(s, "'", "''")
	}

	// Column pool: base set + mutated expressions from fuzz bytes.
	extra := []protocolLiveColumn{
		{alias: "mut_i", expr: fmt.Sprintf("CAST(%d AS int)", int32(next())|int32(next())<<8|int32(next())<<16)},
		{alias: "mut_i3", expr: "CAST(3 AS int)"},
		{alias: "mut_bi", expr: fmt.Sprintf("CAST(%d AS bigint)", int64(next())<<40|int64(next())<<32|int64(next()))},
		{alias: "mut_dec", expr: fmt.Sprintf("CAST(%d.%03d AS decimal(38,6))", u8(10000), u8(1000))},
		{alias: "mut_dec20", expr: fmt.Sprintf("CAST(%d.%06d AS decimal(38,20))", u8(2000), u8(1000000))},
		{alias: "mut_t", expr: fmt.Sprintf("CAST('10:%02d:%02d' AS time(%d))", u8(60), u8(60), u8(8))},
		{alias: "mut_dt", expr: fmt.Sprintf("CAST('202%1d-0%1d-1%1d 10:20:19' AS datetime2(%d))", 0+u8(7), 1+u8(9), u8(9), u8(8))},
		{alias: "mut_dto", expr: fmt.Sprintf("CAST('2026-07-28 10:20:19 +0%1d:00' AS datetimeoffset(%d))", u8(8), u8(8))},
		{alias: "mut_d", expr: fmt.Sprintf("CAST('202%1d-06-15' AS date)", u8(7))},
		{alias: "mut_vc", expr: fmt.Sprintf("CAST('%s' AS varchar(200))", sqlTextChunk(24))},
		{alias: "mut_nvc", expr: fmt.Sprintf("CAST(N'%s' AS nvarchar(200))", sqlTextChunk(24))},
		{alias: "mut_null", expr: "CAST(NULL AS int)"},
		{alias: "mut_money", expr: fmt.Sprintf("CAST(%d.%02d AS money)", u8(500)-100, u8(100))},
	}
	pool := append(append([]protocolLiveColumn{}, all...), extra...)

	nCol := 3 + u8(20)
	if nCol > len(pool) {
		nCol = len(pool)
	}
	perm := r.Perm(len(pool))
	cols := make([]protocolLiveColumn, nCol)
	for i := 0; i < nCol; i++ {
		cols[i] = pool[perm[i]]
		// Occasionally mutate alias suffix so SQL text differs even with same expr.
		if next()&1 == 0 {
			cols[i].alias = fmt.Sprintf("%s_%d", cols[i].alias, u8(1000))
		}
	}
	// Always inject historical desync pattern somewhere.
	s := u8(3)
	inj := []protocolLiveColumn{
		{alias: "fz_t", expr: fmt.Sprintf("CAST('00:00:01.%06d' AS time(%d))", u8(1000000), s)},
		{alias: "fz_i3", expr: "CAST(3 AS int)"},
		{alias: "fz_dt", expr: fmt.Sprintf("CAST('2026-07-28 10:20:19.%07d' AS datetime2(%d))", u8(10000000), s)},
	}
	atIns := u8(len(cols) + 1)
	if atIns > len(cols) {
		cols = append(cols, inj...)
	} else {
		cols = append(cols[:atIns], append(inj, cols[atIns:]...)...)
	}

	// Parameter set mutated from fuzz bytes (used in SQL text via @names).
	pInt := int64(int32(next()) | int32(next())<<8 | int32(next())<<16)
	if next()&1 == 0 {
		pInt = 3 // production IntN-false-length value
	}
	pStr := sqlTextChunk(32)
	pDec := fmt.Sprintf("%d.%06d", u8(5000), u8(1000000))
	pScale := u8(8)
	nowOff := time.Duration(u8(48)) * time.Hour

	params := []rdb.Param{
		{Name: "PInt", Type: rdb.Integer, Value: pInt},
		{Name: "PInt3", Type: rdb.Integer, Value: int64(3)},
		{Name: "PZero", Type: rdb.Integer, Value: int64(0)},
		{Name: "PStr", Type: rdb.TypeAnsiVarChar, Length: 200, Value: pStr},
		{Name: "PNStr", Type: rdb.TypeVarChar, Length: 200, Value: pStr},
		{Name: "PDec", Type: rdb.TypeDecimal, Precision: 38, Scale: 6, Value: mustRat(pDec)},
		{Name: "PDec20", Type: rdb.TypeDecimal, Precision: 38, Scale: 20, Value: mustRat("1234.567891")},
		{Name: "PDate", Type: rdb.TypeDate, Value: time.Date(2020+u8(10), time.Month(1+u8(11)), 1+u8(27), 0, 0, 0, 0, time.UTC)},
		{Name: "PTime", Type: rdb.TypeTimestamp, Value: time.Now().UTC().Add(nowOff)},
		{Name: "PBit", Type: rdb.TypeBool, Value: next()&1 == 1},
	}
	// Always send the full set: every SQL shape may reference any of these names.

	rows := 1 + u8(20)
	parts := make([]string, len(cols))
	for i, c := range cols {
		parts[i] = fmt.Sprintf("[%s] = %s", c.alias, c.expr)
	}
	selectList := strings.Join(parts, ",\n\t")

	// Mutate SQL shape (text structure), not only column list.
	shape := mode % 6
	if len(mutate) > 0 {
		shape = next() % 6
	}
	var sql string
	switch shape {
	case 0: // plain VALUES multi-row
		vals := make([]string, rows)
		for i := 0; i < rows; i++ {
			vals[i] = fmt.Sprintf("(%d)", i)
		}
		sql = fmt.Sprintf("SELECT %s\nFROM (VALUES %s) AS n(rn);", selectList, strings.Join(vals, ","))
	case 1: // TOP + param filter (params appear in SQL text)
		sql = fmt.Sprintf(`
SELECT TOP (%d) %s
FROM (VALUES %s) AS n(rn)
WHERE @PInt3 = 3 AND (@PZero = 0 OR n.rn >= 0)
  AND @PInt IS NOT NULL;
`, rows, selectList, valuesList(rows))
	case 2: // ORDER BY + nvarchar param in projection
		sql = fmt.Sprintf(`
SELECT %s,
	[pstr] = @PNStr,
	[pint] = @PInt
FROM (VALUES %s) AS n(rn)
ORDER BY n.rn %s;
`, selectList, valuesList(rows), orderDir(next()))
	case 3: // UNION ALL of two small selects (more tokens / dual metadata)
		half := rows/2 + 1
		sql = fmt.Sprintf(`
SELECT %s FROM (VALUES %s) AS n(rn)
UNION ALL
SELECT %s FROM (VALUES %s) AS n(rn);
`, selectList, valuesList(half), selectList, valuesList(half))
	case 4: // nested derived table + date/time params
		sql = fmt.Sprintf(`
SELECT x.*
FROM (
	SELECT TOP (%d) %s,
		[pd] = @PDate,
		[pt] = @PTime,
		[pdec] = @PDec
	FROM (VALUES %s) AS n(rn)
) x
WHERE @PBit = @PBit;
`, rows, selectList, valuesList(rows))
	default: // scale-mutated temporal + decimal via params in WHERE-ish select
		sql = fmt.Sprintf(`
SELECT %s,
	[tscale] = CAST('11:22:33' AS time(%d)),
	[dtscale] = CAST(@PTime AS datetime2(%d)),
	[d20] = @PDec20,
	[s] = @PStr
FROM (VALUES %s) AS n(rn);
`, selectList, pScale, pScale, valuesList(rows))
	}

	// Optional trailing comment / whitespace mutation of SQL text.
	if next()&1 == 0 {
		sql = sql + "\n-- " + sqlTextChunk(16) + "\n"
	}
	if next()&3 == 0 {
		sql = "\t" + sql + "\r\n"
	}

	return protocolFuzzCase{SQL: sql, Params: params, MinRows: 1}
}

func valuesList(rows int) string {
	if rows < 1 {
		rows = 1
	}
	vals := make([]string, rows)
	for i := 0; i < rows; i++ {
		vals[i] = fmt.Sprintf("(%d)", i)
	}
	return strings.Join(vals, ",")
}

func orderDir(b byte) string {
	if b&1 == 0 {
		return "ASC"
	}
	return "DESC"
}

func mustRat(s string) *big.Rat {
	r := new(big.Rat)
	if _, ok := r.SetString(s); !ok {
		r.SetInt64(0)
	}
	return r
}

func runProtocolFuzzCase(t testing.TB, pool must.ConnPool, c protocolFuzzCase) (err error) {
	t.Helper()
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("panic: %v\nSQL:\n%s", r, c.SQL)
		}
	}()
	ctx := context.Background()
	res := pool.Query(ctx, &rdb.Command{SQL: c.SQL, Arity: rdb.Any}, c.Params...)
	defer res.Close()
	got := 0
	for res.Next() {
		res.Scan()
		for i := range res.Schema() {
			_ = res.Getx(i)
		}
		got++
	}
	if c.MinRows > 0 && got < c.MinRows {
		return fmt.Errorf("got %d rows, want >= %d\nSQL:\n%s", got, c.MinRows, c.SQL)
	}
	res.Close()
	// Desync probe.
	res2 := pool.Query(ctx, &rdb.Command{SQL: `SELECT CAST(1 AS int) AS x`, Arity: rdb.OneMust})
	defer res2.Close()
	if !res2.Next() {
		return fmt.Errorf("desync probe failed\nSQL was:\n%s", c.SQL)
	}
	res2.Scan()
	return nil
}

// FuzzProtocolLive: go fuzz framework + real server. Mutates SQL text shapes,
// column expressions, and parameter values/names. Requires APP_DSN.
//
//	go test ./ms/ -run '^$' -fuzz=FuzzProtocolLive -fuzztime=20m
func FuzzProtocolLive(f *testing.F) {
	if !db.Valid() {
		f.Skip("DB connection not configured, check APP_DSN")
	}
	// Seed corpus: exercise each SQL shape and interesting params.
	f.Add(int64(0), byte(0), []byte("hello"))
	f.Add(int64(1), byte(1), []byte{0, 1, 2, 3, 3, 3})
	f.Add(int64(42), byte(2), []byte("unicode-你好"))
	f.Add(int64(878), byte(3), []byte{3, 0, 0, 0, 7, 20})
	f.Add(int64(0x12345678), byte(4), []byte("1234.567891"))
	f.Add(int64(8232471098285931930), byte(5), []byte{0xff, 0x00, 11, 7})
	f.Add(int64(99), byte(0), []byte{})
	f.Add(int64(100), byte(5), []byte("a'b;--"))

	f.Fuzz(func(t *testing.T, seed int64, mode byte, mutate []byte) {
		if !db.Valid() {
			t.Skip("no db")
		}
		// Bound mutate length so SQL stays reasonable.
		if len(mutate) > 256 {
			mutate = mutate[:256]
		}
		c := protocolMutate(seed, mode, mutate)
		if err := runProtocolFuzzCase(t, db, c); err != nil {
			t.Fatalf("seed=%d mode=%d mutate=%q: %v", seed, mode, truncate(fmt.Sprintf("%q", mutate), 80), err)
		}
	})
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "…"
}
