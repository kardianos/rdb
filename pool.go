package rdb

import (
	"fmt"
	"runtime"
	"time"

	"context"

	"github.com/kardianos/rdb/internal/pools"
)

const debugConnectionReuse = false

var ErrTimeout = pools.ErrTimeout

// Queryer allows passing either a ConnPool or a Transaction.
type Queryer interface {
	Query(ctx context.Context, cmd *Command, params ...Param) (*Result, error)
}

// Represents a connection or connection configuration to a database.
type ConnPool struct {
	dr   Driver
	conf *Config
	pool *pools.ResourcePool
}

func Open(config *Config) (*ConnPool, error) {
	dr, err := getDriver(config.DriverName)
	if err != nil {
		return nil, err
	}
	if config.Secure && !dr.DriverInfo().SecureConnection {
		return nil, fmt.Errorf("driver %s does not support secure connections", config.DriverName)
	}
	factory := func(ctx context.Context) (pools.Resource, error) {
		if debugConnectionReuse {
			fmt.Println("Conn.Open() NEW")
		}
		conn, err := dr.Open(ctx, config)
		if conn == nil && err == nil {
			return nil, fmt.Errorf("new connection is nil")
		}
		if err != nil {
			return conn, err
		}
		return conn, conn.Reset(config)
	}

	initSize := config.PoolInitCapacity
	maxSize := config.PoolMaxCapacity

	if initSize <= 0 {
		initSize = 2
	}
	if maxSize <= 0 {
		maxSize = 100
	}

	return &ConnPool{
		dr:   dr,
		conf: config,
		pool: pools.NewResourcePool(factory, initSize, maxSize, config.PoolIdleTimeout, 0, nil),
	}, nil
}

func (cp *ConnPool) Close() {
	cp.pool.Close()
}

// Will attempt to connect to the database and disconnect.
// Must not impact any existing connections.
func (cp *ConnPool) Ping(ctx context.Context) error {
	cmd := cp.dr.PingCommand()
	res, err := cp.Query(ctx, cmd)
	if err != nil {
		return err
	}
	return res.Close()
}

// Returns the information specific to the connection.
func (cp *ConnPool) ConnectionInfo(ctx context.Context) (*ConnectionInfo, error) {
	cmd := cp.dr.PingCommand()
	ci := &ConnectionInfo{}
	res, err := cp.query(ctx, false, nil, cmd, &ci)
	if err != nil {
		return nil, err
	}
	return ci, res.Close()
}

func (cp *ConnPool) releaseConn(conn DriverConn, kill bool) error {
	if conn.Status() != StatusReady {
		kill = true
	}
	if life := cp.conf.ConnectionMaxLifetime; life > 0 {
		now := time.Now()
		op := conn.Opened()
		diff := now.Sub(op)
		if diff > life {
			kill = true
		}
	}
	if kill {
		if debugConnectionReuse {
			fmt.Println("Result.Close() CLOSE")
		}
		conn.Close()
		if conn.Available() {
			conn.SetAvailable(false)
			cp.pool.Put(nil)
		}
		return nil
	}
	if debugConnectionReuse {
		fmt.Println("Result.Close() REUSE")
	}
	if conn.Available() {
		err := conn.Reset(cp.conf)
		if err != nil {
			conn.SetAvailable(false)
			cp.pool.Put(nil)
			return err
		}
		conn.SetAvailable(false)
		cp.pool.Put(conn)
	}
	if debugConnectionReuse {
		fmt.Println(cp.pool.StatsJSON())
	}
	return nil
}
func (cp *ConnPool) getConn(ctx context.Context, again bool) (DriverConn, error) {
	var conn DriverConn
	var cancel context.CancelFunc

	connObj, err := cp.pool.Get(ctx)
	if cancel != nil {
		cancel()
	}
	if connObj != nil {
		conn = connObj.(DriverConn)
		conn.SetAvailable(true)
	}
	// Logic to expand the pool capacity up to the max capacity.
	if again && err == pools.ErrTimeout {
		maxCap := cp.pool.MaxCap()
		curCap := cp.pool.Capacity()

		if curCap >= maxCap {
			return cp.getConn(ctx, false)
		}
		curCap += (maxCap / 10)
		if curCap > maxCap {
			curCap = maxCap
		}
		cp.pool.SetCapacity(int(curCap))

		return cp.getConn(ctx, false)
	}
	if err == pools.ErrTimeout {
		err = ErrTimeout
	}
	return conn, err
}

// Perform a query against the database.
// If values are not specified in the Command.Input[...].V, then they
// may be specified in the Value. Order may be used to match the
// existing parameters if the Value.N name is omitted.
func (cp *ConnPool) Query(ctx context.Context, cmd *Command, params ...Param) (*Result, error) {
	return cp.query(ctx, false, nil, cmd, nil, params...)
}

// keepOnClose used to not recycle the DB connection after a query result is done. Used for transactions and connections.
func (cp *ConnPool) query(ctx context.Context, keepOnClose bool, conn DriverConn, cmd *Command, ci **ConnectionInfo, params ...Param) (res *Result, err error) {
	if cmd.Converter != nil {
		for i := range params {
			err = cmd.Converter.ConvertParam(&params[i])
			if err != nil {
				return nil, fmt.Errorf("ConvertParam: %w", err)
			}
		}
	}

	if conn == nil {
		conn, err = cp.getConn(ctx, true)
		if err != nil {
			return nil, fmt.Errorf("getConn: %w", err)
		}
	}
	if ctx == nil {
		ctx = context.Background()
	}

	res = &Result{
		ctx:  ctx,
		conn: conn,
		cp:   cp,
		val: valuer{
			cmd: cmd,
		},
		keepOnClose: keepOnClose,

		closing: make(chan struct{}, 3),
	}

	defer func() {
		if rval := recover(); rval != nil {
			buf := make([]byte, 8000)
			buf = buf[:runtime.Stack(buf, false)]

			err = fmt.Errorf("Panic in database driver: %v\n%s", rval, string(buf))
		}
	}()
	err = conn.Query(ctx, cmd, params, nil, &res.val)
	if ci != nil {
		*ci = conn.ConnectionInfo()
	}

	if err == nil && len(res.val.errorList) != 0 {
		err = res.val.errorList
	}

	// Zero arity check.
	if res.val.cmd.Arity&Zero != 0 {
		defer res.close(false)

		serr := res.conn.NextQuery(ctx)
		if err == nil {
			err = serr
		}
		if err == nil && res.val.rowCount != 0 && !res.val.eof && res.val.cmd.Arity&ArityMust != 0 {
			err = ErrArity
		}
	}
	if err != nil {
		cp.releaseConn(conn, true)
		res.closed = true
	}

	return res, err
}

// Begin starts a Transaction with the default isolation level.
func (cp *ConnPool) Begin(ctx context.Context) (*Transaction, error) {
	return cp.BeginLevel(ctx, LevelDefault)
}

// BeginLevel starts a Transaction with the specified isolation level.
func (cp *ConnPool) BeginLevel(ctx context.Context, level IsolationLevel) (*Transaction, error) {
	conn, err := cp.getConn(ctx, true)
	if err != nil {
		return nil, err
	}

	tran := &Transaction{
		ctx:   ctx,
		cp:    cp,
		conn:  conn,
		level: level,
	}
	err = conn.Begin(ctx, level)
	if err != nil {
		cp.releaseConn(conn, true)
		return nil, err
	}
	return tran, nil
}

// Connection returns a dedicated database connection from the connection pool.
func (cp *ConnPool) Connection(ctx context.Context) (*Connection, error) {
	conn, err := cp.getConn(ctx, true)
	if err != nil {
		return nil, err
	}

	c := &Connection{
		cp:   cp,
		conn: conn,
	}
	return c, nil
}

func (cp *ConnPool) PoolAvailable() (capacity, available int) {
	c, a := cp.pool.Capacity(), cp.pool.Available()
	return int(c), int(a)
}
