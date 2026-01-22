// Copyright 2014 Daniel Theophanes.
// Use of this source code is governed by a zlib-style
// license that can be found in the LICENSE file.

package ms

import (
	"fmt"

	"github.com/kardianos/rdb"
	"github.com/kardianos/rdb/semver"
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
	// typeNCharOld was 0xEF, now consolidated into typeNChar
	typeVarCharOld      driverType = 0x27
	typeBinaryOld       driverType = 0x2D
	typeVarBinaryOld    driverType = 0x25

	typeVarBinary driverType = 0xA5 // Big
	typeVarChar   driverType = 0xA7 // Big
	typeBinary    driverType = 0xAD // Big
	typeChar      driverType = 0xAF // Big
	typeNVarChar driverType = 0xE7
	typeNChar    driverType = 0xEF // NCHARTYPE per TDS spec (was incorrectly 0xDF)
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
	Name      string // Friendly name.
	Fixed     bool   // Fixed lengthed type, no need to read field size.
	Max       bool   // Can this type do "type(max)"?
	IsText    bool   // Is the content text (should it have a collation)?
	NChar     bool   // Does the server expect utf16 encoded text?
	Bytes     bool   // Can the type be treated as a stream of bytes?
	IsPrSc    bool   // Use scale and prec?
	Table     bool   // Send table name after type info.
	Collation bool   // 5 byte CollationPrefix.
	Len       byte   // Length of type of length of length field, in bytes.
	Dt        byte   // Date Time Flags.

	MinVer *semver.Version // Minimum protocol version for type.

	Specific    rdb.Type
	Generic     rdb.Type
	SpecificMap map[byte]rdb.Type
}

// Don't bother with anything before 72 (Server 2005).
var (
	protoVer72  = &semver.Version{InHex: true, Major: 0x2, Minor: 0x72, Patch: 0x0}
	protoVer73A = &semver.Version{InHex: true, Major: 0x3, Minor: 0x73, Patch: 0xA}
	protoVer73B = &semver.Version{InHex: true, Major: 0x3, Minor: 0x73, Patch: 0xB}
	protoVer74  = &semver.Version{InHex: true, Major: 0x4, Minor: 0x74, Patch: 0x0}
)

var typeInfoLookup = map[driverType]typeInfo{
	typeNull: {Name: "NULL", Fixed: true, Len: 0},

	// Some or all of these may be obsolete, may now use the *N types.
	typeByte:          {Name: "Byte", Fixed: true, Len: 1, Specific: rdb.TypeInt8, Generic: rdb.Integer},
	typeBool:          {Name: "Bool", Fixed: true, Len: 1, Specific: rdb.TypeBool, Generic: rdb.Bool},
	typeInt16:         {Name: "Int16", Fixed: true, Len: 2, Specific: rdb.TypeInt16, Generic: rdb.Integer},
	typeInt32:         {Name: "Int32", Fixed: true, Len: 4, Specific: rdb.TypeInt32, Generic: rdb.Integer},
	typeDateTimeSmall: {Name: "DateTimeSmall", Fixed: true, Len: 4, Specific: rdb.TypeTimestamp, Generic: rdb.Integer},
	typeFloat32:       {Name: "Float32", Fixed: true, Len: 4, Specific: rdb.TypeFloat32, Generic: rdb.Float},
	typeMoney:         {Name: "Money", Fixed: true, Len: 8, Specific: rdb.TypeDecimal, Generic: rdb.Decimal},
	typeDateTime:      {Name: "DateTime", Fixed: true, Len: 8, Specific: rdb.TypeTimestamp, Generic: rdb.Time},
	typeFloat64:       {Name: "Float64", Fixed: true, Len: 8, Specific: rdb.TypeFloat64, Generic: rdb.Float},
	typeMoneySmall:    {Name: "MoneySmall", Fixed: true, Len: 4, Specific: rdb.TypeDecimal, Generic: rdb.Decimal},
	typeInt64:         {Name: "Int64", Fixed: true, Len: 8, Specific: rdb.TypeInt64, Generic: rdb.Integer},

	typeGuid: {Name: "GUID", Len: 1, Specific: rdb.TypeUUID, Generic: rdb.Other},
	typeIntN: {Name: "IntN", Len: 1, SpecificMap: map[byte]rdb.Type{
		1: rdb.TypeInt8,
		2: rdb.TypeInt16,
		4: rdb.TypeInt32,
		8: rdb.TypeInt64,
	}, Generic: rdb.Integer},
	typeBitN:    {Name: "BitN", Len: 1, Specific: rdb.TypeBool, Generic: rdb.Bool},
	typeDecimal: {Name: "Decimal", IsPrSc: true, Len: 1, Specific: rdb.TypeDecimal, Generic: rdb.Decimal},
	typeNumeric: {Name: "Numeric", IsPrSc: true, Len: 1, Specific: rdb.TypeDecimal, Generic: rdb.Decimal},

	typeFloatN: {Name: "FloatN", Len: 1, SpecificMap: map[byte]rdb.Type{
		4: rdb.TypeFloat32,
		8: rdb.TypeFloat64,
	}, Generic: rdb.Float},
	typeMoneyN:          {Name: "MoneyN", Len: 1, Specific: rdb.TypeDecimal, Generic: rdb.Decimal},
	typeDateTimeN:       {Name: "DateTimeN", Len: 1, Specific: rdb.TypeTimestamp, MinVer: protoVer72, Generic: rdb.Time},
	typeDateN:           {Name: "DateN", Len: 0, Specific: rdb.TypeDate, Dt: dtDate, MinVer: protoVer73A, Generic: rdb.Time},
	typeTimeN:           {Name: "TimeN", Len: 1, Specific: rdb.TypeTime, Dt: dtTime, MinVer: protoVer73A, Generic: rdb.Time},
	typeDateTime2N:      {Name: "DateTime2N", Len: 1, Specific: rdb.TypeTimestamp, Dt: dtDate | dtTime, MinVer: protoVer73A, Generic: rdb.Time},
	typeDateTimeOffsetN: {Name: "DateTimeOffsetN", Len: 1, Specific: rdb.TypeTimestampz, Dt: dtDate | dtTime | dtZone, MinVer: protoVer73A, Generic: rdb.Time},

	// Probably don't worry about these.
	typeDecimalOld:   {Name: "DecimalOld", Len: 1, Specific: rdb.TypeDecimal, Generic: rdb.Decimal},
	typeNumericOld:   {Name: "NumericOld", Len: 1, Specific: rdb.TypeDecimal, Generic: rdb.Decimal},
	typeCharOld:      {Name: "CharOld", Len: 1, Specific: rdb.TypeAnsiChar, Generic: rdb.Text},
	typeVarCharOld:   {Name: "VarCharOld", Len: 1, Specific: rdb.TypeAnsiVarChar, Generic: rdb.Text},
	typeBinaryOld:    {Name: "BinaryOld", Len: 1, Specific: rdb.TypeBinary, Generic: rdb.Binary},
	typeVarBinaryOld: {Name: "VarBinaryOld", Len: 1, Specific: rdb.TypeBinary, Generic: rdb.Binary},
	// typeNCharOld:     {Name: "NCharOld", Len: 2, Specific: rdb.TypeChar, Generic: rdb.Text},

	typeVarBinary: {Name: "VarBinary(Big)", Bytes: true, Max: true, Len: 2, Specific: rdb.TypeBinary, Generic: rdb.Binary},
	typeVarChar:   {Name: "VarChar(Big)", Bytes: true, IsText: true, Max: true, Len: 2, Specific: rdb.TypeAnsiVarChar, Generic: rdb.Text},
	typeBinary:    {Name: "Binary(Big)", Bytes: true, Max: true, Len: 2, Specific: rdb.TypeBinary, Generic: rdb.Binary},
	typeChar:      {Name: "Char(Big)", Bytes: true, IsText: true, Len: 2, Specific: rdb.TypeAnsiChar, Generic: rdb.Text},
	typeNVarChar: {Name: "NVarChar", Bytes: true, NChar: true, IsText: true, Max: true, Len: 2, Specific: rdb.TypeVarChar, Generic: rdb.Text},
	typeNChar:    {Name: "NChar", Bytes: true, NChar: true, IsText: true, Len: 2, Specific: rdb.TypeChar, Generic: rdb.Text}, // 0xEF per TDS spec
	typeText:      {Name: "Text", Bytes: true, IsText: true, Table: true, Len: 4, Specific: rdb.TypeAnsiText, Generic: rdb.Text},
	typeImage:     {Name: "Image", Bytes: true, Table: true, Len: 4, Specific: rdb.TypeBinary, Generic: rdb.Binary},
	typeNText:     {Name: "NText", Bytes: true, NChar: true, IsText: true, Table: true, Len: 4, Specific: rdb.TypeText, Generic: rdb.Text},

	// The following will be unsupported for a time.
	typeXml:     {Name: "Xml", Max: true, Len: 0, Specific: rdb.TypeXML, Generic: rdb.Other},
	typeUDT:     {Name: "UDT", Max: true, Len: 0, Generic: rdb.Other},
	typeVariant: {Name: "Varient", Len: 4, Generic: rdb.Other},
}

func (value driverType) String() string {
	if info, in := typeInfoLookup[value]; in {
		return info.Name
	}
	return fmt.Sprintf("<UNKNOWN TYPE: 0x%X>", byte(value))
}

type typeWidth struct {
	T       driverType // Driver type.
	W       byte       // Width of length field for mapped typed. Need?
	SqlName string
}

func (t *typeWidth) IsMaxParam(param *rdb.Param) bool {
	info := typeInfoLookup[t.T]
	return info.Max && (param.Length <= 0 || ((info.NChar && param.Length > 4000) || param.Length > 8000))
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
		return fmt.Sprintf("%s(%d)", t.SqlName, param.Length)
	default:
		return t.SqlName
	}
}

func decimalLength(precision int) (byte, error) {
	p := precision
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
	_                    = iota
	TypeOldBool rdb.Type = rdb.TypeDriverThresh + iota
	TypeOldByte
	TypeOldInt16
	TypeOldInt32
	TypeOldInt64
	TypeOldFloat32
	TypeOldFloat64
	TypeOldTD // DateTime
	TypeNumeric
)

var sqlTypeLookup = map[rdb.Type]typeWidth{
	TypeOldBool:    {T: typeBool, SqlName: "bit"},
	TypeOldByte:    {T: typeByte, SqlName: "tinyint"},
	TypeOldInt16:   {T: typeInt16, SqlName: "smallint"},
	TypeOldInt32:   {T: typeInt32, SqlName: "int"},
	TypeOldInt64:   {T: typeInt64, SqlName: "bigint"},
	TypeOldFloat32: {T: typeFloat32, SqlName: "real"},
	TypeOldFloat64: {T: typeFloat64, SqlName: "float"},

	rdb.TypeVarChar: {T: typeNVarChar, SqlName: "nvarchar"},
	rdb.TypeChar:    {T: typeNChar, SqlName: "nchar"},
	rdb.TypeText:    {T: typeNText, SqlName: "ntext"},
	rdb.Text:        {T: typeNVarChar, SqlName: "nvarchar"},

	rdb.TypeAnsiVarChar: {T: typeVarChar, SqlName: "varchar"},
	rdb.TypeAnsiChar:    {T: typeChar, SqlName: "char"},
	rdb.TypeAnsiText:    {T: typeText, SqlName: "text"},

	rdb.TypeBinary: {T: typeVarBinary, SqlName: "varbinary"},
	rdb.Binary:     {T: typeVarBinary, SqlName: "varbinary"},

	rdb.TypeBool: {T: typeBitN, W: 1, SqlName: "bit"},
	rdb.Bool:     {T: typeBitN, W: 1, SqlName: "bit"},

	rdb.TypeInt8:  {T: typeIntN, W: 1, SqlName: "tinyint"},
	rdb.TypeInt16: {T: typeIntN, W: 2, SqlName: "smallint"},
	rdb.TypeInt32: {T: typeIntN, W: 4, SqlName: "int"},
	rdb.TypeInt64: {T: typeIntN, W: 8, SqlName: "bigint"},
	rdb.Integer:   {T: typeIntN, W: 8, SqlName: "bigint"},

	rdb.TypeDecimal: {T: typeDecimal, SqlName: "decimal"},
	TypeNumeric:     {T: typeNumeric, SqlName: "decimal"},
	rdb.Decimal:     {T: typeDecimal, SqlName: "decimal"},

	rdb.TypeFloat32: {T: typeFloatN, W: 4, SqlName: "float"},
	rdb.TypeFloat64: {T: typeFloatN, W: 8, SqlName: "float"},
	rdb.Float:       {T: typeFloatN, W: 8, SqlName: "float"},

	TypeOldTD: {T: typeDateTimeN, W: 8, SqlName: "datetime"},

	rdb.TypeTime:       {T: typeTimeN, W: 5, SqlName: "time"},
	rdb.TypeDate:       {T: typeDateN, W: 3, SqlName: "date"},
	rdb.TypeTimestamp:  {T: typeDateTime2N, W: 8, SqlName: "datetime2"},
	rdb.TypeTimestampz: {T: typeDateTimeOffsetN, W: 10, SqlName: "datetimeoffset"},
	rdb.Time:           {T: typeDateTimeOffsetN, W: 10, SqlName: "datetimeoffset"},

	rdb.TypeMoney: {T: typeMoneyN, W: 8, SqlName: "money"},

	rdb.TypeUUID: {T: typeGuid, SqlName: "uniqueidentifier"},

	rdb.TypeXML: {T: typeXml, SqlName: "xml"},
}
