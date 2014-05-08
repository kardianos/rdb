// Copyright 2014 Daniel Theophanes.
// Use of this source code is governed by a zlib-style
// license that can be found in the LICENSE file.

package rdb

import (
	"net/url"
)

type QueryType byte

const (
	QueryImplicit QueryType = iota
	QueryPrepare            // Create a prepared transaction.
	QueryBegin              // Transaction
)

type ConnStatus byte

const (
	StatusDisconnected ConnStatus = iota
	StatusReady
	StatusQuery
	StatusBulkCopy
)

// Type the database driver must implement.
type Driver interface {
	// Open a database. An actual connection does not need to be established
	// at this time.
	Open(c *Config) (Conn, error)

	// Return information about the database drivers capabilities.
	// Should not reflect any actual server any connections to it.
	DriverMetaInfo() *DriverMeta

	// Return the command to send a NOOP to the server.
	PingCommand() *Command

	// Parse driver specific options into the configuration.
	ParseOptions(KV map[string]interface{}, configOptions url.Values) error
}

// Value type used by the driver to report a field value.
// If a long field, such as a long byte array, it can be chunked
// directly into destination. If the driver is copying from a common
// buffer then the MustCopy field must be true so it is known it must be
// copied out.
type DriverValue struct {
	Value    interface{}
	Null     bool
	MustCopy bool // If the Value is a common driver buffer, set to true.
	More     bool // True if more data is expected for the field.
	Chunked  bool // True if data is sent in chunks.
}

type Conn interface {
	Close()
	ConnectionInfo() (*ConnectionInfo, error)
	Scan() error
	Query(*Command, []Value, QueryType, IsolationLevel, Valuer) error
	// Rollback(savepoint string) error
	// Commit() error
	// SavePoint(name string) error
	Status() ConnStatus
}

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

// Returned from GetN and GetxN.
// Represents a nullable type.
type Nullable struct {
	Null bool        // True if value is null.
	V    interface{} // Value, if any present.
}
