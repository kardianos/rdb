package example

import (
	"bitbucket.org/kardianos/rdb"
	_ "bitbucket.org/kardianos/tds"
	"math/big"
	"reflect"
	"testing"
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
				dec = @dec
		`,
		Arity: rdb.OneOnly,
		Input: []rdb.Param{
			rdb.Param{N: "bt", T: rdb.TypeBool, V: true},
			rdb.Param{N: "bf", T: rdb.TypeBool, V: false},
			rdb.Param{N: "i8", T: rdb.TypeInt8, V: byte(55)},
			rdb.Param{N: "i16", T: rdb.TypeInt16, V: int16(1234)},
			rdb.Param{N: "bb", T: rdb.TypeBinary, L: 0, V: []byte{23, 24, 25, 26, 27}},
			rdb.Param{N: "dec", T: rdb.TypeDecimal, Precision: 38, Scale: 4, V: big.NewRat(1234, 100)},
		},
	}
	_ = big.Int{}

	db := rdb.OpenMust(config)
	defer db.Close()

	var bt, bf bool
	var i8 byte
	var i16 int16
	var bb []byte
	var dec *big.Rat

	res := db.Query(cmd)
	defer res.Close()

	res.PrepAll(&bt, &bf, &i8, &i16, &bb, &dec)

	res.Scan()

	compare := []interface{}{bt, bf, i8, i16, bb, dec}

	for i := range compare {
		in := cmd.Input[i]
		if !reflect.DeepEqual(compare[i], in.V) {
			t.Errorf("Param %s did not round trip: Want (%v) got (%v)", in.N, in.V, compare[i])
		}
	}
	t.Logf("bt: %t, bf: %t", bt, bf)
	t.Logf("i8: %d, i16: %d", i8, i16)
	t.Logf("bb: %v", bb)
	t.Logf("dec: %s", dec.String())
}
