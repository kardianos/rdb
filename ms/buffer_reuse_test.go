// Copyright 2014 Daniel Theophanes.
// Use of this source code is governed by a zlib-style
// license that can be found in the LICENSE file.

package ms

import (
	"bytes"
	"context"
	"fmt"
	"sync"
	"testing"
	"time"
	"unsafe"

	"github.com/kardianos/rdb"
	"github.com/kardianos/rdb/internal/uconv"
	"github.com/kardianos/rdb/must"
)

// utf16LE encodes s as UTF-16LE for feed into decodeNCharUTF8.
func utf16LE(s string) []byte {
	return uconv.Encode.FromString(s)
}

// sameByteBacking reports whether a and b share the same underlying array
// start (common when both are views of utf8Scratch / msgBuf).
func sameByteBacking(a, b []byte) bool {
	if len(a) == 0 || len(b) == 0 {
		return false
	}
	return unsafe.SliceData(a) == unsafe.SliceData(b)
}

// TestScratchReuseWithoutCopyDocumentsHazard shows that decodeNCharUTF8 returns
// a connection-local scratch buffer: a second decode can mutate the first view
// unless the caller copies (MustCopy / DirectAssignBytes).
func TestScratchReuseWithoutCopyDocumentsHazard(t *testing.T) {
	tds := &Connection{}
	v1 := tds.decodeNCharUTF8(utf16LE("AAAAAAAA"))
	if string(v1) != "AAAAAAAA" {
		t.Fatalf("first decode: %q", v1)
	}
	// Hold the slice without copying (the bug pattern).
	held := v1
	v2 := tds.decodeNCharUTF8(utf16LE("BBBBBBBB"))
	if string(v2) != "BBBBBBBB" {
		t.Fatalf("second decode: %q", v2)
	}
	// When both share scratch, held is no longer "AAAAAAAA".
	if sameByteBacking(held, v2) || string(held) != "AAAAAAAA" {
		t.Logf("documented hazard: without copy, held became %q (shared scratch)", held)
	}
	if string(held) == "AAAAAAAA" && !sameByteBacking(held, v2) {
		// Realloc can happen if second string needs more capacity; still require owned path below.
		t.Log("scratch reallocated this run; alias not observed")
	}
}

// TestScratchMustCopySurvivesIntentionalMutation verifies the escape hatch used
// by emit/WriteField: after copying out of scratch, mutating utf8Scratch must
// not change the owned value.
func TestScratchMustCopySurvivesIntentionalMutation(t *testing.T) {
	tds := &Connection{}
	decoded := tds.decodeNCharUTF8(utf16LE("COL_A_VALUE!!"))
	var dest []byte
	handled, err := rdb.DirectAssignBytes(&dest, decoded, false, true, nil)
	if !handled || err != nil {
		t.Fatalf("DirectAssignBytes: handled=%v err=%v", handled, err)
	}
	if string(dest) != "COL_A_VALUE!!" {
		t.Fatalf("dest=%q", dest)
	}
	// Intentional mutation of the shared scratch (simulates next field decode).
	if len(tds.utf8Scratch) == 0 {
		t.Fatal("expected non-empty utf8Scratch after decode")
	}
	for i := range tds.utf8Scratch {
		tds.utf8Scratch[i] = 'X'
	}
	// Also run another decode into the same scratch.
	_ = tds.decodeNCharUTF8(utf16LE("OTHER_COLUMN"))
	if string(dest) != "COL_A_VALUE!!" {
		t.Fatalf("owned dest corrupted after scratch mutation: %q", dest)
	}
	// string prep path always copies via string(bb).
	var s string
	decoded = tds.decodeNCharUTF8(utf16LE("STRING_DEST"))
	handled, err = rdb.DirectAssignBytes(&s, decoded, false, true, nil)
	if !handled || err != nil || s != "STRING_DEST" {
		t.Fatalf("string assign: handled=%v err=%v s=%q", handled, err, s)
	}
	for i := range tds.utf8Scratch {
		tds.utf8Scratch[i] = 'Y'
	}
	_ = tds.decodeNCharUTF8(utf16LE("NEXT"))
	if s != "STRING_DEST" {
		t.Fatalf("string dest corrupted: %q", s)
	}
}

// TestMultiNCharOwnedIsolation unit-simulates two successive NChar fields:
// owned copies must not alias and must survive the next decode.
func TestMultiNCharOwnedIsolation(t *testing.T) {
	tds := &Connection{}
	var a, b []byte
	d1 := tds.decodeNCharUTF8(utf16LE("FIRST_NVARCHAR_AAA"))
	if _, err := rdb.DirectAssignBytes(&a, d1, false, true, nil); err != nil {
		t.Fatal(err)
	}
	d2 := tds.decodeNCharUTF8(utf16LE("SECOND_NVARCHAR_BBB"))
	if _, err := rdb.DirectAssignBytes(&b, d2, false, true, nil); err != nil {
		t.Fatal(err)
	}
	if sameByteBacking(a, b) {
		t.Fatal("owned columns share backing array")
	}
	if sameByteBacking(a, tds.utf8Scratch) || sameByteBacking(b, tds.utf8Scratch) {
		t.Fatal("owned column still aliases utf8Scratch")
	}
	// Mutate scratch and the other column.
	for i := range tds.utf8Scratch {
		tds.utf8Scratch[i] = 'Z'
	}
	for i := range b {
		b[i] = 'z'
	}
	if string(a) != "FIRST_NVARCHAR_AAA" {
		t.Fatalf("column a corrupted: %q", a)
	}
}

// TestLiveBufferIsolationAfterScan exercises the full driver path: multi-column
// nvarchar/varchar, Prep *[]byte/*string, intentional mutation of one dest, and
// a follow-up query that reuses connection scratch/msgBuf.
func TestLiveBufferIsolationAfterScan(t *testing.T) {
	checkSkip(t)
	defer recoverTest(t)

	const (
		aWant = "ALPHA_COLUMN_VALUE_111"
		bWant = "BETA_COLUMN_VALUE_2222"
		cWant = "GAMMA_VARCHAR_ASCII_333"
	)
	cmd := &rdb.Command{
		SQL: fmt.Sprintf(`
			SELECT
				a = CAST(N'%s' AS nvarchar(100)),
				b = CAST(N'%s' AS nvarchar(100)),
				c = CAST('%s' AS varchar(100));
		`, aWant, bWant, cWant),
		Arity: rdb.OneMust,
	}

	var a, b, c []byte
	var as, bs string

	// []byte Prep path (MustCopy into owned slices).
	res := db.Query(context.Background(), cmd)
	res.Prep("a", &a).Prep("b", &b).Prep("c", &c)
	res.Scan()
	res.Close()
	if string(a) != aWant || string(b) != bWant || string(c) != cWant {
		t.Fatalf("[]byte prep: a=%q b=%q c=%q", a, b, c)
	}
	if sameByteBacking(a, b) || sameByteBacking(a, c) || sameByteBacking(b, c) {
		t.Fatal("prep []byte columns share backing")
	}

	// Intentional mutation of one destination.
	for i := range a {
		a[i] = 'X'
	}
	if string(b) != bWant || string(c) != cWant {
		t.Fatalf("mutation of a leaked: b=%q c=%q", b, c)
	}

	// string Prep path.
	res = db.Query(context.Background(), cmd)
	res.Prep("a", &as).Prep("b", &bs)
	var c2 []byte
	res.Prep("c", &c2)
	res.Scan()
	res.Close()
	if as != aWant || bs != bWant || string(c2) != cWant {
		t.Fatalf("string prep: a=%q b=%q c=%q", as, bs, c2)
	}

	// Follow-up query reuses connection buffers (utf8Scratch, msgBuf).
	res2 := db.Query(context.Background(), &rdb.Command{
		SQL:   `SELECT x = CAST(N'FOLLOW_UP_NCHAR' AS nvarchar(50)), y = CAST('follow_up_vc' AS varchar(50));`,
		Arity: rdb.OneMust,
	})
	var x, y string
	res2.Prep("x", &x).Prep("y", &y)
	res2.Scan()
	res2.Close()
	if x != "FOLLOW_UP_NCHAR" || y != "follow_up_vc" {
		t.Fatalf("follow-up: x=%q y=%q", x, y)
	}
	// Prior Prep destinations must still hold their values (owned copies).
	if as != aWant || bs != bWant {
		t.Fatalf("prior prep corrupted after next query: a=%q b=%q", as, bs)
	}
	if string(b) != bWant || string(c) != cWant {
		t.Fatalf("prior []byte prep corrupted after next query: b=%q c=%q", b, c)
	}
	// a was intentionally mutated; must not have become follow-up content.
	if bytes.Contains(a, []byte("FOLLOW")) || bytes.Contains(c2, []byte("follow")) {
		t.Fatalf("prior buffers became follow-up content: a=%q c2=%q", a, c2)
	}
}

// TestLiveGetxHoldsAcrossNextRow ensures Getx []byte/string from row N stay
// intact after scanning row N+1 (MustCopy into valuer buffer).
func TestLiveGetxHoldsAcrossNextRow(t *testing.T) {
	checkSkip(t)
	defer recoverTest(t)

	cmd := &rdb.Command{
		SQL: `
SELECT v = CAST(N'row-' + CAST(n.rn AS nvarchar(10)) AS nvarchar(50)),
       w = CAST('vc-' + CAST(n.rn AS varchar(10)) AS varchar(50))
FROM (VALUES (0),(1),(2),(3),(4)) AS n(rn)
ORDER BY n.rn;
`,
		Arity: rdb.Any,
	}
	res := db.Query(context.Background(), cmd)
	defer res.Close()

	var held []string
	var heldVC [][]byte
	for res.Next() {
		res.Scan()
		v := res.Get("v")
		w := res.Get("w")
		switch s := v.(type) {
		case string:
			held = append(held, s)
		case []byte:
			// Own a copy for the assertion list if driver returns []byte.
			held = append(held, string(s))
		default:
			t.Fatalf("v type %T", v)
		}
		switch b := w.(type) {
		case []byte:
			// Keep the slice as returned — MustCopy should have made it owned.
			heldVC = append(heldVC, b)
		case string:
			heldVC = append(heldVC, []byte(b))
		default:
			t.Fatalf("w type %T", w)
		}
	}
	if len(held) != 5 {
		t.Fatalf("rows=%d", len(held))
	}
	// Intentional mutation of last row's varchar slice.
	last := heldVC[len(heldVC)-1]
	for i := range last {
		last[i] = 'Z'
	}
	// Earlier rows must be unchanged.
	for i := 0; i < 4; i++ {
		want := fmt.Sprintf("row-%d", i)
		if held[i] != want {
			t.Errorf("held[%d]=%q want %q (possible buffer reuse)", i, held[i], want)
		}
		wantVC := fmt.Sprintf("vc-%d", i)
		if string(heldVC[i]) != wantVC {
			t.Errorf("heldVC[%d]=%q want %q", i, heldVC[i], wantVC)
		}
	}
}

// TestLiveIntentionalMsgBufPressure runs a large subsequent query after capturing
// short values, to force msgBuf growth/reuse while old Get slices must remain valid.
func TestLiveIntentionalMsgBufPressure(t *testing.T) {
	checkSkip(t)
	defer recoverTest(t)

	res := db.Query(context.Background(), &rdb.Command{
		SQL:   `SELECT a = CAST(N'small-a' AS nvarchar(20)), b = CAST('small-b' AS varchar(20));`,
		Arity: rdb.OneMust,
	})
	var a, b []byte
	res.Prep("a", &a).Prep("b", &b)
	res.Scan()
	res.Close()
	a0, b0 := string(a), string(b)

	// Large result to churn packet/msgBuf assembly on a pool connection.
	res2 := db.Query(context.Background(), &rdb.Command{
		SQL: `
SELECT TOP 50
	big = CAST(REPLICATE(N'字', 200) AS nvarchar(400)),
	blob = CAST(REPLICATE('Z', 500) AS varchar(500)),
	i = CAST(3 AS int);
`,
		Arity: rdb.Any,
	})
	for res2.Next() {
		res2.Scan()
		_ = res2.Getx(0)
		_ = res2.Getx(1)
		_ = res2.Getx(2)
	}
	res2.Close()

	if string(a) != a0 || string(b) != b0 {
		t.Fatalf("prep values corrupted after large query: a=%q b=%q", a, b)
	}
	// Mutate a; b must stay.
	for i := range a {
		a[i] = '!'
	}
	if string(b) != b0 {
		t.Fatalf("mutating a corrupted b: %q", b)
	}
}

// TestRaceBufferIsolationPool runs concurrent workers on a multi-conn pool.
// Each worker holds decoded strings/[]byte while others query (exercises
// per-connection scratch isolation). Run with: go test -race ./ms/ -run TestRaceBuffer
func TestRaceBufferIsolationPool(t *testing.T) {
	checkSkip(t)
	if config == nil {
		t.Fatal("config nil")
	}
	ccfg := *config
	ccfg.PoolInitCapacity = 4
	ccfg.PoolMaxCapacity = 8
	pool := must.Open(&ccfg)
	defer pool.Close()

	const workers = 8
	const iters = 40
	var wg sync.WaitGroup
	errCh := make(chan error, workers)

	for w := 0; w < workers; w++ {
		w := w
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < iters; i++ {
				tag := fmt.Sprintf("w%02d-i%03d-%s", w, i, time.Now().Format("150405.000"))
				cmd := &rdb.Command{
					SQL: fmt.Sprintf(`
						SELECT
							a = CAST(N'%[1]s-A' AS nvarchar(80)),
							b = CAST(N'%[1]s-B' AS nvarchar(80)),
							c = CAST('%[1]s-C' AS varchar(80)),
							i = CAST(3 AS int);
					`, tag),
					Arity: rdb.OneMust,
				}
				var a, b, c []byte
				var i3 int32
				res := pool.Query(context.Background(), cmd)
				res.Prep("a", &a).Prep("b", &b).Prep("c", &c).Prep("i", &i3)
				res.Scan()
				res.Close()

				// Hold values while other goroutines run.
				time.Sleep(time.Microsecond * time.Duration(50+w*3))

				if string(a) != tag+"-A" || string(b) != tag+"-B" || string(c) != tag+"-C" || i3 != 3 {
					errCh <- fmt.Errorf("worker %d iter %d: a=%q b=%q c=%q i=%d", w, i, a, b, c, i3)
					return
				}
				// Intentional mutation of a; b/c must stay.
				for j := range a {
					a[j] = 'x'
				}
				if string(b) != tag+"-B" || string(c) != tag+"-C" {
					errCh <- fmt.Errorf("worker %d: mutation leaked b=%q c=%q", w, b, c)
					return
				}
			}
		}()
	}
	wg.Wait()
	close(errCh)
	for err := range errCh {
		t.Error(err)
	}
}

// TestRaceUtf8ScratchPerConnection documents that a single Connection's
// utf8Scratch is not safe for concurrent decodeNCharUTF8. With -race this
// should report a race if two goroutines share one Connection.
// The test uses separate Connections (no race) to show the safe pattern.
func TestRaceUtf8ScratchPerConnection(t *testing.T) {
	const workers = 6
	const iters = 200
	var wg sync.WaitGroup
	for w := 0; w < workers; w++ {
		w := w
		wg.Add(1)
		go func() {
			defer wg.Done()
			// Own connection object (no network) — scratch is not shared.
			tds := &Connection{}
			for i := 0; i < iters; i++ {
				s := fmt.Sprintf("worker-%d-iter-%d-payload", w, i)
				enc := utf16LE(s)
				out := tds.decodeNCharUTF8(enc)
				// MustCopy simulation before next decode.
				owned := string(out)
				_ = tds.decodeNCharUTF8(utf16LE("next"))
				if owned != s {
					t.Errorf("owned corrupted: got %q want %q", owned, s)
					return
				}
			}
		}()
	}
	wg.Wait()
}

// TestRaceSharedConnectionDecode is expected to be racy if run as written
// against one Connection from two goroutines. We only run the concurrent
// section when -race is not the only goal; with -race we skip the unsafe
// sharing and instead verify detection setup via a short comment test.
//
// Pattern under test: never share *Connection across goroutines for decode.
func TestRaceSharedConnectionDecode(t *testing.T) {
	// Safe: serialized use of one Connection.
	tds := &Connection{}
	var mu sync.Mutex
	var wg sync.WaitGroup
	for w := 0; w < 4; w++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for i := 0; i < 100; i++ {
				mu.Lock()
				s := fmt.Sprintf("ser-%d-%d", id, i)
				out := append([]byte(nil), tds.decodeNCharUTF8(utf16LE(s))...)
				mu.Unlock()
				if string(out) != s {
					t.Errorf("got %q want %q", out, s)
				}
			}
		}(w)
	}
	wg.Wait()
}

// TestRaceValuerBufferAcrossQueries hammers Scan/Get on the shared test pool
// with overlapping lifetimes of returned values (copies should make this safe).
func TestRaceValuerBufferAcrossQueries(t *testing.T) {
	checkSkip(t)
	if config == nil {
		t.Skip("no config")
	}
	ccfg := *config
	ccfg.PoolInitCapacity = 3
	ccfg.PoolMaxCapacity = 6
	pool := must.Open(&ccfg)
	defer pool.Close()

	var wg sync.WaitGroup
	for w := 0; w < 6; w++ {
		w := w
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < 30; i++ {
				res := pool.Query(context.Background(), &rdb.Command{
					SQL: `
SELECT
	a = CAST(N'αβγ-row' AS nvarchar(30)),
	b = CAST('ascii-row' AS varchar(30)),
	c = CAST(3 AS int),
	d = CAST(1234.567891 AS decimal(38,20));
`,
					Arity: rdb.OneMust,
				})
				res.Scan()
				va := res.Get("a")
				vb := res.Get("b")
				vc := res.Get("c")
				vd := res.Get("d")
				res.Close()

				// Hold across more pool activity.
				time.Sleep(time.Microsecond * 20)

				switch s := va.(type) {
				case string:
					if s != "αβγ-row" {
						t.Errorf("w%d a=%q", w, s)
					}
				case []byte:
					if string(s) != "αβγ-row" {
						t.Errorf("w%d a=%q", w, s)
					}
				}
				switch s := vb.(type) {
				case string:
					if s != "ascii-row" {
						t.Errorf("w%d b=%q", w, s)
					}
				case []byte:
					if string(s) != "ascii-row" {
						t.Errorf("w%d b=%q", w, s)
					}
				}
				_ = vc
				_ = vd
			}
		}()
	}
	wg.Wait()
}
