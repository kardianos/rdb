package ms

import "encoding/binary"

// Collation represents a TDS 5-byte COLLATION structure.
//
// Layout (little-endian uint32 + 1 byte):
//   - Bits 0-19:  LCID (locale identifier)
//   - Bits 20-24: ColFlags (fIgnoreCase, fIgnoreAccent, fIgnoreWidth, fIgnoreKana, fBinary)
//   - Bit  25:    fBinary2
//   - Bit  26:    fUTF8
//   - Bit  27:    FRESERVEDBIT
//   - Bits 28-31: Version
//   - Byte 4:     SortID
type Collation struct {
	LCID    uint32
	Flags   byte
	Version byte
	SortID  byte
}

const collFlagUTF8 = 0x40 // bit 6 of Flags byte (bit 26 of the 32-bit word)

// ParseCollation decodes a 5-byte TDS COLLATION into its components.
func ParseCollation(b [5]byte) Collation {
	v := binary.LittleEndian.Uint32(b[:4])
	return Collation{
		LCID:    v & 0xFFFFF,        // bits 0-19
		Flags:   byte((v >> 20) & 0xFF), // bits 20-27
		Version: byte((v >> 28) & 0x0F), // bits 28-31
		SortID:  b[4],
	}
}

// IsUTF8 reports whether the fUTF8 bit is set, indicating UTF-8 encoding
// for varchar/char data.
func (c Collation) IsUTF8() bool {
	return c.Flags&collFlagUTF8 != 0
}

// Encode serializes the Collation back to a 5-byte TDS representation.
func (c Collation) Encode() [5]byte {
	var b [5]byte
	v := c.LCID & 0xFFFFF
	v |= uint32(c.Flags) << 20
	v |= uint32(c.Version&0x0F) << 28
	binary.LittleEndian.PutUint32(b[:4], v)
	b[4] = c.SortID
	return b
}

// DefaultUTF8Collation returns the collation for Latin1_General_100_CI_AS_SC_UTF8.
// LCID 0x0409 (1033) = English, Flags = CI + fUTF8, Version = 2 (100-series), SortID = 0.
func DefaultUTF8Collation() Collation {
	return Collation{
		LCID:    0x0409,
		Flags:   0x40 | 0x01, // fUTF8 | fIgnoreCase
		Version: 2,           // 100-series collation
		SortID:  0,
	}
}

// DefaultCollation returns the hardcoded SQL_Latin1_General_CP1_CI_AS collation
// that was previously used for all parameters.
func DefaultCollation() Collation {
	return ParseCollation([5]byte{0x09, 0x04, 0xD0, 0x00, 0x34})
}
