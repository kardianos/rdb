// Copyright 2014 Daniel Theophanes.
// Use of this source code is governed by a zlib-style
// license that can be found in the LICENSE file.

package rdb

// Opt is an optional (nullable) value that does not use interface{} boxing.
// Valid is false for SQL NULL; V then holds the zero value of T.
//
// Use as a struct field with a db tag:
//
//	type Row struct {
//	    Name Opt[string] `db:"name"`
//	}
//
// Scanners detect Opt via the RDBOpt method (not package path), so detection
// works for any instantiation Opt[T] under reflection.
type Opt[T any] struct {
	V     T
	Valid bool
}

// RDBOpt marks this type as an rdb optional field for reflection-based planners
// (table.Query / Stream). Value receiver so Opt[T] field types implement it.
func (Opt[T]) RDBOpt() {}

// Set stores a non-null value.
func (o *Opt[T]) Set(v T) {
	o.V = v
	o.Valid = true
}

// SetNull marks the value as SQL NULL and clears V.
func (o *Opt[T]) SetNull() {
	var z T
	o.V = z
	o.Valid = false
}

// Get returns V and whether the value is non-null.
func (o Opt[T]) Get() (T, bool) {
	return o.V, o.Valid
}

// NullFlagPrep pairs a Prep destination with a bool that is set true when the
// column is SQL NULL. Value must be a pointer to the payload field (*int32,
// *string, etc.). The bool is typically addressed via a null:"col" struct tag.
//
//	type Row struct {
//	    Name     string `db:"name"`
//	    NameNull bool   `null:"name"`
//	}
//
// Drivers assign through DirectAssign*; table.Query builds NullFlagPrep when
// planning null-tagged fields.
type NullFlagPrep struct {
	Value any   // *T payload
	Null  *bool // set true on SQL NULL, false on non-null
}
