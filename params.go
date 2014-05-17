// Copyright 2014 Daniel Theophanes.
// Use of this source code is governed by a zlib-style
// license that can be found in the LICENSE file.

package rdb

// If the N (Name) field is not specified is not specified, then the order
// of the parameter should be used if the driver supports it.
type Param struct {
	// Optional parameter name.
	// All parameter names MUST NOT begin with a leading symbol. If required by
	// the backend the driver should insert.
	N string

	// Parameter Type. Drivers may be able to infer this type.
	// Check the driver documentation used for more information.
	T SqlType

	// Paremeter Length. Useful for variable length types that may check truncation.
	L int

	// Value for input parameter.
	// If the value is an io.Reader it will read the value directly to the wire.
	V interface{}

	// Set to true if the parameter is an output parameter.
	// If true, the value member should be provided through a pointer.
	Out bool

	// The following fields may go away.
	Null      bool
	Scale     int
	Precision int
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

	// Convert this field with a function.
	Convert ConvertOutput
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

type ConvertInputIn struct {
	// Allows converting based on a column name suffix like
	// "create_time" -> TypeTD, may attempt to convert a string to a time.Time
	// value.
	Name string

	// Out parameters must be able to be converted bi-directionally.
	// Will be true if value is returning from the database.
	Returning bool

	// The SqlType as defined in the Param.
	Type SqlType

	// The input value.
	Value interface{}
}

// Convert parameter type to another type or simply set the SqlType.
// Ordered array in *Config and *Command.
// The values in *Command are evaluated before any in *Config.
type ConvertInput func(in ConvertInputIn) (handled bool, outType SqlType, out interface{}, err error)

// Convert output field.
// In *Config contains a map[SqlType]ConvertOutput map,
// also in per column Field output.
type ConvertOutput func(*SqlColumn, interface{}) (interface{}, error)

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

	// An ordered list of input converters.
	InputConverters []ConvertInput
}
