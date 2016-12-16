// Copyright 2014 Daniel Theophanes.
// Use of this source code is governed by a zlib-style
// license that can be found in the LICENSE file.

package rdb

import "time"

// If the N (Name) field is not specified is not specified, then the order
// of the parameter should be used if the driver supports it.
type Param struct {
	// Optional parameter name.
	// All parameter names MUST NOT begin with a leading symbol. If required by
	// the backend the driver should insert.
	Name string

	// Parameter Type. Drivers may be able to infer this type.
	// Check the driver documentation used for more information.
	Type Type

	// Paremeter Length. Useful for variable length types that may check truncation.
	Length int

	// Value for input parameter.
	// If the value is an io.Reader it will read the value directly to the wire.
	Value interface{}

	// Set to true if the parameter is an output parameter.
	// If true, the value member should be provided through a pointer.
	Out bool

	// The following fields may go away.
	Null      bool
	Scale     int
	Precision int
}

// Information about the column as reported by the database.
type Column struct {
	Name      string // Columnn name.
	Index     int    // Column zero based index as appearing in result.
	Type      Type   // The data type as reported from the driver.
	Generic   Type   // The generic data type as reported from the driver.
	Length    int    // The length of the column as it makes sense per type.
	Unlimit   bool   // Provides near unlimited length.
	Nullable  bool   // True if the column type can be null.
	Key       bool   // True if the column is part of the key.
	Serial    bool   // True if the column is auto-incrementing.
	Precision int    // For decimal types, the precision.
	Scale     int    // For types with scale, including decimal.
}

// Returned from GetN and GetxN.
// Represents a nullable type.
type Nullable struct {
	Null  bool        // True if value is null.
	Value interface{} // Value, if any present.
}

// If the command output fields are specified, the Field output can help manage
// how the result rows are copied to.
type Field struct {
	Name string // Optional Field Name.

	// Value to report if the driver reports a null value.
	Null interface{}
}

type IsolationLevel byte

const (
	LevelDefault IsolationLevel = iota
	LevelReadUncommited
	LevelReadCommited
	LevelWriteCommited
	LevelRepeatableRead
	LevelSerializable
	LevelSnapshot
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

// Return strings as.
type StringMethod byte

const (
	StringAsString StringMethod = iota
	StringAsBytes
)

// Command represents a SQL command and can be used from many different
// queries at the same time.
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

	// Optional fields to specify output marshal.
	Fields []Field

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

	// Converter returns conversion functions per column or parameter. If nil
	// no conversion is performed.
	Converter Converter

	// Optional name of the command. May be used if logging.
	Name string

	// QueryTimeout is the timeout for a query this command is executed in.
	// This is ignored if zero. If this is zero, the connection configuration
	// QueryTimeout field will be used if non-zero.
	QueryTimeout time.Duration

	// The duraction before the query is automatically closed. If negative,
	// autoclose is disabled. If zero, the default is used. If positive that
	// value is used.
	AutoClose time.Duration

	// Return strings as strings or bytes.
	StringMethod StringMethod
}
