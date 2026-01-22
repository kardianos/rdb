// Copyright 2014 Daniel Theophanes.
// Use of this source code is governed by a zlib-style
// license that can be found in the LICENSE file.

package ms

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"math"
	"math/big"
	"time"
	"unicode/utf16"
)

// MS-BINXML token constants.
const (
	// Structure tokens.
	binxmlXMLDECL       = 0xFE
	binxmlENCODING      = 0xFD
	binxmlDOCTYPEDECL   = 0xFC
	binxmlSYSTEM        = 0xFB
	binxmlPUBLIC        = 0xFA
	binxmlSUBSET        = 0xF9
	binxmlELEMENT       = 0xF8
	binxmlENDELEMENT    = 0xF7
	binxmlATTRIBUTE     = 0xF6
	binxmlENDATTRIBUTES = 0xF5
	binxmlPI            = 0xF4
	binxmlCOMMENT       = 0xF3
	binxmlCDATA         = 0xF2
	binxmlCDATAEND      = 0xF1
	binxmlNAMEDEF       = 0xF0
	binxmlQNAMEDEF      = 0xEF
	binxmlNEST          = 0xEC
	binxmlENDNEST       = 0xEB
	binxmlEXTN          = 0xEA
	binxmlFLUSHNAMES    = 0xE9

	// SQL atomic value types.
	binxmlSQLSMALLINT     = 0x01
	binxmlSQLINT          = 0x02
	binxmlSQLREAL         = 0x03
	binxmlSQLFLOAT        = 0x04
	binxmlSQLMONEY        = 0x05
	binxmlSQLBIT          = 0x06
	binxmlSQLTINYINT      = 0x07
	binxmlSQLBIGINT       = 0x08
	binxmlSQLUUID         = 0x09
	binxmlSQLDECIMAL      = 0x0A
	binxmlSQLNUMERIC      = 0x0B
	binxmlSQLBINARY       = 0x0C
	binxmlSQLCHAR         = 0x0D
	binxmlSQLNCHAR        = 0x0E
	binxmlSQLVARBINARY    = 0x0F
	binxmlSQLVARCHAR      = 0x10
	binxmlSQLNVARCHAR     = 0x11
	binxmlSQLDATETIME     = 0x12
	binxmlSQLSMALLDATETIM = 0x13
	binxmlSQLSMALLMONEY   = 0x14
	binxmlSQLTEXT         = 0x16
	binxmlSQLIMAGE        = 0x17
	binxmlSQLNTEXT        = 0x18
	binxmlSQLUDT          = 0x1B

	// XSD atomic value types.
	binxmlXSDTIMEOFFSET     = 0x7A
	binxmlXSDDATETIMEOFFSET = 0x7B
	binxmlXSDDATEOFFSET     = 0x7C
	binxmlXSDTIME2          = 0x7D
	binxmlXSDDATETIME2      = 0x7E
	binxmlXSDDATE2          = 0x7F
	binxmlXSDTIME           = 0x81
	binxmlXSDDATETIME       = 0x82
	binxmlXSDDATE           = 0x83
	binxmlXSDBINHEX         = 0x84
	binxmlXSDBASE64         = 0x85
	binxmlXSDBOOLEAN        = 0x86
	binxmlXSDDECIMAL        = 0x87
	binxmlXSDBYTE           = 0x88
	binxmlXSDUNSIGNEDSHORT  = 0x89
	binxmlXSDUNSIGNEDINT    = 0x8A
	binxmlXSDUNSIGNEDLONG   = 0x8B
	binxmlXSDQNAME          = 0x8C
)

// binxmlQName represents a qualified name (namespace URI, prefix, local name).
type binxmlQName struct {
	namespaceURI string
	prefix       string
	localName    string
}

// binxmlDecoder decodes MS-BINXML format to text XML.
type binxmlDecoder struct {
	r       *bytes.Reader
	names   []string      // name table (index 0 = empty string)
	qnames  []binxmlQName // qname table (index 0 = invalid)
	version byte
	out     bytes.Buffer
}

// decodeBinXML converts MS-BINXML binary data to text XML.
func decodeBinXML(data []byte) ([]byte, error) {
	if len(data) < 5 {
		return nil, fmt.Errorf("binxml: data too short")
	}

	d := &binxmlDecoder{
		r:      bytes.NewReader(data),
		names:  []string{""},      // index 0 = empty string
		qnames: []binxmlQName{{}}, // index 0 = invalid
	}

	if err := d.decodeDocument(); err != nil {
		return nil, err
	}

	return d.out.Bytes(), nil
}

func (d *binxmlDecoder) decodeDocument() error {
	// Read signature.
	sig := make([]byte, 2)
	if _, err := io.ReadFull(d.r, sig); err != nil {
		return fmt.Errorf("binxml: failed to read signature: %w", err)
	}
	if sig[0] != 0xDF || sig[1] != 0xFF {
		return fmt.Errorf("binxml: invalid signature: %X %X", sig[0], sig[1])
	}

	// Read version.
	ver, err := d.r.ReadByte()
	if err != nil {
		return fmt.Errorf("binxml: failed to read version: %w", err)
	}
	if ver != 1 && ver != 2 && ver != 0 {
		return fmt.Errorf("binxml: unsupported version: %d", ver)
	}
	if ver == 0 {
		ver = 1 // Treat 0 as version 1.
	}
	d.version = ver

	// Read encoding (must be UTF-16LE = 0x04B0 = 1200).
	enc := make([]byte, 2)
	if _, err := io.ReadFull(d.r, enc); err != nil {
		return fmt.Errorf("binxml: failed to read encoding: %w", err)
	}
	// encoding is little-endian: 0xB0 0x04 = 1200
	if enc[0] != 0xB0 || enc[1] != 0x04 {
		return fmt.Errorf("binxml: unsupported encoding: %X %X", enc[0], enc[1])
	}

	// Process content.
	return d.decodeContent(true)
}

func (d *binxmlDecoder) decodeContent(isDocument bool) error {
	for {
		b, err := d.r.ReadByte()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}

		switch b {
		case binxmlXMLDECL:
			if err := d.decodeXMLDecl(); err != nil {
				return err
			}
		case binxmlDOCTYPEDECL:
			if err := d.decodeDoctypeDecl(); err != nil {
				return err
			}
		case binxmlNAMEDEF:
			if err := d.decodeNameDef(); err != nil {
				return err
			}
		case binxmlQNAMEDEF:
			if err := d.decodeQNameDef(); err != nil {
				return err
			}
		case binxmlELEMENT:
			if err := d.decodeElement(); err != nil {
				return err
			}
		case binxmlPI:
			if err := d.decodePI(); err != nil {
				return err
			}
		case binxmlCOMMENT:
			if err := d.decodeComment(); err != nil {
				return err
			}
		case binxmlCDATA:
			if err := d.decodeCDATA(); err != nil {
				return err
			}
		case binxmlEXTN:
			if err := d.decodeExtension(); err != nil {
				return err
			}
		case binxmlFLUSHNAMES:
			// Reset name tables.
			d.names = []string{""}
			d.qnames = []binxmlQName{{}}
		case binxmlNEST:
			// Nested document - save state, decode, restore.
			savedNames := d.names
			savedQNames := d.qnames
			d.names = []string{""}
			d.qnames = []binxmlQName{{}}
			if err := d.decodeDocument(); err != nil {
				return err
			}
			d.names = savedNames
			d.qnames = savedQNames
		case binxmlENDNEST:
			return nil
		case binxmlENDELEMENT:
			// End of element - caller handles this.
			d.r.UnreadByte()
			return nil
		default:
			// Check if it's an atomic value.
			if isAtomicValueToken(b) {
				d.r.UnreadByte()
				if err := d.decodeAtomicValue(); err != nil {
					return err
				}
			} else {
				return fmt.Errorf("binxml: unexpected token 0x%02X at position %d", b, d.r.Size()-int64(d.r.Len()))
			}
		}
	}
}

func (d *binxmlDecoder) decodeXMLDecl() error {
	// Read version text.
	version, err := d.readTextData()
	if err != nil {
		return err
	}

	d.out.WriteString("<?xml version=\"")
	d.out.WriteString(escapeXMLAttr(version))
	d.out.WriteString("\"")

	// Check for encoding token.
	b, err := d.r.ReadByte()
	if err != nil {
		return err
	}
	if b == binxmlENCODING {
		encoding, err := d.readTextData()
		if err != nil {
			return err
		}
		d.out.WriteString(" encoding=\"")
		d.out.WriteString(escapeXMLAttr(encoding))
		d.out.WriteString("\"")
		b, err = d.r.ReadByte()
		if err != nil {
			return err
		}
	}

	// Standalone byte.
	switch b {
	case 0x01:
		d.out.WriteString(" standalone=\"yes\"")
	case 0x02:
		d.out.WriteString(" standalone=\"no\"")
	}

	d.out.WriteString("?>")
	return nil
}

func (d *binxmlDecoder) decodeDoctypeDecl() error {
	// Read name.
	name, err := d.readTextData()
	if err != nil {
		return err
	}

	d.out.WriteString("<!DOCTYPE ")
	d.out.WriteString(name)

	// Check for SYSTEM, PUBLIC, SUBSET tokens.
	for {
		b, err := d.r.ReadByte()
		if err != nil {
			return err
		}

		switch b {
		case binxmlSYSTEM:
			systemID, err := d.readTextData()
			if err != nil {
				return err
			}
			d.out.WriteString(" SYSTEM \"")
			d.out.WriteString(escapeXMLAttr(systemID))
			d.out.WriteString("\"")
		case binxmlPUBLIC:
			publicID, err := d.readTextData()
			if err != nil {
				return err
			}
			d.out.WriteString(" PUBLIC \"")
			d.out.WriteString(escapeXMLAttr(publicID))
			d.out.WriteString("\"")
		case binxmlSUBSET:
			subset, err := d.readTextData()
			if err != nil {
				return err
			}
			d.out.WriteString(" [")
			d.out.WriteString(subset)
			d.out.WriteString("]")
		default:
			d.r.UnreadByte()
			d.out.WriteString(">")
			return nil
		}
	}
}

func (d *binxmlDecoder) decodeNameDef() error {
	name, err := d.readTextData()
	if err != nil {
		return err
	}
	d.names = append(d.names, name)
	return nil
}

func (d *binxmlDecoder) decodeQNameDef() error {
	nsIndex, err := d.readMB32()
	if err != nil {
		return err
	}
	prefixIndex, err := d.readMB32()
	if err != nil {
		return err
	}
	localIndex, err := d.readMB32()
	if err != nil {
		return err
	}

	qn := binxmlQName{
		namespaceURI: d.getName(int(nsIndex)),
		prefix:       d.getName(int(prefixIndex)),
		localName:    d.getName(int(localIndex)),
	}
	d.qnames = append(d.qnames, qn)
	return nil
}

func (d *binxmlDecoder) decodeElement() error {
	qnIndex, err := d.readMB32()
	if err != nil {
		return err
	}

	qn := d.getQName(int(qnIndex))

	d.out.WriteString("<")
	d.writeQName(qn)

	// Process attributes and content.
	hasAttributes := false
	for {
		b, err := d.r.ReadByte()
		if err != nil {
			return err
		}

		switch b {
		case binxmlNAMEDEF:
			if err := d.decodeNameDef(); err != nil {
				return err
			}
		case binxmlQNAMEDEF:
			if err := d.decodeQNameDef(); err != nil {
				return err
			}
		case binxmlATTRIBUTE:
			hasAttributes = true
			if err := d.decodeAttribute(); err != nil {
				return err
			}
		case binxmlENDATTRIBUTES:
			d.out.WriteString(">")
			// Decode element content.
			if err := d.decodeContent(false); err != nil {
				return err
			}
			// Read ENDELEMENT.
			end, err := d.r.ReadByte()
			if err != nil {
				return err
			}
			if end != binxmlENDELEMENT {
				return fmt.Errorf("binxml: expected ENDELEMENT, got 0x%02X", end)
			}
			d.out.WriteString("</")
			d.writeQName(qn)
			d.out.WriteString(">")
			return nil
		case binxmlENDELEMENT:
			// Self-closing element (no attributes case).
			if hasAttributes {
				d.out.WriteString(">")
				d.out.WriteString("</")
				d.writeQName(qn)
				d.out.WriteString(">")
			} else {
				d.out.WriteString("/>")
			}
			return nil
		case binxmlEXTN:
			if err := d.decodeExtension(); err != nil {
				return err
			}
		case binxmlFLUSHNAMES:
			d.names = []string{""}
			d.qnames = []binxmlQName{{}}
		default:
			// Might be content starting (no ENDATTRIBUTES for empty element).
			d.r.UnreadByte()
			if !hasAttributes {
				d.out.WriteString("/>")
				return nil
			}
			return fmt.Errorf("binxml: unexpected token 0x%02X in element", b)
		}
	}
}

func (d *binxmlDecoder) decodeAttribute() error {
	qnIndex, err := d.readMB32()
	if err != nil {
		return err
	}

	qn := d.getQName(int(qnIndex))

	d.out.WriteString(" ")
	d.writeQName(qn)
	d.out.WriteString("=\"")

	// Read attribute value (may have metadata before it).
	for {
		b, err := d.r.ReadByte()
		if err != nil {
			return err
		}

		switch b {
		case binxmlNAMEDEF:
			if err := d.decodeNameDef(); err != nil {
				return err
			}
		case binxmlQNAMEDEF:
			if err := d.decodeQNameDef(); err != nil {
				return err
			}
		case binxmlEXTN:
			if err := d.decodeExtension(); err != nil {
				return err
			}
		case binxmlFLUSHNAMES:
			d.names = []string{""}
			d.qnames = []binxmlQName{{}}
		case binxmlATTRIBUTE, binxmlENDATTRIBUTES, binxmlENDELEMENT:
			// End of attribute value.
			d.r.UnreadByte()
			d.out.WriteString("\"")
			return nil
		default:
			if isAtomicValueToken(b) {
				d.r.UnreadByte()
				if err := d.decodeAtomicValueForAttr(); err != nil {
					return err
				}
			} else {
				d.r.UnreadByte()
				d.out.WriteString("\"")
				return nil
			}
		}
	}
}

func (d *binxmlDecoder) decodePI() error {
	targetIndex, err := d.readMB32()
	if err != nil {
		return err
	}
	data, err := d.readTextData()
	if err != nil {
		return err
	}

	d.out.WriteString("<?")
	d.out.WriteString(d.getName(int(targetIndex)))
	if data != "" {
		d.out.WriteString(" ")
		d.out.WriteString(data)
	}
	d.out.WriteString("?>")
	return nil
}

func (d *binxmlDecoder) decodeComment() error {
	text, err := d.readTextData()
	if err != nil {
		return err
	}
	d.out.WriteString("<!--")
	d.out.WriteString(text)
	d.out.WriteString("-->")
	return nil
}

func (d *binxmlDecoder) decodeCDATA() error {
	d.out.WriteString("<![CDATA[")
	for {
		text, err := d.readTextData()
		if err != nil {
			return err
		}
		d.out.WriteString(text)

		b, err := d.r.ReadByte()
		if err != nil {
			return err
		}
		if b == binxmlCDATAEND {
			break
		}
		if b != binxmlCDATA {
			return fmt.Errorf("binxml: expected CDATA or CDATAEND, got 0x%02X", b)
		}
	}
	d.out.WriteString("]]>")
	return nil
}

func (d *binxmlDecoder) decodeExtension() error {
	length, err := d.readMB32()
	if err != nil {
		return err
	}
	// Skip extension data.
	_, err = d.r.Seek(int64(length), io.SeekCurrent)
	return err
}

func (d *binxmlDecoder) decodeAtomicValue() error {
	b, err := d.r.ReadByte()
	if err != nil {
		return err
	}

	s, err := d.readAtomicValueString(b)
	if err != nil {
		return err
	}
	d.out.WriteString(escapeXMLText(s))
	return nil
}

func (d *binxmlDecoder) decodeAtomicValueForAttr() error {
	b, err := d.r.ReadByte()
	if err != nil {
		return err
	}

	s, err := d.readAtomicValueString(b)
	if err != nil {
		return err
	}
	d.out.WriteString(escapeXMLAttr(s))
	return nil
}

func (d *binxmlDecoder) readAtomicValueString(token byte) (string, error) {
	switch token {
	case binxmlSQLBIT:
		b, err := d.r.ReadByte()
		if err != nil {
			return "", err
		}
		if b == 0 {
			return "0", nil
		}
		return "1", nil

	case binxmlSQLTINYINT, binxmlXSDBYTE:
		b, err := d.r.ReadByte()
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("%d", b), nil

	case binxmlSQLSMALLINT:
		bb := make([]byte, 2)
		if _, err := io.ReadFull(d.r, bb); err != nil {
			return "", err
		}
		return fmt.Sprintf("%d", int16(binary.LittleEndian.Uint16(bb))), nil

	case binxmlXSDUNSIGNEDSHORT:
		bb := make([]byte, 2)
		if _, err := io.ReadFull(d.r, bb); err != nil {
			return "", err
		}
		return fmt.Sprintf("%d", binary.LittleEndian.Uint16(bb)), nil

	case binxmlSQLINT:
		bb := make([]byte, 4)
		if _, err := io.ReadFull(d.r, bb); err != nil {
			return "", err
		}
		return fmt.Sprintf("%d", int32(binary.LittleEndian.Uint32(bb))), nil

	case binxmlXSDUNSIGNEDINT:
		bb := make([]byte, 4)
		if _, err := io.ReadFull(d.r, bb); err != nil {
			return "", err
		}
		return fmt.Sprintf("%d", binary.LittleEndian.Uint32(bb)), nil

	case binxmlSQLBIGINT:
		bb := make([]byte, 8)
		if _, err := io.ReadFull(d.r, bb); err != nil {
			return "", err
		}
		return fmt.Sprintf("%d", int64(binary.LittleEndian.Uint64(bb))), nil

	case binxmlXSDUNSIGNEDLONG:
		bb := make([]byte, 8)
		if _, err := io.ReadFull(d.r, bb); err != nil {
			return "", err
		}
		return fmt.Sprintf("%d", binary.LittleEndian.Uint64(bb)), nil

	case binxmlSQLREAL:
		bb := make([]byte, 4)
		if _, err := io.ReadFull(d.r, bb); err != nil {
			return "", err
		}
		f := math.Float32frombits(binary.LittleEndian.Uint32(bb))
		return fmt.Sprintf("%v", f), nil

	case binxmlSQLFLOAT:
		bb := make([]byte, 8)
		if _, err := io.ReadFull(d.r, bb); err != nil {
			return "", err
		}
		f := math.Float64frombits(binary.LittleEndian.Uint64(bb))
		return fmt.Sprintf("%v", f), nil

	case binxmlSQLMONEY:
		bb := make([]byte, 8)
		if _, err := io.ReadFull(d.r, bb); err != nil {
			return "", err
		}
		val := int64(binary.LittleEndian.Uint64(bb))
		r := big.NewRat(val, 10000)
		return r.FloatString(4), nil

	case binxmlSQLSMALLMONEY:
		bb := make([]byte, 4)
		if _, err := io.ReadFull(d.r, bb); err != nil {
			return "", err
		}
		val := int64(int32(binary.LittleEndian.Uint32(bb)))
		r := big.NewRat(val, 10000)
		return r.FloatString(4), nil

	case binxmlSQLDECIMAL, binxmlSQLNUMERIC, binxmlXSDDECIMAL:
		return d.readDecimal()

	case binxmlSQLNCHAR:
		return d.readTextData()

	case binxmlSQLNVARCHAR, binxmlSQLNTEXT:
		return d.readTextData64()

	case binxmlSQLCHAR:
		return d.readCodePageText()

	case binxmlSQLVARCHAR, binxmlSQLTEXT:
		return d.readCodePageText64()

	case binxmlSQLBINARY, binxmlSQLUDT, binxmlXSDBINHEX:
		data, err := d.readBlob()
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("%X", data), nil

	case binxmlSQLVARBINARY, binxmlSQLIMAGE, binxmlXSDBASE64:
		data, err := d.readBlob64()
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("%X", data), nil

	case binxmlSQLUUID:
		bb := make([]byte, 16)
		if _, err := io.ReadFull(d.r, bb); err != nil {
			return "", err
		}
		// GUID format with byte swapping.
		return fmt.Sprintf("%08X-%04X-%04X-%02X%02X-%012X",
			binary.LittleEndian.Uint32(bb[0:4]),
			binary.LittleEndian.Uint16(bb[4:6]),
			binary.LittleEndian.Uint16(bb[6:8]),
			bb[8], bb[9],
			bb[10:16]), nil

	case binxmlSQLDATETIME:
		bb := make([]byte, 8)
		if _, err := io.ReadFull(d.r, bb); err != nil {
			return "", err
		}
		days := int32(binary.LittleEndian.Uint32(bb[0:4]))
		ticks := binary.LittleEndian.Uint32(bb[4:8])
		// Base date is 1900-01-01, ticks are 1/300th of a second.
		t := time.Date(1900, 1, 1, 0, 0, 0, 0, time.UTC)
		t = t.AddDate(0, 0, int(days))
		t = t.Add(time.Duration(ticks) * time.Second / 300)
		return t.Format("2006-01-02T15:04:05.000"), nil

	case binxmlSQLSMALLDATETIM:
		bb := make([]byte, 4)
		if _, err := io.ReadFull(d.r, bb); err != nil {
			return "", err
		}
		days := binary.LittleEndian.Uint16(bb[0:2])
		mins := binary.LittleEndian.Uint16(bb[2:4])
		t := time.Date(1900, 1, 1, 0, 0, 0, 0, time.UTC)
		t = t.AddDate(0, 0, int(days))
		t = t.Add(time.Duration(mins) * time.Minute)
		return t.Format("2006-01-02T15:04:00"), nil

	case binxmlXSDBOOLEAN:
		b, err := d.r.ReadByte()
		if err != nil {
			return "", err
		}
		if b == 0 {
			return "false", nil
		}
		return "true", nil

	case binxmlXSDDATE:
		return d.readXSDDate()

	case binxmlXSDDATETIME:
		return d.readXSDDateTime()

	case binxmlXSDTIME:
		return d.readXSDTime()

	case binxmlXSDDATE2:
		return d.readSqlDate()

	case binxmlXSDDATETIME2, binxmlXSDTIME2:
		return d.readSqlDateTime2(token == binxmlXSDTIME2)

	case binxmlXSDDATEOFFSET, binxmlXSDDATETIMEOFFSET, binxmlXSDTIMEOFFSET:
		return d.readSqlDateTimeOffset(token)

	case binxmlXSDQNAME:
		idx, err := d.readMB32()
		if err != nil {
			return "", err
		}
		qn := d.getQName(int(idx))
		if qn.prefix != "" {
			return qn.prefix + ":" + qn.localName, nil
		}
		return qn.localName, nil

	default:
		return "", fmt.Errorf("binxml: unknown atomic value token 0x%02X", token)
	}
}

func (d *binxmlDecoder) readDecimal() (string, error) {
	length, err := d.readMB32()
	if err != nil {
		return "", err
	}
	if length < 3 {
		return "", fmt.Errorf("binxml: decimal too short")
	}

	prec, err := d.r.ReadByte()
	if err != nil {
		return "", err
	}
	scale, err := d.r.ReadByte()
	if err != nil {
		return "", err
	}
	sign, err := d.r.ReadByte()
	if err != nil {
		return "", err
	}

	valueLen := int(length) - 3
	bb := make([]byte, valueLen)
	if _, err := io.ReadFull(d.r, bb); err != nil {
		return "", err
	}

	// Value is little-endian, need to reverse for big.Int.
	for i, j := 0, len(bb)-1; i < j; i, j = i+1, j-1 {
		bb[i], bb[j] = bb[j], bb[i]
	}

	integer := new(big.Int).SetBytes(bb)
	if sign == 0 {
		integer.Neg(integer)
	}

	r := new(big.Rat).SetInt(integer)
	divisor := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(scale)), nil)
	r.Quo(r, new(big.Rat).SetInt(divisor))

	_ = prec // precision not used for formatting
	return r.FloatString(int(scale)), nil
}

func (d *binxmlDecoder) readXSDDate() (string, error) {
	bb := make([]byte, 8)
	if _, err := io.ReadFull(d.r, bb); err != nil {
		return "", err
	}
	value := int64(binary.LittleEndian.Uint64(bb))
	// Lower 2 bits = 1, rest encodes date with timezone.
	value >>= 2

	_ = (value % (60 * 29)) - 60*14 // timezone adjustment (not used in output)
	value /= 60 * 29

	day := int(value%31) + 1
	value /= 31
	month := int(value%12) + 1
	value /= 12
	year := int(value) - 9999

	return fmt.Sprintf("%04d-%02d-%02d", year, month, day), nil
}

func (d *binxmlDecoder) readXSDDateTime() (string, error) {
	bb := make([]byte, 8)
	if _, err := io.ReadFull(d.r, bb); err != nil {
		return "", err
	}
	value := int64(binary.LittleEndian.Uint64(bb))
	// Lower 2 bits = 2.
	value >>= 2

	ms := int(value % 1000)
	value /= 1000
	sec := int(value % 60)
	value /= 60
	min := int(value % 60)
	value /= 60
	hour := int(value % 24)
	value /= 24
	day := int(value%31) + 1
	value /= 31
	month := int(value%12) + 1
	value /= 12
	year := int(value) - 9999

	return fmt.Sprintf("%04d-%02d-%02dT%02d:%02d:%02d.%03d", year, month, day, hour, min, sec, ms), nil
}

func (d *binxmlDecoder) readXSDTime() (string, error) {
	bb := make([]byte, 8)
	if _, err := io.ReadFull(d.r, bb); err != nil {
		return "", err
	}
	value := int64(binary.LittleEndian.Uint64(bb))
	// Lower 2 bits = 0.
	value >>= 2

	ms := int(value % 1000)
	value /= 1000
	sec := int(value % 60)
	value /= 60
	min := int(value % 60)
	value /= 60
	hour := int(value)

	return fmt.Sprintf("%02d:%02d:%02d.%03d", hour, min, sec, ms), nil
}

func (d *binxmlDecoder) readSqlDate() (string, error) {
	bb := make([]byte, 3)
	if _, err := io.ReadFull(d.r, bb); err != nil {
		return "", err
	}
	days := int(bb[0]) | int(bb[1])<<8 | int(bb[2])<<16
	t := time.Date(1, 1, 1, 0, 0, 0, 0, time.UTC).AddDate(0, 0, days)
	return t.Format("2006-01-02"), nil
}

func (d *binxmlDecoder) readSqlDateTime2(timeOnly bool) (string, error) {
	prec, err := d.r.ReadByte()
	if err != nil {
		return "", err
	}

	// Time bytes based on precision.
	var timeBytes int
	switch {
	case prec <= 2:
		timeBytes = 3
	case prec <= 4:
		timeBytes = 4
	default:
		timeBytes = 5
	}

	timeBB := make([]byte, timeBytes)
	if _, err := io.ReadFull(d.r, timeBB); err != nil {
		return "", err
	}

	dateBB := make([]byte, 3)
	if _, err := io.ReadFull(d.r, dateBB); err != nil {
		return "", err
	}

	// Decode time.
	var timeVal uint64
	for i := 0; i < timeBytes; i++ {
		timeVal |= uint64(timeBB[i]) << (8 * i)
	}
	scale := int64(1)
	for i := 0; i < int(prec); i++ {
		scale *= 10
	}
	ns := int64(timeVal) * (1000000000 / scale)

	// Decode date.
	days := int(dateBB[0]) | int(dateBB[1])<<8 | int(dateBB[2])<<16

	t := time.Date(1, 1, 1, 0, 0, 0, 0, time.UTC).AddDate(0, 0, days)
	t = t.Add(time.Duration(ns))

	if timeOnly {
		return t.Format("15:04:05.999999999"), nil
	}
	return t.Format("2006-01-02T15:04:05.999999999"), nil
}

func (d *binxmlDecoder) readSqlDateTimeOffset(token byte) (string, error) {
	prec, err := d.r.ReadByte()
	if err != nil {
		return "", err
	}

	// Time bytes based on precision.
	var timeBytes int
	switch {
	case prec <= 2:
		timeBytes = 3
	case prec <= 4:
		timeBytes = 4
	default:
		timeBytes = 5
	}

	timeBB := make([]byte, timeBytes)
	if _, err := io.ReadFull(d.r, timeBB); err != nil {
		return "", err
	}

	dateBB := make([]byte, 3)
	if _, err := io.ReadFull(d.r, dateBB); err != nil {
		return "", err
	}

	tzBB := make([]byte, 2)
	if _, err := io.ReadFull(d.r, tzBB); err != nil {
		return "", err
	}
	tzOffset := int16(binary.LittleEndian.Uint16(tzBB))

	// Decode time.
	var timeVal uint64
	for i := 0; i < timeBytes; i++ {
		timeVal |= uint64(timeBB[i]) << (8 * i)
	}
	scale := int64(1)
	for i := 0; i < int(prec); i++ {
		scale *= 10
	}
	ns := int64(timeVal) * (1000000000 / scale)

	// Decode date.
	days := int(dateBB[0]) | int(dateBB[1])<<8 | int(dateBB[2])<<16

	t := time.Date(1, 1, 1, 0, 0, 0, 0, time.UTC).AddDate(0, 0, days)
	t = t.Add(time.Duration(ns))

	// Format timezone.
	tzStr := ""
	if tzOffset == 0 {
		tzStr = "Z"
	} else {
		sign := "+"
		if tzOffset < 0 {
			sign = "-"
			tzOffset = -tzOffset
		}
		tzStr = fmt.Sprintf("%s%02d:%02d", sign, tzOffset/60, tzOffset%60)
	}

	switch token {
	case binxmlXSDTIMEOFFSET:
		return t.Format("15:04:05.999999999") + tzStr, nil
	case binxmlXSDDATEOFFSET:
		return t.Format("2006-01-02") + tzStr, nil
	default:
		return t.Format("2006-01-02T15:04:05.999999999") + tzStr, nil
	}
}

func (d *binxmlDecoder) readTextData() (string, error) {
	length, err := d.readMB32()
	if err != nil {
		return "", err
	}
	if length == 0 {
		return "", nil
	}
	// Length is in UTF-16 characters, so multiply by 2 for bytes.
	bb := make([]byte, int(length)*2)
	if _, err := io.ReadFull(d.r, bb); err != nil {
		return "", err
	}
	return decodeUTF16LE(bb), nil
}

func (d *binxmlDecoder) readTextData64() (string, error) {
	length, err := d.readMB64()
	if err != nil {
		return "", err
	}
	if length == 0 {
		return "", nil
	}
	// Length is in UTF-16 characters, so multiply by 2 for bytes.
	bb := make([]byte, int(length)*2)
	if _, err := io.ReadFull(d.r, bb); err != nil {
		return "", err
	}
	return decodeUTF16LE(bb), nil
}

func (d *binxmlDecoder) readCodePageText() (string, error) {
	length, err := d.readMB32()
	if err != nil {
		return "", err
	}
	if length < 4 {
		return "", nil
	}
	// First 4 bytes are code page.
	cpBB := make([]byte, 4)
	if _, err := io.ReadFull(d.r, cpBB); err != nil {
		return "", err
	}
	codePage := binary.LittleEndian.Uint32(cpBB)

	dataLen := int(length) - 4
	if dataLen == 0 {
		return "", nil
	}
	bb := make([]byte, dataLen)
	if _, err := io.ReadFull(d.r, bb); err != nil {
		return "", err
	}

	if codePage == 1200 {
		// UTF-16LE.
		return decodeUTF16LE(bb), nil
	}
	// For other code pages, just treat as Latin-1 (best effort).
	return string(bb), nil
}

func (d *binxmlDecoder) readCodePageText64() (string, error) {
	length, err := d.readMB64()
	if err != nil {
		return "", err
	}
	if length < 4 {
		return "", nil
	}
	// First 4 bytes are code page.
	cpBB := make([]byte, 4)
	if _, err := io.ReadFull(d.r, cpBB); err != nil {
		return "", err
	}
	codePage := binary.LittleEndian.Uint32(cpBB)

	dataLen := int(length) - 4
	if dataLen == 0 {
		return "", nil
	}
	bb := make([]byte, dataLen)
	if _, err := io.ReadFull(d.r, bb); err != nil {
		return "", err
	}

	if codePage == 1200 {
		// UTF-16LE.
		return decodeUTF16LE(bb), nil
	}
	// For other code pages, just treat as Latin-1 (best effort).
	return string(bb), nil
}

func (d *binxmlDecoder) readBlob() ([]byte, error) {
	length, err := d.readMB32()
	if err != nil {
		return nil, err
	}
	if length == 0 {
		return nil, nil
	}
	bb := make([]byte, int(length))
	if _, err := io.ReadFull(d.r, bb); err != nil {
		return nil, err
	}
	return bb, nil
}

func (d *binxmlDecoder) readBlob64() ([]byte, error) {
	length, err := d.readMB64()
	if err != nil {
		return nil, err
	}
	if length == 0 {
		return nil, nil
	}
	bb := make([]byte, int(length))
	if _, err := io.ReadFull(d.r, bb); err != nil {
		return nil, err
	}
	return bb, nil
}

func (d *binxmlDecoder) readMB32() (uint32, error) {
	var result uint32
	var shift uint
	for i := 0; i < 5; i++ {
		b, err := d.r.ReadByte()
		if err != nil {
			return 0, err
		}
		result |= uint32(b&0x7F) << shift
		if b < 0x80 {
			return result, nil
		}
		shift += 7
	}
	return 0, fmt.Errorf("binxml: mb32 overflow")
}

func (d *binxmlDecoder) readMB64() (uint64, error) {
	var result uint64
	var shift uint
	for i := 0; i < 10; i++ {
		b, err := d.r.ReadByte()
		if err != nil {
			return 0, err
		}
		result |= uint64(b&0x7F) << shift
		if b < 0x80 {
			return result, nil
		}
		shift += 7
	}
	return 0, fmt.Errorf("binxml: mb64 overflow")
}

func (d *binxmlDecoder) getName(index int) string {
	if index < 0 || index >= len(d.names) {
		return ""
	}
	return d.names[index]
}

func (d *binxmlDecoder) getQName(index int) binxmlQName {
	if index < 0 || index >= len(d.qnames) {
		return binxmlQName{}
	}
	return d.qnames[index]
}

func (d *binxmlDecoder) writeQName(qn binxmlQName) {
	if qn.prefix != "" {
		d.out.WriteString(qn.prefix)
		d.out.WriteString(":")
	}
	d.out.WriteString(qn.localName)
}

func isAtomicValueToken(b byte) bool {
	// SQL types: 0x01-0x14, 0x16-0x18, 0x1B
	// XSD types: 0x7A-0x8C (with gaps)
	return (b >= 0x01 && b <= 0x14) ||
		(b >= 0x16 && b <= 0x18) ||
		b == 0x1B ||
		(b >= 0x7A && b <= 0x8C)
}

func decodeUTF16LE(b []byte) string {
	if len(b)%2 != 0 {
		b = b[:len(b)-1]
	}
	u16 := make([]uint16, len(b)/2)
	for i := 0; i < len(u16); i++ {
		u16[i] = binary.LittleEndian.Uint16(b[i*2:])
	}
	return string(utf16.Decode(u16))
}

func escapeXMLText(s string) string {
	var buf bytes.Buffer
	for _, r := range s {
		switch r {
		case '<':
			buf.WriteString("&lt;")
		case '>':
			buf.WriteString("&gt;")
		case '&':
			buf.WriteString("&amp;")
		default:
			buf.WriteRune(r)
		}
	}
	return buf.String()
}

func escapeXMLAttr(s string) string {
	var buf bytes.Buffer
	for _, r := range s {
		switch r {
		case '<':
			buf.WriteString("&lt;")
		case '>':
			buf.WriteString("&gt;")
		case '&':
			buf.WriteString("&amp;")
		case '"':
			buf.WriteString("&quot;")
		case '\'':
			buf.WriteString("&apos;")
		default:
			buf.WriteRune(r)
		}
	}
	return buf.String()
}
