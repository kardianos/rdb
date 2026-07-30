// Copyright 2014 Daniel Theophanes.
// Use of this source code is governed by a zlib-style
// license that can be found in the LICENSE file.

package table

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/kardianos/rdb"
)

func TestDoNilFn(t *testing.T) {
	err := Do(context.Background(), fakeQueryer{}, &rdb.Command{SQL: "x"}, nil)
	if err == nil {
		t.Fatal("expected error for nil fn")
	}
}

func TestDoPropagatesQueryError(t *testing.T) {
	want := errors.New("query boom")
	err := Do(context.Background(), fakeQueryer{err: want}, &rdb.Command{SQL: "x"}, func(h Handle) error {
		t.Fatal("fn should not run")
		return nil
	})
	if !errors.Is(err, want) {
		t.Fatalf("err=%v want %v", err, want)
	}
}

func TestSliceNilHandle(t *testing.T) {
	_, err := Slice[struct {
		A int `db:"a"`
	}](Handle{})
	if !errors.Is(err, ErrNilHandle) {
		t.Fatalf("err=%v want %v", err, ErrNilHandle)
	}
}

func TestSeqNilHandle(t *testing.T) {
	var got error
	for _, err := range Seq[struct {
		A int `db:"a"`
	}](Handle{}) {
		got = err
		break
	}
	if !errors.Is(got, ErrNilHandle) {
		t.Fatalf("err=%v want %v", got, ErrNilHandle)
	}
}

func TestNextNilHandle(t *testing.T) {
	if err := (Handle{}).Next(); !errors.Is(err, ErrNilHandle) {
		t.Fatalf("err=%v want %v", err, ErrNilHandle)
	}
}

func TestHandleMultiSetSliceThenSeq(t *testing.T) {
	pool := openScriptedPool(t, []scriptedSet{
		{
			cols: []*rdb.Column{
				{Name: "id", Index: 0},
				{Name: "name", Index: 1},
			},
			rows: [][]any{
				{int32(1), "alice"},
				{int32(2), "bob"},
			},
		},
		{
			cols: []*rdb.Column{
				{Name: "sku", Index: 0},
				{Name: "qty", Index: 1},
			},
			rows: [][]any{
				{"A", int32(10)},
				{"B", int32(20)},
				{"C", int32(30)},
			},
		},
	})
	defer pool.Close()

	type user struct {
		ID   int32  `db:"id"`
		Name string `db:"name"`
	}
	type order struct {
		SKU string `db:"sku"`
		Qty int32  `db:"qty"`
	}

	var users []user
	var orders []order
	err := Do(context.Background(), pool, &rdb.Command{SQL: "multi", Arity: rdb.Any}, func(h Handle) error {
		var err error
		users, err = Slice[user](h)
		if err != nil {
			return err
		}
		if err := h.Next(); err != nil {
			return err
		}
		for o, err := range Seq[order](h) {
			if err != nil {
				return err
			}
			orders = append(orders, o)
		}
		if err := h.Next(); !errors.Is(err, ErrNoMoreResults) {
			t.Fatalf("third Next: err=%v want ErrNoMoreResults", err)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}

	if len(users) != 2 || users[0] != (user{1, "alice"}) || users[1] != (user{2, "bob"}) {
		t.Fatalf("users=%v", users)
	}
	if len(orders) != 3 || orders[0] != (order{"A", 10}) || orders[2] != (order{"C", 30}) {
		t.Fatalf("orders=%v", orders)
	}
}

func TestHandleSeqPartialThenNext(t *testing.T) {
	pool := openScriptedPool(t, []scriptedSet{
		{
			cols: []*rdb.Column{{Name: "n", Index: 0}},
			rows: [][]any{{int32(1)}, {int32(2)}, {int32(3)}},
		},
		{
			cols: []*rdb.Column{{Name: "s", Index: 0}},
			rows: [][]any{{"done"}},
		},
	})
	defer pool.Close()

	type rowN struct {
		N int32 `db:"n"`
	}
	type rowS struct {
		S string `db:"s"`
	}

	var first []int32
	var second []string
	err := Do(context.Background(), pool, &rdb.Command{SQL: "partial", Arity: rdb.Any}, func(h Handle) error {
		for r, err := range Seq[rowN](h) {
			if err != nil {
				return err
			}
			first = append(first, r.N)
			if r.N == 1 {
				break // leave rows 2 and 3; Seq must Drain
			}
		}
		if err := h.Next(); err != nil {
			return err
		}
		for r, err := range Seq[rowS](h) {
			if err != nil {
				return err
			}
			second = append(second, r.S)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(first) != 1 || first[0] != 1 {
		t.Fatalf("first=%v want [1]", first)
	}
	if len(second) != 1 || second[0] != "done" {
		t.Fatalf("second=%v want [done]", second)
	}
}

func TestQueryUsesFirstSetOnly(t *testing.T) {
	pool := openScriptedPool(t, []scriptedSet{
		{
			cols: []*rdb.Column{{Name: "n", Index: 0}},
			rows: [][]any{{int32(7)}},
		},
		{
			cols: []*rdb.Column{{Name: "s", Index: 0}},
			rows: [][]any{{"ignored"}},
		},
	})
	defer pool.Close()

	type rowN struct {
		N int32 `db:"n"`
	}
	got, err := Query[rowN](context.Background(), pool, &rdb.Command{SQL: "q", Arity: rdb.Any})
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0].N != 7 {
		t.Fatalf("got=%v", got)
	}
}

func TestStreamUsesFirstSetOnly(t *testing.T) {
	pool := openScriptedPool(t, []scriptedSet{
		{
			cols: []*rdb.Column{{Name: "n", Index: 0}},
			rows: [][]any{{int32(1)}, {int32(2)}},
		},
		{
			cols: []*rdb.Column{{Name: "s", Index: 0}},
			rows: [][]any{{"ignored"}},
		},
	})
	defer pool.Close()

	type rowN struct {
		N int32 `db:"n"`
	}
	var got []int32
	for r, err := range Stream[rowN](context.Background(), pool, &rdb.Command{SQL: "s", Arity: rdb.Any}) {
		if err != nil {
			t.Fatal(err)
		}
		got = append(got, r.N)
	}
	if len(got) != 2 || got[0] != 1 || got[1] != 2 {
		t.Fatalf("got=%v", got)
	}
}

func TestSliceEmptySet(t *testing.T) {
	pool := openScriptedPool(t, []scriptedSet{
		{
			cols: []*rdb.Column{{Name: "n", Index: 0}},
			rows: nil,
		},
		{
			cols: []*rdb.Column{{Name: "s", Index: 0}},
			rows: [][]any{{"x"}},
		},
	})
	defer pool.Close()

	type rowN struct {
		N int32 `db:"n"`
	}
	type rowS struct {
		S string `db:"s"`
	}

	err := Do(context.Background(), pool, &rdb.Command{SQL: "empty", Arity: rdb.Any}, func(h Handle) error {
		a, err := Slice[rowN](h)
		if err != nil {
			return err
		}
		if len(a) != 0 {
			t.Fatalf("first set len=%d", len(a))
		}
		if err := h.Next(); err != nil {
			return err
		}
		b, err := Slice[rowS](h)
		if err != nil {
			return err
		}
		if len(b) != 1 || b[0].S != "x" {
			t.Fatalf("second=%v", b)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}

// --- scripted multi-result fake driver (pool-backed *rdb.Result) ---

type scriptedSet struct {
	cols []*rdb.Column
	rows [][]any
}

type scriptedDriver struct {
	mu   sync.Mutex
	next []scriptedSet
}

func (d *scriptedDriver) DriverInfo() *rdb.DriverInfo {
	return &rdb.DriverInfo{
		DriverSupport: rdb.DriverSupport{MultipleResult: true},
	}
}

func (d *scriptedDriver) PingCommand() *rdb.Command {
	return &rdb.Command{Arity: rdb.Zero}
}

func (d *scriptedDriver) Open(ctx context.Context, c *rdb.Config) (rdb.DriverConn, error) {
	d.mu.Lock()
	sets := d.next
	d.mu.Unlock()
	return &scriptedConn{sets: sets, status: rdb.StatusReady, opened: time.Now()}, nil
}

func (d *scriptedDriver) setNext(sets []scriptedSet) {
	d.mu.Lock()
	d.next = sets
	d.mu.Unlock()
}

type scriptedConn struct {
	sets   []scriptedSet
	setIdx int
	rowIdx int
	status rdb.DriverConnStatus
	val    rdb.DriverValuer
	avail  bool
	opened time.Time
}

func (c *scriptedConn) Query(ctx context.Context, cmd *rdb.Command, params []rdb.Param, preparedToken interface{}, val rdb.DriverValuer) error {
	c.val = val
	c.setIdx = 0
	c.rowIdx = 0
	return c.loadSet()
}

func (c *scriptedConn) loadSet() error {
	if c.setIdx >= len(c.sets) {
		c.status = rdb.StatusResultDone
		return nil
	}
	set := c.sets[c.setIdx]
	if err := c.val.Columns(set.cols); err != nil {
		return err
	}
	c.rowIdx = 0
	if len(set.rows) == 0 {
		c.status = rdb.StatusResultDone
	} else {
		c.status = rdb.StatusQuery
	}
	return nil
}

func (c *scriptedConn) Scan(ctx context.Context) error {
	if c.status != rdb.StatusQuery {
		return nil
	}
	set := c.sets[c.setIdx]
	if c.rowIdx >= len(set.rows) {
		c.status = rdb.StatusResultDone
		return nil
	}
	row := set.rows[c.rowIdx]
	for i, col := range set.cols {
		var v any
		if i < len(row) {
			v = row[i]
		}
		dv := &rdb.DriverValue{Value: v, Null: v == nil}
		if err := c.val.WriteField(col, dv, nil); err != nil {
			return err
		}
	}
	c.val.RowScanned()
	c.rowIdx++
	if c.rowIdx >= len(set.rows) {
		c.status = rdb.StatusResultDone
	}
	return nil
}

func (c *scriptedConn) NextResult(ctx context.Context) (bool, error) {
	if c.setIdx+1 >= len(c.sets) {
		return false, nil
	}
	c.setIdx++
	if err := c.loadSet(); err != nil {
		return false, err
	}
	return true, nil
}

func (c *scriptedConn) NextQuery(ctx context.Context) error {
	c.status = rdb.StatusReady
	if c.val != nil {
		return c.val.Done()
	}
	return nil
}

func (c *scriptedConn) Close()                            {}
func (c *scriptedConn) Available() bool                   { return c.avail }
func (c *scriptedConn) SetAvailable(a bool)               { c.avail = a }
func (c *scriptedConn) ConnectionInfo() *rdb.ConnectionInfo { return nil }
func (c *scriptedConn) Opened() time.Time                 { return c.opened }
func (c *scriptedConn) Status() rdb.DriverConnStatus      { return c.status }
func (c *scriptedConn) Reset(config *rdb.Config) error    { c.status = rdb.StatusReady; return nil }
func (c *scriptedConn) Prepare(*rdb.Command) (interface{}, error) {
	return nil, nil
}
func (c *scriptedConn) Unprepare(interface{}) error { return nil }
func (c *scriptedConn) Begin(context.Context, rdb.IsolationLevel) error {
	return nil
}
func (c *scriptedConn) Rollback(string) error                 { return nil }
func (c *scriptedConn) Commit(context.Context) error          { return nil }
func (c *scriptedConn) SavePoint(context.Context, string) error { return nil }

var (
	scriptedDrv     = &scriptedDriver{}
	scriptedRegister sync.Once
)

func openScriptedPool(t *testing.T, sets []scriptedSet) *rdb.ConnPool {
	t.Helper()
	scriptedRegister.Do(func() {
		rdb.Register("table_scripted", scriptedDrv)
	})
	scriptedDrv.setNext(sets)
	pool, err := rdb.Open(&rdb.Config{
		DriverName:       "table_scripted",
		PoolInitCapacity: 1,
		PoolMaxCapacity:  2,
		SoftWait:         50 * time.Millisecond,
	})
	if err != nil {
		t.Fatal(err)
	}
	return pool
}
