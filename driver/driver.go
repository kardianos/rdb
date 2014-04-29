package driver

import ()

type DriverServer interface {
	Open(c *Config) (DriverDatabase, error)
	DriverMetaInfo(driverName string) *DriverMeta
}

type DriverDatabase interface {
	Close() error
	Query(cmd *Command, vv ...Value) (DriverResult, error)
	Transaction(iso IsolationLevel) (DriverTransaction, error)
}

type DriverTransaction interface {
	Query(cmd *Command, vv ...Value) (DriverResult, error)
	Commit() error
	Rollback() error
}

type DriverResult interface {
	Close() error
	Schema() (*Schema, error)

	ScanPrep() (eof bool, err error)
	Prep(name string, value interface{})
	PrepAll(values ...interface{}) error

	ScanBuffer() (eof bool, err error)
	Get(name string) (interface{}, error)
	Getx(index int) (interface{}, error)
}
