package rdb

type Result struct {
	Code int
	ErrorList
}

// Arrange for (*Command).Scan[...].V to hold fresh pointer each time.
// Returns io.EOF when last row has been scanned.
func (res *Result) ScanPrep() (eof bool, err error) {
	return true, nil
}
func (res *Result) Prep(name string, value interface{}) *Result {
	return res
}
func (res *Result) PrepAll(values ...interface{}) *Result {
	return res
}

// Scans the row into a buffer that can be fetched
// Returns io.EOF when last row has been scanned.
func (res *Result) ScanBuffer() (eof bool, err error) {
	return true, nil
}

// Use with ScanBuffer[M]().
func (res *Result) Get(name string) interface{} {
	return nil
}

// Use with ScanBuffer[M]().
func (res *Result) Getx(index int) interface{} {
	return nil
}

func (res *Result) Schema() *Schema {
	return nil
}

// TODO: Fill out schema.
type Schema struct {
	Columns []Param
}
