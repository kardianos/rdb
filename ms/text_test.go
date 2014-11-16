package ms

import (
	"encoding/hex"
	"fmt"
	"strings"
	"testing"

	"bitbucket.org/kardianos/rdb"
)

func TestShortText(t *testing.T) {
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
		&TI{Name: "AsText", Limit: false},
		// &TI{Name: "AsNText", Limit: false},
		&TI{Name: "AsVarChar", Limit: true},
		&TI{Name: "AsNVarChar", Limit: true},
		&TI{Name: "AsVarCharMax", Limit: false},
		&TI{Name: "AsNVarCharMax", Limit: false},
	}

	cmd := &rdb.Command{
		Sql: fmt.Sprintf(`
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

	res := db.Query(cmd)
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
		&TI{Name: "AsText", Limit: false},
	}

	cmd := &rdb.Command{
		Sql: fmt.Sprintf(`
			select
				AsText = cast('%[1]s' as Text)
		`, testText, testLimit),
		Arity: rdb.OneMust,
	}

	res := db.Query(cmd)
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
