// Copyright 2014 Daniel Theophanes.
// Use of this source code is governed by a zlib-style
// license that can be found in the LICENSE file.

package rdb

// ColumnConverter converts between value types.
type ColumnConverter func(column *Column, nullable *Nullable) error

// ColumnConverter is called once per column or output parameter to fetch the ColumnConverter function.
// Should the function ColumnConverter be nil, no conversion is performed.
// ConvertParam is called once per input parameter.
type Converter interface {
	ColumnConverter(column *Column) ColumnConverter
	ConvertParam(param *Param) error
}
