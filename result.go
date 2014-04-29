package rdb

type Result struct {
	Code int
}

// Arrange for (*Command).Scan[...].V to hold fresh pointer each time.
// Returns io.EOF when last row has been scanned.
func (res *Result) ScanPrep() (eof bool, err error) {
	return true, nil
}

// Use with ScanPrep().
func (res *Result) Prep(name string, value interface{}) error {
	return nil
}

// Use with ScanPrep().
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

// Get the result schema.
func (res *Result) Schema() (*Schema, error) {
	return nil, nil
}

// TODO: Fill out schema.
type Schema struct {
	Columns []SqlColumn
}
