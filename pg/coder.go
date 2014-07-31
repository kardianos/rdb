// Copyright 2014 Daniel Theophanes.
// Use of this source code is governed by a zlib-style
// license that can be found in the rdb LICENSE file.

package pg

import (
	"encoding/hex"
	"fmt"
	"strconv"

	"bitbucket.org/kardianos/rdb"
)

func decodeField(col *Column, read *reader) (*rdb.DriverValue, error) {
	val := &rdb.DriverValue{}
	var err error
	if false {
		fmt.Printf("Field: %v, Format: %v, Type Mod: %v, Generic: %v\n%s", col.Oid, col.Format, col.TypeMod, col.Generic, hex.Dump(read.Bytea(read.Length)))
		return val, err
	}
	switch col.Generic {
	case rdb.Integer:
		var num int64
		snum := string(read.Bytea(read.Length))
		switch col.Type {
		case rdb.TypeInt8:
			num, err = strconv.ParseInt(snum, 10, 8)
			val.Value = int8(num)
		case rdb.TypeInt16:
			num, err = strconv.ParseInt(snum, 10, 16)
			val.Value = int16(num)
		case rdb.TypeInt32:
			num, err = strconv.ParseInt(snum, 10, 32)
			val.Value = int32(num)
		case rdb.TypeInt64:
			val.Value, err = strconv.ParseInt(snum, 10, 64)
		default:
			err = fmt.Errorf("Unhandled Integer type: %d", col.Type)
		}
	case rdb.Text:
		val.Value = string(read.Bytea(read.Length))
	case rdb.Binary:
		str := read.Bytea(read.Length)
		if len(str) <= 2 {
			val.Value = []byte{}
		} else {
			str = str[2:]

			bb := make([]byte, hex.DecodedLen(len(str)))
			var n int
			n, err = hex.Decode(bb, str)
			val.Value = bb[:n]
		}
	default:
		err = fmt.Errorf("Unhandled generic type: %d", col.Generic)
	}
	if debug {
		fmt.Printf("Field Value (%s): %v\n", col.Name, val.Value)
	}
	return val, err
}
