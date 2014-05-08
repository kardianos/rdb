// Copyright 2014 Daniel Theophanes.
// Use of this source code is governed by a zlib-style
// license that can be found in the LICENSE file.

package rdb

// The Transaction API is unstable.
// Represents a transaction in progress.
type Transaction struct {
}

func (tran *Transaction) Query(cmd *Command, vv ...Value) (*Result, error) {
	return nil, nil
}
func (tran *Transaction) Commit() error {
	return nil
}
func (tran *Transaction) Rollback() error {
	return nil
}

// Get the panic'ing version that doesn't return errors.
func (tran *Transaction) Must() TransactionMust {
	return TransactionMust{}
}
