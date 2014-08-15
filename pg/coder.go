// Copyright 2014 Daniel Theophanes.
// Use of this source code is governed by a zlib-style
// license that can be found in the rdb LICENSE file.

package pg

import (
	"encoding/hex"
	"fmt"
	"math/big"
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

func encodeField(param *rdb.Param, write *writer) error {
	// Encode int32 length, then body.

	switch param.Type {
	case rdb.TypeInt8, rdb.TypeInt16, rdb.TypeInt32, rdb.TypeInt64, rdb.Integer,
		rdb.TypeFloat32, rdb.TypeFloat64, rdb.Float,
		rdb.TypeDecimal, rdb.Decimal:
		var eval int64
		var sval string
		switch val := param.Value.(type) {
		case int8:
			eval = int64(val)
		case int16:
			eval = int64(val)
		case int32:
			eval = int64(val)
		case int64:
			eval = int64(val)
		case float32:
			eval = int64(val)
		case float64:
			eval = int64(val)
		case *big.Rat:
			sval = val.String()

		case *int8:
			eval = int64(*val)
		case *int16:
			eval = int64(*val)
		case *int32:
			eval = int64(*val)
		case *int64:
			eval = int64(*val)
		case *float32:
			eval = int64(*val)
		case *float64:
			eval = int64(*val)
		case **big.Rat:
			sval = val.String()
		default:
			return fmt.Errorf("Unsupported value type: %T", param.Value)
		}
		if len(sval) == 0 {
			sval = strconv.FormatInt(eval, 10)
		}
		write.Int32(int32(len(sval)))
		write.StringNoTerm(sval)
	case rdb.TypeText, rdb.TypeAnsiText, rdb.Text,
		rdb.TypeVarChar, rdb.TypeAnsiVarChar, rdb.TypeChar, rdb.TypeAnsiChar,
		rdb.TypeBinary, rdb.Binary:
		switch val := param.Value.(type) {
		case string:
			write.Int32(int32(len(val)))
			write.StringNoTerm(val)
		case []byte:
			write.Int32(int32(len(val)))
			write.Bytea(val)
		case rune:
			sval := string(val)
			write.Int32(int32(len(sval)))
			write.StringNoTerm(sval)

		case *string:
			write.Int32(int32(len(*val)))
			write.StringNoTerm(*val)
		case *[]byte:
			write.Int32(int32(len(*val)))
			write.Bytea(*val)
		case *rune:
			sval := string(*val)
			write.Int32(int32(len(sval)))
			write.StringNoTerm(sval)
		default:
			return fmt.Errorf("Unsupported value type: %T", param.Value)
		}
	default:
		return fmt.Errorf("Unsupported type: %v", param.Type)
	}

	return nil
}
