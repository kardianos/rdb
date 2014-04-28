package rdb

import (
	"reflect"
)

const (
	SqlTypeDriverThresh = 0x00010000
)

// Sql Type constants are not represented in all database systems.
// Names were chosen to afford the best understanding from the go language
// and not from the sql standard.j
const (
	SqlTypeUnknown SqlType = iota // Zero default value.

	SqlTypeNull // A special "type" that can indicate a null value.

	// Driver defaults for text varying lengths.
	SqlTypeString     // Unicode text. Some drivers call this ntext or nvarchar.
	SqlTypeAnsiString // Ansi text. Some drivers call this just text or varchar.
	SqlTypeBinary     // Just a string of bytes.

	// Specific character data types.
	SqlTypeText        // Unicode text with varying length. Also nvarchar.
	SqlTypeAnsiText    // Ansi text with varying length. Also varchar.
	SqlTypeVarChar     // Unicode text with varying length. Also nvarchar.
	SqlTypeAnsiVarChar // Ansi text with varying length. Also varchar.
	SqlTypeChar        // Unicode text with fixed length. Also nchar.
	SqlTypeAnsiChar    // Ansi text with fixed length. Also char.

	SqlTypeBool   // Also bit.
	SqlTypeUint8  // Also unsigned tiny int.
	SqlTypeUint16 // Also unsigned small int.
	SqlTypeUint32 // Also unsigned int.
	SqlTypeUint64 // Also unsigned big int.
	SqlTypeInt8   // Also tiny int.
	SqlTypeInt16  // Also small int.
	SqlTypeInt32  // Also int.
	SqlTypeInt64  // Also big int.

	// Auto-increment integer.
	SqlTypeSerial16
	SqlTypeSerial32
	SqlTypeSerial64

	SqlTypeFloat32 // Floating point number.
	SqlTypeFloat64 // Floating point number, "double" width.

	SqlTypeDecimal // Exact number with specified scale and precision.
	SqlTypeMoney

	SqlTypeTime         // Contains time, date, and time zone.
	SqlTypeDuration     // Contains a span of time.
	SqlTypeOnlyTime     // Only contains time of day.
	SqlTypeOnlyDate     // Only contains a date.
	SqlTypeOnlyDateTime // Only contains a date and time, no time zone.

	SqlTypeUUID // Also uniqueidentifier or GUID.

	SqlTypeEnum
	SqlTypeRange
	SqlTypeArray
	SqlTypeJson
	SqlTypeXml
	SqlTypeTable
)

// Each driver should define its own SqlType over value SqlTypeDriverThresh (65536).
// Values over SqlTypeDriverThresh establish thier own namespace for types.
// Driver types are often limited to 16 bits so that leaves enough space Open
// for more then one type spaces or user types.
type SqlType uint32

// Returns true if this is a driver specific type.
func (t SqlType) Driver() bool {
	return t >= 0x00010000
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

type TypeMarshal struct {
	From NativeType
	To   SqlType
}
type TypeUnmarshal struct {
	From SqlType
	To   NativeType
}

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
