// Copyright 2014 Daniel Theophanes.
// Use of this source code is governed by a zlib-style
// license that can be found in the LICENSE file.

package ms

import (
	"math/big"
	"reflect"
	"testing"

	"bitbucket.org/kardianos/rdb"
)

func TestNumber(t *testing.T) {
	defer func() {
		if re := recover(); re != nil {
			if localError, is := re.(rdb.MustError); is {
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

	openConnPool()

	var bt, bf bool
	var i8 byte
	var i16 int16
	var bb []byte
	var dec *big.Rat
	var fl32 float32
	var fl64 float64

	params := []rdb.Param{
		{N: "bt", T: rdb.TypeBool, V: true},
		{N: "bf", T: rdb.TypeBool, V: false},
		{N: "i8", T: rdb.TypeInt8, V: byte(55)},
		{N: "i16", T: rdb.TypeInt16, V: int16(1234)},
		{N: "bb", T: rdb.TypeBinary, L: 0, V: []byte{23, 24, 25, 26, 27}},
		{N: "dec", T: rdb.TypeDecimal, Precision: 38, Scale: 4, V: big.NewRat(1234, 100)},
		{N: "fl32", T: rdb.TypeFloat32, V: float32(45.67)},
		{N: "fl64", T: rdb.TypeFloat64, V: float64(89.1011)},
	}

	res := db.Query(cmd, params...)
	defer res.Close()

	res.PrepAll(&bt, &bf, &i8, &i16, &bb, &dec, &fl32, &fl64)

	res.Scan()

	compare := []interface{}{bt, bf, i8, i16, bb, dec, fl32, fl64}

	for i := range compare {
		in := params[i]
		if !reflect.DeepEqual(compare[i], in.V) {
			t.Errorf("Param %s did not round trip: Want (%v) got (%v)", in.N, in.V, compare[i])
		}
	}
}
