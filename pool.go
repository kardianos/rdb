package rdb

import (
	"fmt"
	"runtime"
	"sync"
)

// Represents a connection or connection configuration to a database.
type ConnPool struct {
	dr   Driver
	conf *Config

	pool *sync.Pool
}

func (cp *ConnPool) Close() error {
	// TODO: Close all active connections.
	return nil
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

// Perform a query against the database.
// If values are not specified in the Command.Input[...].V, then they
// may be specified in the Value. Order may be used to match the
// existing parameters if the Value.N name is omitted.
func (cp *ConnPool) Query(cmd *Command, vv ...Value) (*Result, error) {
	// TODO: Use a better pool.
	connObj := cp.pool.Get()
	var conn Conn
	var err error
	if connObj == nil {
		if debugConnectionReuse {
			fmt.Println("Conn.Open() NEW")
		}
		conn, err = cp.dr.Open(cp.conf)
		if err != nil {
			return nil, err
		}
		runtime.SetFinalizer(conn, (Conn).Close)
	} else {
		if debugConnectionReuse {
			fmt.Println("Conn.Open() REUSE")
		}
		conn = connObj.(Conn)
		if conn == nil {
			panic("nil conn")
		}
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
	err = conn.Query(cmd, vv, QueryImplicit, IsoLevelDefault, &res.val)
	if err != nil {
		return res, err
	}

	return res, nil
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
