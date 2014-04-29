// By Daniel Theophanes

// SQL Relational Database Client.
/*
	Named Parameters
	Inspect driver support
	Driver callbacks - events
	Manage data type mapping
		Handle custom data types
	For result fields, can write to io.Writer
	For input parameters, can read from io.Reader
	View schema
		Column Names
		Mapped data types
		Driver data types
		Column attributes (nullable, length, precision)
	Bulk Insert
	Handle cases where idle DB connection is reset and must be reconneted to.
		Automatically re-preparing any prepared statements.
	Handle Transactions
		Commit
		Rollback
		Save point
	Handle different isolation levels
	Provide a unified method of logging:
		Full executed sql statements with all parameter values.
			(Parameters map optionally be opmitted from log).
		Any errors that occur.
		Time taken for execution. Useful for ongoing QOS. Associate with query.
		(Can name Commands)
	Provide a standard SQL syntax error structure that can be inspected:
		Sql text
		Line number
		Error text
		Can have multiple SqlErrors for a given query.
	Set active collation.

*/
package rdb

// If the N (Name) field is not specified is not specified, then the order
// of the parameter should be used if the driver supports it.
type Param struct {
	N string  // Optional Parameter Name.
	T SqlType // Parameter Type.
	L int     // Paremeter Length.

	// Value for input parameter.
	// If the value is an io.Reader it will read the value directly to the wire.
	// If this satisfies the Filler interface the value will be fetched from
	// that interface.
	V interface{}

	Null      bool
	Scale     int
	Precision int
}

// If the input parameter value isn't populated in the command,
// the value can be filled in at the time of query.
// If the N (Name) field is not specified, then the order of the
// parameters or values are used if the driver supports it.
type Value struct {
	N string // Parameter Name.

	// Value for input parameter.
	// If the value is an io.Reader it will read the value directly to the wire.
	V interface{}

	Param *Param
}

type SqlColumn struct {
	Name     string
	Index    int
	SqlType  uint8
	Length   uint32
	Unlimit  bool // Provides near unlimited length.
	Nullable bool
}

// If the command output fields are specified, the Field output can help manage
// how the result rows are copied to.
type Field struct {
	N         string     // Optional Field Name.
	Type      NativeType // TODO: Should this stay or go?
	NullValue interface{}
}

// Type panic'ed with after calling a Must method.
type MustError struct {
	Err error
}

func (err MustError) Error() string {
	return err.Err.Error()
}
