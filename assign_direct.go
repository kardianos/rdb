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

// DirectAssignInt writes v into prep without an intermediate interface box.
// Returns handled=false if prep is not a known integer destination (caller should fall back).
func DirectAssignInt(prep interface{}, v int64, null bool, defaultNull interface{}) (handled bool, err error) {
	if null {
		if defaultNull != nil {
			return true, AssignValue(nil, Nullable{Value: defaultNull}, prep, nil)
		}
		return true, assignNullPrep(prep)
	}
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
	if null {
		if defaultNull != nil {
			return true, AssignValue(nil, Nullable{Value: defaultNull}, prep, nil)
		}
		return true, assignNullPrep(prep)
	}
	switch p := prep.(type) {
	case *bool:
		*p = v
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
	if null {
		if defaultNull != nil {
			return true, AssignValue(nil, Nullable{Value: defaultNull}, prep, nil)
		}
		return true, assignNullPrep(prep)
	}
	switch p := prep.(type) {
	case *float32:
		*p = float32(v)
	case *float64:
		*p = v
	case *Nullable:
		p.Null = false
		p.Value = v
	default:
		return false, nil
	}
	return true, nil
}

// DirectAssignBytes writes bb into prep. If mustCopy is true and prep retains
// the bytes (*[]byte, *Nullable), a copy is made. For *string, string(bb) copies.
// For io.Writer, bb is written as-is (caller must not mutate until Write returns).
func DirectAssignBytes(prep interface{}, bb []byte, null bool, mustCopy bool, defaultNull interface{}) (handled bool, err error) {
	if null {
		if defaultNull != nil {
			return true, AssignValue(nil, Nullable{Value: defaultNull}, prep, nil)
		}
		return true, assignNullPrep(prep)
	}
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
	if null {
		if defaultNull != nil {
			return true, AssignValue(nil, Nullable{Value: defaultNull}, prep, nil)
		}
		return true, assignNullPrep(prep)
	}
	switch p := prep.(type) {
	case *string:
		*p = s
	case *[]byte:
		*p = []byte(s)
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
	if null {
		if defaultNull != nil {
			return true, AssignValue(nil, Nullable{Value: defaultNull}, prep, nil)
		}
		return true, assignNullPrep(prep)
	}
	switch p := prep.(type) {
	case *time.Time:
		*p = t
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
	if null {
		if defaultNull != nil {
			return true, AssignValue(nil, Nullable{Value: defaultNull}, prep, nil)
		}
		return true, assignNullPrep(prep)
	}
	switch p := prep.(type) {
	case **big.Rat:
		*p = r
	case *big.Rat:
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
	if null {
		if defaultNull != nil {
			return true, AssignValue(nil, Nullable{Value: defaultNull}, prep, nil)
		}
		return true, assignNullPrep(prep)
	}
	switch p := prep.(type) {
	case *time.Duration:
		*p = d
	case *Nullable:
		p.Null = false
		p.Value = d
	default:
		return false, nil
	}
	return true, nil
}

func assignNullPrep(prep interface{}) error {
	if nullable, is := prep.(*Nullable); is {
		*nullable = Nullable{Null: true}
		return nil
	}
	return ErrScanNull
}
