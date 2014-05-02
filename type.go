// Copyright 2014 Daniel Theophanes.
// Use of this source code is governed by a zlib-style
// license that can be found in the LICENSE file.

package rdb

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
	return t >= TypeDriverThresh
}

// Sql Type constants are not represented in all database systems.
// Names were chosen to afford the best understanding from the go language
// and not from the sql standard.
const (
	TypeUnknown SqlType = iota // Zero default value.

	TypeNull // A special "type" that can indicate a null value.

	// Driver defaults for text varying lengths.
	// PostgreSQL will use "text", Tds will use "nvarchar", Oracle will use "nvarchar2".
	// In greenfield development it is suggested to use this type text rather then another type.
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

	// Auto-incrementing integer.
	TypeSerial16
	TypeSerial32
	TypeSerial64

	TypeFloat32 // Floating point number, also real.
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
