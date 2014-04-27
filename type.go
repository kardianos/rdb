package rdb

import (
	"reflect"
)

// Each driver should define its own SqlType over value 255.
type SqlType uint32

// Returns true if this is a driver specific type.
func (t SqlType) Driver() bool {
	return t > 255
}
func (t SqlType) String() string {
	return ""
}

type NativeType uint32

var nativeTypeMap = make(map[reflect.Type]NativeType)
var nativeTypeNext uint32 = 256

// Returns true if this is a driver specific type.
func (t NativeType) User() bool {
	return t > 255
}
func (t NativeType) String() string {
	return ""
}

func RegisterNativeType(value interface{}) NativeType {
	t := reflect.TypeOf(value)
	nt, in := nativeTypeMap[t]
	if !in {
		nt = NativeType(nativeTypeNext)
		nativeTypeMap[t] = nt
	}
	return nt
}

type TypeMarshal func(t NativeType) SqlType
type TypeUnmarshal func(t SqlType) NativeType

// TODO: Need type notions.
// * Generic Sql Type
// * Specific Driver Type
// * Native Language Type
// * Mapping the above three
// These type mappings should be able to be set in a hierarchical manner.
// * Config
// * Database
// * Command
// * Parameter
// Ideal to make these mappings reusable.

// Generic Sql Type and Specific Driver type should be the same Go type.
// There should be an inline flag that specifies if the type is specific to
// the driver or to the generic sql type.
// Generic sql type should have both text and ntext. If differentiating
// between text and varchar is needed, the specific driver mapping should be used.
// Context determines what type the driver specific type refers to.
