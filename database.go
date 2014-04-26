package rdb

type Marshal struct {
}

type Command struct {
	Sql    string
	Zero   bool
	One    bool
	Input  []Param
	Output []*Marshal
}

type Database struct {
}

func Open(c *Config) (*Database, error) {
	return nil, nil
}

func (db *Database) Close() error {
	return nil
}

func (db *Database) Query(*Command, ...*Value) (*Result, error) {
	return nil, nil
}

func (db *Database) QueryM(*Command, ...*Value) *Result {
	return nil
}

// Map columns to (*Command).Input, for each row map values.
func (db *Database) BulkInsert(*Command) *BulkInsert {
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
