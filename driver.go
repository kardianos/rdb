// Copyright 2014 Daniel Theophanes.
// Use of this source code is governed by a zlib-style
// license that can be found in the LICENSE file.

package rdb

import (
	"net/url"
)

// TODO: Add states for transactions.
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
	// Close the underlying connection to the database.
	Close()

	// Return version information regarding the currently connected server.
	ConnectionInfo() (*ConnectionInfo, error)

	// Read the next row from the connection. For each field in the row
	// call the Valuer.WriteField(...) method. Propagate the reportRow field.
	Scan(reportRow bool) error

	// The isolation level is set by the command.
	// Should return "PreparedTokenNotValid" if the preparedToken was not recognized.
	Query(cmd *Command, params []Param, preparedToken interface{}, val Valuer) error

	Status() ConnStatus

	// Happy Path:
	//  * Interface wants to prepare command, but doesn't have token.
	//  * Interface sends a conn Prepare requests and gets a token.
	//  * Interface uses token in query.
	//  * After some use, Interface unprepares command.
	//
	// Re-prepare Path:
	//  * Interface has a token and attempts to use it in query.
	//  * Query returns "the token is not valid" error (server restart?).
	//  * Interface attempts to re-prepare query.
	//  * Interface uses new token in Query. If that fails again, it should fail the query.
	Prepare(*Command) (preparedToken interface{}, err error)
	Unprepare(preparedToken interface{}) (err error)

	Begin() error
	Rollback(savepoint string) error
	Commit() error
	SavePoint(name string) error
}
