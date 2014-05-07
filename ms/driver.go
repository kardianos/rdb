// Copyright 2014 Daniel Theophanes.
// Use of this source code is governed by a zlib-style
// license that can be found in the LICENSE file.

package ms

import (
	"bitbucket.org/kardianos/rdb"
	"bitbucket.org/kardianos/rdb/semver"
	"fmt"
	"net"
	"net/url"
	"sync"
)

func init() {
	rdb.Register("ms", &Driver{})
}

type Driver struct{}

func (dr *Driver) Open(c *rdb.Config) (rdb.Database, error) {

	return &Database{
		conf: c,
	}, nil
}
func (dr *Driver) DriverMetaInfo() *rdb.DriverMeta {
	return &rdb.DriverMeta{
		DriverSupport: rdb.DriverSupport{
			NamedParameter:   true,
			FluidType:        false,
			MultipleResult:   false,
			SecureConnection: false,
			BulkInsert:       false,
			Notification:     false,
			UserDataTypes:    false,
		},
	}
}

func (dr *Driver) ParseOptions(KV map[string]interface{}, configOptions url.Values) error {
	return nil
}

type Database struct {
	conf *rdb.Config

	// Lock info fields.
	lockInfo sync.Mutex

	Server, Protocol *semver.Version
}

func (db *Database) Close() error {
	return nil
}

var pingCommand = &rdb.Command{
	Sql:   "select top 0 1;",
	Arity: rdb.ZeroMust,
}

func (db *Database) Ping() error {
	_, err := db.Query(pingCommand)
	return err
}

func (db *Database) ConnectionInfo() (*rdb.ConnectionInfo, error) {
	db.lockInfo.Lock()
	needVersion := db.Server != nil
	db.lockInfo.Unlock()

	if needVersion {
		// Connect if it has not already.
		err := db.Ping()
		if err != nil {
			return nil, err
		}
		db.lockInfo.Lock()
		needVersion := db.Server != nil
		db.lockInfo.Unlock()

		if needVersion {
			return nil, fmt.Errorf("Connection is not filling out information.")
		}
	}
	db.lockInfo.Lock()
	defer db.lockInfo.Unlock()

	return &rdb.ConnectionInfo{
		Server:   db.Server,
		Protocol: db.Protocol,
	}, nil
}

func (db *Database) Query(cmd *rdb.Command, vv ...rdb.Value) (rdb.Result, error) {
	port := 1433
	c := db.conf
	if c.Port != 0 {
		port = c.Port
	}
	hostname := "localhost"
	if len(c.Hostname) != 0 && c.Hostname != "." {
		hostname = c.Hostname
	}
	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", hostname, port))
	if err != nil {
		return nil, err
	}

	tds := NewConnection(conn)

	_, err = tds.Open(c)
	if err != nil {
		tds.Close()
		return nil, err
	}

	db.lockInfo.Lock()
	db.Server = tds.ProductVersion
	db.Protocol = tds.ProtocolVersion
	db.lockInfo.Unlock()

	param := make([]*rdb.Param, len(cmd.Input))
	for i := range cmd.Input {
		param[i] = &cmd.Input[i]
	}
	fields := make([]*rdb.Field, len(cmd.Output))
	for i := range cmd.Output {
		fields[i] = &cmd.Output[i]
	}

	res, err := tds.Execute(cmd.Sql, cmd.TruncLongText, cmd.Arity, param, vv, fields)
	if err != nil {
		return res, err
	}
	if len(res.Errors) != 0 {
		err = res.Errors
		res.Close()
		return res, err
	}
	if res.arity&rdb.Zero != 0 {
		defer res.Close()

		err = res.Process(false)
		if !res.EOF && res.arity&rdb.ArityMust != 0 && err == nil {
			err = arityError
			fmt.Printf("ERR: %d\n", res.rowCount)
		}
	}
	return res, err
}

// Not supported in this driver yet.
func (db *Database) Transaction(iso rdb.IsolationLevel) (rdb.Transaction, error) {
	panic("Transactions Unsupported")
}

func (db *Database) Must() rdb.DatabaseMust {
	return rdb.DatabaseMust{NormalDatabase: db}
}

func (r *Result) Prep(name string, value interface{}) error {
	col, found := r.ColumnLookup[name]
	if !found {
		return rdb.ErrorColumnNotFound{At: "Prep", Name: name}
	}
	r.prep[col.Index] = value
	return nil
}
func (r *Result) PrepAll(values ...interface{}) error {
	for i := range values {
		if i >= len(r.Columns) {
			return rdb.ErrorColumnNotFound{At: "PrepAll", Index: i}
		}
		r.prep[i] = values[i]
	}
	return nil
}
func (r *Result) Prepx(index int, value interface{}) error {
	if index < 0 || index >= len(r.Columns) {
		return rdb.ErrorColumnNotFound{At: "Prepx", Index: index}
	}
	r.prep[index] = value
	return nil
}
func (r *Result) Get(name string) (interface{}, error) {
	col, found := r.ColumnLookup[name]
	if !found {
		return nil, rdb.ErrorColumnNotFound{At: "Get", Name: name}
	}
	bv := r.buffer[col.Index]
	if bv == nil {
		return nil, nil
	}
	return bv.Value, nil
}
func (r *Result) Getx(index int) (interface{}, error) {
	if index < 0 || index >= len(r.Columns) {
		return nil, rdb.ErrorColumnNotFound{At: "Getx", Index: index}
	}
	bv := r.buffer[index]
	if bv == nil {
		return nil, nil
	}
	return bv.Value, nil
}
func (r *Result) GetN(name string) (rdb.Nullable, error) {
	col, found := r.ColumnLookup[name]
	if !found {
		return rdb.Nullable{}, rdb.ErrorColumnNotFound{At: "GetN", Name: name}
	}
	bv := r.buffer[col.Index]
	return rdb.Nullable{
		Null: bv.Null,
		V:    bv.Value,
	}, nil
}
func (r *Result) GetxN(index int) (rdb.Nullable, error) {
	if index < 0 || index >= len(r.Columns) {
		return rdb.Nullable{}, rdb.ErrorColumnNotFound{At: "GetxN", Index: index}
	}
	bv := r.buffer[index]
	return rdb.Nullable{
		Null: bv.Null,
		V:    bv.Value,
	}, nil
}

// Fetch the table schema.
func (res *Result) Schema() ([]*rdb.SqlColumn, error) {
	if res.Columns == nil {
		return nil, fmt.Errorf("Must call schema after running query.")
	}
	sch := make([]*rdb.SqlColumn, len(res.Columns))
	for i, drCol := range res.Columns {
		sch[i] = &drCol.SqlColumn
	}
	return sch, nil
}

func (res *Result) Must() rdb.ResultMust {
	return rdb.ResultMust{NormalResult: res}
}

// For each needed field, call Prep() or PrepAll() to prepare
// value pointers for scanning. To scan prepared fields call ScanPrep().
func (res *Result) Scan() (more bool, err error) {
	err = res.Process(true)
	if res.EOF {
		cerr := res.Close()
		if err == nil {
			err = cerr
		}
	}
	return !res.EOF, err
}

type Transaction struct {
}

func (tran *Transaction) Query(cmd *rdb.Command, vv ...rdb.Value) (rdb.Result, error) {
	return nil, nil
}
func (tran *Transaction) Commit() error {
	return nil
}
func (tran *Transaction) Rollback() error {
	return nil
}

func (tran *Transaction) Must() rdb.TransactionMust {
	return rdb.TransactionMust{NormalTransaction: tran}
}
