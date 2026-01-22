package uconv

import (
	"bytes"
	"testing"
)

type decodeTest struct {
	in  []byte
	out []rune
}

var decodeTests = []decodeTest{
	{[]byte{0, 1, 0, 2, 0, 3, 0, 4}, []rune{1, 2, 3, 4}},
	{[]byte{0xff, 0xff, 0xd8, 0x00, 0xdc, 0x00, 0xd8, 0x00, 0xdc, 0x01, 0xd8, 0x08, 0xdf, 0x45, 0xdb, 0xff, 0xdf, 0xff},
		[]rune{0xffff, 0x10000, 0x10001, 0x12345, 0x10ffff}},
	{[]byte{0xd8, 0x00, 0x00, 'a'}, []rune{0xfffd, 'a'}},
	{[]byte{0xdf, 0xff}, []rune{0xfffd}},
}

func TestEndian(t *testing.T) {
	var n1, n2 uint16

	n1 = 0x0034
	n2 = combineBytesBE(0x00, 0x34)
	if n1 != n2 {
		t.Errorf("Want: 0x%X, Got: 0x%X", n1, n2)
	}
}

func TestDecode(t *testing.T) {
	var decode Utf16BeTo8
	for _, tt := range decodeTests {
		out := decode.ToString(tt.in)
		i := 0
		for _, r3 := range out {
			if r3 != tt.out[i] {
				t.Errorf("Decode3(%x) = %x; want %x", tt.in, out, tt.out)
			}
			i++
		}
	}
}

func TestCombineBytesLE(t *testing.T) {
	tests := []struct {
		b1, b2 byte
		want   uint16
	}{
		{0x34, 0x12, 0x1234},
		{0x00, 0x00, 0x0000},
		{0xFF, 0xFF, 0xFFFF},
		{0x01, 0x00, 0x0001},
		{0x00, 0x01, 0x0100},
	}
	for _, tt := range tests {
		got := combineBytesLE(tt.b1, tt.b2)
		if got != tt.want {
			t.Errorf("combineBytesLE(0x%02X, 0x%02X) = 0x%04X; want 0x%04X", tt.b1, tt.b2, got, tt.want)
		}
	}
}

func TestDecodeRune(t *testing.T) {
	tests := []struct {
		r1, r2 rune
		want   rune
	}{
		{0xD800, 0xDC00, 0x10000},      // valid surrogate pair
		{0xD800, 0xDC01, 0x10001},      // valid surrogate pair
		{0xDBFF, 0xDFFF, 0x10FFFF},     // max valid surrogate pair
		{0xD834, 0xDD1E, 0x1D11E},      // G clef (U+1D11E)
		{0x0041, 0xDC00, replacementChar}, // invalid: r1 not in surrogate range
		{0xD800, 0x0041, replacementChar}, // invalid: r2 not in surrogate range
	}
	for _, tt := range tests {
		got := decodeRune(tt.r1, tt.r2)
		if got != tt.want {
			t.Errorf("decodeRune(0x%04X, 0x%04X) = 0x%04X; want 0x%04X", tt.r1, tt.r2, got, tt.want)
		}
	}
}

// Tests for Utf8To16Le

func TestUtf8To16Le_FromBytes(t *testing.T) {
	var enc Utf8To16Le
	tests := []struct {
		in   string
		want []byte
	}{
		{"A", []byte{0x41, 0x00}},
		{"AB", []byte{0x41, 0x00, 0x42, 0x00}},
		{"hello", []byte{0x68, 0x00, 0x65, 0x00, 0x6c, 0x00, 0x6c, 0x00, 0x6f, 0x00}},
		{"", []byte{}},
		// Unicode characters
		{"日", []byte{0xe5, 0x65}}, // U+65E5
		// Surrogate pair (emoji: 😀 U+1F600)
		{"😀", []byte{0x3d, 0xd8, 0x00, 0xde}},
	}
	for _, tt := range tests {
		got := enc.FromBytes([]byte(tt.in))
		if !bytes.Equal(got, tt.want) {
			t.Errorf("FromBytes(%q) = %v; want %v", tt.in, got, tt.want)
		}
	}
}

func TestUtf8To16Le_FromString(t *testing.T) {
	var enc Utf8To16Le
	input := "test"
	got := enc.FromString(input)
	want := enc.FromBytes([]byte(input))
	if !bytes.Equal(got, want) {
		t.Errorf("FromString(%q) = %v; want %v", input, got, want)
	}
}

func TestUtf8To16Le_ToBuffer(t *testing.T) {
	var enc Utf8To16Le
	input := []byte("hello")
	var buf bytes.Buffer
	enc.ToBuffer(input, &buf)
	want := enc.FromBytes(input)
	if !bytes.Equal(buf.Bytes(), want) {
		t.Errorf("ToBuffer(%q) = %v; want %v", input, buf.Bytes(), want)
	}
}

// Tests for Utf16LeTo8

func TestUtf16LeTo8_ToBytes(t *testing.T) {
	var dec Utf16LeTo8
	tests := []struct {
		in      []byte
		want    string
		wantNil bool
	}{
		{[]byte{0x41, 0x00}, "A", false},
		{[]byte{0x41, 0x00, 0x42, 0x00}, "AB", false},
		{[]byte{0x68, 0x00, 0x65, 0x00, 0x6c, 0x00, 0x6c, 0x00, 0x6f, 0x00}, "hello", false},
		{[]byte{}, "", true},
		{[]byte{0x00}, "", true}, // too short
		// Unicode character
		{[]byte{0xe5, 0x65}, "日", false}, // U+65E5
		// Surrogate pair (emoji: 😀 U+1F600)
		{[]byte{0x3d, 0xd8, 0x00, 0xde}, "😀", false},
	}
	for _, tt := range tests {
		got := dec.ToBytes(tt.in)
		if tt.wantNil {
			if got != nil {
				t.Errorf("ToBytes(%v) = %v; want nil", tt.in, got)
			}
		} else if string(got) != tt.want {
			t.Errorf("ToBytes(%v) = %q; want %q", tt.in, string(got), tt.want)
		}
	}
}

func TestUtf16LeTo8_ToString(t *testing.T) {
	var dec Utf16LeTo8
	input := []byte{0x68, 0x00, 0x69, 0x00} // "hi"
	got := dec.ToString(input)
	if got != "hi" {
		t.Errorf("ToString(%v) = %q; want %q", input, got, "hi")
	}
}

func TestUtf16LeTo8_ToBuffer(t *testing.T) {
	var dec Utf16LeTo8
	input := []byte{0x68, 0x00, 0x69, 0x00} // "hi"
	var buf bytes.Buffer
	dec.ToBuffer(input, &buf)
	if buf.String() != "hi" {
		t.Errorf("ToBuffer(%v) = %q; want %q", input, buf.String(), "hi")
	}
}

func TestUtf16LeTo8_ToBuffer_Empty(t *testing.T) {
	var dec Utf16LeTo8
	var buf bytes.Buffer
	dec.ToBuffer([]byte{}, &buf)
	if buf.Len() != 0 {
		t.Errorf("ToBuffer empty should produce empty buffer, got %d bytes", buf.Len())
	}
	dec.ToBuffer([]byte{0x00}, &buf)
	if buf.Len() != 0 {
		t.Errorf("ToBuffer single byte should produce empty buffer, got %d bytes", buf.Len())
	}
}

func TestUtf16LeTo8_ToBuffer_InvalidSurrogate(t *testing.T) {
	var dec Utf16LeTo8
	// Invalid surrogate: high surrogate without valid low surrogate
	input := []byte{0x00, 0xd8, 0x41, 0x00} // D800 followed by 'A'
	var buf bytes.Buffer
	dec.ToBuffer(input, &buf)
	// Should get replacement char + 'A'
	got := buf.String()
	if got != "\uFFFDA" {
		t.Errorf("ToBuffer invalid surrogate = %q; want %q", got, "\uFFFDA")
	}
}

func TestUtf16LeTo8_ToBuffer_SurrogateAtEnd(t *testing.T) {
	var dec Utf16LeTo8
	// High surrogate at end (no room for low surrogate)
	input := []byte{0x00, 0xd8} // Just D800
	var buf bytes.Buffer
	dec.ToBuffer(input, &buf)
	got := buf.String()
	if got != "\uFFFD" {
		t.Errorf("ToBuffer surrogate at end = %q; want %q", got, "\uFFFD")
	}
}

func TestUtf16LeTo8_Prefix1(t *testing.T) {
	var dec Utf16LeTo8
	// Create data: length byte (2 chars = 4 bytes) + "hi" in UTF-16 LE
	data := []byte{0x02, 0x68, 0x00, 0x69, 0x00}
	pos := 0
	reader := func(n int) []byte {
		result := data[pos : pos+n]
		pos += n
		return result
	}
	length, str := dec.Prefix1(reader)
	if length != 5 { // 1 + 4
		t.Errorf("Prefix1 length = %d; want 5", length)
	}
	if str != "hi" {
		t.Errorf("Prefix1 string = %q; want %q", str, "hi")
	}
}

func TestUtf16LeTo8_Prefix2(t *testing.T) {
	var dec Utf16LeTo8
	// Create data: 2-byte length (2 chars = 4 bytes) + "hi" in UTF-16 LE
	data := []byte{0x02, 0x00, 0x68, 0x00, 0x69, 0x00}
	pos := 0
	reader := func(n int) []byte {
		result := data[pos : pos+n]
		pos += n
		return result
	}
	length, str := dec.Prefix2(reader)
	if length != 6 { // 2 + 4
		t.Errorf("Prefix2 length = %d; want 6", length)
	}
	if str != "hi" {
		t.Errorf("Prefix2 string = %q; want %q", str, "hi")
	}
}

// Tests for Utf16BeTo8

func TestUtf16BeTo8_ToBytes(t *testing.T) {
	var dec Utf16BeTo8
	tests := []struct {
		in      []byte
		want    string
		wantNil bool
	}{
		{[]byte{0x00, 0x41}, "A", false},
		{[]byte{0x00, 0x41, 0x00, 0x42}, "AB", false},
		{[]byte{}, "", true},
		{[]byte{0x00}, "", true}, // too short
		// Unicode character
		{[]byte{0x65, 0xe5}, "日", false}, // U+65E5
	}
	for _, tt := range tests {
		got := dec.ToBytes(tt.in)
		if tt.wantNil {
			if got != nil {
				t.Errorf("ToBytes(%v) = %v; want nil", tt.in, got)
			}
		} else if string(got) != tt.want {
			t.Errorf("ToBytes(%v) = %q; want %q", tt.in, string(got), tt.want)
		}
	}
}

func TestUtf16BeTo8_ToBuffer(t *testing.T) {
	var dec Utf16BeTo8
	input := []byte{0x00, 0x68, 0x00, 0x69} // "hi"
	var buf bytes.Buffer
	dec.ToBuffer(input, &buf)
	if buf.String() != "hi" {
		t.Errorf("ToBuffer(%v) = %q; want %q", input, buf.String(), "hi")
	}
}

func TestUtf16BeTo8_ToBuffer_Empty(t *testing.T) {
	var dec Utf16BeTo8
	var buf bytes.Buffer
	dec.ToBuffer([]byte{}, &buf)
	if buf.Len() != 0 {
		t.Errorf("ToBuffer empty should produce empty buffer, got %d bytes", buf.Len())
	}
}

func TestUtf16BeTo8_Prefix1(t *testing.T) {
	var dec Utf16BeTo8
	// Create data: length byte (2 chars = 4 bytes) + "hi" in UTF-16 BE
	data := []byte{0x02, 0x00, 0x68, 0x00, 0x69}
	pos := 0
	reader := func(n int) []byte {
		result := data[pos : pos+n]
		pos += n
		return result
	}
	length, str := dec.Prefix1(reader)
	if length != 5 { // 1 + 4
		t.Errorf("Prefix1 length = %d; want 5", length)
	}
	if str != "hi" {
		t.Errorf("Prefix1 string = %q; want %q", str, "hi")
	}
}

func TestUtf16BeTo8_Prefix2(t *testing.T) {
	var dec Utf16BeTo8
	// Create data: 2-byte length (2 chars = 4 bytes) + "hi" in UTF-16 BE
	data := []byte{0x02, 0x00, 0x00, 0x68, 0x00, 0x69}
	pos := 0
	reader := func(n int) []byte {
		result := data[pos : pos+n]
		pos += n
		return result
	}
	length, str := dec.Prefix2(reader)
	if length != 6 { // 2 + 4
		t.Errorf("Prefix2 length = %d; want 6", length)
	}
	if str != "hi" {
		t.Errorf("Prefix2 string = %q; want %q", str, "hi")
	}
}

// Test roundtrip encoding/decoding

func TestRoundtrip(t *testing.T) {
	var enc Utf8To16Le
	var dec Utf16LeTo8
	tests := []string{
		"hello",
		"Hello, 世界",
		"😀🎉",
		"",
	}
	for _, tt := range tests {
		if tt == "" {
			continue // empty string won't roundtrip due to nil return
		}
		encoded := enc.FromString(tt)
		decoded := dec.ToString(encoded)
		if decoded != tt {
			t.Errorf("Roundtrip(%q): encoded=%v, decoded=%q", tt, encoded, decoded)
		}
	}
}

// Test global Encode/Decode vars

func TestGlobalEncodeVar(t *testing.T) {
	got := Encode.FromString("A")
	want := []byte{0x41, 0x00}
	if !bytes.Equal(got, want) {
		t.Errorf("Encode.FromString(\"A\") = %v; want %v", got, want)
	}
}

func TestGlobalDecodeVar(t *testing.T) {
	got := Decode.ToString([]byte{0x41, 0x00})
	if got != "A" {
		t.Errorf("Decode.ToString = %q; want \"A\"", got)
	}
}
