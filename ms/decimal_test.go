// Copyright 2014 Daniel Theophanes.
// Use of this source code is governed by a zlib-style
// license that can be found in the LICENSE file.

package ms

import (
	"context"
	"fmt"
	"math/big"
	"testing"

	"github.com/kardianos/rdb"
)

// TestPow10 covers all legal TDS decimal scales. Regression for getMult's int64
// overflow at scale >= 19 (10^19 does not fit in int64).
func TestPow10(t *testing.T) {
	for scale := 0; scale <= 38; scale++ {
		got := pow10(scale)
		want := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(scale)), nil)
		if got.Cmp(want) != 0 {
			t.Errorf("pow10(%d) = %s, want %s", scale, got, want)
		}
	}
}

// TestDecimalWireHighScale reproduces the production decode bug:
// decimal(38,20) value 1234.567891 came back as ~15896.516 when the driver
// divided by an overflowed int64 multiplier (getMult(20) == 7766279631452241920).
func TestDecimalWireHighScale(t *testing.T) {
	const scale = 20
	in, ok := new(big.Rat).SetString("1234.567891")
	if !ok {
		t.Fatal("SetString failed")
	}

	// Show what the old getMult path produced (documents the bug).
	scaled := new(big.Int).Set(in.Num())
	scaled.Mul(scaled, pow10(scale))
	scaled.Div(scaled, in.Denom())
	oldMult := getMult(scale) // intentionally wrong for scale 20
	garbled := new(big.Rat).SetInt(scaled)
	garbled.Quo(garbled, new(big.Rat).SetInt64(oldMult))
	if g, w := garbled.FloatString(3), "15896.516"; g != w {
		// If getMult is ever fixed to panic/clamp, this self-check can be removed.
		t.Logf("historical overflow decode: got %s (expected %s with int64 wrap)", g, w)
	} else {
		t.Logf("confirmed historical garble via getMult(%d)=%d → %s", scale, oldMult, g)
	}

	payload, err := encodeDecimalWire(in, scale)
	if err != nil {
		t.Fatalf("encode: %v", err)
	}
	out := decodeDecimalWire(payload, scale)

	if out.Cmp(in) != 0 {
		t.Fatalf("round-trip: got %s, want %s", out.FloatString(scale), in.FloatString(scale))
	}
	// Explicit check against the reported value.
	if out.FloatString(6) != "1234.567891" {
		t.Fatalf("got %s, want 1234.567891", out.FloatString(6))
	}
}

func TestDecimalWireRoundTrip(t *testing.T) {
	cases := []struct {
		name  string
		value string
		scale int
	}{
		{"s0", "1234567890", 0},
		{"s6_money", "1234.567891", 6},
		{"s15", "1.234567890123456", 15},
		{"s18_max_int64_mult", "1.234567890123456789", 18},
		{"s19_overflow_boundary", "1234.567891", 19},
		{"s20_repro", "1234.567891", 20},
		{"s28", "1234567890.123456789012345678", 28},
		{"s38", "0.23456789012345678901234567890123456789", 38},
		{"negative_s20", "-1234.567891", 20},
		{"zero_s20", "0", 20},
		{"small_frac_s20", "0.00000000000000000001", 20},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			in, ok := new(big.Rat).SetString(tc.value)
			if !ok {
				t.Fatalf("SetString(%q) failed", tc.value)
			}
			payload, err := encodeDecimalWire(in, tc.scale)
			if err != nil {
				t.Fatalf("encode: %v", err)
			}
			out := decodeDecimalWire(payload, tc.scale)
			// Compare at the declared scale (truncation toward zero on encode).
			if out.FloatString(tc.scale) != in.FloatString(tc.scale) {
				t.Fatalf("got %s, want %s (payload=%x)", out.FloatString(tc.scale), in.FloatString(tc.scale), payload)
			}
		})
	}
}

// TestDecimalHighScaleSQL exercises decimal(38,20) against a live server when available.
func TestDecimalHighScaleSQL(t *testing.T) {
	checkSkip(t)
	defer recoverTest(t)

	cases := []struct {
		name      string
		precision int
		scale     int
		value     string
	}{
		{"p38_s6_money", 38, 6, "1234.567891"},
		{"p38_s15", 38, 15, "1234.567891012345"},
		{"p38_s20_repro", 38, 20, "1234.567891"},
		{"p38_s28", 38, 28, "1234.567891"},
		// decimal(38,38) has no integer digits; value must be |v| < 1.
		{"p38_s38", 38, 38, "0.23456789012345678901234567890123456789"},
		{"p38_s20_neg", 38, 20, "-999.00000000000000000001"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			defer recoverTest(t)

			in, ok := new(big.Rat).SetString(tc.value)
			if !ok {
				t.Fatalf("SetString(%q) failed", tc.value)
			}

			cmd := &rdb.Command{
				SQL: fmt.Sprintf(`
					DECLARE @v decimal(%d,%d) = @input;
					SELECT
						v = @v,
						s = CONVERT(varchar(100), @v),
						literal = CAST('%s' AS decimal(%d,%d));
				`, tc.precision, tc.scale, tc.value, tc.precision, tc.scale),
				Arity: rdb.OneMust,
			}

			res := db.Query(context.Background(), cmd,
				rdb.Param{Name: "input", Type: rdb.TypeDecimal, Precision: tc.precision, Scale: tc.scale, Value: in},
			)
			defer res.Close()

			res.Scan()
			gotParam := res.Get("v")
			gotLiteral := res.Get("literal")
			gotStr := res.Get("s")

			want := in.FloatString(tc.scale)
			for _, name := range []string{"v", "literal"} {
				var got interface{}
				if name == "v" {
					got = gotParam
				} else {
					got = gotLiteral
				}
				r, ok := got.(*big.Rat)
				if !ok {
					t.Errorf("%s: unexpected type %T", name, got)
					continue
				}
				if r.FloatString(tc.scale) != want {
					t.Errorf("%s: got %s, want %s (sql string=%v)", name, r.FloatString(tc.scale), want, gotStr)
				}
			}
		})
	}
}
