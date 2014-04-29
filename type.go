package rdb

import (
	"reflect"
)

const (
	TypeDriverThresh = 0x00010000
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

// Sql Type constants are not represented in all database systems.
// Names were chosen to afford the best understanding from the go language
// and not from the sql standard.j
const (
	TypeUnknown SqlType = iota // Zero default value.

	TypeNull // A special "type" that can indicate a null value.

	// Driver defaults for text varying lengths.
	TypeString     // Unicode text. Some drivers call this ntext or nvarchar.
	TypeAnsiString // Ansi text. Some drivers call this just text or varchar.
	TypeBinary     // Just a string of bytes.

	// Specific character data types.
	TypeText        // Unicode text with varying length. Also nvarchar.
	TypeAnsiText    // Ansi text with varying length. Also varchar.
	TypeVarChar     // Unicode text with varying length. Also nvarchar.
	TypeAnsiVarChar // Ansi text with varying length. Also varchar.
	TypeChar        // Unicode text with fixed length. Also nchar.
	TypeAnsiChar    // Ansi text with fixed length. Also char.

	TypeBool   // Also bit.
	TypeUint8  // Also unsigned tiny int.
	TypeUint16 // Also unsigned small int.
	TypeUint32 // Also unsigned int.
	TypeUint64 // Also unsigned big int.
	TypeInt8   // Also tiny int.
	TypeInt16  // Also small int.
	TypeInt32  // Also int.
	TypeInt64  // Also big int.

	// Auto-increment integer.
	TypeSerial16
	TypeSerial32
	TypeSerial64

	TypeFloat32 // Floating point number.
	TypeFloat64 // Floating point number, "double" width.

	TypeDecimal // Exact number with specified scale and precision.
	TypeMoney

	TypeTime         // Contains time, date, and time zone.
	TypeDuration     // Contains a span of time.
	TypeOnlyTime     // Only contains time of day.
	TypeOnlyDate     // Only contains a date.
	TypeOnlyDateTime // Only contains a date and time, no time zone.

	TypeUUID // Also uniqueidentifier or GUID.

	TypeEnum
	TypeRange
	TypeArray
	TypeJson
	TypeXml
	TypeTable
)

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
