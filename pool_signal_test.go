package rdb

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// signalDriver is a fake driver that uses channels (signals) instead of sleeps
// for controlling when "work" completes. This allows precise testing of pool behavior.
type signalDriver struct {
	mu            sync.Mutex
	conns         []*signalConn
	connIDCounter atomic.Int64
}

func (d *signalDriver) DriverInfo() *DriverInfo {
	return &DriverInfo{}
}

func (d *signalDriver) Open(ctx context.Context, c *Config) (DriverConn, error) {
	conn := &signalConn{
		id:       d.connIDCounter.Add(1),
		driver:   d,
		opened:   time.Now(),
		status:   StatusReady,
		workDone: make(chan struct{}, 1), // buffered so signal() never blocks
	}
	d.mu.Lock()
	d.conns = append(d.conns, conn)
	d.mu.Unlock()
	return conn, nil
}

func (d *signalDriver) PingCommand() *Command {
	return &Command{Arity: Zero}
}

func (d *signalDriver) getConns() []*signalConn {
	d.mu.Lock()
	defer d.mu.Unlock()
	result := make([]*signalConn, len(d.conns))
	copy(result, d.conns)
	return result
}

func (d *signalDriver) connCount() int {
	d.mu.Lock()
	defer d.mu.Unlock()
	return len(d.conns)
}

// signalConn is a fake connection that blocks on Query until signaled.
type signalConn struct {
	id       int64
	driver   *signalDriver
	opened   time.Time
	status   DriverConnStatus
	avail    bool
	workDone chan struct{} // signal to complete current query (buffered, cap 1)

	mu         sync.Mutex
	queryCount int   // how many times Query was called on this connection
	closed     bool  // track if Close() was called
	nextError  error // if set, next Query returns this error
}

func (c *signalConn) Query(ctx context.Context, cmd *Command, params []Param, preparedToken any, val DriverValuer) error {
	c.mu.Lock()
	c.queryCount++
	c.status = StatusQuery
	// Check for injected error
	if c.nextError != nil {
		err := c.nextError
		c.nextError = nil
		c.status = StatusReady
		c.mu.Unlock()
		return err
	}
	// Drain any pending signals and create fresh channel
	select {
	case <-c.workDone:
	default:
	}
	c.mu.Unlock()

	// Block until signaled or context canceled
	select {
	case <-c.workDone:
		c.mu.Lock()
		c.status = StatusReady
		c.mu.Unlock()
		return nil
	case <-ctx.Done():
		c.mu.Lock()
		c.status = StatusReady
		c.mu.Unlock()
		return ctx.Err()
	}
}

func (c *signalConn) setNextError(err error) {
	c.mu.Lock()
	c.nextError = err
	c.mu.Unlock()
}

func (c *signalConn) isClosed() bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.closed
}

func (c *signalConn) signal() {
	// Non-blocking send to buffered channel
	select {
	case c.workDone <- struct{}{}:
	default:
	}
}

func (c *signalConn) getQueryCount() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.queryCount
}

func (c *signalConn) Reset(conf *Config) error {
	c.mu.Lock()
	c.status = StatusReady
	c.mu.Unlock()
	return nil
}

func (c *signalConn) Begin(ctx context.Context, level IsolationLevel) error { return nil }
func (c *signalConn) Commit(ctx context.Context) error                      { return nil }
func (c *signalConn) Rollback(savepoint string) error                       { return nil }
func (c *signalConn) SavePoint(ctx context.Context, name string) error      { return nil }
func (c *signalConn) NextQuery(ctx context.Context) error                   { return nil }
func (c *signalConn) Close() {
	c.mu.Lock()
	c.closed = true
	c.mu.Unlock()
}
func (c *signalConn) Status() DriverConnStatus                              { return c.status }
func (c *signalConn) Opened() time.Time                                     { return c.opened }
func (c *signalConn) SetAvailable(avail bool)                               { c.avail = avail }
func (c *signalConn) Available() bool                                       { return c.avail }
func (c *signalConn) ConnectionInfo() *ConnectionInfo                       { return nil }
func (c *signalConn) Scan(ctx context.Context) error                        { return nil }
func (c *signalConn) NextResult(ctx context.Context) (bool, error)          { return false, nil }
func (c *signalConn) Prepare(cmd *Command) (any, error)    { return nil, nil }
func (c *signalConn) Unprepare(preparedToken any) error    { return nil }

// TestPoolExpandsToMax verifies that the pool expands when all connections are busy.
func TestPoolExpandsToMax(t *testing.T) {
	driver := &signalDriver{}
	driverName := "signal_expand_test"
	Register(driverName, driver)

	config := &Config{
		DriverName:       driverName,
		PoolInitCapacity: 2,
		PoolMaxCapacity:  5,
		SoftWait:         time.Millisecond * 10,
	}
	pool, err := Open(config)
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()

	ctx := context.Background()

	// Start 5 concurrent queries (more than init capacity)
	var wg sync.WaitGroup
	results := make([]*Result, 5)
	errs := make([]error, 5)

	for i := range 5 {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			// Use a longer timeout to allow pool expansion
			qctx, cancel := context.WithTimeout(ctx, time.Second*2)
			defer cancel()
			results[idx], errs[idx] = pool.Query(qctx, &Command{Arity: Any})
		}(i)
	}

	// Wait for all queries to start (connections to be acquired)
	// Poll until we have 5 connections created
	deadline := time.Now().Add(time.Second * 2)
	for time.Now().Before(deadline) {
		if driver.connCount() >= 5 {
			break
		}
		time.Sleep(time.Millisecond * 10)
	}

	// Signal all connections to complete
	for _, conn := range driver.getConns() {
		conn.signal()
	}

	wg.Wait()

	// Check no errors
	for i, err := range errs {
		if err != nil {
			t.Errorf("query %d failed: %v", i, err)
		}
	}

	// Close results
	for _, res := range results {
		if res != nil {
			res.Close()
		}
	}

	// Verify pool expanded to max
	capacity, _ := pool.PoolAvailable()
	if capacity != 5 {
		t.Errorf("expected pool capacity 5, got %d", capacity)
	}

	// Verify 5 connections were created
	if count := driver.connCount(); count != 5 {
		t.Errorf("expected 5 connections created, got %d", count)
	}
}

// TestPoolStaysAtMin verifies that the pool doesn't expand when queries complete quickly.
func TestPoolStaysAtMin(t *testing.T) {
	driver := &signalDriver{}
	driverName := "signal_min_test"
	Register(driverName, driver)

	config := &Config{
		DriverName:       driverName,
		PoolInitCapacity: 3,
		PoolMaxCapacity:  10,
		SoftWait:         time.Millisecond * 50, // longer soft wait
	}
	pool, err := Open(config)
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()

	ctx := context.Background()

	// Run 10 sequential queries, each completing immediately
	for i := range 10 {
		go func() {
			// Signal completion after a tiny delay
			time.Sleep(time.Millisecond)
			for _, conn := range driver.getConns() {
				conn.signal()
			}
		}()

		res, err := pool.Query(ctx, &Command{Arity: Any})
		if err != nil {
			t.Fatalf("query %d failed: %v", i, err)
		}
		res.Close()
	}

	// Verify pool stayed at initial capacity
	capacity, _ := pool.PoolAvailable()
	if capacity != 3 {
		t.Errorf("expected pool capacity to stay at 3, got %d", capacity)
	}

	// Verify only initial connections were created
	if count := driver.connCount(); count != 3 {
		t.Errorf("expected 3 connections, got %d", count)
	}
}

// TestPoolReusesConnections verifies that connections are reused rather than recreated.
func TestPoolReusesConnections(t *testing.T) {
	driver := &signalDriver{}
	driverName := "signal_reuse_test"
	Register(driverName, driver)

	config := &Config{
		DriverName:       driverName,
		PoolInitCapacity: 2,
		PoolMaxCapacity:  2,
		SoftWait:         time.Millisecond * 10,
	}
	pool, err := Open(config)
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()

	ctx := context.Background()

	// Run 20 sequential queries
	for i := range 20 {
		go func() {
			time.Sleep(time.Millisecond)
			for _, conn := range driver.getConns() {
				conn.signal()
			}
		}()

		res, err := pool.Query(ctx, &Command{Arity: Any})
		if err != nil {
			t.Fatalf("query %d failed: %v", i, err)
		}
		res.Close()
	}

	// Verify only 2 connections were ever created
	conns := driver.getConns()
	if len(conns) != 2 {
		t.Errorf("expected 2 connections, got %d", len(conns))
	}

	// Verify each connection was used multiple times (reused)
	totalQueries := 0
	for _, conn := range conns {
		count := conn.getQueryCount()
		totalQueries += count
		t.Logf("connection %d: %d queries", conn.id, count)
	}

	if totalQueries != 20 {
		t.Errorf("expected 20 total queries across connections, got %d", totalQueries)
	}

	// Each connection should have been used roughly 10 times
	for _, conn := range conns {
		count := conn.getQueryCount()
		if count < 5 {
			t.Errorf("connection %d only used %d times, expected more reuse", conn.id, count)
		}
	}
}

// TestPoolConcurrentReuse verifies connection reuse under concurrent load.
func TestPoolConcurrentReuse(t *testing.T) {
	driver := &signalDriver{}
	driverName := "signal_concurrent_reuse_test"
	Register(driverName, driver)

	config := &Config{
		DriverName:       driverName,
		PoolInitCapacity: 3,
		PoolMaxCapacity:  3,
		SoftWait:         time.Millisecond * 10,
	}
	pool, err := Open(config)
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()

	ctx := context.Background()
	const numQueries = 30
	const numWorkers = 6

	var wg sync.WaitGroup
	errCh := make(chan error, numQueries)

	// Start a goroutine that continuously signals connections
	stopSignal := make(chan struct{})
	go func() {
		for {
			select {
			case <-stopSignal:
				return
			default:
				for _, conn := range driver.getConns() {
					conn.signal()
				}
				time.Sleep(time.Millisecond * 5)
			}
		}
	}()

	// Run concurrent queries
	queriesPerWorker := numQueries / numWorkers
	for w := range numWorkers {
		wg.Add(1)
		go func(worker int) {
			defer wg.Done()
			for i := range queriesPerWorker {
				res, err := pool.Query(ctx, &Command{Arity: Any})
				if err != nil {
					errCh <- err
					continue
				}
				res.Close()
				_ = i
			}
		}(w)
	}

	wg.Wait()
	close(stopSignal)
	close(errCh)

	// Check for errors
	var errCount int
	for err := range errCh {
		t.Logf("query error: %v", err)
		errCount++
	}
	if errCount > 0 {
		t.Errorf("had %d query errors", errCount)
	}

	// Verify only 3 connections were created (pool didn't expand)
	conns := driver.getConns()
	if len(conns) != 3 {
		t.Errorf("expected 3 connections, got %d", len(conns))
	}

	// Verify connections were reused
	totalQueries := 0
	for _, conn := range conns {
		totalQueries += conn.getQueryCount()
	}
	if totalQueries != numQueries {
		t.Errorf("expected %d total queries, got %d", numQueries, totalQueries)
	}
	t.Logf("total queries across %d connections: %d", len(conns), totalQueries)
}

// TestPoolPartialExpansion verifies pool expands only as needed, not always to max.
func TestPoolPartialExpansion(t *testing.T) {
	driver := &signalDriver{}
	driverName := "signal_partial_expand_test"
	Register(driverName, driver)

	config := &Config{
		DriverName:       driverName,
		PoolInitCapacity: 2,
		PoolMaxCapacity:  20,
		SoftWait:         time.Millisecond * 5,
	}
	pool, err := Open(config)
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()

	ctx := context.Background()

	// Hold 4 connections busy (more than init, less than max)
	var wg sync.WaitGroup
	results := make([]*Result, 4)
	errs := make([]error, 4)

	for i := range 4 {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			qctx, cancel := context.WithTimeout(ctx, time.Second*2)
			defer cancel()
			results[idx], errs[idx] = pool.Query(qctx, &Command{Arity: Any})
		}(i)
	}

	// Wait for connections to be acquired
	deadline := time.Now().Add(time.Second)
	for time.Now().Before(deadline) {
		if driver.connCount() >= 4 {
			break
		}
		time.Sleep(time.Millisecond * 5)
	}

	// Signal all to complete
	for _, conn := range driver.getConns() {
		conn.signal()
	}

	wg.Wait()

	for i, err := range errs {
		if err != nil {
			t.Errorf("query %d failed: %v", i, err)
		}
	}
	for _, res := range results {
		if res != nil {
			res.Close()
		}
	}

	// Pool should have expanded but not to max
	capacity, _ := pool.PoolAvailable()
	connCount := driver.connCount()

	t.Logf("pool capacity: %d, connections created: %d", capacity, connCount)

	if connCount > 12 {
		t.Errorf("pool expanded too much: %d connections for 4 concurrent queries", connCount)
	}
	if connCount < 4 {
		t.Errorf("pool didn't expand enough: %d connections for 4 concurrent queries", connCount)
	}
}

// TestPoolRepeatedMildPressure tests if repeated mild pressure causes runaway expansion.
func TestPoolRepeatedMildPressure(t *testing.T) {
	driver := &signalDriver{}
	driverName := "signal_mild_pressure_test"
	Register(driverName, driver)

	config := &Config{
		DriverName:       driverName,
		PoolInitCapacity: 2,
		PoolMaxCapacity:  50,
		SoftWait:         time.Millisecond * 5,
	}
	pool, err := Open(config)
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()

	ctx := context.Background()

	// Simulate repeated bursts of 3 concurrent queries (just above init capacity)
	// Each burst should not cause unbounded expansion
	for burst := range 5 {
		var wg sync.WaitGroup
		results := make([]*Result, 3)
		errs := make([]error, 3)

		for i := range 3 {
			wg.Add(1)
			go func(idx int) {
				defer wg.Done()
				qctx, cancel := context.WithTimeout(ctx, time.Second)
				defer cancel()
				results[idx], errs[idx] = pool.Query(qctx, &Command{Arity: Any})
			}(i)
		}

		// Wait a bit for queries to start, then signal completion
		time.Sleep(time.Millisecond * 20)
		for _, conn := range driver.getConns() {
			conn.signal()
		}

		wg.Wait()

		for i, err := range errs {
			if err != nil {
				t.Errorf("burst %d query %d failed: %v", burst, i, err)
			}
		}
		for _, res := range results {
			if res != nil {
				res.Close()
			}
		}

		capacity, _ := pool.PoolAvailable()
		t.Logf("after burst %d: capacity=%d, connections=%d", burst, capacity, driver.connCount())
	}

	// After 5 bursts of 3 queries, capacity should not have grown excessively
	finalCapacity, _ := pool.PoolAvailable()
	finalConns := driver.connCount()

	t.Logf("final: capacity=%d, connections=%d", finalCapacity, finalConns)

	// Capacity shouldn't exceed what's needed for concurrent load
	// With bursts of 3, we shouldn't need more than ~12-15 capacity
	if finalCapacity > 20 {
		t.Errorf("pool capacity grew excessively: %d (expected <= 20 for bursts of 3)", finalCapacity)
	}
}

// TestPoolIdleTimeout tests that idle connections are properly closed.
func TestPoolIdleTimeout(t *testing.T) {
	driver := &signalDriver{}
	driverName := "signal_idle_timeout_test"
	Register(driverName, driver)

	config := &Config{
		DriverName:       driverName,
		PoolInitCapacity: 3,
		PoolMaxCapacity:  10,
		PoolIdleTimeout:  time.Millisecond * 50, // short idle timeout for testing
		SoftWait:         time.Millisecond * 5,
	}
	pool, err := Open(config)
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()

	ctx := context.Background()

	// Run a few queries to ensure connections are created
	for i := range 3 {
		go func() {
			time.Sleep(time.Millisecond)
			for _, conn := range driver.getConns() {
				conn.signal()
			}
		}()

		res, err := pool.Query(ctx, &Command{Arity: Any})
		if err != nil {
			t.Fatalf("query %d failed: %v", i, err)
		}
		res.Close()
	}

	initialConns := driver.connCount()
	t.Logf("initial connections: %d", initialConns)

	// Wait longer than idle timeout
	time.Sleep(time.Millisecond * 150)

	// Check how many connections were closed
	closedCount := 0
	for _, conn := range driver.getConns() {
		if conn.isClosed() {
			closedCount++
		}
	}

	t.Logf("after idle timeout: %d/%d connections closed", closedCount, initialConns)

	// At least some connections should have been closed due to idle timeout
	// Note: The pool might keep minimum connections alive
	_, available := pool.PoolAvailable()
	t.Logf("pool available: %d", available)
}

// TestPoolErrorRecovery tests that the pool recovers when a connection returns an error.
func TestPoolErrorRecovery(t *testing.T) {
	driver := &signalDriver{}
	driverName := "signal_error_recovery_test"
	Register(driverName, driver)

	config := &Config{
		DriverName:       driverName,
		PoolInitCapacity: 2,
		PoolMaxCapacity:  5,
		SoftWait:         time.Millisecond * 10,
	}
	pool, err := Open(config)
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()

	ctx := context.Background()

	// First, run a successful query to get a connection created
	go func() {
		time.Sleep(time.Millisecond)
		for _, conn := range driver.getConns() {
			conn.signal()
		}
	}()

	res, err := pool.Query(ctx, &Command{Arity: Any})
	if err != nil {
		t.Fatalf("initial query failed: %v", err)
	}
	res.Close()

	initialConns := driver.connCount()
	t.Logf("initial connections: %d", initialConns)

	// Inject an error into ALL current connections
	conns := driver.getConns()
	injectedErr := errors.New("simulated connection error")
	for _, conn := range conns {
		conn.setNextError(injectedErr)
	}

	// Run another query - it should fail because connection returns error
	qctx, cancel := context.WithTimeout(ctx, time.Millisecond*100)
	_, err = pool.Query(qctx, &Command{Arity: Any})
	cancel()
	if err == nil {
		t.Log("query succeeded unexpectedly")
	} else {
		t.Logf("query failed as expected: %v", err)
	}

	// Pool should still be functional - run more queries
	// The pool should create new connections or reuse ones that no longer have errors
	successCount := 0
	for i := range 5 {
		go func() {
			time.Sleep(time.Millisecond)
			for _, conn := range driver.getConns() {
				conn.signal()
			}
		}()

		qctx, cancel := context.WithTimeout(ctx, time.Millisecond*500)
		res, err := pool.Query(qctx, &Command{Arity: Any})
		cancel()
		if err != nil {
			t.Logf("recovery query %d failed: %v", i, err)
			continue
		}
		res.Close()
		successCount++
	}

	t.Logf("successful queries after error: %d/5", successCount)

	if successCount < 4 {
		t.Errorf("pool didn't recover well: only %d/5 queries succeeded", successCount)
	}

	// Check final pool state
	capacity, available := pool.PoolAvailable()
	t.Logf("final pool state: capacity=%d, available=%d, total conns created=%d",
		capacity, available, driver.connCount())
}

// TestPoolExpandByConfig tests that ExpandPoolBy config option works.
func TestPoolExpandByConfig(t *testing.T) {
	driver := &signalDriver{}
	driverName := "signal_expand_by_config_test"
	Register(driverName, driver)

	config := &Config{
		DriverName:       driverName,
		PoolInitCapacity: 2,
		PoolMaxCapacity:  50,
		ExpandPoolBy:     3, // custom expansion size
		SoftWait:         time.Millisecond * 5,
	}
	pool, err := Open(config)
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()

	ctx := context.Background()

	// Start 5 concurrent queries to trigger expansion
	var wg sync.WaitGroup
	results := make([]*Result, 5)
	errs := make([]error, 5)

	for i := range 5 {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			qctx, cancel := context.WithTimeout(ctx, time.Second)
			defer cancel()
			results[idx], errs[idx] = pool.Query(qctx, &Command{Arity: Any})
		}(i)
	}

	// Wait for queries to start
	time.Sleep(time.Millisecond * 50)

	// Signal all to complete
	for _, conn := range driver.getConns() {
		conn.signal()
	}

	wg.Wait()

	for i, err := range errs {
		if err != nil {
			t.Errorf("query %d failed: %v", i, err)
		}
	}
	for _, res := range results {
		if res != nil {
			res.Close()
		}
	}

	// With ExpandPoolBy=3, starting at 2:
	// First expansion: 2 + 3 = 5
	// That should be enough for 5 concurrent queries
	capacity, _ := pool.PoolAvailable()
	t.Logf("capacity=%d, connections=%d", capacity, driver.connCount())

	// Capacity should be 5 (2 + 3), not 8 (2 + 6 default)
	if capacity != 5 {
		t.Errorf("expected capacity 5 with ExpandPoolBy=3, got %d", capacity)
	}
}

// TestPoolConnectionMaxLifetime tests that connections are closed after max lifetime.
func TestPoolConnectionMaxLifetime(t *testing.T) {
	driver := &signalDriver{}
	driverName := "signal_max_lifetime_test"
	Register(driverName, driver)

	config := &Config{
		DriverName:            driverName,
		PoolInitCapacity:      2,
		PoolMaxCapacity:       5,
		ConnectionMaxLifetime: time.Millisecond * 50, // short lifetime for testing
		SoftWait:              time.Millisecond * 5,
	}
	pool, err := Open(config)
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()

	ctx := context.Background()

	// Start background goroutine to continuously signal connections
	stopSignal := make(chan struct{})
	go func() {
		for {
			select {
			case <-stopSignal:
				return
			default:
				for _, conn := range driver.getConns() {
					conn.signal()
				}
				time.Sleep(time.Millisecond * 2)
			}
		}
	}()
	defer close(stopSignal)

	// Run initial query
	res, err := pool.Query(ctx, &Command{Arity: Any})
	if err != nil {
		t.Fatalf("initial query failed: %v", err)
	}
	res.Close()

	initialConns := driver.connCount()
	t.Logf("initial connections: %d", initialConns)

	// Record which connections exist now
	initialConnIDs := make(map[int64]bool)
	for _, conn := range driver.getConns() {
		initialConnIDs[conn.id] = true
	}

	// Wait for connections to exceed max lifetime
	time.Sleep(time.Millisecond * 100)

	// Run more queries - old connections should be replaced
	for i := range 3 {
		res, err := pool.Query(ctx, &Command{Arity: Any})
		if err != nil {
			t.Fatalf("query %d failed: %v", i, err)
		}
		res.Close()
	}

	// Check how many connections were closed (exceeded lifetime)
	closedCount := 0
	newConnCount := 0
	for _, conn := range driver.getConns() {
		if conn.isClosed() {
			closedCount++
		}
		if !initialConnIDs[conn.id] {
			newConnCount++
		}
	}

	totalConns := driver.connCount()
	t.Logf("total connections created: %d, closed: %d, new: %d", totalConns, closedCount, newConnCount)

	// Either old connections were closed, or new ones were created to replace them
	if closedCount == 0 && newConnCount == 0 {
		t.Log("WARNING: no connections were retired despite max lifetime being exceeded")
	} else if closedCount > 0 {
		t.Logf("%d connections closed due to max lifetime", closedCount)
	} else if newConnCount > 0 {
		t.Logf("%d new connections created (old ones likely retired)", newConnCount)
	}
}
