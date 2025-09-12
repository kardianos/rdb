package table

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"reflect"
	"testing"

	"github.com/kardianos/rdb"
)

func TestUnmarshalStruct(t *testing.T) {
	// Test struct with various data types and tags.
	type Detail struct {
		Key   string `json:"key"`
		Value int    `json:"value"`
	}
	type TestStruct struct {
		XMLName xml.Name `xml:"TS" db:"-"`
		ID      int64    `db:"user_id"`
		Name    string   `db:"full_name"`
		Details []Detail `db:",json"`
	}

	// Create schema
	schema := []*rdb.Column{
		{Name: "user_id", Type: rdb.TypeInt64},
		{Name: "full_name", Type: rdb.Text},
		{Name: "Details", Type: rdb.Text},      // JSON stored as string.
		{Name: "extra_column", Type: rdb.Text}, // Extra column to be ignored.
	}

	// Create buffer and set schema
	buf := &Buffer{Name: "test_table"}
	if err := buf.SetSchema(schema); err != nil {
		t.Fatalf("SetSchema failed: %v", err)
	}

	// Create sample JSON data
	details1 := []Detail{
		{Key: "age", Value: 30},
		{Key: "score", Value: 95},
	}
	details2 := []Detail{
		{Key: "level", Value: 5},
	}
	marshal := func(v any) []byte {
		buf := &bytes.Buffer{}
		c := json.NewEncoder(buf)
		c.SetEscapeHTML(false)
		err := c.Encode(v)
		if err != nil {
			t.Fatal(err)
		}
		return buf.Bytes()
	}
	jsonData1 := marshal(details1)
	jsonData2 := marshal(details2)

	// Add rows to buffer. JSON can be either a string or []byte.
	buf.AddRow(int64(1), "Alice Smith", string(jsonData1), "extra1")
	buf.AddRow(int64(2), "Bob Jones", jsonData2, "extra2")

	// Unmarshal into slice of TestStruct.
	result, err := UnmarshalStruct[TestStruct](buf)
	if err != nil {
		t.Fatalf("UnmarshalStruct failed: %v", err)
	}

	// Expected output
	expected := []TestStruct{
		{
			ID:   1,
			Name: "Alice Smith",
			Details: []Detail{
				{Key: "age", Value: 30},
				{Key: "score", Value: 95},
			},
		},
		{
			ID:   2,
			Name: "Bob Jones",
			Details: []Detail{
				{Key: "level", Value: 5},
			},
		},
	}

	// Compare results.
	if len(result) != len(expected) {
		t.Errorf("Expected %d rows, got %d", len(expected), len(result))
	}

	for i, got := range result {
		if !reflect.DeepEqual(got, expected[i]) {
			t.Errorf("Row %d mismatch:\nGot: %+v\nExpected: %+v", i, got, expected[i])
		}
	}

	// Test with nil buffer.
	_, err = UnmarshalStruct[TestStruct](nil)
	if err == nil {
		t.Error("Expected error for nil buffer, got none")
	}

	// Test with invalid JSON.
	buf = &Buffer{Name: "test_table"}
	if err := buf.SetSchema([]*rdb.Column{{Name: "user_details", Type: rdb.Text}}); err != nil {
		t.Fatalf("SetSchema failed: %v", err)
	}
	buf.AddRow("invalid json")
	_, err = UnmarshalStruct[TestStruct](buf)
	if err == nil {
		t.Error("Expected JSON unmarshal error, got none")
	}
}
