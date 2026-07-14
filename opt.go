// Copyright 2014 Daniel Theophanes.
// Use of this source code is governed by a zlib-style
// license that can be found in the LICENSE file.

package rdb

// Opt is an optional (nullable) value that does not use interface{} boxing.
// Valid is false for SQL NULL; V then holds the zero value of T.
//
// Use as a struct field by value with a db tag (not *Opt[T]):
//
//	type Row struct {
//	    Name Opt[string] `db:"name"`
//	}
//
// table.Query / Stream detect optionals structurally (exported fields V and
// Valid bool). Prefer Opt when nullability should be self-contained on one
// field. Note that embedding Valid bool next to T can add padding (looser
// packing than a plain T plus a separate null:"…" bool placed carefully).
//
// For denser layouts when you do not need a self-contained optional:
//
//	type Row struct {
//	    Name     string `db:"name"`
//	    NameNull bool   `null:"name"`
//	}
type Opt[T any] struct {
	V     T
	Valid bool
}

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
