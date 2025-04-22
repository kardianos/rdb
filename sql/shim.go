// Copyright 2014 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package sql is API compatible with "database/sql" but uses the rdb driver and connection pool.
// The sql package must be used in conjunction with a database driver. See http://golang.org/s/sqldrivers for a list of drivers.
// For more usage examples, see the wiki page at http://golang.org/s/sqlwiki.
package sql

import (
	"context"
	"errors"

	"github.com/kardianos/rdb"
	"github.com/kardianos/rdb/sql/driver"
)

// ErrNoRows is returned by Scan when QueryRow doesn't return a row. In such a case, QueryRow returns a placeholder *Row value that defers this error until a Scan.
var ErrNoRows = errors.New("sql: no rows in result set")

var ErrTxDone = errors.New("sql: Transaction has already been committed or rolled back")

// DB is a database handle representing a pool of zero or more underlying connections. It's safe for concurrent use by multiple goroutines.
//
// The sql package creates and frees connections automatically; it also maintains a free pool of idle connections. If the database has a concept of per-connection state, such state can only be reliably observed within a transaction. Once DB.Begin is called, the returned Tx is bound to a single connection. Once Commit or Rollback is called on the transaction, that transaction's connection is returned to DB's idle connection pool. The pool size can be controlled with SetMaxIdleConns.
type DB struct {
	pool *rdb.ConnPool
}

func NewDB(pool *rdb.ConnPool) *DB {
	return &DB{
		pool: pool,
	}
}

func (db *DB) Normal() *rdb.ConnPool {
	return db.pool
}

// Open opens a database specified by its database driver name and a driver-specific data source name, usually consisting of at least a database name and connection information.
//
// Most users will open a database via a driver-specific connection helper function that returns a *DB. No database drivers are included in the Go standard library. See http://golang.org/s/sqldrivers for a list of third-party drivers.
//
// Open may just validate its arguments without creating a connection to the database. To verify that the data source name is valid, call Ping.
//
// The returned DB is safe for concurrent use by multiple goroutines and maintains its own pool of idle connections. Thus, the Open function should be called just once. It is rarely necessary to close a DB.
func Open(config *rdb.Config) (*DB, error) {
	pool, err := rdb.Open(config)
	if err != nil {
		return nil, err
	}
	return NewDB(pool), nil
}

// Begin starts a transaction. The isolation level is dependent on the driver.
func (db *DB) Begin() (*Tx, error) {
	tran, err := db.pool.Begin(context.Background())
	if err != nil {
		return nil, err
	}
	return &Tx{
		tran: tran,
	}, nil
}

// Close closes the database, releasing any open resources.
//
// It is rare to Close a DB, as the DB handle is meant to be long-lived and shared between many goroutines.
func (db *DB) Close() error {
	db.pool.Close()
	return nil
}

// A Result summarizes an executed SQL command.
type result struct {
	res *rdb.Result
}

func (r result) LastInsertId() (int64, error) {
	return 0, nil
}
func (r result) RowsAffected() (int64, error) {
	return int64(r.res.RowsAffected()), nil
}

func prepParams(args []interface{}) []rdb.Param {
	params := make([]rdb.Param, len(args))
	for i, value := range args {
		params[i].Value = value
	}
	return params
}
func prep(arity rdb.Arity, query string, args []interface{}) (*rdb.Command, []rdb.Param) {
	params := make([]rdb.Param, len(args))
	for i, value := range args {
		params[i].Value = value
	}
	return &rdb.Command{
		SQL:   query,
		Arity: arity,
	}, params
}

// Exec executes a query without returning any rows. The args are for any placeholder parameters in the query.
func (db *DB) Exec(query string, args ...interface{}) (Result, error) {
	ctx := context.Background()
	cmd, params := prep(rdb.Zero, query, args)
	res, err := db.pool.Query(ctx, cmd, params...)
	if err != nil {
		return nil, err
	}
	return result{res: res}, nil
}

// Ping verifies a connection to the database is still alive, establishing a connection if necessary.
func (db *DB) Ping() error {
	ctx := context.Background()
	return db.pool.Ping(ctx)
}

// Prepare creates a prepared statement for later queries or executions. Multiple queries or executions may be run concurrently from the returned statement.
func (db *DB) Prepare(query string) (*Stmt, error) {
	return &Stmt{
		q: db.pool,
		cmd: &rdb.Command{
			SQL: query,
		},
	}, nil
}

// Query executes a query that returns rows, typically a SELECT. The args are for any placeholder parameters in the query.
func (db *DB) Query(query string, args ...interface{}) (*Rows, error) {
	ctx := context.Background()
	cmd, params := prep(rdb.Any, query, args)
	res, err := db.pool.Query(ctx, cmd, params...)
	if err != nil {
		return nil, err
	}
	return &Rows{res: res}, nil
}

// QueryRow executes a query that is expected to return at most one row. QueryRow always return a non-nil value. Errors are deferred until Row's Scan method is called.
func (db *DB) QueryRow(query string, args ...interface{}) *Row {
	ctx := context.Background()
	cmd, params := prep(rdb.One, query, args)
	res, err := db.pool.Query(ctx, cmd, params...)
	row := &Row{res: res}
	if err != nil {
		if err == rdb.ErrArity {
			err = ErrNoRows
		}
		row.err = err
	}
	return row
}

// SetMaxIdleConns is not supported directly.
// Set the PoolInitCapacity field in the rdb.Config.
func (db *DB) SetMaxIdleConns(n int) {
	panic("Set PoolInitCapacity in rdb.Config.")
}

// SetMaxOpenConns is not supported directly.
// Set the PoolMaxCapacity field in the rdb.Config.
func (db *DB) SetMaxOpenConns(n int) {
	panic("Set PoolMaxCapacity in rdb.Config.")
}

// NullBool represents a bool that may be null. NullBool implements the Scanner interface so it can be used as a scan destination, similar to NullString.
type NullBool struct {
	Bool  bool
	Valid bool // Valid is true if Bool is not NULL
}

// Scan implements the Scanner interface.
func (n *NullBool) Scan(value interface{}) error {
	return nil
}

// Value implements the driver Valuer interface.
func (n NullBool) Value() (driver.Value, error) {
	return nil, nil
}

// NullFloat64 represents a float64 that may be null. NullFloat64 implements the Scanner interface so it can be used as a scan destination, similar to NullString.
type NullFloat64 struct {
	Float64 float64
	Valid   bool // Valid is true if Float64 is not NULL
}

// Scan implements the Scanner interface.
func (n *NullFloat64) Scan(value interface{}) error {
	return nil
}

// Value implements the driver Valuer interface.
func (n NullFloat64) Value() (driver.Value, error) {
	return nil, nil
}

// NullInt64 represents an int64 that may be null. NullInt64 implements the Scanner interface so it can be used as a scan destination, similar to NullString.
type NullInt64 struct {
	Int64 int64
	Valid bool // Valid is true if Int64 is not NULL
}

// Scan implements the Scanner interface.
func (n *NullInt64) Scan(value interface{}) error {
	return nil
}

// Value implements the driver Valuer interface.
func (n NullInt64) Value() (driver.Value, error) {
	return nil, nil
}

// NullString represents a string that may be null. NullString implements the Scanner interface so it can be used as a scan destination:
//
//	var s NullString
//	err := db.QueryRow("SELECT name FROM foo WHERE id=?", id).Scan(&s)
//	...
//	if s.Valid {
//	   // use s.String
//	} else {
//	  // NULL value
//	}
type NullString struct {
	String string
	Valid  bool // Valid is true if String is not NULL
}

// Scan implements the Scanner interface.
func (ns *NullString) Scan(value interface{}) error {
	return nil
}

// Value implements the driver Valuer interface.
func (ns NullString) Value() (driver.Value, error) {
	return nil, nil
}

// RawBytes is a byte slice that holds a reference to memory owned by the database itself. After a Scan into a RawBytes, the slice is only valid until the next call to Next, Scan, or Close.
type RawBytes []byte

// A Result summarizes an executed SQL command.
type Result interface {
	// LastInsertId returns the integer generated by the database
	// in response to a command. Typically this will be from an
	// "auto increment" column when inserting a new row. Not all
	// databases support this feature, and the syntax of such
	// statements varies.
	LastInsertId() (int64, error)

	// RowsAffected returns the number of rows affected by an
	// update, insert, or delete. Not every database or database
	// driver may support this.
	RowsAffected() (int64, error)
}

// Row is the result of calling QueryRow to select a single row.
type Row struct {
	res *rdb.Result
	err error
}

// Scan copies the columns from the matched row into the values pointed at by dest. If more than one row matches the query, Scan uses the first row and discards the rest. If no row matches the query, Scan returns ErrNoRows.
func (r *Row) Scan(dest ...interface{}) error {
	if r.err != nil {
		return r.err
	}
	return r.res.Scan(dest...)
}

// Rows is the result of a query. Its cursor starts before the first row of the result set. Use Next to advance through the rows:

// rows, err := db.Query("SELECT ...")
//
//	...
//	defer rows.Close()
//	for rows.Next() {
//	    var id int
//	    var name string
//	    err = rows.Scan(&id, &name)
//	    ...
//	}
//	err = rows.Err() // get any error encountered during iteration
//	...
type Rows struct {
	res *rdb.Result
}

// Close closes the Rows, preventing further enumeration. If Next returns false, the Rows are closed automatically and it will suffice to check the result of Err. Close is idempotent and does not affect the result of Err.
func (rs *Rows) Close() error {
	return rs.res.Close()
}

// Columns returns the column names. Columns returns an error if the rows are closed, or if the rows are from QueryRow and there was a deferred error.
func (rs *Rows) Columns() ([]string, error) {
	schema := rs.res.Schema()
	names := make([]string, len(schema))
	for i, col := range schema {
		names[i] = col.Name
	}
	return names, nil
}

// Err returns the error, if any, that was encountered during iteration. Err may be called after an explicit or implicit Close.
func (rs *Rows) Err() error {
	return nil
}

// Next prepares the next result row for reading with the Scan method. It returns true on success, or false if there is no next result row or an error happened while preparing it. Err should be consulted to distinguish between the two cases.
//
// Every call to Scan, even the first one, must be preceded by a call to Next.
func (rs *Rows) Next() bool {
	return rs.res.Next()
}

// Scan copies the columns in the current row into the values pointed at by dest.
//
// If an argument has type *[]byte, Scan saves in that argument a copy of the corresponding data. The copy is owned by the caller and can be modified and held indefinitely. The copy can be avoided by using an argument of type *RawBytes instead; see the documentation for RawBytes for restrictions on its use.
//
// If an argument has type *interface{}, Scan copies the value provided by the underlying driver without conversion. If the value is of type []byte, a copy is made and the caller owns the result.
func (rs *Rows) Scan(dest ...interface{}) error {
	return rs.res.Scan(dest...)
}

// Scanner is an interface used by Scan.
type Scanner interface {
	// Scan assigns a value from a database driver.
	//
	// The src value will be of one of the following restricted
	// set of types:
	//
	//    int64
	//    float64
	//    bool
	//    []byte
	//    string
	//    time.Time
	//    nil - for NULL values
	//
	// An error should be returned if the value can not be stored
	// without loss of information.
	Scan(src interface{}) error
}

// Stmt is a prepared statement. Stmt is safe for concurrent use by multiple goroutines.
type Stmt struct {
	q   rdb.Queryer
	cmd *rdb.Command
}

// Close closes the statement.
func (s *Stmt) Close() error {
	// TODO: Unprepare any statement.
	return nil
}

// Exec executes a prepared statement with the given arguments and returns a Result summarizing the effect of the statement.
func (s *Stmt) Exec(args ...interface{}) (Result, error) {
	ctx := context.Background()
	res, err := s.q.Query(ctx, s.cmd, prepParams(args)...)
	if err != nil {
		return nil, err
	}
	return result{res: res}, res.Close()
}

// Query executes a prepared query statement with the given arguments and returns the query results as a *Rows.
func (s *Stmt) Query(args ...interface{}) (*Rows, error) {
	ctx := context.Background()
	res, err := s.q.Query(ctx, s.cmd, prepParams(args)...)
	if err != nil {
		return nil, err
	}
	return &Rows{res: res}, nil
}

// QueryRow executes a prepared query statement with the given arguments. If an error occurs during the execution of the statement, that error will be returned by a call to Scan on the returned *Row, which is always non-nil. If the query selects no rows, the *Row's Scan will return ErrNoRows. Otherwise, the *Row's Scan scans the first selected row and discards the rest.
//
// Example usage:
//
//	var name string
//	err := nameByUseridStmt.QueryRow(id).Scan(&name)
//	type Tx
//	type Tx struct {
//	    // contains filtered or unexported fields
//	}
//
// Tx is an in-progress database transaction.
//
// A transaction must end with a call to Commit or Rollback.
//
// After a call to Commit or Rollback, all operations on the transaction fail with ErrTxDone.
func (s *Stmt) QueryRow(args ...interface{}) *Row {
	ctx := context.Background()
	res, err := s.q.Query(ctx, s.cmd, prepParams(args)...)
	row := &Row{res: res}
	if err != nil {
		if err == rdb.ErrArity {
			err = ErrNoRows
		}
		row.err = err
	}
	return row
}

type Tx struct {
	tran *rdb.Transaction
}

func (tx *Tx) Normal() *rdb.Transaction {
	return tx.tran
}

// Commit commits the transaction.
func (tx *Tx) Commit() error {
	return tx.tran.Commit()
}

// Exec executes a query that doesn't return rows. For example: an INSERT and UPDATE.
func (tx *Tx) Exec(query string, args ...interface{}) (Result, error) {
	ctx := context.Background()
	cmd, params := prep(rdb.Zero, query, args)
	res, err := tx.tran.Query(ctx, cmd, params...)
	if err != nil {
		return nil, err
	}
	return result{res: res}, nil
}

// Prepare creates a prepared statement for use within a transaction.
//
// The returned statement operates within the transaction and can no longer be used once the transaction has been committed or rolled back.
//
// To use an existing prepared statement on this transaction, see Tx.Stmt.
func (tx *Tx) Prepare(query string) (*Stmt, error) {
	return &Stmt{
		q: tx.tran,
		cmd: &rdb.Command{
			SQL: query,
		},
	}, nil
}

// Query executes a query that returns rows, typically a SELECT.
func (tx *Tx) Query(query string, args ...interface{}) (*Rows, error) {
	ctx := context.Background()
	cmd, params := prep(rdb.Any, query, args)
	res, err := tx.tran.Query(ctx, cmd, params...)
	if err != nil {
		return nil, err
	}
	return &Rows{res: res}, nil
}

// QueryRow executes a query that is expected to return at most one row. QueryRow always return a non-nil value. Errors are deferred until Row's Scan method is called.
func (tx *Tx) QueryRow(query string, args ...interface{}) *Row {
	ctx := context.Background()
	cmd, params := prep(rdb.Any, query, args)
	res, err := tx.tran.Query(ctx, cmd, params...)
	row := &Row{res: res}
	if err != nil {
		if err == rdb.ErrArity {
			err = ErrNoRows
		}
		row.err = err
	}
	return row
}

// Rollback aborts the transaction.
func (tx *Tx) Rollback() error {
	return tx.tran.Rollback()
}

// Stmt returns a transaction-specific prepared statement from an existing statement.
//
// Example:
//
//	updateMoney, err := db.Prepare("UPDATE balance SET money=money+? WHERE id=?")
//	...
//	tx, err := db.Begin()
//	...
//	res, err := tx.Stmt(updateMoney).Exec(123.45, 98293203)
func (tx *Tx) Stmt(stmt *Stmt) *Stmt {
	ns := *stmt
	s := &ns
	s.q = tx.tran
	return s
}
