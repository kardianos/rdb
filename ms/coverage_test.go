// Copyright 2014 Daniel Theophanes.
// Use of this source code is governed by a zlib-style
// license that can be found in the LICENSE file.

package ms

import (
	"context"
	"reflect"
	"testing"
	"time"

	"github.com/kardianos/rdb"
)

// TestIntegerConversions tests various integer type conversions
func TestIntegerConversions(t *testing.T) {
	checkSkip(t)
	if parallel {
		t.Parallel()
	}
	defer assertFreeConns(t)
	defer recoverTest(t)

	// Test various integer types and conversions
	list := []struct {
		name  string
		t     rdb.Type
		value interface{}
		want  interface{}
	}{
		// int8/byte
		{name: "int8_from_int8", t: rdb.TypeInt8, value: int8(42), want: int8(42)},
		{name: "int8_from_int16", t: rdb.TypeInt8, value: int16(42), want: int8(42)},
		{name: "int8_from_int32", t: rdb.TypeInt8, value: int32(42), want: int8(42)},
		{name: "int8_from_int64", t: rdb.TypeInt8, value: int64(42), want: int8(42)},
		{name: "int8_from_int", t: rdb.TypeInt8, value: int(42), want: int8(42)},
		{name: "int8_from_uint", t: rdb.TypeInt8, value: uint(42), want: int8(42)},
		{name: "int8_from_uint16", t: rdb.TypeInt8, value: uint16(42), want: int8(42)},
		{name: "int8_from_uint32", t: rdb.TypeInt8, value: uint32(42), want: int8(42)},
		{name: "int8_from_uint64", t: rdb.TypeInt8, value: uint64(42), want: int8(42)},
		{name: "int8_from_float32", t: rdb.TypeInt8, value: float32(42.0), want: int8(42)},
		{name: "int8_from_float64", t: rdb.TypeInt8, value: float64(42.0), want: int8(42)},
		{name: "int8_from_string", t: rdb.TypeInt8, value: "42", want: int8(42)},

		// int16
		{name: "int16_from_int8", t: rdb.TypeInt16, value: int8(42), want: int16(42)},
		{name: "int16_from_byte", t: rdb.TypeInt16, value: byte(42), want: int16(42)},
		{name: "int16_from_int16", t: rdb.TypeInt16, value: int16(1234), want: int16(1234)},
		{name: "int16_from_int32", t: rdb.TypeInt16, value: int32(1234), want: int16(1234)},
		{name: "int16_from_int64", t: rdb.TypeInt16, value: int64(1234), want: int16(1234)},
		{name: "int16_from_int", t: rdb.TypeInt16, value: int(1234), want: int16(1234)},
		{name: "int16_from_uint", t: rdb.TypeInt16, value: uint(1234), want: int16(1234)},
		{name: "int16_from_float32", t: rdb.TypeInt16, value: float32(1234.0), want: int16(1234)},
		{name: "int16_from_float64", t: rdb.TypeInt16, value: float64(1234.0), want: int16(1234)},
		{name: "int16_from_string", t: rdb.TypeInt16, value: "1234", want: int16(1234)},

		// int32
		{name: "int32_from_int8", t: rdb.TypeInt32, value: int8(42), want: int32(42)},
		{name: "int32_from_byte", t: rdb.TypeInt32, value: byte(42), want: int32(42)},
		{name: "int32_from_int16", t: rdb.TypeInt32, value: int16(1234), want: int32(1234)},
		{name: "int32_from_uint16", t: rdb.TypeInt32, value: uint16(1234), want: int32(1234)},
		{name: "int32_from_int32", t: rdb.TypeInt32, value: int32(123456), want: int32(123456)},
		{name: "int32_from_uint32", t: rdb.TypeInt32, value: uint32(123456), want: int32(123456)},
		{name: "int32_from_int64", t: rdb.TypeInt32, value: int64(123456), want: int32(123456)},
		{name: "int32_from_uint64", t: rdb.TypeInt32, value: uint64(123456), want: int32(123456)},
		{name: "int32_from_int", t: rdb.TypeInt32, value: int(123456), want: int32(123456)},
		{name: "int32_from_uint", t: rdb.TypeInt32, value: uint(123456), want: int32(123456)},
		{name: "int32_from_float32", t: rdb.TypeInt32, value: float32(123456.0), want: int32(123456)},
		{name: "int32_from_float64", t: rdb.TypeInt32, value: float64(123456.0), want: int32(123456)},
		{name: "int32_from_string", t: rdb.TypeInt32, value: "123456", want: int32(123456)},

		// int64
		{name: "int64_from_int8", t: rdb.Integer, value: int8(42), want: int64(42)},
		{name: "int64_from_byte", t: rdb.Integer, value: byte(42), want: int64(42)},
		{name: "int64_from_int16", t: rdb.Integer, value: int16(1234), want: int64(1234)},
		{name: "int64_from_uint16", t: rdb.Integer, value: uint16(1234), want: int64(1234)},
		{name: "int64_from_int32", t: rdb.Integer, value: int32(123456), want: int64(123456)},
		{name: "int64_from_uint32", t: rdb.Integer, value: uint32(123456), want: int64(123456)},
		{name: "int64_from_int64", t: rdb.Integer, value: int64(1234567890), want: int64(1234567890)},
		{name: "int64_from_uint64", t: rdb.Integer, value: uint64(1234567890), want: int64(1234567890)},
		{name: "int64_from_int", t: rdb.Integer, value: int(1234567890), want: int64(1234567890)},
		{name: "int64_from_uint", t: rdb.Integer, value: uint(1234567890), want: int64(1234567890)},
		{name: "int64_from_float32", t: rdb.Integer, value: float32(1234567.0), want: int64(1234567)},
		{name: "int64_from_float64", t: rdb.Integer, value: float64(1234567890.0), want: int64(1234567890)},
		{name: "int64_from_string", t: rdb.Integer, value: "1234567890", want: int64(1234567890)},
	}

	for _, item := range list {
		t.Run(item.name, func(t *testing.T) {
			cmd := &rdb.Command{
				SQL:   `select v = @v`,
				Arity: rdb.OneMust,
			}

			res := db.Query(context.Background(), cmd, rdb.Param{Name: "v", Type: item.t, Value: item.value})
			defer res.Close()

			res.Scan()
			got := res.Getx(0)

			if !reflect.DeepEqual(got, item.want) {
				t.Errorf("got %v (%T), want %v (%T)", got, got, item.want, item.want)
			}
		})
	}
}

// TestPointerValues tests passing pointer values
func TestPointerValues(t *testing.T) {
	checkSkip(t)
	if parallel {
		t.Parallel()
	}
	defer assertFreeConns(t)
	defer recoverTest(t)

	i8 := int8(42)
	i16 := int16(1234)
	i32 := int32(123456)
	i64 := int64(1234567890)
	u16 := uint16(1234)
	u32 := uint32(123456)
	u64 := uint64(1234567890)
	f32 := float32(45.67)
	f64 := float64(89.1011)
	b := true
	by := byte(42)
	i := int(123456)
	u := uint(123456)

	list := []struct {
		name  string
		t     rdb.Type
		value interface{}
		want  interface{}
	}{
		// Pointer types for int8
		{name: "ptr_int8", t: rdb.TypeInt8, value: &i8, want: int8(42)},
		{name: "ptr_byte", t: rdb.TypeInt8, value: &by, want: int8(42)},
		{name: "ptr_int16_to_int8", t: rdb.TypeInt8, value: &i16, want: int8(-46)}, // truncation (1234 & 0xFF = 210 = -46 as int8)

		// Pointer types for int16
		{name: "ptr_int8_to_int16", t: rdb.TypeInt16, value: &i8, want: int16(42)},
		{name: "ptr_int16", t: rdb.TypeInt16, value: &i16, want: int16(1234)},
		{name: "ptr_uint16_to_int16", t: rdb.TypeInt16, value: &u16, want: int16(1234)},
		{name: "ptr_int32_to_int16", t: rdb.TypeInt16, value: &i32, want: int16(-7616)}, // truncation
		{name: "ptr_f32_to_int16", t: rdb.TypeInt16, value: &f32, want: int16(45)},
		{name: "ptr_f64_to_int16", t: rdb.TypeInt16, value: &f64, want: int16(89)},

		// Pointer types for int32
		{name: "ptr_int8_to_int32", t: rdb.TypeInt32, value: &i8, want: int32(42)},
		{name: "ptr_int16_to_int32", t: rdb.TypeInt32, value: &i16, want: int32(1234)},
		{name: "ptr_int32", t: rdb.TypeInt32, value: &i32, want: int32(123456)},
		{name: "ptr_uint32_to_int32", t: rdb.TypeInt32, value: &u32, want: int32(123456)},
		{name: "ptr_int_to_int32", t: rdb.TypeInt32, value: &i, want: int32(123456)},
		{name: "ptr_uint_to_int32", t: rdb.TypeInt32, value: &u, want: int32(123456)},
		{name: "ptr_f32_to_int32", t: rdb.TypeInt32, value: &f32, want: int32(45)},
		{name: "ptr_f64_to_int32", t: rdb.TypeInt32, value: &f64, want: int32(89)},

		// Pointer types for int64
		{name: "ptr_int8_to_int64", t: rdb.Integer, value: &i8, want: int64(42)},
		{name: "ptr_int16_to_int64", t: rdb.Integer, value: &i16, want: int64(1234)},
		{name: "ptr_int32_to_int64", t: rdb.Integer, value: &i32, want: int64(123456)},
		{name: "ptr_int64", t: rdb.Integer, value: &i64, want: int64(1234567890)},
		{name: "ptr_uint64_to_int64", t: rdb.Integer, value: &u64, want: int64(1234567890)},
		{name: "ptr_int_to_int64", t: rdb.Integer, value: &i, want: int64(123456)},
		{name: "ptr_uint_to_int64", t: rdb.Integer, value: &u, want: int64(123456)},
		{name: "ptr_f32_to_int64", t: rdb.Integer, value: &f32, want: int64(45)},
		{name: "ptr_f64_to_int64", t: rdb.Integer, value: &f64, want: int64(89)},

		// Pointer types for float - SQL Server returns float32 as float64
		// The value sent as float32(45.67) becomes 45.66999816894531 when converted
		{name: "ptr_f32", t: rdb.TypeFloat32, value: &f32, want: float64(f32)}, // Returns as float64
		{name: "ptr_f64", t: rdb.TypeFloat64, value: &f64, want: f64},
		{name: "ptr_i8_to_f64", t: rdb.TypeFloat64, value: &i8, want: float64(42)},
		{name: "ptr_i16_to_f64", t: rdb.TypeFloat64, value: &i16, want: float64(1234)},
		{name: "ptr_i32_to_f64", t: rdb.TypeFloat64, value: &i32, want: float64(123456)},
		{name: "ptr_i64_to_f64", t: rdb.TypeFloat64, value: &i64, want: float64(1234567890)},
		{name: "ptr_u16_to_f64", t: rdb.TypeFloat64, value: &u16, want: float64(1234)},
		{name: "ptr_u32_to_f64", t: rdb.TypeFloat64, value: &u32, want: float64(123456)},
		{name: "ptr_u64_to_f64", t: rdb.TypeFloat64, value: &u64, want: float64(1234567890)},
		{name: "ptr_by_to_f64", t: rdb.TypeFloat64, value: &by, want: float64(42)},

		// Bool pointer
		{name: "ptr_bool", t: rdb.TypeBool, value: &b, want: true},
	}

	for _, item := range list {
		t.Run(item.name, func(t *testing.T) {
			cmd := &rdb.Command{
				SQL:   `select v = @v`,
				Arity: rdb.OneMust,
			}

			res := db.Query(context.Background(), cmd, rdb.Param{Name: "v", Type: item.t, Value: item.value})
			defer res.Close()

			res.Scan()
			got := res.Getx(0)

			if !reflect.DeepEqual(got, item.want) {
				t.Errorf("got %v (%T), want %v (%T)", got, got, item.want, item.want)
			}
		})
	}
}

// TestNullValues tests NULL value handling for different types
func TestNullValues(t *testing.T) {
	checkSkip(t)
	if parallel {
		t.Parallel()
	}
	defer assertFreeConns(t)
	defer recoverTest(t)

	list := []struct {
		name string
		t    rdb.Type
	}{
		{name: "null_int8", t: rdb.TypeInt8},
		{name: "null_int16", t: rdb.TypeInt16},
		{name: "null_int32", t: rdb.TypeInt32},
		{name: "null_int64", t: rdb.Integer},
		{name: "null_float32", t: rdb.TypeFloat32},
		{name: "null_float64", t: rdb.TypeFloat64},
		{name: "null_bool", t: rdb.TypeBool},
		{name: "null_text", t: rdb.Text},
		{name: "null_binary", t: rdb.Binary},
	}

	for _, item := range list {
		t.Run(item.name, func(t *testing.T) {
			cmd := &rdb.Command{
				SQL:   `select v = @v`,
				Arity: rdb.OneMust,
			}

			res := db.Query(context.Background(), cmd, rdb.Param{Name: "v", Type: item.t, Value: nil, Null: true})
			defer res.Close()

			res.Scan()
			got := res.Getx(0)

			if got != nil {
				t.Errorf("got %v (%T), want nil", got, got)
			}
		})
	}
}

// TestTextTypes tests various text type handling
func TestTextTypes(t *testing.T) {
	checkSkip(t)
	if parallel {
		t.Parallel()
	}
	defer assertFreeConns(t)
	defer recoverTest(t)

	list := []struct {
		name   string
		t      rdb.Type
		value  interface{}
		length int
		want   string
	}{
		{name: "nvarchar_string", t: rdb.TypeVarChar, value: "hello world", length: 100, want: "hello world"},
		{name: "nvarchar_bytes", t: rdb.TypeVarChar, value: []byte("hello bytes"), length: 100, want: "hello bytes"},
		{name: "nvarchar_empty", t: rdb.TypeVarChar, value: "", length: 100, want: ""},
		// nchar is fixed-length, so it gets padded with spaces
		{name: "nchar_string", t: rdb.TypeChar, value: "hello", length: 10, want: "hello     "},
		{name: "varchar_ansi", t: rdb.TypeAnsiVarChar, value: "ansi text", length: 100, want: "ansi text"},
	}

	for _, item := range list {
		t.Run(item.name, func(t *testing.T) {
			cmd := &rdb.Command{
				SQL:   `select v = @v`,
				Arity: rdb.OneMust,
			}

			res := db.Query(context.Background(), cmd, rdb.Param{Name: "v", Type: item.t, Value: item.value, Length: item.length})
			defer res.Close()

			res.Scan()
			got := res.Getx(0)

			gotStr, ok := got.(string)
			if !ok {
				if gotBytes, ok := got.([]byte); ok {
					gotStr = string(gotBytes)
				} else {
					t.Fatalf("unexpected type: %T", got)
				}
			}

			if gotStr != item.want {
				t.Errorf("got %q, want %q", gotStr, item.want)
			}
		})
	}
}

// TestMaxLengthText tests text with max length (varchar(max), etc.)
func TestMaxLengthText(t *testing.T) {
	checkSkip(t)
	if parallel {
		t.Parallel()
	}
	defer assertFreeConns(t)
	defer recoverTest(t)

	// Generate a moderately long string (within typical limits)
	longStr := ""
	for range 100 {
		longStr += "abcdefghij"
	}

	list := []struct {
		name   string
		t      rdb.Type
		value  interface{}
		length int
	}{
		{name: "nvarchar_max_string", t: rdb.TypeVarChar, value: longStr, length: 0}, // 0 means max
		{name: "nvarchar_max_bytes", t: rdb.TypeVarChar, value: []byte(longStr), length: 0},
		{name: "varbinary_max", t: rdb.TypeBinary, value: []byte(longStr), length: 0},
		{name: "varchar_max_ansi", t: rdb.TypeAnsiVarChar, value: longStr, length: 0},
	}

	for _, item := range list {
		t.Run(item.name, func(t *testing.T) {
			cmd := &rdb.Command{
				SQL:   `select v = @v`,
				Arity: rdb.OneMust,
			}

			res := db.Query(context.Background(), cmd, rdb.Param{Name: "v", Type: item.t, Value: item.value, Length: item.length})
			defer res.Close()

			res.Scan()
			got := res.Getx(0)

			var gotStr string
			switch g := got.(type) {
			case string:
				gotStr = g
			case []byte:
				gotStr = string(g)
			default:
				t.Fatalf("unexpected type: %T", got)
			}

			var wantStr string
			switch w := item.value.(type) {
			case string:
				wantStr = w
			case []byte:
				wantStr = string(w)
			}

			if gotStr != wantStr {
				t.Errorf("length mismatch: got %d, want %d", len(gotStr), len(wantStr))
			}
		})
	}
}

// TestStringPointers tests string and byte slice pointers
func TestStringPointers(t *testing.T) {
	checkSkip(t)
	if parallel {
		t.Parallel()
	}
	defer assertFreeConns(t)
	defer recoverTest(t)

	str := "hello pointer"
	bytes := []byte("hello bytes pointer")

	list := []struct {
		name   string
		t      rdb.Type
		value  interface{}
		length int
		want   string
	}{
		{name: "ptr_string", t: rdb.TypeVarChar, value: &str, length: 100, want: str},
		{name: "ptr_bytes", t: rdb.TypeVarChar, value: &bytes, length: 100, want: string(bytes)},
	}

	for _, item := range list {
		t.Run(item.name, func(t *testing.T) {
			cmd := &rdb.Command{
				SQL:   `select v = @v`,
				Arity: rdb.OneMust,
			}

			res := db.Query(context.Background(), cmd, rdb.Param{Name: "v", Type: item.t, Value: item.value, Length: item.length})
			defer res.Close()

			res.Scan()
			got := res.Getx(0)

			gotStr, ok := got.(string)
			if !ok {
				if gotBytes, ok := got.([]byte); ok {
					gotStr = string(gotBytes)
				} else {
					t.Fatalf("unexpected type: %T", got)
				}
			}

			if gotStr != item.want {
				t.Errorf("got %q, want %q", gotStr, item.want)
			}
		})
	}
}

// TestFloatConversions tests float type handling
func TestFloatConversions(t *testing.T) {
	checkSkip(t)
	if parallel {
		t.Parallel()
	}
	defer assertFreeConns(t)
	defer recoverTest(t)

	list := []struct {
		name  string
		t     rdb.Type
		value interface{}
		want  float64
	}{
		{name: "f64_from_f32", t: rdb.TypeFloat64, value: float32(45.5), want: float64(float32(45.5))},
		{name: "f64_from_f64", t: rdb.TypeFloat64, value: float64(89.125), want: 89.125},
		{name: "f64_from_byte", t: rdb.TypeFloat64, value: byte(42), want: 42.0},
		{name: "f64_from_int8", t: rdb.TypeFloat64, value: int8(-42), want: -42.0},
		{name: "f64_from_int16", t: rdb.TypeFloat64, value: int16(1234), want: 1234.0},
		{name: "f64_from_uint16", t: rdb.TypeFloat64, value: uint16(1234), want: 1234.0},
		{name: "f64_from_int32", t: rdb.TypeFloat64, value: int32(123456), want: 123456.0},
		{name: "f64_from_uint32", t: rdb.TypeFloat64, value: uint32(123456), want: 123456.0},
		{name: "f64_from_int64", t: rdb.TypeFloat64, value: int64(1234567890), want: 1234567890.0},
		{name: "f64_from_uint64", t: rdb.TypeFloat64, value: uint64(1234567890), want: 1234567890.0},

		{name: "f32_from_f32", t: rdb.TypeFloat32, value: float32(45.5), want: float64(float32(45.5))},
		{name: "f32_from_f64", t: rdb.TypeFloat32, value: float64(45.5), want: float64(float32(45.5))},
	}

	for _, item := range list {
		t.Run(item.name, func(t *testing.T) {
			cmd := &rdb.Command{
				SQL:   `select v = @v`,
				Arity: rdb.OneMust,
			}

			res := db.Query(context.Background(), cmd, rdb.Param{Name: "v", Type: item.t, Value: item.value})
			defer res.Close()

			res.Scan()
			got := res.Getx(0)

			var gotF float64
			switch g := got.(type) {
			case float32:
				gotF = float64(g)
			case float64:
				gotF = g
			default:
				t.Fatalf("unexpected type: %T", got)
			}

			if gotF != item.want {
				t.Errorf("got %v, want %v", gotF, item.want)
			}
		})
	}
}

// TestRdbNullValue tests using rdb.Null explicitly
func TestRdbNullValue(t *testing.T) {
	checkSkip(t)
	if parallel {
		t.Parallel()
	}
	defer assertFreeConns(t)
	defer recoverTest(t)

	cmd := &rdb.Command{
		SQL:   `select v = @v`,
		Arity: rdb.OneMust,
	}

	// Test rdb.Null for various types
	types := []rdb.Type{
		rdb.Integer,
		rdb.TypeFloat64,
		rdb.Text,
		rdb.Binary,
		rdb.TypeBool,
	}

	typeNames := []string{"Integer", "Float64", "Text", "Binary", "Bool"}
	for i, typ := range types {
		t.Run(typeNames[i], func(t *testing.T) {
			res := db.Query(context.Background(), cmd, rdb.Param{Name: "v", Type: typ, Value: rdb.Null, Length: 100})
			defer res.Close()

			res.Scan()
			got := res.Getx(0)

			if got != nil {
				t.Errorf("expected nil for rdb.Null, got %v (%T)", got, got)
			}
		})
	}
}

// TestEmptyBinarySlice tests empty byte slices
func TestEmptyBinarySlice(t *testing.T) {
	checkSkip(t)
	if parallel {
		t.Parallel()
	}
	defer assertFreeConns(t)
	defer recoverTest(t)

	cmd := &rdb.Command{
		SQL:   `select v = @v, len = datalength(@v)`,
		Arity: rdb.OneMust,
	}

	// Test with empty slice - should get empty slice back, not nil
	res := db.Query(context.Background(), cmd, rdb.Param{Name: "v", Type: rdb.Binary, Value: []byte{}, Length: 100})
	defer res.Close()

	res.Scan()
	got := res.Get("v")
	length := res.Get("len")

	if got == nil {
		t.Error("expected empty byte slice, got nil")
	}
	if bytes, ok := got.([]byte); ok {
		if len(bytes) != 0 {
			t.Errorf("expected empty slice, got length %d", len(bytes))
		}
	}
	if length != nil && length.(int32) != 0 {
		t.Errorf("expected length 0, got %v", length)
	}
}

// TestEmptyString tests empty strings
func TestEmptyString(t *testing.T) {
	checkSkip(t)
	if parallel {
		t.Parallel()
	}
	defer assertFreeConns(t)
	defer recoverTest(t)

	cmd := &rdb.Command{
		SQL:   `select v = @v, len = len(@v)`,
		Arity: rdb.OneMust,
	}

	// Test with empty string
	res := db.Query(context.Background(), cmd, rdb.Param{Name: "v", Type: rdb.TypeVarChar, Value: "", Length: 100})
	defer res.Close()

	res.Scan()
	got := res.Get("v")
	length := res.Get("len")

	if got == nil {
		t.Error("expected empty string, got nil")
	}

	gotStr, ok := got.(string)
	if !ok {
		if gotBytes, ok := got.([]byte); ok {
			gotStr = string(gotBytes)
		}
	}
	if gotStr != "" {
		t.Errorf("expected empty string, got %q", gotStr)
	}
	if length != nil && length.(int32) != 0 {
		t.Errorf("expected length 0, got %v", length)
	}
}

// TestTimeDuration tests time.Duration encoding for Time type
func TestTimeDuration(t *testing.T) {
	checkSkip(t)
	if parallel {
		t.Parallel()
	}
	defer assertFreeConns(t)
	defer recoverTest(t)

	list := []struct {
		name  string
		value time.Duration
		want  string
	}{
		{name: "dur_1_hour", value: time.Hour, want: "01:00:00.0000000"},
		{name: "dur_30_min", value: 30 * time.Minute, want: "00:30:00.0000000"},
		{name: "dur_45_sec", value: 45 * time.Second, want: "00:00:45.0000000"},
		{name: "dur_1h30m", value: time.Hour + 30*time.Minute, want: "01:30:00.0000000"},
		{name: "dur_500ms", value: 500 * time.Millisecond, want: "00:00:00.5000000"},
	}

	for _, item := range list {
		t.Run(item.name, func(t *testing.T) {
			cmd := &rdb.Command{
				SQL:   `select v = convert(nvarchar(50), @v, 121)`,
				Arity: rdb.OneMust,
			}

			res := db.Query(context.Background(), cmd, rdb.Param{Name: "v", Type: rdb.TypeTime, Value: item.value})
			defer res.Close()

			res.Scan()
			got := res.Getx(0)

			gotStr, ok := got.(string)
			if !ok {
				if gotBytes, ok := got.([]byte); ok {
					gotStr = string(gotBytes)
				} else {
					t.Fatalf("unexpected type: %T", got)
				}
			}

			if gotStr != item.want {
				t.Errorf("got %q, want %q", gotStr, item.want)
			}
		})
	}
}

// Custom type to test reflection paths
type customInt int32
type customString string
type customBool bool
type customFloat float64

// TestCustomTypes tests reflection-based type handling
func TestCustomTypes(t *testing.T) {
	checkSkip(t)
	if parallel {
		t.Parallel()
	}
	defer assertFreeConns(t)
	defer recoverTest(t)

	list := []struct {
		name  string
		t     rdb.Type
		value interface{}
		want  interface{}
	}{
		{name: "custom_int_to_int32", t: rdb.TypeInt32, value: customInt(12345), want: int32(12345)},
		{name: "custom_int_to_int64", t: rdb.Integer, value: customInt(12345), want: int64(12345)},
		{name: "custom_float", t: rdb.TypeFloat64, value: customFloat(45.67), want: float64(45.67)},
		{name: "custom_bool_true", t: rdb.TypeBool, value: customBool(true), want: true},
		{name: "custom_bool_false", t: rdb.TypeBool, value: customBool(false), want: false},
	}

	for _, item := range list {
		t.Run(item.name, func(t *testing.T) {
			cmd := &rdb.Command{
				SQL:   `select v = @v`,
				Arity: rdb.OneMust,
			}

			res := db.Query(context.Background(), cmd, rdb.Param{Name: "v", Type: item.t, Value: item.value})
			defer res.Close()

			res.Scan()
			got := res.Getx(0)

			if !reflect.DeepEqual(got, item.want) {
				t.Errorf("got %v (%T), want %v (%T)", got, got, item.want, item.want)
			}
		})
	}
}

// TestCustomStringType tests custom string type via reflection
func TestCustomStringType(t *testing.T) {
	checkSkip(t)
	if parallel {
		t.Parallel()
	}
	defer assertFreeConns(t)
	defer recoverTest(t)

	cmd := &rdb.Command{
		SQL:   `select v = @v`,
		Arity: rdb.OneMust,
	}

	cs := customString("custom string value")
	res := db.Query(context.Background(), cmd, rdb.Param{Name: "v", Type: rdb.TypeVarChar, Value: cs, Length: 100})
	defer res.Close()

	res.Scan()
	got := res.Getx(0)

	gotStr, ok := got.(string)
	if !ok {
		if gotBytes, ok := got.([]byte); ok {
			gotStr = string(gotBytes)
		} else {
			t.Fatalf("unexpected type: %T", got)
		}
	}

	if gotStr != string(cs) {
		t.Errorf("got %q, want %q", gotStr, string(cs))
	}
}

// TestStringToIntConversion tests string-to-integer conversion via reflection
func TestStringToIntConversionReflect(t *testing.T) {
	checkSkip(t)
	if parallel {
		t.Parallel()
	}
	defer assertFreeConns(t)
	defer recoverTest(t)

	type stringNum string

	list := []struct {
		name  string
		t     rdb.Type
		value interface{}
		want  interface{}
	}{
		// Note: Custom string type via reflection returns int8 instead of byte
		{name: "custom_str_to_int8", t: rdb.TypeInt8, value: stringNum("42"), want: int8(42)},
		{name: "custom_str_to_int16", t: rdb.TypeInt16, value: stringNum("1234"), want: int16(1234)},
		{name: "custom_str_to_int32", t: rdb.TypeInt32, value: stringNum("123456"), want: int32(123456)},
		{name: "custom_str_to_int64", t: rdb.Integer, value: stringNum("1234567890"), want: int64(1234567890)},
	}

	for _, item := range list {
		t.Run(item.name, func(t *testing.T) {
			cmd := &rdb.Command{
				SQL:   `select v = @v`,
				Arity: rdb.OneMust,
			}

			res := db.Query(context.Background(), cmd, rdb.Param{Name: "v", Type: item.t, Value: item.value})
			defer res.Close()

			res.Scan()
			got := res.Getx(0)

			if !reflect.DeepEqual(got, item.want) {
				t.Errorf("got %v (%T), want %v (%T)", got, got, item.want, item.want)
			}
		})
	}
}
