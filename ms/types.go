// Copyright 2014 Daniel Theophanes.
// Use of this source code is governed by a zlib-style
// license that can be found in the LICENSE file.

package ms

import (
	"bitbucket.org/kardianos/rdb"
	"bitbucket.org/kardianos/rdb/semver"
	"fmt"
)

type driverType byte

/*
NULLTYPE = %x1F ; Null
INT1TYPE = %x30 ; TinyInt
BITTYPE = %x32 ; Bit
INT2TYPE = %x34 ; SmallInt
INT4TYPE = %x38 ; Int
DATETIM4TYPE = %x3A ; SmallDateTime
FLT4TYPE = %x3B ; Real
MONEYTYPE = %x3C ; Money
DATETIMETYPE = %x3D ; DateTime
FLT8TYPE = %x3E ; Float
MONEY4TYPE = %x7A ; SmallMoney
INT8TYPE = %x7F ; BigInt
*/
const (
	// Fixed length data types.
	typeNull          driverType = 0x1F
	typeByte          driverType = 0x30
	typeBool          driverType = 0x32
	typeInt16         driverType = 0x34
	typeInt32         driverType = 0x38
	typeDateTimeSmall driverType = 0x3A
	typeFloat32       driverType = 0x3B
	typeMoney         driverType = 0x3C
	typeDateTime      driverType = 0x3D
	typeFloat64       driverType = 0x3E
	typeMoneySmall    driverType = 0x7A
	typeInt64         driverType = 0x7F
)

/*
GUIDTYPE = %x24 ; UniqueIdentifier
INTNTYPE = %x26 ; (see below)
DECIMALTYPE = %x37 ; Decimal (legacy support)
NUMERICTYPE = %x3F ; Numeric (legacy support)
BITNTYPE = %x68 ; (see below)
DECIMALNTYPE = %x6A ; Decimal
NUMERICNTYPE = %x6C ; Numeric

FLTNTYPE = %x6D ; (see below)
MONEYNTYPE = %x6E ; (see below)
DATETIMNTYPE = %x6F ; (see below)
DATENTYPE = %x28 ; (introduced in TDS 7.3)
TIMENTYPE = %x29 ; (introduced in TDS 7.3)
DATETIME2NTYPE = %x2A ; (introduced in TDS 7.3)
DATETIMEOFFSETNTYPE = %x2B ; (introduced in TDS 7.3)
CHARTYPE = %x2F ; Char (legacy support)
VARCHARTYPE = %x27 ; VarChar (legacy support)
BINARYTYPE = %x2D ; Binary (legacy support)
VARBINARYTYPE = %x25 ; VarBinary (legacy support)

BIGVARBINTYPE = %xA5 ; VarBinary
BIGVARCHRTYPE = %xA7 ; VarChar
BIGBINARYTYPE = %xAD ; Binary
BIGCHARTYPE = %xAF ; Char
NVARCHARTYPE = %xE7 ; NVarChar
NCHARTYPE = %xEF ; NChar
XMLTYPE = %xF1 ; XML (introduced in TDS 7.2)
UDTTYPE = %xF0 ; CLR-UDT (introduced in TDS 7.2)
TEXTTYPE = %x23 ; Text
IMAGETYPE = %x22 ; Image
NTEXTTYPE = %x63 ; NText
SSVARIANTTYPE = %x62 ; Sql_Variant (introduced in TDS 7.2)
*/
const (
	typeGuid       driverType = 0x24
	typeIntN       driverType = 0x26
	typeDecimalOld driverType = 0x37
	typeNumericOld driverType = 0x3F
	typeBitN       driverType = 0x68
	typeDecimal    driverType = 0x6A
	typeNumeric    driverType = 0x6C

	typeFloatN          driverType = 0x6D
	typeMoneyN          driverType = 0x6E
	typeDateTimeN       driverType = 0x6F
	typeDateN           driverType = 0x28
	typeTimeN           driverType = 0x29
	typeDateTime2N      driverType = 0x2A
	typeDateTimeOffsetN driverType = 0x2B
	typeCharOld         driverType = 0x2F
	typeVarCharOld      driverType = 0x27
	typeBinaryOld       driverType = 0x2D
	typeVarBinaryOld    driverType = 0x25

	typeVarBinary driverType = 0xA5 // Big
	typeVarChar   driverType = 0xA7 // Big
	typeBinary    driverType = 0xAD // Big
	typeChar      driverType = 0xAF // Big
	typeNVarChar  driverType = 0xE7
	typeNChar     driverType = 0xDF
	typeXml       driverType = 0xF1
	typeUDT       driverType = 0xF0
	typeText      driverType = 0x23
	typeImage     driverType = 0x22
	typeNText     driverType = 0x63
	typeVariant   driverType = 0x62
)

const (
	dtTime = 1
	dtDate = 2
	dtZone = 4
)

type typeInfo struct {
	Name   string // Friendly name.
	Fixed  bool   // Fixed lengthed type, no need to read field size.
	Max    bool   // Can this type do "type(max)"?
	IsText bool   // Is the content text (should it have a collation)?
	Len    byte   // Length of type of length of length field, in bytes.
	NChar  bool   // Does the server expect utf16 encoded text?
	Bytes  bool   // Can the type be treated as a stream of bytes?
	IsPrSc bool   // Use scale and prec?
	Dt     byte   // Date Time Flags.

	MinVer *semver.Version // Minimum protocol version for type.
}

// Don't bother with anything before 72 (Server 2005).
var (
	protoVer72  = &semver.Version{InHex: true, Major: 0x2, Minor: 0x72, Patch: 0x0}
	protoVer73A = &semver.Version{InHex: true, Major: 0x3, Minor: 0x73, Patch: 0xA}
	protoVer73B = &semver.Version{InHex: true, Major: 0x3, Minor: 0x73, Patch: 0xB}
	protoVer74  = &semver.Version{InHex: true, Major: 0x4, Minor: 0x74, Patch: 0x0}
)

var typeInfoLookup = map[driverType]typeInfo{
	typeNull: typeInfo{Name: "NULL", Fixed: true, Len: 0},

	// Some or all of these may be obsolete, may now use the *N types.
	typeByte:          typeInfo{Name: "Byte", Fixed: true, Len: 1},
	typeBool:          typeInfo{Name: "Bool", Fixed: true, Len: 1},
	typeInt16:         typeInfo{Name: "Int16", Fixed: true, Len: 2},
	typeInt32:         typeInfo{Name: "Int32", Fixed: true, Len: 4},
	typeDateTimeSmall: typeInfo{Name: "DateTimeSmall", Fixed: true, Len: 4},
	typeFloat32:       typeInfo{Name: "Float32", Fixed: true, Len: 4},
	typeMoney:         typeInfo{Name: "Money", Fixed: true, Len: 8},
	typeDateTime:      typeInfo{Name: "DateTime", Fixed: true, Len: 4},
	typeFloat64:       typeInfo{Name: "Float64", Fixed: true, Len: 8},
	typeMoneySmall:    typeInfo{Name: "MoneySmall", Fixed: true, Len: 4},
	typeInt64:         typeInfo{Name: "Int64", Fixed: true, Len: 8},

	typeGuid:    typeInfo{Name: "GUID", Len: 1},
	typeIntN:    typeInfo{Name: "IntN", Len: 1},
	typeBitN:    typeInfo{Name: "BitN", Len: 1},
	typeDecimal: typeInfo{Name: "Decimal", IsPrSc: true, Len: 1},
	typeNumeric: typeInfo{Name: "Numeric", IsPrSc: true, Len: 1},

	typeFloatN:          typeInfo{Name: "FloatN", Len: 1},
	typeMoneyN:          typeInfo{Name: "MoneyN", Len: 1},
	typeDateTimeN:       typeInfo{Name: "DateTimeN", Len: 1, MinVer: protoVer72},
	typeDateN:           typeInfo{Name: "DateN", Len: 0, Dt: dtDate, MinVer: protoVer73A},
	typeTimeN:           typeInfo{Name: "TimeN", Len: 1, Dt: dtTime, MinVer: protoVer73A},
	typeDateTime2N:      typeInfo{Name: "DateTime2N", Len: 1, Dt: dtDate | dtTime, MinVer: protoVer73A},
	typeDateTimeOffsetN: typeInfo{Name: "DateTimeOffsetN", Len: 1, Dt: dtDate | dtTime | dtZone, MinVer: protoVer73A},

	// Probably don't worry about these.
	typeDecimalOld:   typeInfo{Name: "DecimalOld", Len: 1},
	typeNumericOld:   typeInfo{Name: "NumericOld", Len: 1},
	typeCharOld:      typeInfo{Name: "CharOld", Len: 1},
	typeVarCharOld:   typeInfo{Name: "VarCharOld", Len: 1},
	typeBinaryOld:    typeInfo{Name: "BinaryOld", Len: 1},
	typeVarBinaryOld: typeInfo{Name: "VarBinaryOld", Len: 1},

	typeVarBinary: typeInfo{Name: "VarBinary(Big)", Bytes: true, Max: true, Len: 2},
	typeVarChar:   typeInfo{Name: "VarChar(Big)", Bytes: true, IsText: true, Max: true, Len: 2},
	typeBinary:    typeInfo{Name: "Binary(Big)", Bytes: true, Max: true, Len: 2},
	typeChar:      typeInfo{Name: "Char(Big)", Bytes: true, IsText: true, Len: 2},
	typeNVarChar:  typeInfo{Name: "NVarChar", Bytes: true, NChar: true, IsText: true, Max: true, Len: 2},
	typeNChar:     typeInfo{Name: "NChar", Bytes: true, NChar: true, IsText: true, Len: 2},
	typeText:      typeInfo{Name: "Text", Bytes: true, IsText: true, Len: 4},
	typeImage:     typeInfo{Name: "Image", Bytes: true, Len: 4},
	typeNText:     typeInfo{Name: "NText", Bytes: true, NChar: true, IsText: true, Len: 4},

	// The following will be unsupported for a time.
	typeXml:     typeInfo{Name: "Xml", Max: true, Len: 4},
	typeUDT:     typeInfo{Name: "UDT", Max: true, Len: 0},
	typeVariant: typeInfo{Name: "Varient", Len: 4},
}

func (value driverType) String() string {
	if info, in := typeInfoLookup[value]; in {
		return info.Name
	}
	return fmt.Sprintf("<UNKNOWN TYPE: 0x%X>", value)
}

type typeWidth struct {
	T       driverType // Driver type.
	W       byte       // Width of length field for mapped typed. Need?
	SqlName string
}

func (t *typeWidth) IsMaxParam(param *rdb.Param) bool {
	info := typeInfoLookup[t.T]
	return info.Max && (param.L <= 0 || ((info.NChar && param.L > 4000) || param.L > 8000))
}

func (t *typeWidth) TypeString(param *rdb.Param) string {
	info := typeInfoLookup[t.T]
	switch {
	case info.IsPrSc:
		return fmt.Sprintf("%s(%d,%d)", t.SqlName, param.Precision, param.Scale)
	case info.Bytes:
		if t.IsMaxParam(param) {
			return fmt.Sprintf("%s(max)", t.SqlName)
		}
		return fmt.Sprintf("%s(%d)", t.SqlName, param.L)
	default:
		return t.SqlName
	}
}

func decimalLength(param *rdb.Param) (byte, error) {
	p := param.Precision
	switch {
	case 1 <= p && p <= 9:
		return 4 + 1, nil
	case 10 <= p && p <= 19:
		return 8 + 1, nil
	case 20 <= p && p <= 28:
		return 12 + 1, nil
	case 29 <= p && p <= 38:
		return 16 + 1, nil
	default:
		return 0, fmt.Errorf("Requested precision(%d) not in range [1,38]", p)
	}
}

func reverseBytes(bb []byte) {
	var a, b int
	b = len(bb) - 1
	for a < b {
		bb[a], bb[b] = bb[b], bb[a]
		a++
		b--
	}
}

func getMult(scale int) int64 {
	mult := int64(1)
	for _ = range make([]struct{}, scale) {
		mult *= 10
	}
	return mult
}

const (
	_                          = iota
	TypeOldBool    rdb.SqlType = rdb.TypeDriverThresh + iota
	TypeOldByte    rdb.SqlType = rdb.TypeDriverThresh + iota
	TypeOldInt16   rdb.SqlType = rdb.TypeDriverThresh + iota
	TypeOldInt32   rdb.SqlType = rdb.TypeDriverThresh + iota
	TypeOldInt64   rdb.SqlType = rdb.TypeDriverThresh + iota
	TypeOldFloat32 rdb.SqlType = rdb.TypeDriverThresh + iota
	TypeOldFloat64 rdb.SqlType = rdb.TypeDriverThresh + iota
	TypeOldTD      rdb.SqlType = rdb.TypeDriverThresh + iota // DateTime
	TypeNumeric    rdb.SqlType = rdb.TypeDriverThresh + iota
)

var sqlTypeLookup = map[rdb.SqlType]*typeWidth{
	rdb.TypeNull: &typeWidth{T: typeNull},

	TypeOldBool:    &typeWidth{T: typeBool, SqlName: "bit"},
	TypeOldByte:    &typeWidth{T: typeByte, SqlName: "tinyint"},
	TypeOldInt16:   &typeWidth{T: typeInt16, SqlName: "smallint"},
	TypeOldInt32:   &typeWidth{T: typeInt32, SqlName: "int"},
	TypeOldInt64:   &typeWidth{T: typeInt64, SqlName: "bigint"},
	TypeOldFloat32: &typeWidth{T: typeFloat32, SqlName: "real"},
	TypeOldFloat64: &typeWidth{T: typeFloat64, SqlName: "float"},

	rdb.TypeString:  &typeWidth{T: typeNVarChar, SqlName: "nvarchar"},
	rdb.TypeVarChar: &typeWidth{T: typeNVarChar, SqlName: "nvarchar"},
	rdb.TypeChar:    &typeWidth{T: typeNChar, SqlName: "nchar"},
	rdb.TypeText:    &typeWidth{T: typeNText, SqlName: "ntext"},

	rdb.TypeAnsiString:  &typeWidth{T: typeVarChar, SqlName: "varchar"},
	rdb.TypeAnsiVarChar: &typeWidth{T: typeVarChar, SqlName: "varchar"},
	rdb.TypeAnsiChar:    &typeWidth{T: typeChar, SqlName: "char"},
	rdb.TypeAnsiText:    &typeWidth{T: typeText, SqlName: "text"},

	rdb.TypeBinary: &typeWidth{T: typeVarBinary, SqlName: "varbinary"},

	rdb.TypeBool:  &typeWidth{T: typeBitN, W: 1, SqlName: "bit"},
	rdb.TypeInt8:  &typeWidth{T: typeIntN, W: 1, SqlName: "tinyint"},
	rdb.TypeInt16: &typeWidth{T: typeIntN, W: 2, SqlName: "smallint"},
	rdb.TypeInt32: &typeWidth{T: typeIntN, W: 4, SqlName: "int"},
	rdb.TypeInt64: &typeWidth{T: typeIntN, W: 8, SqlName: "bigint"},

	rdb.TypeDecimal: &typeWidth{T: typeDecimal, SqlName: "decimal"},
	TypeNumeric:     &typeWidth{T: typeNumeric, SqlName: "decimal"},

	rdb.TypeFloat32: &typeWidth{T: typeFloatN, W: 4, SqlName: "float"},
	rdb.TypeFloat64: &typeWidth{T: typeFloatN, W: 8, SqlName: "float"},

	TypeOldTD: &typeWidth{T: typeDateTimeN, W: 8, SqlName: "datetime"},

	rdb.TypeTime: &typeWidth{T: typeTimeN, W: 5, SqlName: "time"},
	rdb.TypeDate: &typeWidth{T: typeDateN, W: 3, SqlName: "date"},
	rdb.TypeTD:   &typeWidth{T: typeDateTime2N, W: 8, SqlName: "datetime2"},
	rdb.TypeTDZ:  &typeWidth{T: typeDateTimeOffsetN, W: 10, SqlName: "datetimeoffset"},
}

type driverTypeWidth struct {
	T driverType
	W uint8
}

var driverTypeLookup = map[driverType]map[uint8]rdb.SqlType{}

func init() {
	for st, tw := range sqlTypeLookup {
		wLookup, found := driverTypeLookup[tw.T]
		if !found {
			wLookup = make(map[uint8]rdb.SqlType)
			driverTypeLookup[tw.T] = wLookup
		}
		wLookup[tw.W] = st
	}
}

func lookupSqlType(dt driverType, w uint8) (rdb.SqlType, error) {
	wLookup, found := driverTypeLookup[dt]
	if !found {
		return rdb.TypeUnknown, fmt.Errorf("Unknown driver type: 0x%X", uint8(dt))
	}
	st, found := wLookup[w]
	if !found {
		return wLookup[0], nil
	}
	return st, nil
}
