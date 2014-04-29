package rdb

type IsolationLevel byte

const (
	IsoLevelDefault IsolationLevel = iota
	IsoLevelReadUncommited
	IsoLevelReadCommited
	IsoLevelWriteCommited
	IsoLevelRepeatableRead
	IsoLevelSerializable
	IsoLevelSnapshot
)

type Arity byte

// The number of rows to expect from a command.
const (
	Many Arity = iota
	One
	Zero
	OneOnly
	ZeroOnly
)

// Command represents a SQL command and can be used from many different
// queries at the same time, so long as the input parameter values
// "Input[N].V (Value)" are not set in the Param struct but passed in with
// the actual query as Value.
type Command struct {
	// The SQL to be used in the command.
	Sql string

	// Number of rows expected.
	//   If Arity is One or OneOnly, only the first row is returned.
	//   If Arity is OneOnly, if more results are returned an error is returned.
	//   If Arity is Zero or ZeroOnly, no rows are returned.
	//   If Arity is ZeroOnnly, if any results are returned an error is returned.
	Arity Arity
	Input []Param

	// Optional fields to specify output marshal.
	Output []*Field

	// If set to true silently truncates text longer then the field.
	// If this is set to false text truncation will result in an error.
	TruncLongText bool

	// Optional name of the command. May be used if logging.
	Name string
}

type Database struct {
}

func Open(c *Config) (*Database, error) {
	return nil, nil
}

func (db *Database) Close() error {
	return nil
}

// Input parameter values can either be specified in the paremeter definition
// or on each query. If the value is not put in the parameter definition
// then the command instance may be reused for every query.
func (db *Database) Query(cmd *Command, vv ...Value) (*Result, error) {
	return nil, nil
}

// Same as Query but will panic on an error.
func (db *Database) Transaction(iso IsolationLevel) (*Transaction, error) {
	return nil, nil
}

type Transaction struct {
}

// Input parameter values can either be specified in the paremeter definition
// or on each query. If the value is not put in the parameter definition
// then the command instance may be reused for every query.
func (db *Transaction) Query(cmd *Command, vv ...Value) (*Result, error) {
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
