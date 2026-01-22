// Copyright 2014 Daniel Theophanes.
// Use of this source code is governed by a zlib-style
// license that can be found in the LICENSE file.

package ms

import (
	"context"
	"math/big"
	"reflect"
	"strings"
	"testing"

	"github.com/kardianos/rdb"
)

func TestNumber(t *testing.T) {
	checkSkip(t)
	defer recoverTest(t)

	cmd := &rdb.Command{
		SQL: `
			select
				bt = @bt, bf = @bf,
				i8 = @i8, i16 = @i16, i64 = @i64,
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

	var i64 int64 = 1234567

	params := []rdb.Param{
		{Name: "bt", Type: rdb.TypeBool, Value: true},
		{Name: "bf", Type: rdb.TypeBool, Value: false},
		{Name: "i8", Type: rdb.TypeInt8, Value: byte(55)},
		{Name: "i16", Type: rdb.TypeInt16, Value: int16(1234)},
		{Name: "i64", Type: rdb.Integer, Value: i64},
		{Name: "bb", Type: rdb.Binary, Length: 0, Value: []byte{23, 24, 25, 26, 27}},
		{Name: "dec", Type: rdb.TypeDecimal, Precision: 38, Scale: 4, Value: big.NewRat(1234, 100)},
		{Name: "fl32", Type: rdb.TypeFloat32, Value: float32(45.67)},
		{Name: "fl64", Type: rdb.TypeFloat64, Value: float64(89.1011)},
	}

	res := db.Query(context.Background(), cmd, params...)
	defer res.Close()

	res.Scan(&bt, &bf, &i8, &i16, &i64, &bb, &dec, &fl32, &fl64)

	compare := []interface{}{bt, bf, i8, i16, i64, bb, dec, fl32, fl64}

	for i := range compare {
		in := params[i]
		if !reflect.DeepEqual(compare[i], in.Value) {
			t.Errorf("Param %s did not round trip: Want (%v) got (%v)", in.Name, in.Value, compare[i])
		}
	}
}

func TestDecimal(t *testing.T) {
	checkSkip(t)
	defer recoverTest(t)

	cmd := &rdb.Command{
		SQL: `
declare @ld decimal(38,3) = @d;
select
	d = @ld,
	s = cast(@ld as varchar(100))
;
`,
		Arity: rdb.OneMust,
	}

	var dec *big.Rat
	var sdec string

	dIn := &big.Rat{}
	// dIn.SetString("1.035")
	dIn = big.NewRat(4661225614328463, 4503599627370496)
	params := []rdb.Param{
		{Name: "d", Type: rdb.TypeDecimal, Precision: 38, Scale: 6, Value: dIn},
	}

	res := db.Query(context.Background(), cmd, params...)
	defer res.Close()

	res.Scan(&dec, &sdec)

	if dec.FloatString(3) != dIn.FloatString(3) {
		t.Errorf("D: %v, S: %v, In: %v", dec.FloatString(3), sdec, dIn.FloatString(3))
	}
}
func TestDecimal2(t *testing.T) {
	checkSkip(t)
	defer assertFreeConns(t)
	defer recoverTest(t)

	list := []struct {
		Name  string
		Input interface{}
		Scale int
		Want  string
	}{
		{
			Name:  "bad scale",
			Input: big.NewRat(35840000000000003, 1000000000000000),
			Scale: 2,
			Want:  "35.84",
		},
		{
			Name:  "NULL1",
			Input: nil,
			Scale: 2,
			Want:  "NULL",
		},
		{
			Name:  "NULL2",
			Input: rdb.Null,
			Scale: 2,
			Want:  "NULL",
		},
	}

	cmd := &rdb.Command{
		SQL: `
declare @D decimal(36,2) = @V;
select S = isnull(convert(nvarchar(100), @D), 'NULL');
`,
		Arity: rdb.OneMust,
	}
	for _, item := range list {
		t.Run(item.Name, func(t *testing.T) {
			res := db.Query(context.Background(), cmd,
				rdb.Param{Name: "V", Type: rdb.Decimal, Precision: 38, Scale: item.Scale, Value: item.Input},
			)
			defer res.Close()

			if res.Next() == false {
				t.Fatal("expected row")
			}
			var got string
			res.Scan(&got)

			if g, w := got, item.Want; g != w {
				t.Fatalf("got: %q, want: %q", g, w)
			}
		})
	}
}

func TestBytesValue(t *testing.T) {
	checkSkip(t)
	defer recoverTest(t)

	cmd := &rdb.Command{
		SQL: `
			select @bytesEmpty, @bytesOne
		`,
		Arity: rdb.OneMust,
	}

	bytesEmpty, bytesOne := []byte{}, []byte{01}
	var bytesEmptyOut, bytesOneOut []byte

	params := []rdb.Param{
		{Name: "bytesEmpty", Type: rdb.Binary, Value: bytesEmpty},
		{Name: "bytesOne", Type: rdb.Binary, Value: bytesOne},
	}

	res := db.Query(context.Background(), cmd, params...)
	defer res.Close()

	res.Scan(&bytesEmptyOut, &bytesOneOut)

	compare := []interface{}{bytesEmptyOut, bytesOneOut}

	for i := range compare {
		in := params[i]
		if !reflect.DeepEqual(compare[i], in.Value) {
			t.Errorf("Param %s did not round trip: Want (%v) got (%v)", in.Name, in.Value, compare[i])
		}
	}
}

func TestNullNumbers(t *testing.T) {
	checkSkip(t)
	defer recoverTest(t)

	cmd := &rdb.Command{
		SQL: `
			select @decimal
		`,
		Arity: rdb.OneMust,
	}

	params := []rdb.Param{
		{Name: "decimal", Type: rdb.Decimal, Value: nil, Precision: 38, Scale: 6, Null: true},
	}

	res := db.Query(context.Background(), cmd, params...)
	defer res.Close()

	res.Scan()
	val := res.Getx(0)

	if val != nil {
		t.Fatalf("Rat should be nil: %v", val)
	}
}

func TestGUID(t *testing.T) {
	checkSkip(t)
	defer recoverTest(t)

	cmd := &rdb.Command{
		SQL: `
			select newid();
		`,
		Arity: rdb.OneMust,
	}

	res := db.Query(context.Background(), cmd)
	defer res.Close()

	res.Scan()
	val := res.Getx(0)

	if val == nil {
		t.Fatalf("GUID should not be nil: %v", val)
	}
	t.Log(val)
}

func TestGUIDParam(t *testing.T) {
	checkSkip(t)
	defer recoverTest(t)

	testUUID := "12345678-1234-5678-9ABC-DEF012345678"

	// UUID bytes in RFC 4122 order (big-endian for Data1/2/3).
	testBytes := [16]byte{
		0x12, 0x34, 0x56, 0x78, // Data1
		0x12, 0x34, // Data2
		0x56, 0x78, // Data3
		0x9A, 0xBC, // Data4 (first 2 bytes)
		0xDE, 0xF0, 0x12, 0x34, 0x56, 0x78, // Data4 (last 6 bytes)
	}

	// Custom GUID type for reflection test.
	type CustomGUID [16]byte
	customGUID := CustomGUID(testBytes)

	tests := []struct {
		name  string
		value interface{}
	}{
		{"string", testUUID},
		{"slice", testBytes[:]},
		{"array", testBytes},
		{"array_ptr", &testBytes},
		{"custom_type", customGUID},
		{"custom_type_ptr", &customGUID},
	}

	cmd := &rdb.Command{
		SQL: `
			select @guid;
		`,
		Arity: rdb.OneMust,
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer recoverTest(t)

			params := []rdb.Param{
				{Name: "guid", Type: rdb.TypeUUID, Value: tt.value},
			}

			res := db.Query(context.Background(), cmd, params...)
			defer res.Close()

			res.Scan()
			val := res.Getx(0)

			if val == nil {
				t.Fatalf("GUID should not be nil: %v", val)
			}
			// GUID string comparison is case-insensitive.
			if got, want := val.(string), testUUID; !strings.EqualFold(got, want) {
				t.Errorf("GUID mismatch: got %q, want %q", got, want)
			}
		})
	}
}

func TestMoney(t *testing.T) {
	checkSkip(t)
	defer recoverTest(t)

	cmd := &rdb.Command{
		SQL: `
			select
				m = @m,
				sm = @sm,
				mCast = cast(123.4567 as money),
				smCast = cast(67.89 as smallmoney);
		`,
		Arity: rdb.OneMust,
	}

	mIn := big.NewRat(12345, 100)  // 123.45
	smIn := big.NewRat(6789, 100)  // 67.89

	params := []rdb.Param{
		{Name: "m", Type: rdb.TypeMoney, Value: mIn},
		{Name: "sm", Type: rdb.TypeMoney, Value: smIn},
	}

	res := db.Query(context.Background(), cmd, params...)
	defer res.Close()

	var m, sm, mCast, smCast *big.Rat
	res.Scan(&m, &sm, &mCast, &smCast)

	if m.Cmp(mIn) != 0 {
		t.Errorf("m: got %s, want %s", m.FloatString(4), mIn.FloatString(4))
	}
	if sm.Cmp(smIn) != 0 {
		t.Errorf("sm: got %s, want %s", sm.FloatString(4), smIn.FloatString(4))
	}

	wantMCast := big.NewRat(1234567, 10000) // 123.4567
	if mCast.Cmp(wantMCast) != 0 {
		t.Errorf("mCast: got %s, want %s", mCast.FloatString(4), wantMCast.FloatString(4))
	}

	wantSMCast := big.NewRat(6789, 100) // 67.89
	if smCast.Cmp(wantSMCast) != 0 {
		t.Errorf("smCast: got %s, want %s", smCast.FloatString(4), wantSMCast.FloatString(4))
	}
}

func TestMoneyNull(t *testing.T) {
	checkSkip(t)
	defer recoverTest(t)

	cmd := &rdb.Command{
		SQL: `
			select @m;
		`,
		Arity: rdb.OneMust,
	}

	params := []rdb.Param{
		{Name: "m", Type: rdb.TypeMoney, Value: nil, Null: true},
	}

	res := db.Query(context.Background(), cmd, params...)
	defer res.Close()

	res.Scan()
	val := res.Getx(0)

	if val != nil {
		t.Fatalf("Money should be nil: %v", val)
	}
}

func TestSystemTables(t *testing.T) {
	checkSkip(t)

	// Test selecting from various system tables to ensure all column types decode correctly.
	queries := []string{
		"select top 1 * from sys.tables",
		"select top 1 * from sys.columns",
		"select top 1 * from sys.objects",
		"select top 1 * from sys.types",
		"select top 1 * from sys.schemas",
		"select top 1 * from sys.databases",
		"select top 1 * from sys.indexes",
		"select top 1 * from sys.index_columns",
		"select top 1 * from sys.foreign_keys",
		"select top 1 * from sys.foreign_key_columns",
		"select top 1 * from sys.key_constraints",
		"select top 1 * from sys.check_constraints",
		"select top 1 * from sys.default_constraints",
		"select top 1 * from sys.procedures",
		"select top 1 * from sys.views",
		"select top 1 * from sys.triggers",
		"select top 1 * from sys.parameters",
		"select top 1 * from sys.sql_modules",
		"select top 1 * from sys.dm_exec_sessions",
		// "select top 1 * from sys.dm_exec_connections", // Contains sql_variant type, not yet supported
		"select top 1 * from INFORMATION_SCHEMA.TABLES",
		"select top 1 * from INFORMATION_SCHEMA.COLUMNS",
		"select top 1 * from INFORMATION_SCHEMA.ROUTINES",
		"select top 1 * from INFORMATION_SCHEMA.VIEWS",
		"select top 1 * from INFORMATION_SCHEMA.SCHEMATA",
	}

	for _, sql := range queries {
		t.Run(sql, func(t *testing.T) {
			defer recoverTest(t)

			cmd := &rdb.Command{
				SQL:   sql,
				Arity: rdb.ZeroMust,
			}

			res := db.Query(context.Background(), cmd)
			defer res.Close()

			// Iterate through all rows and columns to ensure decoding works.
			for res.Next() {
				schema := res.Schema()
				for i := range schema {
					_ = res.Getx(i) // Just read the value, don't care about the result.
				}
			}
		})
	}
}
