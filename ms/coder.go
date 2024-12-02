// Copyright 2014 Daniel Theophanes.
// Use of this source code is governed by a zlib-style
// license that can be found in the LICENSE file.

package ms

import (
	"encoding/binary"
	"fmt"
	"io"
	"math"
	"math/big"
	"reflect"
	"strconv"
	"time"

	"github.com/kardianos/rdb"
	"github.com/kardianos/rdb/internal/uconv"
	"github.com/kardianos/rdb/semver"
)

var minDateTime = time.Date(1753, time.January, 1, 0, 0, 0, 0, time.UTC)

var zeroDateTime = time.Date(1900, time.January, 1, 0, 0, 0, 0, time.UTC)

func zeroDateN(loc *time.Location) time.Time {
	return time.Date(1, time.January, 1, 0, 0, 0, 0, loc)
}

func encodeType(w *PacketWriter, ti paramTypeInfo, param *rdb.Param) error {
	// Start TYPE_INFO.
	// Write the type of field this is.
	w.WriteByte(byte(ti.T))

	switch {
	case ti.Bytes:
		var typeLength uint32
		if ti.IsMaxParam(param) {
			typeLength = 0xFFFF
		} else {
			typeLength = uint32(param.Length)
			if ti.NChar {
				// Double the stated length if utf16 sized text.
				typeLength += typeLength
			}
		}

		// Write type length. This is different then the field length.
		switch ti.Len {
		case 1:
			w.WriteByte(uint8(typeLength))
		case 2:
			w.WriteUint16(uint16(typeLength))
		case 4:
			w.WriteUint32(typeLength)
		}
		if ti.IsText {
			// TODO: Handle collation.
			collation := []byte{0x09, 0x04, 0xD0, 0x00, 0x34}
			w.WriteBuffer(collation)
		}
	case ti.T == typeIntN:
		w.WriteByte(ti.W) // TYPE_INFO width.
	case ti.T == typeBitN:
		w.WriteByte(ti.W) // TYPE_INFO width.
	case ti.IsPrSc:
		typeLength, err := decimalLength(param.Precision)
		if err != nil {
			return err
		}
		w.WriteByte(typeLength) // TYPE_INFO width.
		w.WriteByte(byte(param.Precision))
		w.WriteByte(byte(param.Scale))
	case ti.T == typeFloatN:
		w.WriteByte(ti.W) // TYPE_INFO width.
	case ti.T == typeDateTimeN:
		w.WriteByte(ti.W) // TYPE_INFO width.
	case ti.Dt != 0:
		switch ti.Len {
		case 0:
		case 1:
			w.WriteByte(7) // TYPE_INFO scale.
		}
	}
	return nil
}

const (
	textNULL    = 0xFFFFFFFFFFFFFFFF
	textUnknown = 0xFFFFFFFFFFFFFFFE
)

func encodeValue(w *PacketWriter, ti paramTypeInfo, param *rdb.Param, truncValues bool, value interface{}) error {
	var nullValue bool
	if value == rdb.Null || value == nil || param.Null {
		nullValue = true
	}

	switch {
	default:
		return fmt.Errorf("unhandled type for param @%s", param.Name)
	case ti.Bytes:
		// End TYPE_INFO.
		// Start ParamLenData.
		// This uses type info, but lengths refer to the actual field value.

		switch v := value.(type) {
		case *string:
			value = *v
		case *[]byte:
			value = *v
		}
		var typeLength uint32
		var maxLen bool
		if ti.IsMaxParam(param) {
			typeLength = 0xFFFF
			maxLen = true
		} else {
			typeLength = uint32(param.Length)
			if ti.NChar {
				// Double the stated length if utf16 sized text.
				typeLength += typeLength
			}
		}
		if maxLen {
			if nullValue {
				w.WriteUint64(textNULL)
				return nil
			}
			if reader, ok := value.(io.Reader); ok {
				// Size Unknown.
				var bb = make([]byte, 4000)
				w.WriteUint64(textUnknown)
				var err error
				var n int
				for {
					n, err = reader.Read(bb)
					if err != nil {
						if err == io.EOF {
							break
						}
						return err
					}

					writeBb := bb[n:]
					if ti.NChar {
						writeBb = uconv.Encode.FromBytes(bb)
					}
					w.WriteUint32(uint32(len(writeBb)))
					// Use w.Write() to write to buffer and attempt to send.
					// This should prevent the internal buffer from getting too big.
					_, err = w.Write(writeBb)
					if err != nil {
						return fmt.Errorf("failed to write internal buffer: %w", err)
					}
				}
				// Write end of field.
				w.WriteUint32(0)
				return nil
			}
			var writeBb []byte
			switch v := value.(type) {
			case string:
				if ti.NChar {
					writeBb = uconv.Encode.FromString(v)
				} else {
					writeBb = []byte(v)
				}
			case []byte:
				if ti.NChar {
					writeBb = uconv.Encode.FromBytes(v)
				} else {
					writeBb = v
				}
			default:
				rv := reflect.ValueOf(v)
				k := rv.Kind()
				switch k {
				default:
					return fmt.Errorf("max unsupported type: %[1]T=%[1]s, kind=%[2]v", value, k)
				case reflect.Int32:
					s := string(rune(rv.Int()))
					if ti.NChar {
						writeBb = uconv.Encode.FromString(s)
					} else {
						writeBb = []byte(s)
					}
				case reflect.String:
					if ti.NChar {
						writeBb = uconv.Encode.FromString(rv.String())
					} else {
						writeBb = []byte(rv.String())
					}
				case reflect.Slice:
					if ti.NChar {
						writeBb = uconv.Encode.FromBytes(rv.Bytes())
					} else {
						writeBb = rv.Bytes()
					}
				}
			}
			if writeBb == nil {
				w.WriteUint64(textNULL)
				return nil
			}
			if len(writeBb) == 0 {
				w.WriteUint64(0)
				w.WriteUint32(0)
				return nil
			}

			w.WriteUint64(textUnknown)
			w.WriteUint32(uint32(len(writeBb)))
			w.WriteBuffer(writeBb)
			w.WriteUint32(0)
			return nil

		}
		if nullValue {
			w.WriteUint16(0xFFFF) // Field length.
			return nil
		}

		// A non-max value.
		var writeBb []byte
		switch {
		case ti.Bytes:
			switch v := value.(type) {
			case string:
				if ti.NChar {
					writeBb = uconv.Encode.FromString(v)
				} else {
					writeBb = []byte(v)
				}
			case []byte:
				if ti.NChar {
					writeBb = uconv.Encode.FromBytes(v)
				} else {
					writeBb = v
				}
			default:
				rv := reflect.ValueOf(v)
				k := rv.Kind()
				switch k {
				default:
					return fmt.Errorf("len unsupported type: %[1]T=%[1]s, kind=%[2]v", value, k)
				case reflect.Int32:
					s := string(rune(rv.Int()))
					if ti.NChar {
						writeBb = uconv.Encode.FromString(s)
					} else {
						writeBb = []byte(s)
					}
				case reflect.String:
					if ti.NChar {
						writeBb = uconv.Encode.FromString(rv.String())
					} else {
						writeBb = []byte(rv.String())
					}
				case reflect.Array:
					if ti.NChar {
						writeBb = uconv.Encode.FromBytes(rv.Bytes())
					} else {
						writeBb = rv.Bytes()
					}
				}
			}

			fieldLen := uint16(len(writeBb))
			if uint32(fieldLen) > typeLength {
				if !truncValues {
					return InputToolong{DataLen: uint32(fieldLen), TypeLen: typeLength}
				}
				writeBb = writeBb[:int(typeLength)]
				fieldLen = uint16(typeLength)
			}
			w.WriteUint16(fieldLen) // Field length.
			w.WriteBuffer(writeBb)
		default:
			return fmt.Errorf("unsupported type: %[1]T=%[1]s", ti.Name)
		}
		return nil
	case ti.T == typeIntN:
		if nullValue {
			w.WriteByte(0)
			return nil
		}
		w.WriteByte(ti.W) // Row field width.
		switch ti.W {
		case 1:
			switch v := value.(type) {
			case int8:
				w.WriteByte(byte(v))
			case byte:
				w.WriteByte(v)
			case int16:
				w.WriteByte(byte(v))
			case uint16:
				w.WriteByte(byte(v))
			case int32:
				w.WriteByte(byte(v))
			case uint32:
				w.WriteByte(byte(v))
			case int64:
				w.WriteByte(byte(v))
			case uint64:
				w.WriteByte(byte(v))
			case int:
				w.WriteByte(byte(v))
			case uint:
				w.WriteByte(byte(v))
			case float32:
				w.WriteByte(byte(v))
			case float64:
				w.WriteByte(byte(v))
			case string:
				iv, err := strconv.ParseInt(v, 10, 8)
				if err != nil {
					return fmt.Errorf("cannot convert string to int8 for param %q", param.Name)
				}
				w.WriteByte(byte(iv))

			case *int8:
				w.WriteByte(byte(*v))
			case *byte:
				w.WriteByte(*v)
			case *int16:
				w.WriteByte(byte(*v))
			case *uint16:
				w.WriteByte(byte(*v))
			case *int32:
				w.WriteByte(byte(*v))
			case *uint32:
				w.WriteByte(byte(*v))
			case *int64:
				w.WriteByte(byte(*v))
			case *uint64:
				w.WriteByte(byte(*v))
			case *int:
				w.WriteByte(byte(*v))
			case *uint:
				w.WriteByte(byte(*v))
			case *float32:
				w.WriteByte(byte(*v))
			case *float64:
				w.WriteByte(byte(*v))
			default:
				rv := reflect.ValueOf(v)
				switch {
				default:
					return fmt.Errorf("need byte or smaller for param @%s", param.Name)
				case rv.CanInt():
					w.WriteByte(byte(rv.Int()))
				case rv.Kind() == reflect.String:
					iv, err := strconv.ParseInt(rv.String(), 10, 8)
					if err != nil {
						return fmt.Errorf("cannot convert string to int8 for param %q", param.Name)
					}
					w.WriteByte(byte(iv))
				}
			}
		case 2:
			switch v := value.(type) {
			case int8:
				w.WriteUint16(uint16(v))
			case byte:
				w.WriteUint16(uint16(v))
			case int16:
				w.WriteUint16(uint16(v))
			case uint16:
				w.WriteUint16(uint16(v))
			case int32:
				w.WriteUint16(uint16(v))
			case uint32:
				w.WriteUint16(uint16(v))
			case int64:
				w.WriteUint16(uint16(v))
			case uint64:
				w.WriteUint16(uint16(v))
			case int:
				w.WriteUint16(uint16(v))
			case uint:
				w.WriteUint16(uint16(v))
			case float32:
				w.WriteUint16(uint16(v))
			case float64:
				w.WriteUint16(uint16(v))
			case string:
				iv, err := strconv.ParseInt(v, 10, 16)
				if err != nil {
					return fmt.Errorf("cannot convert string to int16 for param %q", param.Name)
				}
				w.WriteUint16(uint16(iv))

			case *int8:
				w.WriteUint16(uint16(*v))
			case *byte:
				w.WriteUint16(uint16(*v))
			case *int16:
				w.WriteUint16(uint16(*v))
			case *uint16:
				w.WriteUint16(uint16(*v))
			case *int32:
				w.WriteUint16(uint16(*v))
			case *uint32:
				w.WriteUint16(uint16(*v))
			case *int64:
				w.WriteUint16(uint16(*v))
			case *uint64:
				w.WriteUint16(uint16(*v))
			case *int:
				w.WriteUint16(uint16(*v))
			case *uint:
				w.WriteUint16(uint16(*v))
			case *float32:
				w.WriteUint16(uint16(*v))
			case *float64:
				w.WriteUint16(uint16(*v))
			default:
				rv := reflect.ValueOf(v)
				switch {
				default:
					return fmt.Errorf("need uint16 or smaller for param @%s", param.Name)
				case rv.CanInt():
					w.WriteUint16(uint16(rv.Int()))
				case rv.Kind() == reflect.String:
					iv, err := strconv.ParseInt(rv.String(), 10, 16)
					if err != nil {
						return fmt.Errorf("cannot convert string to int16 for param %q", param.Name)
					}
					w.WriteUint16(uint16(iv))
				}
			}
		case 4:
			switch v := value.(type) {
			case int8:
				w.WriteUint32(uint32(v))
			case byte:
				w.WriteUint32(uint32(v))
			case int16:
				w.WriteUint32(uint32(v))
			case uint16:
				w.WriteUint32(uint32(v))
			case int32:
				w.WriteUint32(uint32(v))
			case uint32:
				w.WriteUint32(uint32(v))
			case int64:
				w.WriteUint32(uint32(v))
			case uint64:
				w.WriteUint32(uint32(v))
			case int:
				w.WriteUint32(uint32(v))
			case uint:
				w.WriteUint32(uint32(v))
			case float32:
				w.WriteUint32(uint32(v))
			case float64:
				w.WriteUint32(uint32(v))
			case string:
				iv, err := strconv.ParseInt(v, 10, 32)
				if err != nil {
					return fmt.Errorf("cannot convert string to int32 for param %q", param.Name)
				}
				w.WriteUint32(uint32(iv))

			case *int8:
				w.WriteUint32(uint32(*v))
			case *byte:
				w.WriteUint32(uint32(*v))
			case *int16:
				w.WriteUint32(uint32(*v))
			case *uint16:
				w.WriteUint32(uint32(*v))
			case *int32:
				w.WriteUint32(uint32(*v))
			case *uint32:
				w.WriteUint32(uint32(*v))
			case *int64:
				w.WriteUint32(uint32(*v))
			case *uint64:
				w.WriteUint32(uint32(*v))
			case *int:
				w.WriteUint32(uint32(*v))
			case *uint:
				w.WriteUint32(uint32(*v))
			case *float32:
				w.WriteUint32(uint32(*v))
			case *float64:
				w.WriteUint32(uint32(*v))
			default:
				rv := reflect.ValueOf(v)
				switch {
				default:
					return fmt.Errorf("need uint32 or smaller for param @%s", param.Name)
				case rv.CanInt():
					w.WriteUint32(uint32(rv.Int()))
				case rv.Kind() == reflect.String:
					iv, err := strconv.ParseInt(rv.String(), 10, 32)
					if err != nil {
						return fmt.Errorf("cannot convert string to int32 for param %q", param.Name)
					}
					w.WriteUint32(uint32(iv))
				}
			}
		case 8:
			switch v := value.(type) {
			case int8:
				w.WriteUint64(uint64(v))
			case byte:
				w.WriteUint64(uint64(v))
			case int16:
				w.WriteUint64(uint64(v))
			case uint16:
				w.WriteUint64(uint64(v))
			case int32:
				w.WriteUint64(uint64(v))
			case uint32:
				w.WriteUint64(uint64(v))
			case int64:
				w.WriteUint64(uint64(v))
			case uint64:
				w.WriteUint64(uint64(v))
			case int:
				w.WriteUint64(uint64(v))
			case uint:
				w.WriteUint64(uint64(v))
			case float32:
				w.WriteUint64(uint64(v))
			case float64:
				w.WriteUint64(uint64(v))
			case string:
				iv, err := strconv.ParseInt(v, 10, 64)
				if err != nil {
					return fmt.Errorf("cannot convert string to int64 for param %q", param.Name)
				}
				w.WriteUint64(uint64(iv))

			case *int8:
				w.WriteUint64(uint64(*v))
			case *byte:
				w.WriteUint64(uint64(*v))
			case *int16:
				w.WriteUint64(uint64(*v))
			case *uint16:
				w.WriteUint64(uint64(*v))
			case *int32:
				w.WriteUint64(uint64(*v))
			case *uint32:
				w.WriteUint64(uint64(*v))
			case *int64:
				w.WriteUint64(uint64(*v))
			case *uint64:
				w.WriteUint64(uint64(*v))
			case *int:
				w.WriteUint64(uint64(*v))
			case *uint:
				w.WriteUint64(uint64(*v))
			case *float32:
				w.WriteUint64(uint64(*v))
			case *float64:
				w.WriteUint64(uint64(*v))
			default:
				rv := reflect.ValueOf(v)
				switch {
				default:
					return fmt.Errorf("need uint64 or smaller for param @%s", param.Name)
				case rv.CanInt():
					w.WriteUint64(uint64(rv.Int()))
				case rv.Kind() == reflect.String:
					iv, err := strconv.ParseInt(rv.String(), 10, 64)
					if err != nil {
						return fmt.Errorf("cannot convert string to int64 for param %q", param.Name)
					}
					w.WriteUint64(uint64(iv))
				}
			}
		}
		return nil
	case ti.T == typeBitN:
		if nullValue {
			w.WriteByte(0)
			return nil
		}
		w.WriteByte(ti.W) // Row field width.

		switch v := value.(type) {
		case *bool:
			value = *v
		}
		var writeValue byte
		switch v := value.(type) {
		case bool:
			if v {
				writeValue = 1
			}
		default:
			rv := reflect.ValueOf(v)
			if rv.Kind() == reflect.Bool {
				if rv.Bool() {
					writeValue = 1
				}
			} else {
				return fmt.Errorf("need bool for param @%s", param.Name)
			}
		}
		w.WriteByte(writeValue)
		return nil
	case ti.IsPrSc:
		// byte type
		// byte length (5, 9, 13, 17)
		// byte prec
		// byte scale
		if nullValue {
			w.Write([]byte{0})
			return nil
		}
		var pv big.Rat
		var rv *big.Rat
		type rater interface {
			Rat(r *big.Rat) *big.Rat
		}
		switch v := value.(type) {
		default:
			// Support github.com/woodsbury/decimal128
			if r, ok := v.(rater); ok {
				r.Rat(&pv)
			}
			return fmt.Errorf("need *big.Rat for param @%s", param.Name)
		case **big.Rat:
			if v == nil || *v == nil {
				w.Write([]byte{0})
				return nil
			}
			pv = **v
		case *big.Rat:
			if v == nil {
				w.Write([]byte{0})
				return nil
			}
			pv = *v
		}
		rv = &pv

		sign := byte(0)
		if rv.Sign() >= 0 {
			sign = 1
		}

		// Num / Denom == Integer * 10^(-S)
		// Mult = 10^(-S)
		// Num / Denom == Integer * Mult
		// Num * Mult / Denom == Integer
		mult := getMult(param.Scale)
		num := rv.Num()
		denom := rv.Denom()
		num.Mul(num, big.NewInt(mult))
		num.Div(num, denom)
		bb := num.Bytes()
		if len(bb) > 16 {
			return fmt.Errorf("decimal value of (%s) too large for param %s %s", rv.String(), param.Name, ti.TypeString(param))
		}
		// Big.Bytes writes out in big-endian.
		// Want little endian so reverse bytes.
		reverseBytes(bb)
		dataLen := 0
		switch {
		case len(bb) <= 4:
			dataLen = 4
		case len(bb) <= 8:
			dataLen = 8
		case len(bb) <= 12:
			dataLen = 12
		case len(bb) <= 16:
			dataLen = 16
		}
		filler := make([]byte, dataLen-len(bb))
		w.WriteByte(byte(dataLen + 1))
		w.WriteByte(sign)
		w.WriteBuffer(bb)
		if len(filler) > 0 {
			w.WriteBuffer(filler)
		}
		return nil
	case ti.T == typeFloatN:
		if nullValue {
			w.WriteByte(0)
			return nil
		}
		w.WriteByte(ti.W) // Row field width.

		var writeValue float64
		switch v := value.(type) {
		case float32:
			writeValue = float64(v)
		case float64:
			writeValue = float64(v)
		case byte:
			writeValue = float64(v)
		case int8:
			writeValue = float64(v)
		case uint16:
			writeValue = float64(v)
		case int16:
			writeValue = float64(v)
		case uint32:
			writeValue = float64(v)
		case int32:
			writeValue = float64(v)
		case uint64:
			writeValue = float64(v)
		case int64:
			writeValue = float64(v)
		case *big.Rat:
			writeValue, _ = v.Float64()

		case *float32:
			writeValue = float64(*v)
		case *float64:
			writeValue = float64(*v)
		case *byte:
			writeValue = float64(*v)
		case *int8:
			writeValue = float64(*v)
		case *uint16:
			writeValue = float64(*v)
		case *int16:
			writeValue = float64(*v)
		case *uint32:
			writeValue = float64(*v)
		case *int32:
			writeValue = float64(*v)
		case *uint64:
			writeValue = float64(*v)
		case *int64:
			writeValue = float64(*v)
		case **big.Rat:
			writeValue, _ = (*v).Float64()
		default:
			rv := reflect.ValueOf(v)
			if rv.CanFloat() {
				writeValue = rv.Float()
			} else {
				return fmt.Errorf("need numeric for param @%s", param.Name)
			}
		}

		if ti.W == 4 {
			w.WriteUint32(math.Float32bits(float32(writeValue)))
		} else {
			w.WriteUint64(math.Float64bits(writeValue))
		}
		return nil
	case ti.T == typeDateTimeN:
		if nullValue {
			w.WriteByte(0)
			return nil
		}
		w.WriteByte(ti.W) // Row field width.

		switch v := value.(type) {
		case *time.Time:
			value = *v
		}
		switch v := value.(type) {
		case time.Time:
			if v.Before(minDateTime) {
				return fmt.Errorf("time for @%s must be after %s", param.Name, minDateTime.String())
			}
			vNoTime := time.Date(v.Year(), v.Month(), v.Day(), 0, 0, 0, 0, v.Location())
			day := int64(vNoTime.Sub(zeroDateTime).Hours()) / 24
			sec := v.Sub(vNoTime).Seconds()
			w.WriteUint32(uint32(day))
			w.WriteUint32(uint32(sec * 300))
		default:
			return fmt.Errorf("need time.Time for param @%s", param.Name)
		}
		return nil
	case ti.Dt != 0:
		if nullValue {
			w.WriteByte(0)
			return nil
		}
		var v time.Time
		var dur time.Duration

		switch input := value.(type) {
		case *time.Time:
			if input != nil {
				v = *input
			} else {
				nullValue = true
			}
		case time.Time:
			v = input
		case time.Duration:
			dur = input
		default:
			return fmt.Errorf("need time.Time for param @%s", param.Name)
		}
		if nullValue {
			w.WriteByte(0)
			return nil
		}

		w.WriteByte(ti.W) // Row field width.

		// Days from 0001-01-01 in Gregorian.
		gregorianDays := func(year, yearday int) int {
			year0 := year - 1
			return year0*365 + year0/4 - year0/100 + year0/400 + yearday - 1
		}

		dateTime2 := func(t time.Time) (days int, seconds int, ns int) {
			// Days from 0001-01-01 (in same TZ as t).
			days = gregorianDays(t.Year(), t.YearDay())
			seconds = t.Second() + t.Minute()*60 + t.Hour()*60*60
			ns = t.Nanosecond()
			if days < 0 {
				days = 0
				seconds = 0
				ns = 0
			}
			max := gregorianDays(9999, 365)
			if days > max {
				days = max
				seconds = 59 + 59*60 + 23*60*60
				ns = 999999900
			}
			return
		}
		encodeTimeInt := func(w *PacketWriter, seconds, ns, scale int) {
			ns_total := int64(seconds)*1000*1000*1000 + int64(ns)
			t := ns_total / int64(math.Pow10(int(scale)*-1)*1e9)
			w.WriteByte(byte(t))
			w.WriteByte(byte(t >> 8))
			w.WriteByte(byte(t >> 16))
			w.WriteByte(byte(t >> 24))
			w.WriteByte(byte(t >> 32))
		}

		const timeScale = 7
		_, offset := v.Zone()
		if (ti.Dt & dtZone) != 0 {
			v = v.UTC()
		}
		days, seconds, ns := dateTime2(v)

		if dur > 0 {
			const NANO = 1_000_000_000
			nsX := dur.Nanoseconds()
			sX := nsX / NANO
			ns = int(nsX - (sX * NANO))
			seconds = int(sX)
		}

		if (ti.Dt & dtTime) != 0 {
			encodeTimeInt(w, seconds, ns, timeScale)
		}

		if (ti.Dt & dtDate) != 0 {
			w.WriteByte(byte(days))
			w.WriteByte(byte(days >> 8))
			w.WriteByte(byte(days >> 16))
		}

		if (ti.Dt & dtZone) != 0 {
			w.WriteUint16(uint16(offset / 60))
		}
		return nil
	}
}

type paramTypeInfo struct {
	typeWidth
	typeInfo
}

func getParamTypeInfo(tdsVer *semver.Version, paramType rdb.Type) (paramTypeInfo, error) {
	var ti paramTypeInfo
	var found bool

	ti.typeWidth, found = sqlTypeLookup[paramType]
	if !found {
		return ti, fmt.Errorf("sql type not setup: %d", paramType)
	}
	ti.typeInfo, found = typeInfoLookup[driverType(ti.T)]
	if !found {
		return ti, fmt.Errorf("Driver type not found: %d", ti.T)
	}
	if ti.MinVer != nil && tdsVer.Comp(ti.MinVer) < 0 {
		return ti, fmt.Errorf("param type %s does not work with %s", ti.SqlName, tdsVer.String())
	}
	return ti, nil
}

func encodeParam(w *PacketWriter, truncValues bool, tdsVer *semver.Version, param *rdb.Param, value interface{}) error {
	// Write field name.
	if len(param.Name) == 0 {
		w.WriteByte(0) // No name. Length zero.
	} else {
		nameUtf16 := uconv.Encode.FromString("@" + param.Name)
		w.WriteByte(byte(len(nameUtf16) / 2))
		w.WriteBuffer(nameUtf16)
	}
	// Status flag. 0 = normal, 1 = output parameter.
	if param.Out {
		w.WriteByte(1)
	} else {
		w.WriteByte(0)
	}

	ti, err := getParamTypeInfo(tdsVer, param.Type)
	if err != nil {
		return err
	}

	err = encodeType(w, ti, param)
	if err != nil {
		return err
	}
	return encodeValue(w, ti, param, truncValues, value)
}

type colFlags struct {
	Nullable        bool
	Serial          bool
	Key             bool
	SparseColumnSet bool
	NullableUnknown bool
}

func colFlagsFromSlice(flags []byte) colFlags {
	return colFlags{
		Nullable:        flags[0]&(1<<0) != 0,
		Serial:          flags[0]&(1<<4) != 0,
		SparseColumnSet: flags[1]&(1<<2) != 0,
		Key:             flags[1]&(1<<6) != 0,
		NullableUnknown: flags[1]&(1<<7) != 0,
	}
}
func colFlagsToSlice(cf colFlags) []byte {
	var f0, f1 byte
	if cf.Nullable {
		f0 |= (1 << 0)
	}
	if cf.Serial {
		f0 |= (1 << 4)
	}
	if cf.SparseColumnSet {
		f1 |= (1 << 2)
	}
	if cf.Key {
		f1 |= (1 << 6)
	}
	if cf.NullableUnknown {
		f1 |= (1 << 7)
	}
	return []byte{f0, f1}
}

func decodeColumnInfo(read uconv.PanicReader) *SQLColumn {
	userType := binary.LittleEndian.Uint32(read(4)) // userType
	flags := colFlagsFromSlice(read(2))
	driverType := driverType(read(1)[0])

	info, ok := typeInfoLookup[driverType]
	if !ok {
		panic(recoverError{fmt.Errorf("not a known type: 0x%X (UserType: %d, flags: %v)", int(driverType), userType, flags)})
	}

	/*
		byte-0
		0 fNullable
		1 fCaseSen
		2 usUpdateable (2bit)
		4 fIdentity
		5 fComputed
		6 usReservedODBC (2bit)

		byte-1
		0 fFixedLenCLRType
		1	FRESERVEDBIT
		2	fSparseColumnSet
		3	fEncrypted
		4	usReserved3; (introduced in TDS 7.4)
		5 fHidden
		6 fKey
		7 fNullableUnknown
	*/
	if flags.SparseColumnSet {
		panic(recoverError{fmt.Errorf("sparse column set requested, but not supported")})
	}
	column := &SQLColumn{
		Column: rdb.Column{
			Nullable: flags.Nullable,
			Serial:   flags.Serial,
			Key:      flags.Key,
		},
		code: driverType,
		info: info,
	}

	if info.Fixed {
		column.Length = int(info.Len)
	} else {
		switch info.Len {
		case 1:
			column.Length = int(read(1)[0])
		case 2:
			column.Length = int(binary.LittleEndian.Uint16(read(2)))
			if info.Max && column.Length == 0xffff {
				column.Unlimit = true
			}
		case 4:
			column.Length = int(binary.LittleEndian.Uint32(read(4)))
		}
	}
	column.Column.Type = info.Specific
	if column.Column.Type == 0 && info.SpecificMap != nil {
		column.Column.Type = info.SpecificMap[byte(column.Length)]
	}
	column.Column.Generic = info.Generic

	if info.IsText {
		copy(column.Collation[:], read(5))
	}
	if info.IsPrSc {
		column.Precision = int(read(1)[0])
		column.Scale = int(read(1)[0])
	}

	return column
}

type writeField func(c *rdb.Column, value *rdb.DriverValue, assign rdb.Assigner) error

func (tds *Connection) decodeFieldValue(read uconv.PanicReader, column *SQLColumn, resultWf writeField, reportRow bool) {
	sc := &column.Column
	var err error
	defer func() {
		if err != nil {
			panic(recoverError{err: err})
		}
	}()

	var wf = func(val *rdb.DriverValue) {
		if !reportRow {
			return
		}
		err = resultWf(sc, val, nil)
	}

	if column.Unlimit {
		totalSize := binary.LittleEndian.Uint64(read(8))
		sizeUnknown := false

		if totalSize == textNULL {
			wf(&rdb.DriverValue{
				Null: true,
			})
			return
		}
		if totalSize == textUnknown {
			sizeUnknown = true
		}
		useChunks := false
		first := true
		for {
			chunkSize := int(binary.LittleEndian.Uint32(read(4)))
			if chunkSize == 0 {
				if useChunks || totalSize == 0 {
					wf(&rdb.DriverValue{
						More:    false,
						Chunked: useChunks,
						Value:   []byte{},
					})
				}
				break
			}
			if first {
				useChunks = sizeUnknown || totalSize != uint64(chunkSize)
			}

			var value []byte
			if column.info.NChar {
				// TODO: This could probably be cleaner.
				// Data is chunked in a way that ignores UCS-2 runes.
				// Before decoding to UTF-8, make sure a uint16 rune isn't split between two packets.
				// If it is, save it for the next packet and append it.
				split := (chunkSize+len(tds.ucs2Next))%2 == 1
				if split {
					if len(tds.ucs2Next) != 0 {
						bb := read(chunkSize)
						bb2 := append(tds.ucs2Next, bb[:len(bb)-1]...)
						tds.ucs2Next = bb[len(bb)-1:]
						value = uconv.Decode.ToBytes(bb2)
					} else {
						bb := read(chunkSize)
						value = uconv.Decode.ToBytes(bb[:len(bb)-1])
						tds.ucs2Next = bb[len(bb)-1:]
					}
				} else {
					if len(tds.ucs2Next) != 0 {
						bb := read(chunkSize)
						bb2 := append(tds.ucs2Next, bb...)
						value = uconv.Decode.ToBytes(bb2)
						tds.ucs2Next = nil
					} else {
						value = uconv.Decode.ToBytes(read(chunkSize))
					}
				}
			} else {
				value = make([]byte, chunkSize)
				copy(value, read(chunkSize))
			}
			wf(&rdb.DriverValue{
				More:    useChunks,
				Chunked: useChunks,
				Value:   value,
			})
			first = false
		}
		return
	}

	dataLen := 0
	isNull := false

	if column.info.Table {
		panic(recoverError{err: fmt.Errorf("types Text, NText, and Image are not currently supported, long values do not decode correctly")})
		// Types text, ntext, and image.
		/*
			10 > 16 (meta-data length)
			64 75 6d 6d 79 20 74 65 78 74 70 74 72 00 00 00  > dummy textptr (meta-data)
			64 75 6d 6d 79 54 53 00 > dummyTS (label)
			05 00 00 00 > 5 (data length)
			48 65 6c 6c 6f > Hello (data)
		*/
		metaDataLen := read(1)[0]
		if metaDataLen == 0 {
			isNull = true
		} else {
			read(int(metaDataLen)) // metaData
			read(8)                // label; not sure if this should be hard-coded or scanned till null.
		}
	}

	if column.info.Fixed {
		dataLen = int(column.info.Len)
	} else {
		switch column.info.Len {
		case 0:
			fallthrough
		case 1:
			dataLen = int(read(1)[0])
			isNull = dataLen == 0xFF
		case 2:
			dataLen = int(binary.LittleEndian.Uint16(read(2)))
			isNull = dataLen == 0xFFFF
		case 4:
			dataLen = int(binary.LittleEndian.Uint32(read(4)))
			isNull = dataLen == -1 // 0xFFFFFFFF
		}
	}

	if column.info.Bytes || column.code == typeGuid {
		if isNull {
			wf(&rdb.DriverValue{
				Null: true,
			})
			return
		}
		var value []byte
		if column.info.NChar {
			value = uconv.Decode.ToBytes(read(dataLen))
		} else {
			value = make([]byte, dataLen)
			copy(value, read(dataLen))
		}
		if column.code == typeGuid {
			reverse := func(b []byte) {
				for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
					b[i], b[j] = b[j], b[i]
				}
			}
			bytesToGuidString := func(u []byte) string {
				reverse(u[0:4])
				reverse(u[4:6])
				reverse(u[6:8])
				return fmt.Sprintf("%X-%X-%X-%X-%X", u[0:4], u[4:6], u[6:8], u[8:10], u[10:])
			}
			wf(&rdb.DriverValue{
				Value: bytesToGuidString(value),
			})
			return
		}
		wf(&rdb.DriverValue{
			Value: value,
		})
		return
	}

	if dataLen == 0 || column.code == typeNull {
		wf(&rdb.DriverValue{
			Null: true,
		})
		return
	}

	if column.info.Fixed {
		bb := read(int(column.info.Len))

		var v interface{}
		switch column.code {
		case typeBool:
			v = (bb[0] != 0)
		case typeByte:
			v = bb[0]
		case typeInt16:
			v = int16(binary.LittleEndian.Uint16(bb))
		case typeInt32:
			v = int32(binary.LittleEndian.Uint32(bb))
		case typeInt64:
			v = int64(binary.LittleEndian.Uint64(bb))
		case typeFloat32:
			v = math.Float32frombits(binary.LittleEndian.Uint32(bb))
		case typeFloat64:
			v = math.Float64frombits(binary.LittleEndian.Uint64(bb))
		case typeDateTime:
			dt := time.Duration(binary.LittleEndian.Uint32(bb))
			tm := time.Duration(binary.LittleEndian.Uint32(bb[4:]))
			t := time.Date(1900, time.January, 1, 0, 0, 0, 0, time.UTC)
			v = t.Add(time.Hour*24*dt + time.Millisecond*tm*1000/300)
		default:
			panic(recoverError{fmt.Errorf("unhandled fixed type: %v", column.code)})
		}
		wf(&rdb.DriverValue{
			Value: v,
		})
		return
	}

	if column.code == typeIntN {
		switch dataLen {
		case 1:
			wf(&rdb.DriverValue{
				Value: int8(read(1)[0]),
			})
		case 2:
			wf(&rdb.DriverValue{
				Value: int16(binary.LittleEndian.Uint16(read(2))),
			})
		case 4:
			wf(&rdb.DriverValue{
				Value: int32(binary.LittleEndian.Uint32(read(4))),
			})
		case 8:
			wf(&rdb.DriverValue{
				Value: int64(binary.LittleEndian.Uint64(read(8))),
			})
		default:
			panic(fmt.Errorf("proto error IntN, unknown data len %d", dataLen))
		}
		return
	}

	if column.code == typeBitN {
		switch dataLen {
		case 1:
			v := read(1)[0]
			writeValue := false
			if v != 0 {
				writeValue = true
			}
			wf(&rdb.DriverValue{
				Value: writeValue,
			})
		default:
			panic(fmt.Errorf("proto error BitN, unknown data len %d", dataLen))
		}
		return
	}

	if column.info.IsPrSc {
		all := read(dataLen)
		signByte := all[0]
		intBytes := all[1:]
		reverseBytes(intBytes)

		integer := &big.Int{}
		integer.SetBytes(intBytes)

		mult := getMult(int(column.Scale))

		v := &big.Rat{}
		v.SetInt(integer)
		if signByte == 0 {
			v.Neg(v)
		}
		v.Quo(v, (&big.Rat{}).SetInt64(mult))

		wf(&rdb.DriverValue{
			Value: v,
		})
		return
	}

	if column.code == typeFloatN {
		switch dataLen {
		case 4:
			wf(&rdb.DriverValue{
				Value: math.Float32frombits(binary.LittleEndian.Uint32(read(4))),
			})
			return
		case 8:
			wf(&rdb.DriverValue{
				Value: math.Float64frombits(binary.LittleEndian.Uint64(read(8))),
			})
			return
		default:
			panic(fmt.Errorf("proto error FloatN, unknown data len %d", dataLen))
		}

	}
	if column.code == typeDateTimeN {
		switch dataLen {
		case 8:
			dt := time.Duration(int64(binary.LittleEndian.Uint32(read(4)))*24) * time.Hour
			tmf := float64(binary.LittleEndian.Uint32(read(4)))
			// tmf counts 300 per second, from midnight.
			tm := time.Duration(int64(tmf / 300 * 1000000000))

			v := zeroDateTime.Add(dt).Add(tm)
			wf(&rdb.DriverValue{
				Value: v,
			})
			return
		default:
			panic(fmt.Errorf("proto error DateTimeN, unknown data len %d", dataLen))
		}
	}

	if column.info.Dt != 0 {
		// Time Width:
		// 3 bytes if 0 <= n < = 2.
		// 4 bytes if 3 <= n < = 4.
		// 5 bytes if 5 <= n < = 7.

		var tm time.Duration
		var dt time.Time
		var offset int16
		if (column.info.Dt & dtTime) != 0 {
			bbLen := 5
			if dataLen < 5 {
				bbLen = 4
			}
			if dataLen < 3 {
				bbLen = 3
			}
			bb := read(bbLen)
			full := make([]byte, 8)
			copy(full, bb)
			value := int64(binary.LittleEndian.Uint64(full))
			scale := getMult(int(column.Length))

			tm = time.Duration(1000000000 / scale * value)
			if column.info.Dt == dtTime {
				wf(&rdb.DriverValue{
					Value: tm,
				})
				return
			}
		}
		if (column.info.Dt & dtDate) != 0 {
			bb := read(3)
			full := make([]byte, 4)
			copy(full, bb)
			days := int32(binary.LittleEndian.Uint32(full))
			// time.Duration can't hold more then 290 years at a time.
			// Add days in increments.

			dt = zeroDateN(time.UTC)
			dayChunkCount := int32(250 * 365)
			dayChunk := time.Duration(dayChunkCount*24) * time.Hour
			for days > dayChunkCount {
				dt = dt.Add(dayChunk)
				days -= dayChunkCount
			}
			dt = dt.Add(time.Duration(days*24) * time.Hour)
		}
		if (column.info.Dt & dtZone) != 0 {
			offset = int16(binary.LittleEndian.Uint16(read(2)))
		}
		dt = dt.Add(tm)

		loc := time.UTC
		if offset != 0 {
			hrs := offset / 60
			mins := offset % 60
			loc = time.FixedZone(fmt.Sprintf("UTC %d:%02d", hrs, mins), int(offset)*60)
		}
		dt = time.Date(dt.Year(), dt.Month(), dt.Day(), dt.Hour(), dt.Minute(), dt.Second(), dt.Nanosecond(), time.UTC).In(loc)
		wf(&rdb.DriverValue{
			Value: dt,
		})
		return
	}

	panic(fmt.Errorf("unsupported data type: %s", column.info.Name))
}
