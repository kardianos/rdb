// Copyright 2014 Daniel Theophanes.
// Use of this source code is governed by a zlib-style
// license that can be found in the LICENSE file.

package rdb

import (
	"context"
	"errors"
)

// Although nested transactions are unsupported, savepoints are supported.
// A transaction should end with either a Commit() or Rollback() call.
type Transaction struct {
	ctx  context.Context
	cp   *ConnPool
	conn DriverConn

	done  bool
	level IsolationLevel
}

var errTransactionClosed = errors.New("transaction already closed")

func (tran *Transaction) Query(ctx context.Context, cmd *Command, params ...Param) (*Result, error) {
	if tran.done {
		return nil, errTransactionClosed
	}
	return tran.cp.query(ctx, true, tran.conn, cmd, nil, params...)
}

// Commit commits a one or more queries. If no queries have been run this
// just returns the connection without any action being taken.
func (tran *Transaction) Commit() error {
	if tran.done {
		return errTransactionClosed
	}
	tran.done = true
	err := tran.conn.Commit(tran.ctx)
	tran.cp.releaseConn(tran.conn, tran.conn.Status() != StatusReady)
	return err
}

// Rollback rolls back one or more queries. If no queries have been run this
// just returns the connection without any action being taken.
func (tran *Transaction) Rollback() error {
	return tran.RollbackTo("")
}

// Rollback to an existing savepoint. Commit or Rollback should still
// be called after calling RollbackTo.
func (tran *Transaction) RollbackTo(savepoint string) error {
	if tran.done {
		return errTransactionClosed
	}

	err := tran.conn.Rollback(savepoint)
	if len(savepoint) == 0 {
		tran.done = true
		tran.cp.releaseConn(tran.conn, tran.conn.Status() != StatusReady)
	}
	return err
}

// Create a save point in the transaction.
func (tran *Transaction) SavePoint(name string) error {
	if tran.done {
		return errTransactionClosed
	}
	return tran.conn.SavePoint(tran.ctx, name)
}

// Return true if the transaction has not been either commited or entirely rolled back.
func (tran *Transaction) Active() bool {
	return !tran.done
}
