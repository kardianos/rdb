// Copyright 2014 Daniel Theophanes.
// Use of this source code is governed by a zlib-style
// license that can be found in the LICENSE file.

package ms

import (
	"math/big"
	"reflect"
	"testing"

	"bitbucket.org/kardianos/rdb"
	"bitbucket.org/kardianos/rdb/must"
)

func TestNumber(t *testing.T) {
	defer func() {
		if re := recover(); re != nil {
			if localError, is := re.(must.Error); is {
				t.Errorf("SQL Error: %v", localError)
				return
			}
			panic(re)
		}
	}()

	cmd := &rdb.Command{
		Sql: `
			select
				bt = @bt, bf = @bf,
				i8 = @i8, i16 = @i16,
				bb = @bb,
				dec = @dec,
				fl32 = @fl32,
				fl64 = @fl64
		`,
		Arity: rdb.OneMust,
	}

	var bt, bf bool
	var i8 byte
	var i16 int16
	var bb []byte
	var dec *big.Rat
	var fl32 float32
	var fl64 float64

	params := []rdb.Param{
		{Name: "bt", Type: rdb.TypeBool, Value: true},
		{Name: "bf", Type: rdb.TypeBool, Value: false},
		{Name: "i8", Type: rdb.TypeInt8, Value: byte(55)},
		{Name: "i16", Type: rdb.TypeInt16, Value: int16(1234)},
		{Name: "bb", Type: rdb.Binary, Length: 0, Value: []byte{23, 24, 25, 26, 27}},
		{Name: "dec", Type: rdb.TypeDecimal, Precision: 38, Scale: 4, Value: big.NewRat(1234, 100)},
		{Name: "fl32", Type: rdb.TypeFloat32, Value: float32(45.67)},
		{Name: "fl64", Type: rdb.TypeFloat64, Value: float64(89.1011)},
	}

	res := db.Query(cmd, params...)
	defer res.Close()

	res.Scan(&bt, &bf, &i8, &i16, &bb, &dec, &fl32, &fl64)

	compare := []interface{}{bt, bf, i8, i16, bb, dec, fl32, fl64}

	for i := range compare {
		in := params[i]
		if !reflect.DeepEqual(compare[i], in.Value) {
			t.Errorf("Param %s did not round trip: Want (%v) got (%v)", in.Name, in.Value, compare[i])
		}
	}
}
