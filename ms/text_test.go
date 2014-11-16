package ms

import (
	"fmt"
	"testing"

	"bitbucket.org/kardianos/rdb"
)

func TestText(t *testing.T) {
	defer recoverTest(t)

	var testText = "Hello"
	var testLimit = 100

	type TI struct {
		Name  string
		Limit bool
		Value string
	}
	var list = []*TI{
		// &TI{Name: "AsText", Limit: false},
		// &TI{Name: "AsNText", Limit: false},
		&TI{Name: "AsVarChar", Limit: true},
		&TI{Name: "AsNVarChar", Limit: true},
		&TI{Name: "AsVarCharMax", Limit: false},
		&TI{Name: "AsNVarCharMax", Limit: false},
	}

	cmd := &rdb.Command{
		Sql: fmt.Sprintf(`
			select
				-- AsText = cast('%[1]s' as Text),
				-- AsNText = cast('%[1]s' as NText),
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
		item.Value = string(res.Get(item.Name).([]byte))
	}
	for _, item := range list {
		var compareTo = testText
		if item.Limit && len(testText) > testLimit && len(item.Value) > testLimit {
			compareTo = testText[:testLimit]
		}

		if item.Value != compareTo {
			t.Errorf("Field %s not correct value. Is: %s", item.Name, item.Value)
		}
	}
}
