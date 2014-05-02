// Copyright 2014 Daniel Theophanes.
// Use of this source code is governed by a zlib-style
// license that can be found in the LICENSE file.

package rdb

import (
	"net/url"
)

// Type the database driver must implement.
type Driver interface {
	// Open a database. An actual connection does not need to be established
	// at this time.
	Open(c *Config) (Database, error)

	// Return information about the database drivers capabilities.
	// Should not reflect any actual server any connections to it.
	DriverMetaInfo() *DriverMeta

	ParseOptions(KV map[string]interface{}, configOptions url.Values) error
}

// Represents a connection or connection configuration to a database.
type Database interface {
	Close() error

	// Will attempt to connect to the database and disconnect.
	// Must not impact any existing connections.
	Ping() error

	// Returns the information specific to the connection.
	// May call Ping() if there has not yet been a connection.
	ConnectionInfo() (*ConnectionInfo, error)

	// Perform a query against the database.
	// If values are not specified in the Command.Input[...].V, then they
	// may be specified in the Value. Order may be used to match the
	// existing parameters if the Value.N name is omitted.
	Query(cmd *Command, vv ...Value) (Result, error)

	// API for tranactions are preliminary. Not a stable API call.
	Transaction(iso IsolationLevel) (Transaction, error)

	// Get the panic'ing version that doesn't return errors.
	Must() DatabaseMust
}

// The Transaction API is unstable.
// Represents a transaction in progress.
type Transaction interface {
	Query(cmd *Command, vv ...Value) (Result, error)
	Commit() error
	Rollback() error

	// Get the panic'ing version that doesn't return errors.
	Must() TransactionMust
}

// Returned from GetN and GetxN.
// Represents a nullable type.
type Nullable struct {
	Null bool        // True if value is null.
	V    interface{} // Value, if any present.
}

// Manages the life cycle of a query.
// The result must automaticly Close() if the command Arity is Zero after
// execution or after the first Scan() if Arity is One.
type Result interface {
	Close() error

	// Fetch the table schema.
	Schema() (*Schema, error)

	// Prepare pointers to values to be populated by name using Prep. After
	// preparing call Scan().
	Prep(name string, value interface{}) error

	// Prepare pointers to values to be populated by index using Prep. After
	// preparing call Scan().
	Prepx(index int, value interface{}) error

	// Prepare pointers to values to be populated by index using Prep. After
	// preparing call Scan().
	PrepAll(values ...interface{}) error

	// Scans the row into a buffer that can be fetched with Get and scans
	// directly into any prepared values.
	// Return value "more" is false if no more rows.
	Scan() (more bool, err error)

	// Use after Scan(). Can only pull fields which have not already been sent
	// into a prepared value.
	Get(name string) (interface{}, error)

	// Use after Scan(). Can only pull fields which have not already been sent
	// into a prepared value.
	Getx(index int) (interface{}, error)

	// Use after Scan(). Can only pull fields which have not already been sent
	// into a prepared value.
	GetN(name string) (Nullable, error)

	// Use after Scan(). Can only pull fields which have not already been sent
	// into a prepared value.
	GetxN(index int) (Nullable, error)

	// Get the panic'ing version that doesn't return errors.
	Must() ResultMust
}
