package rdb

import (
	"fmt"

	"github.com/youtube/vitess/go/pools"
)

const debugConnectionReuse = false

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
	res, err := cp.query(cmd, &ci)
	if err != nil {
		return nil, err
	}
	return ci, res.Close()
}

func (cp *ConnPool) releaseConn(conn Conn, kill bool) error {
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
func (cp *ConnPool) getConn() (Conn, error) {
	var conn Conn
	connObj, err := cp.pool.Get()
	if connObj != nil {
		conn = connObj.(Conn)
	}
	return conn, err
}

// Perform a query against the database.
// If values are not specified in the Command.Input[...].V, then they
// may be specified in the Value. Order may be used to match the
// existing parameters if the Value.N name is omitted.
func (cp *ConnPool) Query(cmd *Command, params ...Param) (*Result, error) {
	return cp.query(cmd, nil, params...)
}

func (cp *ConnPool) query(cmd *Command, ci **ConnectionInfo, params ...Param) (*Result, error) {
	conn, err := cp.getConn()
	if err != nil {
		return nil, err
	}

	res := &Result{
		conn: conn,
		cp:   cp,
		val: valuer{
			arity: cmd.Arity,
		},
	}

	fields := make([]*Field, len(cmd.Fields))
	for i := range cmd.Fields {
		fields[i] = &cmd.Fields[i]
	}

	res.val.initFields = fields
	err = conn.Query(cmd, params, nil, &res.val)

	if ci != nil {
		var ciErr error
		*ci, ciErr = conn.ConnectionInfo()
		if err == nil {
			err = ciErr
		}
	}

	if err == nil && len(res.val.errorList) != 0 {
		err = res.val.errorList
	}

	// Zero arity check.
	if res.val.arity&Zero != 0 {
		defer res.close(false)

		serr := res.conn.Scan(false)
		if err == nil {
			err = serr
		}
		if err == nil && res.val.rowCount != 0 && !res.val.eof && res.val.arity&ArityMust != 0 {
			err = arityError
		}
	}

	return res, err
}

// Begin a Transaction with the default isolation level.
func (cp *ConnPool) Begin() (*Transaction, error) {
	return cp.BeginLevel(LevelDefault)
}
func (cp *ConnPool) BeginLevel(level IsolationLevel) (*Transaction, error) {
	return nil, NotImplemented
}
