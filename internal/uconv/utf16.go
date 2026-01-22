package uconv

import (
	"bytes"
	"encoding/binary"
	"unicode/utf16"
)

const (
	replacementChar = '\uFFFD' // Unicode replacement character.

	// 0xd800-0xdc00 encodes the high 10 bits of a pair.
	// 0xdc00-0xe000 encodes the low 10 bits of a pair.
	// the value is those 20 bits plus 0x10000.
	surr1 = 0xd800
	surr2 = 0xdc00
	surr3 = 0xe000
)

// Utf8 to Utf16 LE:
//  * string -> []byte
//  * []byte -> []byte
//  * []byte (->) *byte.Buffer
// Utf16 LE to Utf8.
//  * []byte -> string
//  * []byte -> []byte
//  * []byte (->) *bytes.Buffer

type PanicReader func(n int) []byte

var Encode Utf8To16Le
var Decode Utf16LeTo8

type Utf8To16Le struct{}

func (code Utf8To16Le) ToBuffer(s []byte, coded *bytes.Buffer) {
	coded.Write(code.FromBytes(s))
}
func (code Utf8To16Le) FromBytes(s []byte) []byte {
	uu := utf16.Encode([]rune(string(s)))
	bb := make([]byte, len(uu)*2)
	for i, u := range uu {
		binary.LittleEndian.PutUint16(bb[i*2:], u)
	}
	return bb
}
func (code Utf8To16Le) FromString(s string) []byte {
	return code.FromBytes([]byte(s))
}

type Utf16LeTo8 struct{}

func (code Utf16LeTo8) ToBuffer(s []byte, coded *bytes.Buffer) {
	if len(s) < 2 {
		return
	}

	var r, r1 uint16
	for i := 0; i+1 < len(s); i += 2 {
		r = combineBytesLE(s[i], s[i+1])
		hasAnother := false
		canSurr := surr1 <= r && r < surr2

		if canSurr && i+3 < len(s) {
			hasAnother = true
			r1 = combineBytesLE(s[i+2], s[i+3])
		}

		switch {
		case canSurr && hasAnother && surr2 <= r1 && r1 < surr3:
			// valid surrogate sequence
			coded.WriteRune(decodeRune(rune(r), rune(r1)))
			i += 2
		case surr1 <= r && r < surr3:
			// invalid surrogate sequence
			coded.WriteRune(rune(replacementChar))
		default:
			// normal rune
			coded.WriteRune(rune(r))
		}
	}
}
func (code Utf16LeTo8) ToString(s []byte) string {
	return string(code.ToBytes(s))
}
func (code Utf16LeTo8) ToBytes(s []byte) []byte {
	if len(s) < 2 {
		return nil
	}
	a := bytes.NewBuffer(make([]byte, 0, len(s)/2))
	code.ToBuffer(s, a)
	return a.Bytes()
}
func (code Utf16LeTo8) Prefix1(r PanicReader) (int, string) {
	textLength := int(r(1)[0]) * 2
	return 1 + textLength, code.ToString(r(textLength))
}
func (code Utf16LeTo8) Prefix2(r PanicReader) (int, string) {
	textLength := int(binary.LittleEndian.Uint16(r(2))) * 2
	return 2 + textLength, code.ToString(r(textLength))
}

type Utf16BeTo8 struct{}

func (code Utf16BeTo8) ToBuffer(s []byte, coded *bytes.Buffer) {
	if len(s) < 2 {
		return
	}

	var r, r1 uint16
	for i := 0; i+1 < len(s); i += 2 {
		r = combineBytesBE(s[i], s[i+1])
		hasAnother := false
		canSurr := surr1 <= r && r < surr2

		if canSurr && i+3 < len(s) {
			hasAnother = true
			r1 = combineBytesBE(s[i+2], s[i+3])
		}

		switch {
		case canSurr && hasAnother && surr2 <= r1 && r1 < surr3:
			// valid surrogate sequence
			coded.WriteRune(decodeRune(rune(r), rune(r1)))
			i += 2
		case surr1 <= r && r < surr3:
			// invalid surrogate sequence
			coded.WriteRune(rune(replacementChar))
		default:
			// normal rune
			coded.WriteRune(rune(r))
		}
	}
}
func (code Utf16BeTo8) ToString(s []byte) string {
	return string(code.ToBytes(s))
}
func (code Utf16BeTo8) ToBytes(s []byte) []byte {
	if len(s) < 2 {
		return nil
	}
	a := bytes.NewBuffer(make([]byte, 0, len(s)/2))
	code.ToBuffer(s, a)
	return a.Bytes()
}
func (code Utf16BeTo8) Prefix1(r PanicReader) (int, string) {
	textLength := int(r(1)[0]) * 2
	return 1 + textLength, code.ToString(r(textLength))
}
func (code Utf16BeTo8) Prefix2(r PanicReader) (int, string) {
	textLength := int(binary.LittleEndian.Uint16(r(2))) * 2
	return 2 + textLength, code.ToString(r(textLength))
}

func decodeRune(r1, r2 rune) rune {
	if surr1 <= r1 && r1 < surr2 && surr2 <= r2 && r2 < surr3 {
		return (r1-surr1)<<10 | (r2 - surr2) + 0x10000
	}
	return replacementChar
}

func combineBytesLE(b1, b2 byte) uint16 {
	return uint16(b1) | (uint16(b2) << 8)
}
func combineBytesBE(b1, b2 byte) uint16 {
	return uint16(b2) | (uint16(b1) << 8)
}
