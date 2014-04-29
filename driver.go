package rdb

type Driver interface {
	Open(c *Config) (Database, error)
	DriverMetaInfo() *DriverMeta
}

type Database interface {
	Close() error
	Query(cmd *Command, vv ...Value) (Result, error)
	Transaction(iso IsolationLevel) (Transaction, error)
}

type Transaction interface {
	Query(cmd *Command, vv ...Value) (Result, error)
	Commit() error
	Rollback() error
}

type Result interface {
	Close() error

	// Fetch the table schema.
	Schema() (*Schema, error)

	// Prepare pointers to values to be populated by name using Prep. After
	// preparing call ScanPrep().
	Prep(name string, value interface{}) error

	// Prepare pointers to values to be populated by index using Prep. After
	// preparing call ScanPrep().
	Prepx(index int, value interface{}) error

	// Prepare pointers to values to be populated by index using Prep. After
	// preparing call ScanPrep().
	PrepAll(values ...interface{}) error

	// Scans the row into a buffer that can be fetched
	// Returns io.EOF when last row has been scanned.
	Scan() (eof bool, err error)

	// Use with ScanBuffer().
	Get(name string) (interface{}, error)

	// Use with ScanBuffer().
	Getx(index int) (interface{}, error)
}
