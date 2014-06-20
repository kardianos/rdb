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
	"time"

	"bitbucket.org/kardianos/rdb"
	"bitbucket.org/kardianos/rdb/ms/uconv"
	"bitbucket.org/kardianos/rdb/semver"
)

var zeroDateTime = time.Date(1900, time.January, 1, 0, 0, 0, 0, time.UTC)
var minDateTime = time.Date(1753, time.January, 1, 0, 0, 0, 0, time.UTC)

var zeroDateN = time.Date(1, time.January, 1, 0, 0, 0, 0, time.UTC)

func encodeParam(w *PacketWriter, truncValues bool, tdsVer *semver.Version, param *rdb.Param, value interface{}) error {
	// TODO: Handle collation.
	// TODO: Check input value length.
	collation := []byte{0x09, 0x04, 0xD0, 0x00, 0x34}

	nullValue := false
	if typeValue, isType := value.(rdb.SqlType); isType && typeValue == rdb.TypeNull {
		nullValue = true
	}

	// Write field name.
	if len(param.N) == 0 {
		w.WriteByte(0) // No name. Length zero.
	} else {
		nameUtf16 := uconv.Encode.FromString("@" + param.N)
		w.WriteByte(byte(len(nameUtf16) / 2))
		w.WriteBuffer(nameUtf16)
	}
	w.WriteByte(0) // Status flag.

	st, found := sqlTypeLookup[param.T]
	if !found {
		return fmt.Errorf("Sql type not setup: %d", param.T)
	}

	info, found := typeInfoLookup[driverType(st.T)]
	if !found {
		return fmt.Errorf("Driver type not found: %d", st.T)
	}

	if info.MinVer != nil && tdsVer.Comp(info.MinVer) < 0 {
		return fmt.Errorf("Param type %s does not work with %s.", st.SqlName, tdsVer.String())
	}

	typeLength := uint32(0)
	writeMaxValue := false

	// Start TYPE_INFO.
	// Write the type of field this is.
	w.WriteByte(byte(st.T))

	if info.Bytes {
		if st.IsMaxParam(param) {
			typeLength = 0xFFFF
			writeMaxValue = true
		} else {
			typeLength = uint32(param.L)
			if info.NChar {
				// Double the stated length if utf16 sized text.
				typeLength += typeLength
			}
		}

		// Write type length. This is different then the field length.
		switch info.Len {
		case 1:
			w.WriteByte(uint8(typeLength))
		case 2:
			w.WriteUint16(uint16(typeLength))
		case 4:
			w.WriteUint32(typeLength)
		}
		if info.IsText {
			w.WriteBuffer(collation)
		}
		// End TYPE_INFO.
		// Start ParamLenData.
		// This uses type info, but lengths refer to the actual field value.

		if writeMaxValue {
			if reader, ok := value.(io.Reader); ok {
				// Size Unknown.
				var bb = make([]byte, 4000)
				w.WriteUint64(0xFFFFFFFFFFFFFFFE)
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
					if info.NChar {
						writeBb = uconv.Encode.FromBytes(bb)
					}
					w.WriteUint32(uint32(len(writeBb)))
					// Use w.Write() to write to buffer and attempt to send.
					// This should prevent the internal buffer from getting too big.
					_, err = w.Write(writeBb)
					if err != nil {
						return err
					}
				}
				// Write end of field.
				w.WriteUint32(0)
				return nil
			}
			var writeBb []byte
			switch v := value.(type) {
			case rdb.SqlType:
				if v != rdb.TypeNull {
					return fmt.Errorf("Unsupported SqlType as a value: %v", v)
				}
			case string:
				if info.NChar {
					writeBb = uconv.Encode.FromString(v)
				} else {
					writeBb = []byte(v)
				}
			case []byte:
				if info.NChar {
					writeBb = uconv.Encode.FromBytes(v)
				} else {
					writeBb = v
				}
			default:
				return fmt.Errorf("Unsupported type: %T", value)
			}
			if writeBb == nil {
				w.WriteUint64(0xFFFFFFFFFFFFFFFF)
				return nil
			}
			if len(writeBb) == 0 {
				w.WriteUint64(0)
				w.WriteUint32(0)
				return nil
			}

			w.WriteUint64(uint64(len(writeBb)))
			w.WriteUint32(uint32(len(writeBb)))
			w.WriteBuffer(writeBb)
			w.WriteUint32(0)
			return nil

		}

		// A non-max value.
		var writeBb []byte
		switch {
		case info.Bytes:
			switch v := value.(type) {
			case string:
				if info.NChar {
					writeBb = uconv.Encode.FromString(v)
				} else {
					writeBb = []byte(v)
				}
			case []byte:
				if info.NChar {
					writeBb = uconv.Encode.FromBytes(v)
				} else {
					writeBb = v
				}
				writeBb = v
			default:
				return fmt.Errorf("Unsupported type: %T", value)
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
			return fmt.Errorf("Unsupported type: %s", info.Name)
		}
		return nil
	}

	if st.T == typeIntN {
		w.WriteByte(st.W) // TYPE_INFO width.

		if nullValue {
			w.WriteByte(0)
			return nil
		}
		w.WriteByte(st.W) // Row field width.
		switch st.W {
		case 1:
			switch v := value.(type) {
			case int8:
				w.WriteByte(byte(v))
			case byte:
				w.WriteByte(v)
			default:
				return fmt.Errorf("Need byte or smaller for param @%s", param.N)
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
			default:
				return fmt.Errorf("Need uint16 or smaller for param @%s", param.N)
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
			default:
				return fmt.Errorf("Need uint32 or smaller for param @%s", param.N)
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
			default:
				return fmt.Errorf("Need uint64 or smaller for param @%s", param.N)
			}
		}
		return nil
	}

	if st.T == typeBitN {
		w.WriteByte(st.W) // TYPE_INFO width.

		if nullValue {
			w.WriteByte(0)
			return nil
		}
		w.WriteByte(st.W) // Row field width.
		switch v := value.(type) {
		case bool:
			writeValue := byte(0)
			if v {
				writeValue = 1
			}
			w.WriteByte(writeValue)
		default:
			return fmt.Errorf("Need bool for param @%s", param.N)
		}
		return nil
	}

	if info.IsPrSc {
		// byte type
		// byte length (5, 9, 13, 17)
		// byte prec
		// byte scale
		typeLength, err := decimalLength(param)
		if err != nil {
			return err
		}
		if nullValue {
			w.WriteByte(0)
			return nil
		}
		w.WriteByte(typeLength) // TYPE_INFO width.
		w.WriteByte(byte(param.Precision))
		w.WriteByte(byte(param.Scale))
		switch v := value.(type) {
		case *big.Rat:
			sign := byte(0)
			if v.Sign() >= 0 {
				sign = 1
			}

			mult := getMult(param.Scale)

			scale := big.NewRat(mult, 1)
			bb := scale.Mul(v, scale).Num().Bytes()
			if len(bb) > 16 {
				return fmt.Errorf("Decimal value of (%s) too large for param %s %s", v.String(), param.N, st.TypeString(param))
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
		default:
			return fmt.Errorf("Need bool for param @%s", param.N)
		}
		return nil
	}

	if st.T == typeFloatN {
		w.WriteByte(st.W) // TYPE_INFO width.
		if nullValue {
			w.WriteByte(0)
			return nil
		}
		w.WriteByte(st.W) // Row field width.

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
		default:
			return fmt.Errorf("Need numeric for param @%s", param.N)
		}

		if st.W == 4 {
			w.WriteUint32(math.Float32bits(float32(writeValue)))
		} else {
			w.WriteUint64(math.Float64bits(writeValue))
		}
		return nil
	}
	if st.T == typeDateTimeN {
		w.WriteByte(st.W) // TYPE_INFO width.
		if nullValue {
			w.WriteByte(0)
			return nil
		}
		w.WriteByte(st.W) // Row field width.

		switch v := value.(type) {
		case time.Time:
			if v.Before(minDateTime) {
				return fmt.Errorf("Time for @%s must be after %s", param.N, minDateTime.String())
			}
			vNoTime := v.Truncate(24 * time.Hour)
			w.WriteUint32(uint32(vNoTime.Sub(zeroDateTime).Hours() / 24))
			w.WriteUint32(uint32(v.Sub(vNoTime).Seconds() * 300))
		default:
			return fmt.Errorf("Need time.Time for param @%s", param.N)
		}
		return nil
	}

	if info.Dt != 0 {
		switch info.Len {
		case 0:
		case 1:
			w.WriteByte(7) // TYPE_INFO scale.
		}
		if nullValue {
			w.WriteByte(0)
			return nil
		}
		w.WriteByte(st.W) // Row field width.

		var v time.Time
		var dur time.Duration
		switch input := value.(type) {
		case time.Time:
			v = input
		case time.Duration:
			dur = input
		default:
			return fmt.Errorf("Need time.Time for param @%s", param.N)
		}

		if (info.Dt & dtTime) != 0 {
			var nano time.Duration
			if dur == 0 {
				v = v.UTC()
				nano += time.Duration(v.Hour()) * time.Hour
				nano += time.Duration(v.Minute()) * time.Minute
				nano += time.Duration(v.Second()) * time.Second
				nano += time.Duration(v.Nanosecond()) * time.Nanosecond
			} else {
				nano = dur
			}
			encoded := nano / 100

			bb := make([]byte, 8)
			binary.LittleEndian.PutUint64(bb, uint64(encoded))
			w.WriteBuffer(bb[:5])
		}
		if (info.Dt & dtDate) != 0 {
			vDate := v.Truncate(time.Hour * 24)
			dt := zeroDateN
			dtNext := dt

			dayChunkCount := uint32(250 * 365)
			dayChunk := time.Duration(dayChunkCount*24) * time.Hour
			var days uint32

			for {
				dtNext = dt.Add(dayChunk)
				if vDate.Before(dtNext) {
					break
				}
				dt = dtNext
				days += dayChunkCount
			}
			days += uint32(vDate.Sub(dt).Hours() / 24)

			bb := make([]byte, 4)
			binary.LittleEndian.PutUint32(bb, days)
			w.WriteBuffer(bb[:3])
		}
		if (info.Dt & dtZone) != 0 {
			_, sec := v.Zone()
			w.WriteUint16(uint16(sec / 60))
		}
		return nil
	}

	return fmt.Errorf("Unhandled type for param @%s", param.N)
}

func decodeColumnInfo(read uconv.PanicReader) *SqlColumn {
	_ = binary.LittleEndian.Uint32(read(4)) // userType

	flags := read(2)

	driverType := driverType(read(1)[0])

	info, ok := typeInfoLookup[driverType]
	if !ok {
		panic(fmt.Sprintf("Not a known type: %d", driverType))
	}

	column := &SqlColumn{
		SqlColumn: rdb.SqlColumn{
			Nullable: flags[0]&(1<<0) != 0,
			Serial:   flags[0]&(1<<4) != 0,
			Key:      flags[1]&(1<<4) != 0,
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
	var err error
	column.SqlColumn.SqlType, err = lookupSqlType(driverType, byte(column.Length))
	if err != nil {
		panic(err)
	}

	if info.IsText {
		copy(column.Collation[:], read(5))
	}
	if info.IsPrSc {
		column.Precision = int(read(1)[0])
		column.Scale = int(read(1)[0])
	}

	_, column.Name = uconv.Decode.Prefix1(read)
	return column
}

func decodeFieldValue(read uconv.PanicReader, column *SqlColumn, result rdb.DriverValuer, reportRow bool) {
	sc := &column.SqlColumn
	var err error
	defer func() {
		if err != nil {
			panic(recoverError{err: err})
		}
	}()

	var wf = func(val *rdb.DriverValue) {
		err = result.WriteField(sc, reportRow, val, nil)
	}

	if column.Unlimit {
		totalSize := binary.LittleEndian.Uint64(read(8))
		sizeUnknown := false

		if totalSize == 0xFFFFFFFFFFFFFFFF {
			wf(&rdb.DriverValue{
				Null: true,
			})
			return
		}
		if totalSize == 0xFFFFFFFFFFFFFFFE {
			sizeUnknown = true
		}
		useChunks := false
		first := true
		for {
			chunkSize := int(binary.LittleEndian.Uint32(read(4)))
			if chunkSize == 0 {
				if useChunks {
					wf(&rdb.DriverValue{
						More:    false,
						Chunked: true,
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
				value = uconv.Decode.ToBytes(read(chunkSize))
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
	nullValue := 0
	if column.info.Fixed {
		dataLen = int(column.info.Len)
	} else {
		switch column.info.Len {
		case 0:
			fallthrough
		case 1:
			dataLen = int(read(1)[0])
			nullValue = 0xFF
		case 2:
			dataLen = int(binary.LittleEndian.Uint16(read(2)))
			nullValue = 0xFFFF
		case 4:
			dataLen = int(binary.LittleEndian.Uint32(read(4)))
			nullValue = 0xFFFFFFFF
		}
	}

	if column.info.Bytes {
		if dataLen == nullValue {
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
		default:
			panic("unhandled fixed type")
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
				Value: read(1)[0],
			})
		case 2:
			wf(&rdb.DriverValue{
				Value: binary.LittleEndian.Uint16(read(2)),
			})
		case 4:
			wf(&rdb.DriverValue{
				Value: binary.LittleEndian.Uint32(read(4)),
			})
		case 8:
			wf(&rdb.DriverValue{
				Value: binary.LittleEndian.Uint64(read(8)),
			})
		default:
			panic("Proto Error IntN")
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
			panic("Proto Error BitN")
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
			panic("Proto Error decode data typeFloatN")
		}

	}
	if column.code == typeDateTimeN {
		switch dataLen {
		case 8:
			dt := time.Duration(int64(binary.LittleEndian.Uint32(read(4)))*24) * time.Hour
			tmf := float64(binary.LittleEndian.Uint32(read(4)))
			// tmf counts 300 per second, from midnight.
			tm := time.Duration(int64(tmf / 300 * 1000000000))

			v := zeroDateTime.Add(dt).Add(tm).Local()
			wf(&rdb.DriverValue{
				Value: v,
			})
			return
		default:
			panic("Proto Error decode data typeDateTimeN")
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

			dt = zeroDateN
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
		// TODO: set offset.
		_ = offset
		loc := time.UTC
		if offset != 0 {
			hrs := offset / 60
			mins := offset - (hrs * 60)
			loc = time.FixedZone(fmt.Sprintf("UTC %d:%02d", hrs, mins), int(offset)*60)
		}
		dt = time.Date(dt.Year(), dt.Month(), dt.Day(), dt.Hour(), dt.Minute(), dt.Second(), dt.Nanosecond(), loc)
		wf(&rdb.DriverValue{
			Value: dt,
		})
		return
	}

	panic(fmt.Errorf("Unsupported data type: %s", column.info.Name))
}
