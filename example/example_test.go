package example

import (
	"bitbucket.org/kardianos/rdb"
	_ "bitbucket.org/kardianos/tds"
	"testing"
)

func TestExample(t *testing.T) {
	defer func() {
		if re := recover(); re != nil {
			if localError, is := re.(rdb.MustError); is {
				t.Errorf("SQL Error: %v", localError)
				return
			}
			panic(re)
		}
	}()
	config := rdb.ParseConfigMust("tds://TESTU@localhost/SqlExpress?db=master")

	cmd := &rdb.Command{
		Sql: `
			select cast('fox' as varchar(7)) as dock, box = cast(@animal as nvarchar(max));
		`,
		Arity: rdb.OneOnly,
		Input: []*rdb.Param{
			&rdb.Param{
				N: "animal",
				T: rdb.TypeString,
			},
		},
	}

	db := rdb.OpenMust(config)
	defer db.Close()

	var dock, box string

	res := db.Query(cmd, rdb.Value{V: "Fish"})
	defer res.Close()

	res.PrepAll(&dock, &box)
	res.Scan()

	t.Logf("Dock: %s, Box: %s", dock, box)
}
