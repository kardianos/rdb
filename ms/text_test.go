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
	t.Skipf("Long text values not supported.")
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
		// {Name: "AsNText", Limit: false},
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
	t.Skipf("Long text values not supported.") // Text that spans packets are not supported.

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
