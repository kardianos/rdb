// Copyright 2014 Daniel Theophanes.
// Use of this source code is governed by a zlib-style
// license that can be found in the LICENSE file.

package rdb

type ResultMust struct {
	norm *Result
}

type ConnPoolMust struct {
	norm *ConnPool
}

type TransactionMust struct {
	norm *Transaction
}

// Get the panic'ing version that doesn't return errors.
func (cp *ConnPool) Must() ConnPoolMust {
	return ConnPoolMust{norm: cp}
}

// Get the panic'ing version that doesn't return errors.
func (r *Result) Must() ResultMust {
	return ResultMust{norm: r}
}

// Get the panic'ing version that doesn't return errors.
func (tran *Transaction) Must() TransactionMust {
	return TransactionMust{norm: tran}
}

// Get the non-panic'ing version of Result.
func (must ResultMust) Normal() *Result {
	return must.norm
}

// Get the non-panic'ing version of Database.
func (must ConnPoolMust) Normal() *ConnPool {
	return must.norm
}

// Get the non-panic'ing version of Transaction.
func (must TransactionMust) Normal() *Transaction {
	return must.norm
}

// Same as ParseConfig() but all errors are returned as a panic(MustError{}).
func ParseConfigMust(connectionString string) *Config {
	config, err := ParseConfig(connectionString)
	if err != nil {
		panic(MustError{Err: err})
	}
	return config
}

// Same as Open() but all errors are returned as a panic(MustError{}).
func OpenMust(c *Config) ConnPoolMust {
	db, err := Open(c)
	if err != nil {
		panic(MustError{Err: err})
	}
	return ConnPoolMust{
		norm: db,
	}
}

func (must ConnPoolMust) Close() {
	must.norm.Close()
}

func (must ConnPoolMust) Ping() {
	err := must.norm.Ping()
	if err != nil {
		panic(MustError{Err: err})
	}
}
func (must ConnPoolMust) ConnectionInfo() *ConnectionInfo {
	ci, err := must.norm.ConnectionInfo()
	if err != nil {
		panic(MustError{Err: err})
	}
	return ci
}

// Input parameter values can either be specified in the paremeter definition
// or on each query. If the value is not put in the parameter definition
// then the command instance may be reused for every query.
func (must ConnPoolMust) Query(cmd *Command, params ...Param) ResultMust {
	res, err := must.norm.Query(cmd, params...)
	if err != nil {
		panic(MustError{Err: err})
	}
	return ResultMust{
		norm: res,
	}
}

// Same as Query but will panic on an error.
func (must ConnPoolMust) Begin() TransactionMust {
	tran, err := must.norm.Begin()
	if err != nil {
		panic(MustError{Err: err})
	}
	return TransactionMust{
		norm: tran,
	}
}

// Same as Query but will panic on an error.
func (must ConnPoolMust) BeginLevel(level IsolationLevel) TransactionMust {
	tran, err := must.norm.BeginLevel(level)
	if err != nil {
		panic(MustError{Err: err})
	}
	return TransactionMust{
		norm: tran,
	}
}

// Input parameter values can either be specified in the paremeter definition
// or on each query. If the value is not put in the parameter definition
// then the command instance may be reused for every query.
func (must TransactionMust) Query(cmd *Command, params ...Param) ResultMust {
	res, err := must.norm.Query(cmd, params...)
	if err != nil {
		panic(MustError{Err: err})
	}
	return ResultMust{
		norm: res,
	}
}

func (must TransactionMust) Commit() {
	err := must.norm.Commit()
	if err != nil {
		panic(MustError{Err: err})
	}
}
func (must TransactionMust) Rollback() {
	err := must.norm.Rollback()
	if err != nil {
		panic(MustError{Err: err})
	}
}
func (must TransactionMust) RollbackTo(savepoint string) {
	err := must.norm.RollbackTo(savepoint)
	if err != nil {
		panic(MustError{Err: err})
	}
}
func (must TransactionMust) SavePoint(name string) {
	err := must.norm.SavePoint(name)
	if err != nil {
		panic(MustError{Err: err})
	}
}
func (must TransactionMust) Active() bool {
	return must.norm.Active()
}

// Make sure the result is closed.
func (must ResultMust) Close() {
	err := must.norm.Close()
	if err != nil {
		panic(MustError{Err: err})
	}
}

func (must ResultMust) Next() (more bool) {
	return must.norm.Next()
}

// For each needed field, call Prep() or PrepAll() to prepare
// value pointers for scanning. To scan prepared fields call Scan().
// Call Scan() before using Get() or Getx().
// Returns false if no more rows.
func (must ResultMust) Scan(values ...interface{}) ResultMust {
	err := must.norm.Scan(values...)
	if err != nil {
		panic(MustError{Err: err})
	}
	return must
}

// Informational messages. Do not call concurrently with Scan() or Done().
func (must ResultMust) Info() []*SqlMessage {
	return must.norm.Info()
}

// Prepare pointers to values to be populated by name using Prep. After
// preparing call Scan().
func (must ResultMust) Prep(name string, value interface{}) ResultMust {
	must.norm.Prep(name, value)
	return must
}

// Prepare pointers to values to be populated by index using Prep. After
// preparing call Scan().
func (must ResultMust) Prepx(index int, value interface{}) ResultMust {
	must.norm.Prepx(index, value)
	return must
}

// Use after Scan(). Can only pull fields which have not already been sent
// into a prepared value.
func (must ResultMust) Get(name string) interface{} {
	return must.norm.Get(name)
}

// Use after Scan(). Can only pull fields which have not already been sent
// into a prepared value.
func (must ResultMust) Getx(index int) interface{} {
	return must.norm.Getx(index)
}

// Use after Scan(). Can only pull fields which have not already been sent
// into a prepared value.
func (must ResultMust) GetN(name string) Nullable {
	return must.norm.GetN(name)
}

// Use after Scan(). Can only pull fields which have not already been sent
// into a prepared value.
func (must ResultMust) GetxN(index int) Nullable {
	return must.norm.GetxN(index)
}

// Use after Scan(). Can only pull fields which have not already been sent
// into a prepared value. Not all fields will be populated if some have
// been prepared.
func (must ResultMust) GetRowN() []Nullable {
	return must.norm.GetRowN()
}

// Fetch the table schema.
func (must ResultMust) Schema() []*SqlColumn {
	return must.norm.Schema()
}
