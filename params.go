// Copyright 2014 Daniel Theophanes.
// Use of this source code is governed by a zlib-style
// license that can be found in the LICENSE file.

package rdb

// If the N (Name) field is not specified is not specified, then the order
// of the parameter should be used if the driver supports it.
type Param struct {
	N string // Optional Parameter Name.

	// Parameter Type. Drivers may be able to infer this type.
	// Check the driver documentation used for more information.
	T SqlType

	// Paremeter Length. Useful for variable length types that may check truncation.
	L int

	// Value for input parameter.
	// Do not use if reusing *Command, use Value.V instead.
	// If the value is an io.Reader it will read the value directly to the wire.
	// If this satisfies the Filler interface the value will be fetched from
	// that interface.
	V interface{}

	// Set to true if the parameter is an output parameter.
	// If true, the value member should be provided through a pointer.
	Out bool

	// The following fields may go away.
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

	// Param is typically only set by the driver.
	// Users should set the name parameter or just rely on the Value index.
	Param *Param
}

// Information about the column as reported by the database.
type SqlColumn struct {
	Name      string  // Columnn name.
	Index     int     // Column zero based index as appearing in result.
	SqlType   SqlType // The data type as reported from the driver.
	Length    int     // The length of the column as it makes sense per type.
	Unlimit   bool    // Provides near unlimited length.
	Nullable  bool    // True if the column type can be null.
	Key       bool    // True if the column is part of the key.
	Serial    bool    // True if the column is auto-incrementing.
	Precision int     // For decimal types, the precision.
	Scale     int     // For types with scale, including decimal.
}

// Returned from GetN and GetxN.
// Represents a nullable type.
type Nullable struct {
	Null bool        // True if value is null.
	V    interface{} // Value, if any present.
}

// If the command output fields are specified, the Field output can help manage
// how the result rows are copied to.
type Field struct {
	N string // Optional Field Name.

	// Value to report if the driver reports a null value.
	NullValue interface{}
}

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
	Any Arity = 0

	ArityMust Arity = 16

	Zero Arity = 1 // Close the result after the query executes.
	One  Arity = 2 // Close the result after one row.

	// Close the result after the query executes,
	// return an error if any rows are returned.
	ZeroMust Arity = Zero | ArityMust

	// Close the result after one row,
	// return an error if more or less then one row is returned.
	OneMust Arity = One | ArityMust
)

// Command represents a SQL command and can be used from many different
// queries at the same time, so long as the input parameter values
// "Input[N].V (Value)" are not set in the Param struct but passed in with
// the actual query as Value.
// The Command MUST be reused if the Prepare field is true.
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
	Output []Field

	// If set to true silently truncates text longer then the field.
	// If this is set to false text truncation will result in an error.
	TruncLongText bool

	// If true the connection will attempt to lookup any cached prepared
	// identifier. If the cached identifier is not found or if it is found
	// to be invalid, it is renewed.
	// When the connection or connection pool is closed any prepared statements
	// are un-prepared.
	Prepare bool

	// Set the isolation level for the query or transaction.
	IsoLevel IsolationLevel

	// Optional name of the command. May be used if logging.
	Name string
}
