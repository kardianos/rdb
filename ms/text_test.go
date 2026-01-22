package ms

import (
	"context"
	"encoding/hex"
	"fmt"
	"strings"
	"testing"

	"github.com/kardianos/rdb"
)

func TestShortText(t *testing.T) {
	checkSkip(t)
	defer recoverTest(t)

	var testText = strings.Repeat("X", 200)
	var testLimit = 100

	type TI struct {
		Name  string
		Limit bool
		Value string
	}
	var list = []*TI{
		{Name: "AsText", Limit: false},
		{Name: "AsNText", Limit: false},
		{Name: "AsVarChar", Limit: true},
		{Name: "AsNVarChar", Limit: true},
		{Name: "AsVarCharMax", Limit: false},
		{Name: "AsNVarCharMax", Limit: false},
	}

	cmd := &rdb.Command{
		SQL: fmt.Sprintf(`
			select
				AsText = cast('%[1]s' as Text),
				AsNText = cast('%[1]s' as NText),
				AsVarChar = cast('%[1]s' as varchar(%[2]d)),
				AsNVarChar = cast('%[1]s' as nvarchar(%[2]d)),
				AsVarCharMax = cast('%[1]s' as varchar(max)),
				AsNVarCharMax = cast('%[1]s' as nvarchar(max))
		`, testText, testLimit),
		Arity: rdb.OneMust,
	}

	res := db.Query(context.Background(), cmd)
	defer res.Close()

	res.Scan()

	for _, item := range list {
		if v, is := res.Get(item.Name).([]byte); is {
			item.Value = string(v)
		}
	}
	for _, item := range list {
		var compareTo = testText
		if item.Limit && len(testText) > testLimit {
			compareTo = testText[:testLimit]
		}

		if item.Value != compareTo {
			// dv := hex.Dump([]byte(item.Value))
			t.Errorf("Field %s not correct value.\n", item.Name)
		}
	}
}
func TestLongText(t *testing.T) {
	checkSkip(t)
	t.Skipf("Long text values not supported - text spanning packets not implemented.")
	defer recoverTest(t)

	var testText = strings.Repeat("X", 50000)
	var testLimit = 100

	type TI struct {
		Name  string
		Limit bool
		Value string
	}
	var list = []*TI{
		{Name: "AsText", Limit: false},
	}

	cmd := &rdb.Command{
		SQL: fmt.Sprintf(`
			select
				AsText = cast('%[1]s' as Text)
		`, testText, testLimit),
		Arity: rdb.OneMust,
	}

	res := db.Query(context.Background(), cmd)
	defer res.Close()

	res.Scan()

	for _, item := range list {
		if v, is := res.Get(item.Name).([]byte); is {
			item.Value = string(v)
		}
	}
	for _, item := range list {
		var compareTo = testText
		if item.Limit && len(testText) > testLimit {
			compareTo = testText[:testLimit]
		}

		if item.Value != compareTo {
			dv := hex.Dump([]byte(item.Value))
			t.Errorf("Field %s not correct value, len: %d\n:%s", item.Name, len(item.Value), dv)
		}
	}
}

func TestImage(t *testing.T) {
	checkSkip(t)
	defer recoverTest(t)

	testData := []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0A}

	cmd := &rdb.Command{
		SQL: `
			select cast(@data as image);
		`,
		Arity: rdb.OneMust,
	}

	params := []rdb.Param{
		{Name: "data", Type: rdb.Binary, Value: testData},
	}

	res := db.Query(context.Background(), cmd, params...)
	defer res.Close()

	res.Scan()
	val := res.Getx(0)

	if val == nil {
		t.Fatalf("Image should not be nil")
	}

	got, ok := val.([]byte)
	if !ok {
		t.Fatalf("Image should be []byte, got %T", val)
	}

	if len(got) != len(testData) {
		t.Errorf("Image length mismatch: got %d, want %d", len(got), len(testData))
	}

	for i := range testData {
		if got[i] != testData[i] {
			t.Errorf("Image byte %d mismatch: got %02x, want %02x", i, got[i], testData[i])
		}
	}
}
