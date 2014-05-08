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
	Close() error
	ConnectionInfo() (*ConnectionInfo, error)
	Scan() error
	Query(*Command, []Value, QueryType, IsolationLevel, Valuer) error
	// Rollback(savepoint string) error
	// Commit() error
	// SavePoint(name string) error
	Status() ConnStatus
}

// Represents a connection or connection configuration to a database.
type ConnPool struct {
	dr   Driver
	conf *Config
}

func (cp *ConnPool) Close() error {
	// Close all active connections.
	return nil
}

// Will attempt to connect to the database and disconnect.
// Must not impact any existing connections.
func (cp *ConnPool) Ping() error {
	return nil
}

// Returns the information specific to the connection.
// May call Ping() if there has not yet been a connection.
func (cp *ConnPool) ConnectionInfo() (*ConnectionInfo, error) {
	// Cache on first connection, then pull from that cache.
	return nil, nil
}

// Perform a query against the database.
// If values are not specified in the Command.Input[...].V, then they
// may be specified in the Value. Order may be used to match the
// existing parameters if the Value.N name is omitted.
func (cp *ConnPool) Query(cmd *Command, vv ...Value) (*Result, error) {
	// TODO: Use actual pool.
	// For now, ignore any pooling option.
	conn, err := cp.dr.Open(cp.conf)
	if err != nil {
		return nil, err
	}
	res := &Result{
		conn: conn,
	}
	err = conn.Query(cmd, vv, QueryImplicit, IsoLevelDefault, &res.val)
	if err != nil {
		return res, err
	}

	fields := make([]*Field, len(cmd.Output))
	for i := range cmd.Output {
		fields[i] = &cmd.Output[i]
	}

	res.val.initFields = fields

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
