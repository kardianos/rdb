package ms

import "testing"

func TestCollationRoundtrip(t *testing.T) {
	tests := []struct {
		name    string
		raw     [5]byte
		wantUTF8 bool
		wantLCID uint32
	}{
		{
			name:     "SQL_Latin1_General_CP1_CI_AS",
			raw:      [5]byte{0x09, 0x04, 0xD0, 0x00, 0x34},
			wantUTF8: false,
			wantLCID: 0x0409,
		},
		{
			name:     "Latin1_General_100_CI_AS_SC_UTF8",
			raw:      DefaultUTF8Collation().Encode(),
			wantUTF8: true,
			wantLCID: 0x0409,
		},
		{
			name:     "zero_collation",
			raw:      [5]byte{0, 0, 0, 0, 0},
			wantUTF8: false,
			wantLCID: 0,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			c := ParseCollation(tc.raw)
			if c.IsUTF8() != tc.wantUTF8 {
				t.Errorf("IsUTF8() = %v, want %v", c.IsUTF8(), tc.wantUTF8)
			}
			if c.LCID != tc.wantLCID {
				t.Errorf("LCID = 0x%X, want 0x%X", c.LCID, tc.wantLCID)
			}
			// Roundtrip: Encode should produce the same bytes.
			got := c.Encode()
			if got != tc.raw {
				t.Errorf("Encode() = %X, want %X", got, tc.raw)
			}
		})
	}
}

func TestCollationUTF8Bit(t *testing.T) {
	// Construct a collation with the UTF-8 bit set manually.
	c := Collation{
		LCID:    1033,
		Flags:   collFlagUTF8,
		Version: 0,
		SortID:  0,
	}
	if !c.IsUTF8() {
		t.Error("expected IsUTF8() == true")
	}

	// Clear the bit.
	c.Flags &^= collFlagUTF8
	if c.IsUTF8() {
		t.Error("expected IsUTF8() == false after clearing bit")
	}
}
