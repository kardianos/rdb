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

const (
	Many Arity = iota
	One
	Zero
	OneOnly
	ZeroOnly
)

type Command struct {
	// The SQL to be used in the command.
	Sql string

	// Number of rows expected.
	//   If Arity is One or OneOnly, only the first row is returned.
	//   If Arity is OneOnly, if more results are returned an error is returned.
	//   If Arity is Zero or ZeroOnly, no rows are returned.
	//   If Arity is ZeroOnnly, if any results are returned an error is returned.
	Arity  Arity
	Input  []Param
	Output []*Field
}

type Database struct {
}

func Open(c *Config) (*Database, error) {
	return nil, nil
}

func (db *Database) Close() error {
	return nil
}

type DatabaseMust struct {
}

func OpenMust(c *Config) *DatabaseMust {
	return nil
}

func (db *DatabaseMust) Close() {
	return
}

// Input parameter values can either be specified in the paremeter definition
// or on each query. If the value is not put in the parameter definition
// then the command instance may be reused for every query.
func (db *Database) Query(cmd *Command, vv ...Value) (*Result, error) {
	return nil, nil
}

// Same as Query but will panic on an error.
func (db *Database) Transaction(iso IsolationLevel) *Transaction {
	return nil
}

// Input parameter values can either be specified in the paremeter definition
// or on each query. If the value is not put in the parameter definition
// then the command instance may be reused for every query.
func (db *DatabaseMust) Query(cmd *Command, vv ...Value) *ResultMust {
	return nil
}

// Same as Query but will panic on an error.
func (db *DatabaseMust) Transaction(iso IsolationLevel) *Transaction {
	return nil
}

type Transaction struct {
}
type TransactionMust struct {
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

// Input parameter values can either be specified in the paremeter definition
// or on each query. If the value is not put in the parameter definition
// then the command instance may be reused for every query.
func (db *TransactionMust) Query(cmd *Command, vv ...Value) *ResultMust {
	return nil
}

func (db *TransactionMust) Commit() {
	return
}
func (db *TransactionMust) Rollback() {
	return
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
