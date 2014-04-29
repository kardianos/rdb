package rdb

import (
	"bitbucket.org/kardianos/rdb/driver"
)

type Database struct {
}

func Open(c *driver.Config) (*Database, error) {
	return nil, nil
}

func (db *Database) Close() error {
	return nil
}

// Input parameter values can either be specified in the paremeter definition
// or on each query. If the value is not put in the parameter definition
// then the command instance may be reused for every query.
func (db *Database) Query(cmd *driver.Command, vv ...driver.Value) (*Result, error) {
	return nil, nil
}

// Same as Query but will panic on an error.
func (db *Database) Transaction(iso driver.IsolationLevel) (*Transaction, error) {
	return nil, nil
}

type Transaction struct {
}

// Input parameter values can either be specified in the paremeter definition
// or on each query. If the value is not put in the parameter definition
// then the command instance may be reused for every query.
func (db *Transaction) Query(cmd *driver.Command, vv ...driver.Value) (*Result, error) {
	return nil, nil
}

func (db *Transaction) Commit() error {
	return nil
}
func (db *Transaction) Rollback() error {
	return nil
}

/*
// Map columns to (*Command).Input, for each row map values.
func (db *Database) BulkInsert(cmd *Command) *BulkInsert {
	return nil
}

type BulkInsert struct {
	BatchSize int
}

func (bi *BulkInsert) NextRow() error {
	return nil
}

func (bi *BulkInsert) Done() error {
	return nil
}
*/
