// Copyright 2014 Daniel Theophanes.
// Use of this source code is governed by a zlib-style
// license that can be found in the LICENSE file.

package rdb

type ResultMust struct {
	NormalResult Result
}

type DatabaseMust struct {
	NormalDatabase Database
}

type TransactionMust struct {
	NormalTransaction Transaction
}

// Get the non-panic'ing version of Result.
func (must ResultMust) Normal() Result {
	return must.NormalResult
}

// Get the non-panic'ing version of Database.
func (must DatabaseMust) Normal() Database {
	return must.NormalDatabase
}

// Get the non-panic'ing version of Transaction.
func (must TransactionMust) Normal() Transaction {
	return must.NormalTransaction
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
func OpenMust(c *Config) DatabaseMust {
	db, err := Open(c)
	if err != nil {
		panic(MustError{Err: err})
	}
	return DatabaseMust{
		NormalDatabase: db,
	}
}

func (must DatabaseMust) Close() {
	err := must.NormalDatabase.Close()
	if err != nil {
		panic(MustError{Err: err})
	}
}

func (must DatabaseMust) Ping() {
	err := must.NormalDatabase.Ping()
	if err != nil {
		panic(MustError{Err: err})
	}
}
func (must DatabaseMust) ConnectionInfo() *ConnectionInfo {
	ci, err := must.NormalDatabase.ConnectionInfo()
	if err != nil {
		panic(MustError{Err: err})
	}
	return ci
}

// Input parameter values can either be specified in the paremeter definition
// or on each query. If the value is not put in the parameter definition
// then the command instance may be reused for every query.
func (must DatabaseMust) Query(cmd *Command, vv ...Value) ResultMust {
	res, err := must.NormalDatabase.Query(cmd, vv...)
	if err != nil {
		panic(MustError{Err: err})
	}
	return ResultMust{
		NormalResult: res,
	}
}

// Same as Query but will panic on an error.
func (must DatabaseMust) Transaction(iso IsolationLevel) TransactionMust {
	tran, err := must.NormalDatabase.Transaction(iso)
	if err != nil {
		panic(MustError{Err: err})
	}
	return TransactionMust{
		NormalTransaction: tran,
	}
}

// Input parameter values can either be specified in the paremeter definition
// or on each query. If the value is not put in the parameter definition
// then the command instance may be reused for every query.
func (must TransactionMust) Query(cmd *Command, vv ...Value) ResultMust {
	res, err := must.NormalTransaction.Query(cmd, vv...)
	if err != nil {
		panic(MustError{Err: err})
	}
	return ResultMust{
		NormalResult: res,
	}
}

func (must TransactionMust) Commit() {
	err := must.NormalTransaction.Commit()
	if err != nil {
		panic(MustError{Err: err})
	}
}
func (must TransactionMust) Rollback() {
	err := must.NormalTransaction.Rollback()
	if err != nil {
		panic(MustError{Err: err})
	}
}

// Make sure the result is closed.
func (must ResultMust) Close() {
	err := must.NormalResult.Close()
	if err != nil {
		panic(MustError{Err: err})
	}
}

// For each needed field, call Prep() or PrepAll() to prepare
// value pointers for scanning. To scan prepared fields call Scan().
// Call Scan() before using Get() or Getx().
// Returns false if no more rows.
func (must ResultMust) Scan() (more bool) {
	eof, err := must.NormalResult.Scan()
	if err != nil {
		panic(MustError{Err: err})
	}
	return eof
}

// Prepare pointers to values to be populated by name using Prep. After
// preparing call Scan().
func (must ResultMust) Prep(name string, value interface{}) ResultMust {
	err := must.NormalResult.Prep(name, value)
	if err != nil {
		panic(MustError{Err: err})
	}
	return must
}

// Prepare pointers to values to be populated by index using Prep. After
// preparing call Scan().
func (must ResultMust) Prepx(index int, value interface{}) ResultMust {
	err := must.NormalResult.Prepx(index, value)
	if err != nil {
		panic(MustError{Err: err})
	}
	return must
}

// Prepare pointers to values to be populated by index using Prep. After
// preparing call Scan().
func (must ResultMust) PrepAll(values ...interface{}) ResultMust {
	err := must.NormalResult.PrepAll(values...)
	if err != nil {
		panic(MustError{Err: err})
	}
	return must
}

// Use after Scan(). Can only pull fields which have not already been sent
// into a prepared value.
func (must ResultMust) Get(name string) interface{} {
	value, err := must.NormalResult.Get(name)
	if err != nil {
		panic(MustError{Err: err})
	}
	return value
}

// Use after Scan(). Can only pull fields which have not already been sent
// into a prepared value.
func (must ResultMust) Getx(index int) interface{} {
	value, err := must.NormalResult.Getx(index)
	if err != nil {
		panic(MustError{Err: err})
	}
	return value
}

// Use after Scan(). Can only pull fields which have not already been sent
// into a prepared value.
func (must ResultMust) GetN(name string) Nullable {
	value, err := must.NormalResult.GetN(name)
	if err != nil {
		panic(MustError{Err: err})
	}
	return value
}

// Use after Scan(). Can only pull fields which have not already been sent
// into a prepared value.
func (must ResultMust) GetxN(index int) Nullable {
	value, err := must.NormalResult.GetxN(index)
	if err != nil {
		panic(MustError{Err: err})
	}
	return value
}

// Fetch the table schema.
func (must ResultMust) Schema() []*SqlColumn {
	schema, err := must.NormalResult.Schema()
	if err != nil {
		panic(MustError{Err: err})
	}
	return schema
}
