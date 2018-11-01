// Copyright 2014 Daniel Theophanes.
// Use of this source code is governed by a zlib-style
// license that can be found in the LICENSE file.

package rdb

import (
	"bitbucket.org/kardianos/rdb/semver"
)

// TODO: Add states for transactions.
type DriverConnStatus byte

const (
	StatusDisconnected DriverConnStatus = iota
	StatusReady
	StatusQuery
	StatusResultDone
	StatusBulkCopy
)

type DriverOption struct {
	Name string

	Description string
	Parse       func(input string) (interface{}, error)
}

type DriverSupport struct {
	// PreparePerConn is set to true if prepared statements are local to
	// each connection. Set to false if prepared statements are global.
	PreparePerConn bool

	NamedParameter   bool // Supports named parameters.
	FluidType        bool // Like SQLite.
	MultipleResult   bool // Supports returning multiple result sets.
	SecureConnection bool // Supports a secure connection.
	BulkInsert       bool // Supports a fast bulk insert method.
	Notification     bool // Supports driver notifications.
	UserDataTypes    bool // Handles user supplied data types.
}

type DriverInfo struct {
	Options []*DriverOption
	DriverSupport
}

type ConnectionInfo struct {
	Server, Protocol *semver.Version
}

// Driver is implemented by the database driver.
type Driver interface {
	// Open a database. An actual connection does not need to be established
	// at this time.
	Open(c *Config) (DriverConn, error)

	// Return information about the database drivers capabilities.
	// Should not reflect any actual server any connections to it.
	DriverInfo() *DriverInfo

	// Return the command to send a NOOP to the server.
	PingCommand() *Command
}

// DriverValue used by the driver to report a field value.
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

// Conn represents a database driver connection.
type DriverConn interface {
	// Close the underlying connection to the database.
	Close()

	Available() bool             // True if not currently in a connection pool.
	SetAvailable(available bool) // Set when adding or removing from connection pool.

	// Return version information regarding the currently connected server.
	ConnectionInfo() *ConnectionInfo

	// Read the next row from the connection. For each field in the row
	// call the Valuer.WriteField(...) method. Propagate the reportRow field.
	Scan() error

	// NextResult advances to the next result if there are multiple results.
	NextResult() (more bool, err error)

	// NextQuery stops the active query and gets the connection for the next one.
	NextQuery() (err error)

	// The isolation level is set by the command.
	// Should return "PreparedTokenNotValid" if the preparedToken was not recognized.
	Query(cmd *Command, params []Param, preparedToken interface{}, val DriverValuer) error

	Status() DriverConnStatus

	// Reset the connection to be ready for next connection.
	Reset(*Config) error

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

	Begin(iso IsolationLevel) error
	Rollback(savepoint string) error
	Commit() error
	SavePoint(name string) error
}
