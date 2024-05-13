// Copyright 2014 Daniel Theophanes.
// Use of this source code is governed by a zlib-style
// license that can be found in the LICENSE file.

package ms

import (
	"context"
	"encoding/hex"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/kardianos/rdb"
)

func TestTimeoutDie(t *testing.T) {
	checkSkip(t)
	if testing.Short() {
		t.Skip()
	}
	if parallel {
		t.Parallel()
	}
	// Handle multiple result sets.
	defer recoverTest(t)

	start := time.Now()

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	res, err := db.Normal().Query(ctx, &rdb.Command{
		SQL: `
-- TestTimeoutDie
waitfor delay '00:00:02';
select 1 as 'ID';
`,
		Arity: rdb.Any,
	})
	defer assertFreeConns(t)
	defer res.Close()

	dur := time.Now().Sub(start)
	t.Log("duration", dur)
	t.Log("error", err)

	if err == nil {
		t.Errorf("Failed to timeout: %v", err)
	}
}

func TestTimeoutLive(t *testing.T) {
	checkSkip(t)
	if testing.Short() {
		t.Skip()
	}
	if parallel {
		t.Parallel()
	}
	// Handle multiple result sets.
	defer recoverTest(t)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	res, err := db.Normal().Query(ctx, &rdb.Command{
		SQL: `
-- TestTimeoutLive
waitfor delay '00:00:01';
select 1 as 'ID';
`,
		Arity: rdb.Any,
	})

	defer assertFreeConns(t)
	defer res.Close()

	if err != nil {
		t.Errorf("Error with query: %v", err)
	}
}

func TestError(t *testing.T) {
	checkSkip(t)
	if parallel {
		t.Parallel()
	}
	// Handle multiple result sets.
	defer recoverTest(t)

	// Packet size divided by 2 (utf-16) minus packet header.
	var longText = strings.Repeat("Hello everyone in the world.\n", 100)[:4096/2-10] // -21

	res1, err := db.Normal().Query(context.Background(), &rdb.Command{
		SQL: `
			select top 0 ID = 0;
		`,
		Arity: rdb.Any,
	}, rdb.Param{Name: "Text", Type: rdb.Text, Value: longText})
	res1.Close()
	assertFreeConns(t)

	if err != nil {
		t.Fatalf("Error with query: %v", err)
	}

	_, err = db.Normal().Query(context.Background(), &rdb.Command{
		SQL: `
			fooBad top 0 ID = 0;
		`,
		Arity: rdb.Any,
	}, rdb.Param{Name: "Text", Type: rdb.Text, Value: longText})
	// res2.Close()
	assertFreeConns(t)

	if err == nil {
		t.Fatalf("Expected error (res2).")
	}

	res3, err := db.Normal().Query(context.Background(), &rdb.Command{
		SQL: `
			select top 1 TX = @Text;
		`,
		Arity: rdb.Any,
	}, rdb.Param{Name: "Text", Type: rdb.Text, Value: longText})
	if err != nil {
		t.Fatalf("Error with query3: %v", err)
	}

	err = res3.Scan()
	if err != nil {
		t.Fatalf("Error doing scan: %v", err)
	}
	tx := string(res3.Get("TX").([]byte))

	res3.Close()
	assertFreeConns(t)

	if tx != longText {
		txDump := hex.Dump([]byte(tx))
		t.Fatalf("Text does not match:\n\n%s\n\n", txDump)
	}

	if err != nil {
		t.Fatalf("Error with query: %v", err)
	}
}

func TestMismatchTypeError(t *testing.T) {
	checkSkip(t)
	if parallel {
		t.Parallel()
	}
	// Handle multiple result sets.
	defer recoverTest(t)

	timeout(t, time.Second*2, func() {
		res1, err := db.Normal().Query(context.Background(), &rdb.Command{
			SQL: `
			select MyString = @MyString;
		`,
			Arity: rdb.Any,
		}, rdb.Param{Name: "Text", Type: rdb.TypeDate, Value: "my text"})
		res1.Close()
		assertFreeConns(t)

		if err == nil {
			t.Errorf("Error missing from query: %v", err)
		}
	})
}

func timeout(t *testing.T, d time.Duration, f func()) {
	done := make(chan struct{})
	tm := time.NewTimer(d)
	go func() {
		f()
		tm.Stop()
		close(done)
	}()
	select {
	case <-tm.C:
		t.Errorf("Query out after %v.", d)
	case <-done:
	}
}

func TestConnectionPoolExhaustion(t *testing.T) {
	checkSkip(t)
	if testing.Short() {
		t.Skip()
	}

	// Handle multiple result sets.
	defer recoverTest(t)

	wait := &sync.WaitGroup{}

	for i := 0; i < 10; i++ {
		wait.Add(1)
		go func() {
			defer wait.Done()

			res, err := db.Normal().Query(context.Background(), &rdb.Command{
				SQL: `
-- TestConnectionPoolExhaustion
waitfor delay '00:00:01';
select ID = 1;
`,
				Arity: rdb.Any,
			})
			if err != nil {
				t.Errorf("Failed to wait for next connection: %v", err)
			}
			res.Close()
		}()
	}
	wc := make(chan struct{}, 3)
	timeoutDur := time.Second * 35
	timeout := time.After(timeoutDur)
	go func() {
		wait.Wait()
		wc <- struct{}{}
	}()
	select {
	case <-wc:
	case <-timeout:
		t.Fatalf("Timeout after %v", timeoutDur)
	}
	assertFreeConns(t)
}

func TestThrowError(t *testing.T) {
	checkSkip(t)
	if testing.Short() {
		t.Skip()
	}

	// Handle multiple result sets.
	defer recoverTest(t)

	res, err := db.Normal().Query(context.Background(), &rdb.Command{
		SQL:   `RAISERROR(N'throw an error', 16, 1);`,
		Arity: rdb.Any,
	})
	if err == nil {
		t.Errorf("Failed to get error")
	} else {
		t.Log(err.Error())
	}
	res.Close()

	assertFreeConns(t)
}
