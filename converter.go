// Copyright 2014 Daniel Theophanes.
// Use of this source code is governed by a zlib-style
// license that can be found in the LICENSE file.

package rdb

// If the value is headed to the server, for instance for a parameter, toServer
// is true. If the value is being received by the client toServer is false.
type Convert func(toServer bool, column *Column, value *Nullable) error

func (c Convert) Convert(toServer bool, column *Column) Convert {
	return c
}

// Convert is called once per column or parameter to fetch the Convert function.
// Should the function Convert be nil, no conversion is performed.
// toServer is true when a value is moving from the client to the server.
type Converter interface {
	Convert(toServer bool, column *Column) Convert
}
