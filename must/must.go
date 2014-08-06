// Copyright 2014 Daniel Theophanes.
// Use of this source code is governed by a zlib-style
// license that can be found in the LICENSE file.

// package must wraps the rdb database interface with one that returns errors
// with a panic.
package must

import (
	"bitbucket.org/kardianos/rdb"
)

// Type panic'ed with after calling a Must method.
type Error struct {
	Err error
}

func (err Error) Error() string {
	return err.Err.Error()
}

type Result struct {
	norm *rdb.Result
}

type ConnPool struct {
	norm *rdb.ConnPool
}

type Transaction struct {
	norm *rdb.Transaction
}

// Get the panic'ing version that doesn't return errors.
func NewConnPool(cp *rdb.ConnPool) ConnPool {
	return ConnPool{norm: cp}
}

// Get the panic'ing version that doesn't return errors.
func NewResult(r *rdb.Result) Result {
	return Result{norm: r}
}

// Get the panic'ing version that doesn't return errors.
func NewTransaction(tran *rdb.Transaction) Transaction {
	return Transaction{norm: tran}
}

// Get the non-panic'ing version of Result.
func (must Result) Normal() *rdb.Result {
	return must.norm
}

// Get the non-panic'ing version of Database.
func (must ConnPool) Normal() *rdb.ConnPool {
	return must.norm
}

// Get the non-panic'ing version of Transaction.
func (must Transaction) Normal() *rdb.Transaction {
	return must.norm
}

// ConfigMust takes the output of the ParseConfig and panics if an error is
// present.
func Config(config *rdb.Config, err error) *rdb.Config {
	if err != nil {
		panic(Error{Err: err})
	}
	return config
}

// Same as Open() but all errors are returned as a panic(MustError{}).
func Open(c *rdb.Config) ConnPool {
	db, err := rdb.Open(c)
	if err != nil {
		panic(Error{Err: err})
	}
	return ConnPool{
		norm: db,
	}
}

func (must ConnPool) Close() {
	must.norm.Close()
}

func (must ConnPool) Ping() {
	err := must.norm.Ping()
	if err != nil {
		panic(Error{Err: err})
	}
}
func (must ConnPool) ConnectionInfo() *rdb.ConnectionInfo {
	ci, err := must.norm.ConnectionInfo()
	if err != nil {
		panic(Error{Err: err})
	}
	return ci
}

// Input parameter values can either be specified in the paremeter definition
// or on each query. If the value is not put in the parameter definition
// then the command instance may be reused for every query.
func (must ConnPool) Query(cmd *rdb.Command, params ...rdb.Param) Result {
	res, err := must.norm.Query(cmd, params...)
	if err != nil {
		panic(Error{Err: err})
	}
	return Result{
		norm: res,
	}
}

// Same as Query but will panic on an error.
func (must ConnPool) Begin() Transaction {
	tran, err := must.norm.Begin()
	if err != nil {
		panic(Error{Err: err})
	}
	return Transaction{
		norm: tran,
	}
}

// Same as Query but will panic on an error.
func (must ConnPool) BeginLevel(level rdb.IsolationLevel) Transaction {
	tran, err := must.norm.BeginLevel(level)
	if err != nil {
		panic(Error{Err: err})
	}
	return Transaction{
		norm: tran,
	}
}

func (must ConnPool) PoolAvailable() (capacity, available int) {
	return must.norm.PoolAvailable()
}

// Input parameter values can either be specified in the paremeter definition
// or on each query. If the value is not put in the parameter definition
// then the command instance may be reused for every query.
func (must Transaction) Query(cmd *rdb.Command, params ...rdb.Param) Result {
	res, err := must.norm.Query(cmd, params...)
	if err != nil {
		panic(Error{Err: err})
	}
	return Result{
		norm: res,
	}
}

func (must Transaction) Commit() {
	err := must.norm.Commit()
	if err != nil {
		panic(Error{Err: err})
	}
}
func (must Transaction) Rollback() {
	err := must.norm.Rollback()
	if err != nil {
		panic(Error{Err: err})
	}
}
func (must Transaction) RollbackTo(savepoint string) {
	err := must.norm.RollbackTo(savepoint)
	if err != nil {
		panic(Error{Err: err})
	}
}
func (must Transaction) SavePoint(name string) {
	err := must.norm.SavePoint(name)
	if err != nil {
		panic(Error{Err: err})
	}
}
func (must Transaction) Active() bool {
	return must.norm.Active()
}

// Make sure the result is closed.
func (must Result) Close() {
	err := must.norm.Close()
	if err != nil {
		panic(Error{Err: err})
	}
}

func (r *Result) RowsAffected() uint64 {
	return r.norm.RowsAffected()
}

func (must Result) Next() (more bool) {
	return must.norm.Next()
}

func (must Result) NextResult() (more bool) {
	more, err := must.norm.NextResult()
	if err != nil {
		panic(Error{Err: err})
	}
	return more
}

// For each needed field, call Prep() or PrepAll() to prepare
// value pointers for scanning. To scan prepared fields call Scan().
// Call Scan() before using Get() or Getx().
// Returns false if no more rows.
func (must Result) Scan(values ...interface{}) Result {
	err := must.norm.Scan(values...)
	if err != nil {
		panic(Error{Err: err})
	}
	return must
}

// Informational messages. Do not call concurrently with Scan() or Done().
func (must Result) Info() []*rdb.Message {
	return must.norm.Info()
}

// Prepare pointers to values to be populated by name using Prep. After
// preparing call Scan().
func (must Result) Prep(name string, value interface{}) Result {
	must.norm.Prep(name, value)
	return must
}

// Prepare pointers to values to be populated by index using Prep. After
// preparing call Scan().
func (must Result) Prepx(index int, value interface{}) Result {
	must.norm.Prepx(index, value)
	return must
}

// Use after Scan(). Can only pull fields which have not already been sent
// into a prepared value.
func (must Result) Get(name string) interface{} {
	return must.norm.Get(name)
}

// Use after Scan(). Can only pull fields which have not already been sent
// into a prepared value.
func (must Result) Getx(index int) interface{} {
	return must.norm.Getx(index)
}

// Use after Scan(). Can only pull fields which have not already been sent
// into a prepared value.
func (must Result) GetN(name string) rdb.Nullable {
	return must.norm.GetN(name)
}

// Use after Scan(). Can only pull fields which have not already been sent
// into a prepared value.
func (must Result) GetxN(index int) rdb.Nullable {
	return must.norm.GetxN(index)
}

// Use after Scan(). Can only pull fields which have not already been sent
// into a prepared value. Not all fields will be populated if some have
// been prepared.
func (must Result) GetRowN() []rdb.Nullable {
	return must.norm.GetRowN()
}

// Fetch the table schema.
func (must Result) Schema() []*rdb.Column {
	return must.norm.Schema()
}
