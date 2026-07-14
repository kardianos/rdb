package uconv

import (
	"bytes"
	"encoding/binary"
	"unicode/utf16"
	"unicode/utf8"
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
	// Encode directly into the buffer without an intermediate slice when possible.
	if len(s) == 0 {
		return
	}
	// Grow once for the worst-case UTF-16LE size (2 bytes per UTF-8 byte).
	coded.Grow(len(s) * 2)
	start := coded.Len()
	// Ensure space by writing zeros then filling; simpler: use Append path.
	var scratch [4]byte
	for i := 0; i < len(s); {
		r, size := utf8.DecodeRune(s[i:])
		i += size
		n := encodeRuneUTF16LE(r, scratch[:])
		coded.Write(scratch[:n])
	}
	_ = start
}

// encodeRuneUTF16LE writes r as UTF-16LE into dst (must have room for 4 bytes).
// Returns the number of bytes written (2 or 4).
func encodeRuneUTF16LE(r rune, dst []byte) int {
	if r < 0 || r > utf8.MaxRune {
		r = replacementChar
	}
	if surr1 <= r && r < surr3 {
		// Lone surrogate — not a valid scalar value.
		r = replacementChar
	}
	if r < 0x10000 {
		dst[0] = byte(r)
		dst[1] = byte(r >> 8)
		return 2
	}
	r1, r2 := utf16.EncodeRune(r)
	dst[0] = byte(r1)
	dst[1] = byte(r1 >> 8)
	dst[2] = byte(r2)
	dst[3] = byte(r2 >> 8)
	return 4
}

func (code Utf8To16Le) FromBytes(s []byte) []byte {
	if len(s) == 0 {
		// Non-nil empty: callers distinguish nil (SQL NULL) from empty string.
		return []byte{}
	}
	// Worst-case UTF-16LE size is 2 bytes per input byte (all ASCII).
	bb := make([]byte, len(s)*2)
	n := 0
	for i := 0; i < len(s); {
		r, size := utf8.DecodeRune(s[i:])
		i += size
		n += encodeRuneUTF16LE(r, bb[n:])
	}
	return bb[:n]
}

func (code Utf8To16Le) FromString(s string) []byte {
	if len(s) == 0 {
		// Non-nil empty: callers distinguish nil (SQL NULL) from empty string.
		return []byte{}
	}
	// Worst-case UTF-16LE size is 2 bytes per input byte (all ASCII).
	bb := make([]byte, len(s)*2)
	n := 0
	for _, r := range s {
		n += encodeRuneUTF16LE(r, bb[n:])
	}
	return bb[:n]
}

// AppendString appends the UTF-16LE encoding of s onto dst and returns the result.
func (code Utf8To16Le) AppendString(dst []byte, s string) []byte {
	for _, r := range s {
		var scratch [4]byte
		n := encodeRuneUTF16LE(r, scratch[:])
		dst = append(dst, scratch[:n]...)
	}
	return dst
}

// AppendBytes appends the UTF-16LE encoding of s onto dst and returns the result.
func (code Utf8To16Le) AppendBytes(dst []byte, s []byte) []byte {
	for i := 0; i < len(s); {
		r, size := utf8.DecodeRune(s[i:])
		i += size
		var scratch [4]byte
		n := encodeRuneUTF16LE(r, scratch[:])
		dst = append(dst, scratch[:n]...)
	}
	return dst
}

type Utf16LeTo8 struct{}

func (code Utf16LeTo8) ToBuffer(s []byte, coded *bytes.Buffer) {
	if len(s) < 2 {
		return
	}
	// Grow for worst-case UTF-8 (3 bytes per UTF-16 code unit for BMP).
	coded.Grow((len(s) / 2) * 3)
	var buf [utf8.UTFMax]byte
	for i := 0; i+1 < len(s); i += 2 {
		r := combineBytesLE(s[i], s[i+1])
		hasAnother := false
		canSurr := surr1 <= r && r < surr2
		var r1 uint16
		if canSurr && i+3 < len(s) {
			hasAnother = true
			r1 = combineBytesLE(s[i+2], s[i+3])
		}

		var rr rune
		switch {
		case canSurr && hasAnother && surr2 <= r1 && r1 < surr3:
			rr = decodeRune(rune(r), rune(r1))
			i += 2
		case surr1 <= r && r < surr3:
			rr = replacementChar
		default:
			rr = rune(r)
		}
		n := utf8.EncodeRune(buf[:], rr)
		coded.Write(buf[:n])
	}
}

func (code Utf16LeTo8) ToString(s []byte) string {
	return string(code.ToBytes(s))
}

func (code Utf16LeTo8) ToBytes(s []byte) []byte {
	if len(s) < 2 {
		return nil
	}
	need := utf16LEDecodedLen(s)
	out := make([]byte, need)
	n := decodeUTF16LEInto(out, s)
	return out[:n]
}

// AppendBytes appends the UTF-8 decoding of UTF-16LE s onto dst.
func (code Utf16LeTo8) AppendBytes(dst []byte, s []byte) []byte {
	if len(s) < 2 {
		return dst
	}
	need := utf16LEDecodedLen(s)
	off := len(dst)
	if cap(dst)-off < need {
		nd := make([]byte, off+need)
		copy(nd, dst)
		dst = nd
	} else {
		dst = dst[:off+need]
	}
	n := decodeUTF16LEInto(dst[off:off+need], s)
	return dst[:off+n]
}

// decodeUTF16LEInto writes UTF-8 of s into out (must be large enough). Returns bytes written.
func decodeUTF16LEInto(out []byte, s []byte) int {
	n := 0
	for i := 0; i+1 < len(s); i += 2 {
		r := combineBytesLE(s[i], s[i+1])
		hasAnother := false
		canSurr := surr1 <= r && r < surr2
		var r1 uint16
		if canSurr && i+3 < len(s) {
			hasAnother = true
			r1 = combineBytesLE(s[i+2], s[i+3])
		}

		var rr rune
		switch {
		case canSurr && hasAnother && surr2 <= r1 && r1 < surr3:
			rr = decodeRune(rune(r), rune(r1))
			i += 2
		case surr1 <= r && r < surr3:
			rr = replacementChar
		default:
			rr = rune(r)
		}
		n += utf8.EncodeRune(out[n:], rr)
	}
	return n
}

// utf16LEDecodedLen returns the exact UTF-8 byte length of decoding s as UTF-16LE.
func utf16LEDecodedLen(s []byte) int {
	n := 0
	for i := 0; i+1 < len(s); i += 2 {
		r := combineBytesLE(s[i], s[i+1])
		hasAnother := false
		canSurr := surr1 <= r && r < surr2
		var r1 uint16
		if canSurr && i+3 < len(s) {
			hasAnother = true
			r1 = combineBytesLE(s[i+2], s[i+3])
		}
		var rr rune
		switch {
		case canSurr && hasAnother && surr2 <= r1 && r1 < surr3:
			rr = decodeRune(rune(r), rune(r1))
			i += 2
		case surr1 <= r && r < surr3:
			rr = replacementChar
		default:
			rr = rune(r)
		}
		rl := utf8.RuneLen(rr)
		if rl < 0 {
			rl = utf8.RuneLen(replacementChar)
		}
		n += rl
	}
	return n
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
	coded.Grow((len(s) / 2) * 3)
	var buf [utf8.UTFMax]byte
	for i := 0; i+1 < len(s); i += 2 {
		r := combineBytesBE(s[i], s[i+1])
		hasAnother := false
		canSurr := surr1 <= r && r < surr2
		var r1 uint16
		if canSurr && i+3 < len(s) {
			hasAnother = true
			r1 = combineBytesBE(s[i+2], s[i+3])
		}

		var rr rune
		switch {
		case canSurr && hasAnother && surr2 <= r1 && r1 < surr3:
			rr = decodeRune(rune(r), rune(r1))
			i += 2
		case surr1 <= r && r < surr3:
			rr = replacementChar
		default:
			rr = rune(r)
		}
		n := utf8.EncodeRune(buf[:], rr)
		coded.Write(buf[:n])
	}
}
func (code Utf16BeTo8) ToString(s []byte) string {
	return string(code.ToBytes(s))
}
func (code Utf16BeTo8) ToBytes(s []byte) []byte {
	if len(s) < 2 {
		return nil
	}
	need := utf16BEDecodedLen(s)
	out := make([]byte, need)
	n := 0
	for i := 0; i+1 < len(s); i += 2 {
		r := combineBytesBE(s[i], s[i+1])
		hasAnother := false
		canSurr := surr1 <= r && r < surr2
		var r1 uint16
		if canSurr && i+3 < len(s) {
			hasAnother = true
			r1 = combineBytesBE(s[i+2], s[i+3])
		}

		var rr rune
		switch {
		case canSurr && hasAnother && surr2 <= r1 && r1 < surr3:
			rr = decodeRune(rune(r), rune(r1))
			i += 2
		case surr1 <= r && r < surr3:
			rr = replacementChar
		default:
			rr = rune(r)
		}
		n += utf8.EncodeRune(out[n:], rr)
	}
	return out[:n]
}

func utf16BEDecodedLen(s []byte) int {
	n := 0
	for i := 0; i+1 < len(s); i += 2 {
		r := combineBytesBE(s[i], s[i+1])
		hasAnother := false
		canSurr := surr1 <= r && r < surr2
		var r1 uint16
		if canSurr && i+3 < len(s) {
			hasAnother = true
			r1 = combineBytesBE(s[i+2], s[i+3])
		}
		var rr rune
		switch {
		case canSurr && hasAnother && surr2 <= r1 && r1 < surr3:
			rr = decodeRune(rune(r), rune(r1))
			i += 2
		case surr1 <= r && r < surr3:
			rr = replacementChar
		default:
			rr = rune(r)
		}
		rl := utf8.RuneLen(rr)
		if rl < 0 {
			rl = utf8.RuneLen(replacementChar)
		}
		n += rl
	}
	return n
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
