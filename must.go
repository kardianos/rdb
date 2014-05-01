package rdb

type ResultMust struct {
	res Result
}

type DatabaseMust struct {
	db Database
}

type TransactionMust struct {
	tran Transaction
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
		db: db,
	}
}

func (must DatabaseMust) Close() {
	err := must.db.Close()
	if err != nil {
		panic(MustError{Err: err})
	}
}

func (must DatabaseMust) Ping() {
	err := must.db.Ping()
	if err != nil {
		panic(MustError{Err: err})
	}
}
func (must DatabaseMust) ConnectionInfo() *ConnectionInfo {
	ci, err := must.db.ConnectionInfo()
	if err != nil {
		panic(MustError{Err: err})
	}
	return ci
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

// Make sure the result is closed.
func (must ResultMust) Close() {
	err := must.res.Close()
	if err != nil {
		panic(MustError{Err: err})
	}
}

// For each needed field, call Prep() or PrepAll() to prepare
// value pointers for scanning. To scan prepared fields call Scan().
// Call Scan() before using Get() or Getx().
func (must ResultMust) Scan() (eof bool) {
	eof, err := must.res.Scan()
	if err != nil {
		panic(MustError{Err: err})
	}
	return eof
}

// Prepare pointers to values to be populated by name using Prep. After
// preparing call Scan().
func (must ResultMust) Prep(name string, value interface{}) ResultMust {
	err := must.res.Prep(name, value)
	if err != nil {
		panic(MustError{Err: err})
	}
	return must
}

// Prepare pointers to values to be populated by index using Prep. After
// preparing call Scan().
func (must ResultMust) PrepAll(values ...interface{}) ResultMust {
	err := must.res.PrepAll(values...)
	if err != nil {
		panic(MustError{Err: err})
	}
	return must
}

// Use after Scan(). Can only pull fields which have not already been sent
// into a prepared value.
func (must ResultMust) Get(name string) interface{} {
	value, err := must.res.Get(name)
	if err != nil {
		panic(MustError{Err: err})
	}
	return value
}

// Use after Scan(). Can only pull fields which have not already been sent
// into a prepared value.
func (must ResultMust) Getx(index int) interface{} {
	value, err := must.res.Getx(index)
	if err != nil {
		panic(MustError{Err: err})
	}
	return value
}

// Fetch the table schema.
func (must ResultMust) Schema() *Schema {
	schema, err := must.res.Schema()
	if err != nil {
		panic(MustError{Err: err})
	}
	return schema
}
