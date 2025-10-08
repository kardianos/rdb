package rdb

import (
	"context"
	"sync"
	"testing"
	"time"
)

// dummyDriver
type dummyDriver struct{}

func (d *dummyDriver) DriverInfo() *DriverInfo {
	return &DriverInfo{}
}

func (d *dummyDriver) Open(ctx context.Context, c *Config) (DriverConn, error) {
	return &dummyConn{
		opened: time.Now(),
	}, nil
}

func (d *dummyDriver) PingCommand() *Command {
	return &Command{
		Arity: Zero,
	}
}

// dummyConn
type dummyConn struct {
	opened time.Time
	status DriverConnStatus
	avail  bool
}

func (c *dummyConn) Query(ctx context.Context, cmd *Command, params []Param, preparedToken interface{}, val DriverValuer) error {
	// This is a non-blocking query. We just check if the context is already expired.
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}
	return nil
}

func (c *dummyConn) Reset(conf *Config) error {
	return nil
}

func (c *dummyConn) Begin(ctx context.Context, level IsolationLevel) error {
	return nil
}

func (c *dummyConn) Commit(ctx context.Context) error {
	return nil
}
func (c *dummyConn) Rollback(savepoint string) error {
	return nil
}
func (c *dummyConn) SavePoint(ctx context.Context, name string) error {
	return nil
}
func (c *dummyConn) NextQuery(ctx context.Context) error {
	return nil
}
func (c *dummyConn) Close() {
}
func (c *dummyConn) Status() DriverConnStatus {
	return c.status
}
func (c *dummyConn) Opened() time.Time {
	return c.opened
}
func (c *dummyConn) SetAvailable(avail bool) {
	c.avail = avail
}
func (c *dummyConn) Available() bool {
	return c.avail
}
func (c *dummyConn) ConnectionInfo() *ConnectionInfo {
	return nil
}
func (c *dummyConn) Scan(ctx context.Context) error {
	return nil
}
func (c *dummyConn) NextResult(ctx context.Context) (more bool, err error) {
	return false, nil
}
func (c *dummyConn) Prepare(cmd *Command) (preparedToken interface{}, err error) {
	return nil, nil
}
func (c *dummyConn) Unprepare(preparedToken interface{}) (err error) {
	return nil
}

func init() {
	Register("pool_test_dummy_final", &dummyDriver{})
}

func TestPoolExpansion(t *testing.T) {
	config := &Config{
		DriverName:       "pool_test_dummy_final",
		PoolInitCapacity: 1,
		PoolMaxCapacity:  5,
		SoftWait:         time.Millisecond * 20,
	}
	pool, err := Open(config)
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()

	// Occupy the single connection in the pool so subsequent requests must wait.
	ctx := context.Background()
	firstConn, err := pool.getConn(ctx, false)
	if err != nil {
		t.Fatal(err)
	}

	var wg sync.WaitGroup
	numGoroutines := 4
	errs := make(chan error, numGoroutines)
	wg.Add(numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func(i int) {
			defer wg.Done()
			ctx, cancel := context.WithTimeout(context.Background(), 800*time.Millisecond)
			defer cancel()

			res, err := pool.Query(ctx, &Command{})
			if err != nil {
				errs <- err
				return
			}
			res.Close()
		}(i)
	}

	wg.Wait()
	close(errs)

	// Release the first connection we took.
	pool.releaseConn(ctx, firstConn, false)

	var errCount int
	for err := range errs {
		if err != nil {
			t.Logf("goroutine failed with: %v", err)
			errCount++
		}
	}

	if errCount > 0 {
		t.Errorf("Expected 0 errors from pool expansion, but got %d", errCount)
	}
}
