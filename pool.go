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
		return dr.Open(config)
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
	res, err := cp.Query(cmd)
	if err != nil {
		return nil, err
	}
	defer res.Close()
	return res.conn.ConnectionInfo()
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
func (cp *ConnPool) Query(cmd *Command, vv ...Value) (*Result, error) {
	conn, err := cp.getConn()
	if err != nil {
		return nil, err
	}

	res := &Result{
		conn: conn,
		cp:   cp,
	}

	fields := make([]*Field, len(cmd.Output))
	for i := range cmd.Output {
		fields[i] = &cmd.Output[i]
	}

	res.val.initFields = fields
	err = conn.Query(cmd, vv, false, IsoLevelDefault, &res.val)

	if err == nil && len(res.val.errors) != 0 {
		err = res.val.errors
	}

	return res, err
}

// API for tranactions are preliminary. Not a stable API call.
func (cp *ConnPool) Transaction(iso IsolationLevel) (*Transaction, error) {
	panic("Not implemented")
	return nil, nil
}

// Get the panic'ing version that doesn't return errors.
func (cp *ConnPool) Must() ConnPoolMust {
	return ConnPoolMust{norm: cp}
}
