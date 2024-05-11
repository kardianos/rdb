// Copyright 2014 Daniel Theophanes.
// Use of this source code is governed by a zlib-style
// license that can be found in the LICENSE file.

package rdb

type NullType struct{}

var Null NullType

const (
	TypeDriverThresh = 0x00010000
)

// Each driver should define its own Type over value TypeDriverThresh (65536).
// Values over TypeDriverThresh establish thier own namespace for types.
// Driver types are often limited to 16 bits so that leaves enough space Open
// for more then one type spaces or user types.
type Type uint32

// Returns true if this is a driver defined type.
func (t Type) Driver() bool {
	return t >= TypeDriverThresh
}

// Generic returns true if the type is a generic type.
func (t Type) Generic() bool {
	return t >= 16 && t < 1024
}

const (
	TypeUnknown Type = 0
)

// Generic SQL types. Can be used in parameters.
// Reported in SqlColumn.Generic.
const (
	Text Type = 16 + iota
	Binary
	Bool
	Integer
	Float
	Decimal
	Time
	Other
)

// Type constants are not represented in all database systems.
// Additional sql types may be recognized per driver, but such types
// must have a vlaue greater then TypeDriverThresh.
const (
	// Driver defaults for text varying lengths.
	// Specific character data types.
	TypeText        Type = 1024 + iota // Unicode text with varying length. Also nvarchar.
	TypeAnsiText                       // Ansi text with varying length. Also varchar.
	TypeVarChar                        // Unicode text with varying length. Also nvarchar.
	TypeAnsiVarChar                    // Ansi text with varying length. Also varchar.
	TypeChar                           // Unicode text with fixed length. Also nchar.
	TypeAnsiChar                       // Ansi text with fixed length. Also char.

	TypeBinary // Byte array.

	TypeBool   // Also bit.
	TypeUint8  // Also unsigned tiny int.
	TypeUint16 // Also unsigned small int.
	TypeUint32 // Also unsigned int.
	TypeUint64 // Also unsigned big int.
	TypeInt8   // Also tiny int.
	TypeInt16  // Also small int.
	TypeInt32  // Also int.
	TypeInt64  // Also big int.

	// Auto-incrementing integer.
	TypeSerial16
	TypeSerial32
	TypeSerial64

	TypeFloat32 // Floating point number, also real.
	TypeFloat64 // Floating point number, "double" width.

	TypeDecimal // Exact number with specified scale and precision.
	TypeMoney

	TypeTimestampz // Contains time, date, and time zone.
	TypeDuration   // Contains a span of time.
	TypeTime       // Only contains time of day.
	TypeDate       // Only contains a date.
	TypeTimestamp  // Only contains a time and, no time zone.

	TypeUUID // Also uniqueidentifier or GUID.

	TypeEnum
	TypeRange
	TypeArray
	TypeJSON
	TypeXML
	TypeTable
)
