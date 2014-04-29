package rdb

type Result struct {
	Code int
}

// Arrange for (*Command).Scan[...].V to hold fresh pointer each time.
// Returns io.EOF when last row has been scanned.
func (res *Result) ScanPrep() (eof bool, err error) {
	return true, nil
}
func (res *Result) Prep(name string, value interface{}) error {
	return nil
}
func (res *Result) PrepAll(values ...interface{}) error {
	return nil
}

// Scans the row into a buffer that can be fetched
// Returns io.EOF when last row has been scanned.
func (res *Result) ScanBuffer() (eof bool, err error) {
	return true, nil
}

// Use with ScanBuffer[M]().
func (res *Result) Get(name string) (interface{}, error) {
	return nil, nil
}

// Use with ScanBuffer[M]().
func (res *Result) Getx(index int) (interface{}, error) {
	return nil, nil
}

func (res *Result) Schema() (*Schema, error) {
	return nil, nil
}

// TODO: Fill out schema.
type Schema struct {
	Columns []Param
}

type ResultMust struct {
	r *Result
}

// Arrange for (*Command).Scan[...].V to hold fresh pointer each time.
// Returns io.EOF when last row has been scanned.
func (res *ResultMust) ScanPrep() (eof bool) {
	return true
}
func (res *ResultMust) Prep(name string, value interface{}) *ResultMust {
	return res
}
func (res *ResultMust) PrepAll(values ...interface{}) *ResultMust {
	return res
}

// Scans the row into a buffer that can be fetched
// Returns io.EOF when last row has been scanned.
func (res *ResultMust) ScanBuffer() (eof bool) {
	return true
}

// Use with ScanBuffer[M]().
func (res *ResultMust) Get(name string) interface{} {
	return nil
}

// Use with ScanBuffer[M]().
func (res *ResultMust) Getx(index int) interface{} {
	return nil
}

func (res *ResultMust) Schema() *Schema {
	return nil
}
