package rdb

type ResultMust struct {
	res *Result
}

type DatabaseMust struct {
	db *Database
}

type TransactionMust struct {
	tran *Transaction
}

// Same as Open() but all errors are returned as a panic(MustError{}).
func OpenMust(c *Config) DatabaseMust {
	db, err := Open(c)
	if err != nil {
		panic(MustError{Err: err})
	}
	return DatabaseMust{
		db: db,
	}
}

func (must DatabaseMust) Close() {
	err := must.db.Close()
	if err != nil {
		panic(MustError{Err: err})
	}
}

// Input parameter values can either be specified in the paremeter definition
// or on each query. If the value is not put in the parameter definition
// then the command instance may be reused for every query.
func (must DatabaseMust) Query(cmd *Command, vv ...Value) ResultMust {
	res, err := must.db.Query(cmd, vv...)
	if err != nil {
		panic(MustError{Err: err})
	}
	return ResultMust{
		res: res,
	}
}

// Same as Query but will panic on an error.
func (must DatabaseMust) Transaction(iso IsolationLevel) TransactionMust {
	tran, err := must.db.Transaction(iso)
	if err != nil {
		panic(MustError{Err: err})
	}
	return TransactionMust{
		tran: tran,
	}
}

// Input parameter values can either be specified in the paremeter definition
// or on each query. If the value is not put in the parameter definition
// then the command instance may be reused for every query.
func (must TransactionMust) Query(cmd *Command, vv ...Value) ResultMust {
	res, err := must.tran.Query(cmd, vv...)
	if err != nil {
		panic(MustError{Err: err})
	}
	return ResultMust{
		res: res,
	}
}

func (must TransactionMust) Commit() {
	err := must.tran.Commit()
	if err != nil {
		panic(MustError{Err: err})
	}
}
func (must TransactionMust) Rollback() {
	err := must.tran.Rollback()
	if err != nil {
		panic(MustError{Err: err})
	}
}

// Arrange for (*Command).Scan[...].V to hold fresh pointer each time.
// Returns io.EOF when last row has been scanned.
func (must ResultMust) ScanPrep() (eof bool) {
	eof, err := must.res.ScanPrep()
	if err != nil {
		panic(MustError{Err: err})
	}
	return eof
}
func (must ResultMust) Prep(name string, value interface{}) ResultMust {
	err := must.res.Prep(name, value)
	if err != nil {
		panic(MustError{Err: err})
	}
	return must
}
func (must ResultMust) PrepAll(values ...interface{}) ResultMust {
	err := must.res.PrepAll(values...)
	if err != nil {
		panic(MustError{Err: err})
	}
	return must
}

// Scans the row into a buffer that can be fetched
// Returns io.EOF when last row has been scanned.
func (must ResultMust) ScanBuffer() (eof bool) {
	eof, err := must.res.ScanBuffer()
	if err != nil {
		panic(MustError{Err: err})
	}
	return eof
}

// Use with ScanBuffer[M]().
func (must ResultMust) Get(name string) interface{} {
	value, err := must.res.Get(name)
	if err != nil {
		panic(MustError{Err: err})
	}
	return value
}

// Use with ScanBuffer[M]().
func (must ResultMust) Getx(index int) interface{} {
	value, err := must.res.Getx(index)
	if err != nil {
		panic(MustError{Err: err})
	}
	return value
}

func (must ResultMust) Schema() *Schema {
	schema, err := must.res.Schema()
	if err != nil {
		panic(MustError{Err: err})
	}
	return schema
}
