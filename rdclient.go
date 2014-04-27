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

type Param struct {
	N string  // Optional Parameter Name.
	T SqlType // Parameter Type.
	L int     // Paremeter Length.

	V interface{} // Value for input parameter.

	Null      bool
	Scale     int
	Precision int
}

type Value struct {
	N string      // Parameter Name.
	V interface{} // Value for input parameter.
}

type WriteProp struct {
	ColumnIndex int
	ColumnCount int
	MustCopy    bool
}

type Field struct {
	N         string // Optional Field Name.
	Type      NativeType
	NullValue interface{}
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
