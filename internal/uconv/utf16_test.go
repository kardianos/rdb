package uconv

import (
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
