package ms

import (
	"bytes"
	"fmt"
	"testing"
)

func TestFlagRoundtrip(t *testing.T) {
	list := []colFlags{
		{Nullable: true, Serial: true, Key: true, SparseColumnSet: true, NullableUnknown: true},
		{},
		{Nullable: true, Serial: false, Key: true, SparseColumnSet: false, NullableUnknown: true},
	}
	for i, f1 := range list {
		t.Run(fmt.Sprintf("index-%02d", i), func(t *testing.T) {
			encoded1 := colFlagsToSlice(f1)
			f2 := colFlagsFromSlice(encoded1)
			encoded2 := colFlagsToSlice(f2)

			if f1 != f2 {
				t.Fatalf("initial %#v != roundtrip %#v", f1, f2)
			}
			if !bytes.Equal(encoded1, encoded2) {
				t.Fatalf("initial %#v != roundtrip %#v", encoded1, encoded2)
			}
		})
	}
}
