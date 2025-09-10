package table

import (
	"encoding/json"
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
		ID      int64    `db:"user_id"`
		Name    string   `db:"full_name"`
		Details []Detail `db:"user_details,json"`
	}

	// Create schema
	schema := []*rdb.Column{
		{Name: "user_id", Type: rdb.TypeInt64},
		{Name: "full_name", Type: rdb.Text},
		{Name: "user_details", Type: rdb.Text}, // JSON stored as string.
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
	jsonData1, _ := json.Marshal(details1)
	jsonData2, _ := json.Marshal(details2)

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
