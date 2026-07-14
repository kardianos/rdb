// Copyright 2014 Daniel Theophanes.
// Use of this source code is governed by a zlib-style
// license that can be found in the LICENSE file.

package rdb

import (
	"io"
	"math/big"
	"time"
)

// DriverValuerPrep is an optional extension of DriverValuer that exposes
// per-column Prep destinations so drivers can assign decoded values without
// boxing through DriverValue.Value.
type DriverValuerPrep interface {
	// PrepAt returns the Prep destination for column index, or nil.
	PrepAt(index int) interface{}
	// HasConverter reports whether a ColumnConverter is registered for index.
	HasConverter(index int) bool
	// FieldNull returns the configured default for SQL NULL on that field, if any.
	FieldNull(index int) interface{}
}

// unwrapFlag returns the payload prep and optional null flag from prep.
// If prep is *NullFlagPrep, flag is non-nil and payload is Value.
func unwrapFlag(prep interface{}) (payload interface{}, flag *bool) {
	if w, ok := prep.(*NullFlagPrep); ok {
		return w.Value, w.Null
	}
	return prep, nil
}

func setFlag(flag *bool, null bool) {
	if flag != nil {
		*flag = null
	}
}

// DirectAssignInt writes v into prep without an intermediate interface box.
// Returns handled=false if prep is not a known integer destination (caller should fall back).
func DirectAssignInt(prep interface{}, v int64, null bool, defaultNull interface{}) (handled bool, err error) {
	prep, flag := unwrapFlag(prep)
	if null {
		return true, assignNullWithFlag(prep, flag, defaultNull)
	}
	setFlag(flag, false)
	switch p := prep.(type) {
	case *int8:
		*p = int8(v)
	case *uint8:
		*p = uint8(v)
	case *int16:
		*p = int16(v)
	case *uint16:
		*p = uint16(v)
	case *int32:
		*p = int32(v)
	case *uint32:
		*p = uint32(v)
	case *int64:
		*p = v
	case *uint64:
		*p = uint64(v)
	case *int:
		*p = int(v)
	case *uint:
		*p = uint(v)
	case *Opt[int8]:
		p.Set(int8(v))
	case *Opt[uint8]:
		p.Set(uint8(v))
	case *Opt[int16]:
		p.Set(int16(v))
	case *Opt[uint16]:
		p.Set(uint16(v))
	case *Opt[int32]:
		p.Set(int32(v))
	case *Opt[uint32]:
		p.Set(uint32(v))
	case *Opt[int64]:
		p.Set(v)
	case *Opt[uint64]:
		p.Set(uint64(v))
	case *Opt[int]:
		p.Set(int(v))
	case *Opt[uint]:
		p.Set(uint(v))
	case *Nullable:
		p.Null = false
		p.Value = v
	default:
		return false, nil
	}
	return true, nil
}

// DirectAssignBool writes v into prep without boxing.
func DirectAssignBool(prep interface{}, v bool, null bool, defaultNull interface{}) (handled bool, err error) {
	prep, flag := unwrapFlag(prep)
	if null {
		return true, assignNullWithFlag(prep, flag, defaultNull)
	}
	setFlag(flag, false)
	switch p := prep.(type) {
	case *bool:
		*p = v
	case *Opt[bool]:
		p.Set(v)
	case *Nullable:
		p.Null = false
		p.Value = v
	default:
		return false, nil
	}
	return true, nil
}

// DirectAssignFloat writes v into prep without boxing.
func DirectAssignFloat(prep interface{}, v float64, null bool, defaultNull interface{}) (handled bool, err error) {
	prep, flag := unwrapFlag(prep)
	if null {
		return true, assignNullWithFlag(prep, flag, defaultNull)
	}
	setFlag(flag, false)
	switch p := prep.(type) {
	case *float32:
		*p = float32(v)
	case *float64:
		*p = v
	case *Opt[float32]:
		p.Set(float32(v))
	case *Opt[float64]:
		p.Set(v)
	case *Nullable:
		p.Null = false
		p.Value = v
	default:
		return false, nil
	}
	return true, nil
}

// DirectAssignBytes writes bb into prep. If mustCopy is true and prep retains
// the bytes (*[]byte, *Nullable, *Opt[[]byte]), a copy is made. For *string /
// *Opt[string], string(bb) copies. For io.Writer, bb is written as-is (caller
// must not mutate until Write returns).
func DirectAssignBytes(prep interface{}, bb []byte, null bool, mustCopy bool, defaultNull interface{}) (handled bool, err error) {
	prep, flag := unwrapFlag(prep)
	if null {
		return true, assignNullWithFlag(prep, flag, defaultNull)
	}
	setFlag(flag, false)
	switch p := prep.(type) {
	case *string:
		*p = string(bb)
	case *[]byte:
		if mustCopy {
			cp := make([]byte, len(bb))
			copy(cp, bb)
			*p = cp
		} else {
			*p = bb
		}
	case *Opt[string]:
		p.Set(string(bb))
	case *Opt[[]byte]:
		if mustCopy {
			cp := make([]byte, len(bb))
			copy(cp, bb)
			p.Set(cp)
		} else {
			p.Set(bb)
		}
	case io.Writer:
		_, err = p.Write(bb)
	case *Nullable:
		p.Null = false
		if mustCopy {
			cp := make([]byte, len(bb))
			copy(cp, bb)
			p.Value = cp
		} else {
			p.Value = bb
		}
	default:
		return false, nil
	}
	return true, err
}

// DirectAssignString writes s into prep.
func DirectAssignString(prep interface{}, s string, null bool, defaultNull interface{}) (handled bool, err error) {
	prep, flag := unwrapFlag(prep)
	if null {
		return true, assignNullWithFlag(prep, flag, defaultNull)
	}
	setFlag(flag, false)
	switch p := prep.(type) {
	case *string:
		*p = s
	case *[]byte:
		*p = []byte(s)
	case *Opt[string]:
		p.Set(s)
	case *Opt[[]byte]:
		p.Set([]byte(s))
	case io.Writer:
		_, err = p.Write([]byte(s))
	case *Nullable:
		p.Null = false
		p.Value = s
	default:
		return false, nil
	}
	return true, err
}

// DirectAssignTime writes t into prep.
func DirectAssignTime(prep interface{}, t time.Time, null bool, defaultNull interface{}) (handled bool, err error) {
	prep, flag := unwrapFlag(prep)
	if null {
		return true, assignNullWithFlag(prep, flag, defaultNull)
	}
	setFlag(flag, false)
	switch p := prep.(type) {
	case *time.Time:
		*p = t
	case *Opt[time.Time]:
		p.Set(t)
	case *Nullable:
		p.Null = false
		p.Value = t
	default:
		return false, nil
	}
	return true, nil
}

// DirectAssignRat writes r into prep.
func DirectAssignRat(prep interface{}, r *big.Rat, null bool, defaultNull interface{}) (handled bool, err error) {
	prep, flag := unwrapFlag(prep)
	if null {
		return true, assignNullWithFlag(prep, flag, defaultNull)
	}
	setFlag(flag, false)
	switch p := prep.(type) {
	case **big.Rat:
		*p = r
	case *big.Rat:
		p.Set(r)
	case *Opt[*big.Rat]:
		p.Set(r)
	case *Nullable:
		p.Null = false
		p.Value = r
	default:
		return false, nil
	}
	return true, nil
}

// DirectAssignDuration writes d into prep (for TIME-only values).
func DirectAssignDuration(prep interface{}, d time.Duration, null bool, defaultNull interface{}) (handled bool, err error) {
	prep, flag := unwrapFlag(prep)
	if null {
		return true, assignNullWithFlag(prep, flag, defaultNull)
	}
	setFlag(flag, false)
	switch p := prep.(type) {
	case *time.Duration:
		*p = d
	case *Opt[time.Duration]:
		p.Set(d)
	case *Nullable:
		p.Null = false
		p.Value = d
	default:
		return false, nil
	}
	return true, nil
}

// assignNullWithFlag marks flag (if any) and nulls the payload prep.
// When a null flag is present, SQL NULL is fully represented by the flag;
// payload zeroing is best-effort and never returns ErrScanNull.
func assignNullWithFlag(prep interface{}, flag *bool, defaultNull interface{}) error {
	setFlag(flag, true)
	if defaultNull != nil {
		return AssignValue(nil, Nullable{Value: defaultNull}, prep, nil)
	}
	if flag != nil {
		if err := assignNullPrep(prep); err != nil {
			_ = zeroPayload(prep)
		}
		return nil
	}
	return assignNullPrep(prep)
}

func assignNullPrep(prep interface{}) error {
	if prep == nil {
		return ErrScanNull
	}
	if w, ok := prep.(*NullFlagPrep); ok {
		setFlag(w.Null, true)
		_ = zeroPayload(w.Value)
		_ = assignNullPrep(w.Value)
		return nil
	}
	if _, ok := prep.(io.Writer); ok {
		return nil
	}
	switch p := prep.(type) {
	case *Nullable:
		*p = Nullable{Null: true}
		return nil
	case *Opt[int8]:
		p.SetNull()
		return nil
	case *Opt[uint8]:
		p.SetNull()
		return nil
	case *Opt[int16]:
		p.SetNull()
		return nil
	case *Opt[uint16]:
		p.SetNull()
		return nil
	case *Opt[int32]:
		p.SetNull()
		return nil
	case *Opt[uint32]:
		p.SetNull()
		return nil
	case *Opt[int64]:
		p.SetNull()
		return nil
	case *Opt[uint64]:
		p.SetNull()
		return nil
	case *Opt[int]:
		p.SetNull()
		return nil
	case *Opt[uint]:
		p.SetNull()
		return nil
	case *Opt[bool]:
		p.SetNull()
		return nil
	case *Opt[float32]:
		p.SetNull()
		return nil
	case *Opt[float64]:
		p.SetNull()
		return nil
	case *Opt[string]:
		p.SetNull()
		return nil
	case *Opt[[]byte]:
		p.SetNull()
		return nil
	case *Opt[time.Time]:
		p.SetNull()
		return nil
	case *Opt[time.Duration]:
		p.SetNull()
		return nil
	case *Opt[*big.Rat]:
		p.SetNull()
		return nil
	}
	return ErrScanNull
}

func zeroPayload(prep interface{}) error {
	if prep == nil {
		return nil
	}
	switch p := prep.(type) {
	case *int8:
		*p = 0
	case *uint8:
		*p = 0
	case *int16:
		*p = 0
	case *uint16:
		*p = 0
	case *int32:
		*p = 0
	case *uint32:
		*p = 0
	case *int64:
		*p = 0
	case *uint64:
		*p = 0
	case *int:
		*p = 0
	case *uint:
		*p = 0
	case *bool:
		*p = false
	case *float32:
		*p = 0
	case *float64:
		*p = 0
	case *string:
		*p = ""
	case *[]byte:
		*p = nil
	case *time.Time:
		*p = time.Time{}
	case *time.Duration:
		*p = 0
	default:
		// Leave as-is for unknown payload types; flag already marks null.
	}
	return nil
}
