// Copyright 2014 Daniel Theophanes.
// Use of this source code is governed by a zlib-style
// license that can be found in the LICENSE file.

package example

import (
	"math/big"
	"reflect"
	"testing"

	"bitbucket.org/kardianos/rdb"
	_ "bitbucket.org/kardianos/tds"
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
	config := rdb.ParseConfigMust("tds://TESTU@localhost/SqlExpress?db=master")

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
		Input: []rdb.Param{
			rdb.Param{N: "bt", T: rdb.TypeBool, V: true},
			rdb.Param{N: "bf", T: rdb.TypeBool, V: false},
			rdb.Param{N: "i8", T: rdb.TypeInt8, V: byte(55)},
			rdb.Param{N: "i16", T: rdb.TypeInt16, V: int16(1234)},
			rdb.Param{N: "bb", T: rdb.TypeBinary, L: 0, V: []byte{23, 24, 25, 26, 27}},
			rdb.Param{N: "dec", T: rdb.TypeDecimal, Precision: 38, Scale: 4, V: big.NewRat(1234, 100)},
			rdb.Param{N: "fl32", T: rdb.TypeFloat32, V: float32(45.67)},
			rdb.Param{N: "fl64", T: rdb.TypeFloat64, V: float64(89.1011)},
		},
	}

	db := rdb.OpenMust(config)
	defer db.Close()

	var bt, bf bool
	var i8 byte
	var i16 int16
	var bb []byte
	var dec *big.Rat
	var fl32 float32
	var fl64 float64

	res := db.Query(cmd)
	defer res.Close()

	res.PrepAll(&bt, &bf, &i8, &i16, &bb, &dec, &fl32, &fl64)

	res.Scan()

	compare := []interface{}{bt, bf, i8, i16, bb, dec, fl32, fl64}

	for i := range compare {
		in := cmd.Input[i]
		if !reflect.DeepEqual(compare[i], in.V) {
			t.Errorf("Param %s did not round trip: Want (%v) got (%v)", in.N, in.V, compare[i])
		}
	}
}
