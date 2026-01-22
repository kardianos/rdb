// Copyright 2014 Daniel Theophanes.
// Use of this source code is governed by a zlib-style
// license that can be found in the LICENSE file.

package ms

import (
	"context"
	"encoding/xml"
	"testing"

	"github.com/kardianos/rdb"
)

// normalizeXML parses and re-encodes XML to normalize whitespace and formatting.
func normalizeXML(t *testing.T, data []byte) string {
	t.Helper()
	var root xmlNode
	if err := xml.Unmarshal(data, &root); err != nil {
		t.Fatalf("failed to parse XML: %v\ndata: %s", err, string(data))
	}
	out, err := xml.Marshal(&root)
	if err != nil {
		t.Fatalf("failed to marshal XML: %v", err)
	}
	return string(out)
}

// checkXML normalizes both got and want XML and compares them.
func checkXML(t *testing.T, got []byte, want string) {
	t.Helper()
	gotNorm := normalizeXML(t, got)
	wantNorm := normalizeXML(t, []byte(want))
	if gotNorm != wantNorm {
		t.Errorf("XML mismatch:\ngot:  %s\nwant: %s", gotNorm, wantNorm)
	}
}

// xmlNode is a generic XML element for normalization.
type xmlNode struct {
	XMLName  xml.Name
	Attrs    []xml.Attr `xml:",any,attr"`
	Content  string     `xml:",chardata"`
	Children []xmlNode  `xml:",any"`
}

func (n *xmlNode) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	n.XMLName = start.Name
	n.Attrs = start.Attr

	for {
		tok, err := d.Token()
		if err != nil {
			return err
		}

		switch t := tok.(type) {
		case xml.StartElement:
			var child xmlNode
			if err := child.UnmarshalXML(d, t); err != nil {
				return err
			}
			n.Children = append(n.Children, child)
		case xml.CharData:
			n.Content += string(t)
		case xml.EndElement:
			return nil
		}
	}
}

func (n xmlNode) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	start.Name = n.XMLName
	start.Attr = n.Attrs

	if err := e.EncodeToken(start); err != nil {
		return err
	}

	if n.Content != "" {
		if err := e.EncodeToken(xml.CharData(n.Content)); err != nil {
			return err
		}
	}

	for _, child := range n.Children {
		if err := e.Encode(child); err != nil {
			return err
		}
	}

	return e.EncodeToken(xml.EndElement{Name: start.Name})
}

func TestDecodeBinXML(t *testing.T) {
	tests := []struct {
		name    string
		binxml  []byte
		want    string
		wantErr bool
	}{
		{
			name: "simple element with text",
			binxml: []byte{
				0xDF, 0xFF, // signature
				0x01,       // version 1
				0xB0, 0x04, // encoding UTF-16LE (1200)
				// NAMEDEF "root" (id 1)
				0xF0, 0x04, 0x72, 0x00, 0x6F, 0x00, 0x6F, 0x00, 0x74, 0x00,
				// QNAMEDEF 0 0 1 (id 1: ns="", prefix="", local="root")
				0xEF, 0x00, 0x00, 0x01,
				// ELEMENT qname=1
				0xF8, 0x01,
				// ENDATTRIBUTES
				0xF5,
				// SQL-NVARCHAR "Hello" (length=5 UTF-16 chars)
				0x11, 0x05, 0x48, 0x00, 0x65, 0x00, 0x6C, 0x00, 0x6C, 0x00, 0x6F, 0x00,
				// ENDELEMENT
				0xF7,
			},
			want: "<root>Hello</root>",
		},
		{
			name: "element with attribute",
			binxml: []byte{
				0xDF, 0xFF, // signature
				0x01,       // version 1
				0xB0, 0x04, // encoding UTF-16LE (1200)
				// NAMEDEF "root" (id 1)
				0xF0, 0x04, 0x72, 0x00, 0x6F, 0x00, 0x6F, 0x00, 0x74, 0x00,
				// QNAMEDEF 0 0 1 (id 1: ns="", prefix="", local="root")
				0xEF, 0x00, 0x00, 0x01,
				// NAMEDEF "attr" (id 2)
				0xF0, 0x04, 0x61, 0x00, 0x74, 0x00, 0x74, 0x00, 0x72, 0x00,
				// QNAMEDEF 0 0 2 (id 2: ns="", prefix="", local="attr")
				0xEF, 0x00, 0x00, 0x02,
				// ELEMENT qname=1
				0xF8, 0x01,
				// ATTRIBUTE qname=2
				0xF6, 0x02,
				// SQL-NVARCHAR "value" (length=5)
				0x11, 0x05, 0x76, 0x00, 0x61, 0x00, 0x6C, 0x00, 0x75, 0x00, 0x65, 0x00,
				// ENDATTRIBUTES
				0xF5,
				// ENDELEMENT
				0xF7,
			},
			want: `<root attr="value"></root>`,
		},
		{
			name: "empty document",
			binxml: []byte{
				0xDF, 0xFF, // signature
				0x01,       // version 1
				0xB0, 0x04, // encoding UTF-16LE (1200)
			},
			want: "",
		},
		{
			name: "invalid signature",
			binxml: []byte{
				0x00, 0x00, // invalid signature
				0x01,
				0xB0, 0x04,
			},
			wantErr: true,
		},
		{
			name:    "too short",
			binxml:  []byte{0xDF, 0xFF},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := decodeBinXML(tt.binxml)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("DecodeBinXML failed: %v", err)
			}
			if tt.want == "" {
				if len(result) != 0 {
					t.Errorf("expected empty result, got: %q", string(result))
				}
				return
			}
			checkXML(t, result, tt.want)
		})
	}
}

func TestXMLRoundTrip(t *testing.T) {
	checkSkip(t)

	tests := []struct {
		name string
		xml  string
	}{
		{
			name: "simple",
			xml:  "<root><item>Hello</item></root>",
		},
		{
			name: "with namespace",
			xml:  `<root xmlns="http://example.com"><item>Data</item></root>`,
		},
		{
			name: "with text content",
			xml:  "<test><data>Hello World</data></test>",
		},
		{
			name: "with attributes",
			xml:  `<root attr="value"><child id="1">Content</child></root>`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name+"/cast", func(t *testing.T) {
			defer recoverTest(t)

			cmd := &rdb.Command{
				SQL:   `select cast(@xml as xml);`,
				Arity: rdb.OneMust,
			}
			params := []rdb.Param{
				{Name: "xml", Type: rdb.Text, Value: tt.xml},
			}

			res := db.Query(context.Background(), cmd, params...)
			defer res.Close()

			res.Scan()
			val := res.Getx(0)

			if val == nil {
				t.Fatal("XML should not be nil")
			}

			xmlBytes, ok := val.([]byte)
			if !ok {
				t.Fatalf("XML should be []byte, got %T", val)
			}

			checkXML(t, xmlBytes, tt.xml)
		})

		t.Run(tt.name+"/param", func(t *testing.T) {
			defer recoverTest(t)

			cmd := &rdb.Command{
				SQL:   `select @xml;`,
				Arity: rdb.OneMust,
			}
			params := []rdb.Param{
				{Name: "xml", Type: rdb.TypeXML, Value: tt.xml},
			}

			res := db.Query(context.Background(), cmd, params...)
			defer res.Close()

			res.Scan()
			val := res.Getx(0)

			if val == nil {
				t.Fatal("XML should not be nil")
			}

			xmlBytes, ok := val.([]byte)
			if !ok {
				t.Fatalf("XML should be []byte, got %T", val)
			}

			checkXML(t, xmlBytes, tt.xml)
		})
	}
}

func TestXMLNull(t *testing.T) {
	checkSkip(t)
	defer recoverTest(t)

	cmd := &rdb.Command{
		SQL:   `select @xml;`,
		Arity: rdb.OneMust,
	}

	params := []rdb.Param{
		{Name: "xml", Type: rdb.TypeXML, Value: nil, Null: true},
	}

	res := db.Query(context.Background(), cmd, params...)
	defer res.Close()

	res.Scan()
	val := res.Getx(0)

	if val != nil {
		t.Fatalf("XML should be nil: %v", val)
	}
}
