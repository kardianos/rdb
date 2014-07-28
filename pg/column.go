// Copyright 2014 Daniel Theophanes.
// Use of this source code is governed by a zlib-style
// license that can be found in the rdb LICENSE file.

package pg

import "bitbucket.org/kardianos/rdb"

type Column struct {
	rdb.Column

	Oid     int32
	TypeMod int32
	Format  int16

	TableID  int32
	ColumnID int16
}
