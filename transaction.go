// Copyright 2014 Daniel Theophanes.
// Use of this source code is governed by a zlib-style
// license that can be found in the LICENSE file.

package rdb

// Although nested transactions are unsupported, savepoints are supported.
// A transaction should end with either a Commit() or Rollback() call.
type Transaction struct {
	done bool
}

func (tran *Transaction) Query(cmd *Command, params ...Param) (*Result, error) {
	return nil, nil
}
func (tran *Transaction) Commit() error {
	tran.done = true
	return nil
}
func (tran *Transaction) Rollback() error {
	tran.done = true
	return nil
}

// Rollback to an existing savepoint. Commit or Rollback should still
// be called after calling RollbackTo.
func (tran *Transaction) RollbackTo(savepoint string) error {
	return nil
}

// Create a save point in the transaction.
func (tran *Transaction) SavePoint(name string) error {
	return nil
}

// Return true if the transaction has not been either commited or entirely rolled back.
func (tran *Transaction) Active() bool {
	return !tran.done
}

// Get the panic'ing version that doesn't return errors.
func (tran *Transaction) Must() TransactionMust {
	return TransactionMust{norm: tran}
}
