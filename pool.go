package rdb

import (
	"fmt"

	"bitbucket.org/kardianos/rdb/third_party/vitess/pools"
)

const debugConnectionReuse = false

// Queryer allows passing either a ConnPool or a Transaction.
type Queryer interface {
	Query(cmd *Command, params ...Param) (*Result, error)
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
		return conn, err
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
	if kill {
		if debugConnectionReuse {
			fmt.Println("Result.Close() CLOSE")
		}
		conn.Close()
		cp.pool.Put(nil)
		return nil
	}
	if debugConnectionReuse {
		fmt.Println("Result.Close() REUSE")
	}
	cp.pool.Put(conn)
	if debugConnectionReuse {
		fmt.Println(cp.pool.StatsJSON())
	}
	return nil
}
func (cp *ConnPool) getConn() (DriverConn, error) {
	var conn DriverConn
	connObj, err := cp.pool.Get()
	if connObj != nil {
		conn = connObj.(DriverConn)
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

func (cp *ConnPool) query(inTran bool, conn DriverConn, cmd *Command, ci **ConnectionInfo, params ...Param) (*Result, error) {
	var err error
	if conn == nil {
		conn, err = cp.getConn()
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
		keepOnClose: inTran,
	}

	if cmd.Converter != nil {
		for i := range params {
			err = cmd.Converter.ConvertParam(&params[i])
			if err != nil {
				return nil, err
			}
		}
	}

	err = conn.Query(cmd, params, nil, &res.val)

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

	return res, err
}

// Begin a Transaction with the default isolation level.
func (cp *ConnPool) Begin() (*Transaction, error) {
	return cp.BeginLevel(LevelDefault)
}
func (cp *ConnPool) BeginLevel(level IsolationLevel) (*Transaction, error) {
	conn, err := cp.getConn()
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

func (cp *ConnPool) PoolAvailable() (capacity, available int) {
	c, a, _, _, _, _ := cp.pool.Stats()
	return int(c), int(a)
}
