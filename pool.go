package rdb

import (
	"fmt"
	"runtime"
	"time"

	"github.com/youtube/vitess/go/pools"
	"golang.org/x/net/context"
)

const debugConnectionReuse = false

var ErrTimeout = pools.ErrTimeout

// Queryer allows passing either a ConnPool or a Transaction.
type Queryer interface {
	Query(cmd *Command, params ...Param) (*Result, error)
}

// Represents a connection or connection configuration to a database.
type ConnPool struct {
	dr   Driver
	conf *Config
	pool *pools.ResourcePool

	OnAutoClose func(sql string)
}

func Open(config *Config) (*ConnPool, error) {
	dr, err := getDriver(config.DriverName)
	if err != nil {
		return nil, err
	}
	if config.Secure && dr.DriverInfo().SecureConnection == false {
		return nil, fmt.Errorf("Driver %s does not support secure connections.", config.DriverName)
	}
	factory := func() (pools.Resource, error) {
		if debugConnectionReuse {
			fmt.Println("Conn.Open() NEW")
		}
		conn, err := dr.Open(config)
		if conn == nil && err == nil {
			return nil, fmt.Errorf("New connection is nil")
		}
		if err != nil {
			return conn, err
		}
		return conn, conn.Reset()
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
		pool: pools.NewResourcePool(factory, initSize, maxSize, config.PoolIdleTimeout),
	}, nil
}

func (cp *ConnPool) Close() {
	cp.pool.Close()
}

// Will attempt to connect to the database and disconnect.
// Must not impact any existing connections.
func (cp *ConnPool) Ping() error {
	cmd := cp.dr.PingCommand()
	res, err := cp.Query(cmd)
	if err != nil {
		return err
	}
	return res.Close()
}

// Returns the information specific to the connection.
func (cp *ConnPool) ConnectionInfo() (*ConnectionInfo, error) {
	cmd := cp.dr.PingCommand()
	ci := &ConnectionInfo{}
	res, err := cp.query(false, nil, cmd, &ci)
	if err != nil {
		return nil, err
	}
	return ci, res.Close()
}

func (cp *ConnPool) releaseConn(conn DriverConn, kill bool) error {
	if !kill && conn.Status() != StatusReady {
		kill = true
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
		err := conn.Reset()
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
func (cp *ConnPool) getConn(again bool) (DriverConn, error) {
	var conn DriverConn
	ctx := context.Background()
	var cancel context.CancelFunc
	var timeout time.Duration

	// Time to wait for an available connection.
	if again {
		timeout = time.Millisecond * 150
	} else {
		timeout = time.Second * 30
	}
	ctx, cancel = context.WithTimeout(ctx, timeout)

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
			return cp.getConn(false)
		}
		curCap += (maxCap / 10)
		if curCap > maxCap {
			curCap = maxCap
		}
		cp.pool.SetCapacity(int(curCap))

		return cp.getConn(false)
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
func (cp *ConnPool) Query(cmd *Command, params ...Param) (*Result, error) {
	return cp.query(false, nil, cmd, nil, params...)
}

// keepOnClose used to not recycle the DB connection after a query result is done. Used for transactions and connections.
func (cp *ConnPool) query(keepOnClose bool, conn DriverConn, cmd *Command, ci **ConnectionInfo, params ...Param) (*Result, error) {
	var err error
	if cmd.Converter != nil {
		for i := range params {
			err = cmd.Converter.ConvertParam(&params[i])
			if err != nil {
				return nil, err
			}
		}
	}

	if conn == nil {
		conn, err = cp.getConn(true)
		if err != nil {
			return nil, err
		}
	}

	res := &Result{
		conn: conn,
		cp:   cp,
		val: valuer{
			cmd: cmd,
		},
		keepOnClose: keepOnClose,

		closing: make(chan struct{}, 3),
	}

	timeout := cmd.QueryTimeout
	if timeout == 0 {
		timeout = cp.conf.QueryTimeout
	}
	// Suspect this is causing an issue with connection state.
	if timeout != 0 {
		// Give the driver time to stop it if possible.
		timeout = timeout + (time.Second * 1)

		done := make(chan struct{})
		tm := time.NewTimer(timeout)
		go func() {
			defer func() {
				if rval := recover(); rval != nil {
					buf := make([]byte, 8000)
					buf = buf[:runtime.Stack(buf, false)]

					err = fmt.Errorf("Panic in database driver: %v\n%s", rval, string(buf))
				}
			}()
			err = conn.Query(cmd, params, nil, &res.val)
			tm.Stop()
			close(done)
		}()
		select {
		case <-tm.C:
			// TODO: There should be a method for aborting an active command.
			conn.Close()
			cp.releaseConn(conn, true)
			return nil, fmt.Errorf("Query timed out after %v.", timeout)
		case <-done:
		}
	} else {
		defer func() {
			if rval := recover(); rval != nil {
				buf := make([]byte, 8000)
				buf = buf[:runtime.Stack(buf, false)]

				err = fmt.Errorf("Panic in database driver: %v\n%s", rval, string(buf))
			}
		}()
		err = conn.Query(cmd, params, nil, &res.val)
	}
	if ci != nil {
		*ci = conn.ConnectionInfo()
	}

	if err == nil && len(res.val.errorList) != 0 {
		err = res.val.errorList
	}

	// Zero arity check.
	if res.val.cmd.Arity&Zero != 0 {
		defer res.close(false)

		serr := res.conn.NextQuery()
		if err == nil {
			err = serr
		}
		if err == nil && res.val.rowCount != 0 && !res.val.eof && res.val.cmd.Arity&ArityMust != 0 {
			err = ArityError
		}
	}
	if err != nil {
		cp.releaseConn(conn, true)
	}

	res.autoClose(time.Second * 25)

	return res, err
}

// Begin starts a Transaction with the default isolation level.
func (cp *ConnPool) Begin() (*Transaction, error) {
	return cp.BeginLevel(LevelDefault)
}

// BeginLevel starts a Transaction with the specified isolation level.
func (cp *ConnPool) BeginLevel(level IsolationLevel) (*Transaction, error) {
	conn, err := cp.getConn(true)
	if err != nil {
		return nil, err
	}

	tran := &Transaction{
		cp:    cp,
		conn:  conn,
		level: level,
	}
	err = conn.Begin(level)
	if err != nil {
		cp.releaseConn(conn, true)
		return nil, err
	}
	return tran, nil
}

// Connection returns a dedicated database connection from the connection pool.
func (cp *ConnPool) Connection() (*Connection, error) {
	conn, err := cp.getConn(true)
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
	c, a, _, _, _, _ := cp.pool.Stats()
	return int(c), int(a)
}
