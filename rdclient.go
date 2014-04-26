// By Daniel Theophanes

// SQL Relational Database Client.
package rdb

/*
	Named Parameters
	Inspect driver support
	Driver callbacks - events
	Manage data type mapping
		Handle custom data types
	View schema
		Column Names
		Mapped data types
		Driver data types
		Column attributes (nullable, length, precision)
	Bulk Insert
	Handle cases where idle DB connection is reset and must be reconneted to.
		Automatically re-preparing any prepared statements.
*/
import (
	`fmt`
)

// TODO: Fill out type information.
type Type struct {
	// TODO: Should driver and user flags be separate? (no)
	// TODO: Should Driver and User flags be encoded into the Code? (yes)
	// TODO: Encode both standard type and driver type at the same time? (yes)
	Driver bool
	User   bool
	Code   uint64
}

type Param struct {
	N string // Parameter Name.
	T Type   // Parameter Type.
	L int    // Paremeter Length.

	V interface{} // Value for input parameter.

	Null      bool
	Scale     int
	Precision int
}

type Value struct {
	N string      // Parameter Name.
	V interface{} // Value for input parameter.
}

type SqlType byte
type DriverType uint32
type NativeType byte

type Marshal func(t NativeType) (SqlType, DriverType)
type Unmarshal func(t DriverType) (SqlType, NativeType)

// TODO: Need type notions.
// * Generic Sql Type
// * Specific Driver Type
// * Native Language Type
// * Mapping the above three
// These type mappings should be able to be set in a hierarchical manner.
// * Config
// * Database
// * Command
// * Parameter
// Ideal to make these mappings reusable.

type WriteProp struct {
	ColumnIndex int
	ColumnCount int
	MustCopy    bool
}

type Field struct {
	N     string // Field Name.
	Write func(wp *WriteProp, bb []byte) (n int, err error)
}

// TODO: Should this exist?
// Should return a pointer to a value.
type Filler interface {
	Fill(p *Param) interface{}
}

type ErrorList []Error

func (err ErrorList) Error() string {
	return fmt.Sprintf("%v", err)
}

// Type panic'ed with after calling a *M() method.
type Error struct {
	Err error
}

func (err *Error) Error() string {
	return err.Err.Error()
}
