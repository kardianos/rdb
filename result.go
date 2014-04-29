package rdb

import (
	"bitbucket.org/kardianos/rdb/driver"
)

type Result struct {
}

// For each needed field, call Prep() or PrepAll() to prepare
// value pointers for scanning. To scan prepared fields call ScanPrep().
func (res *Result) ScanPrep() (eof bool, err error) {
	return true, nil
}

// Prepare pointers to values to be populated by name using Prep. After
// preparing call ScanPrep().
func (res *Result) Prep(name string, value interface{}) error {
	return nil
}

// Prepare pointers to values to be populated by index using Prep. After
// preparing call ScanPrep().
func (res *Result) PrepAll(values ...interface{}) error {
	return nil
}

// Scans the row into a buffer that can be fetched
// Returns io.EOF when last row has been scanned.
func (res *Result) ScanBuffer() (eof bool, err error) {
	return true, nil
}

// Use with ScanBuffer().
func (res *Result) Get(name string) (interface{}, error) {
	return nil, nil
}

// Use with ScanBuffer().
func (res *Result) Getx(index int) (interface{}, error) {
	return nil, nil
}

// Fetch the table schema.
func (res *Result) Schema() (*driver.Schema, error) {
	return nil, nil
}
