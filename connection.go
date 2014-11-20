// Copyright 2014 Daniel Theophanes.
// Use of this source code is governed by a zlib-style
// license that can be found in the LICENSE file.

package rdb

import (
	"errors"
)

type Connection struct {
	cp   *ConnPool
	conn DriverConn
	done bool
}

var connectionClosed = errors.New("Connection already closed.")

// Query executes a Command on the connection.
func (c *Connection) Query(cmd *Command, params ...Param) (*Result, error) {
	if c.done {
		return nil, connectionClosed
	}
	return c.cp.query(true, c.conn, cmd, nil, params...)
}

// Close returns the underlying connection to the Connection Pool.
func (c *Connection) Close() error {
	if c.done {
		return transactionClosed
	}
	c.done = true
	c.cp.releaseConn(c.conn, c.conn.Status() != StatusReady)
	return nil
}

// Return true if the connection has not been closed.
func (c *Connection) Active() bool {
	return !c.done
}
