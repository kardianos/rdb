package rdb

import (
	"testing"
)

func TestApi(t *testing.T) {
	conf, err := ParseConfig(`testmem://u:p@localhost/test`)
	if err != nil {
		t.Errorf("Could not parse configuration string: %v", err)
		return
	}

	defer func() {
		if re := recover(); re != nil {
			if localError, is := re.(MustError); is {
				t.Errorf("SQL Error: %v", localError)
				return
			}
			panic(re)
		}
	}()

	db := OpenMust(conf)

	sql := `
		select
			'Foo' as bar,
			'Fii' as box
		where
			1 = @foo
		;
	`
	var foo = 1

	// This could be static and read-only, shared between all queries.
	cmd := &Command{
		Sql:   sql,
		Arity: One,

		// Convert: GoString -> nvarchar Length=300
		Input: []Param{
			Param{N: "foo", T: TypeString, V: foo},
			// input type is a go string, Length=10 (nvarchar mapping from command convert mapping).
		},
	}

	var bar, box string

	// err := db.QueryRow(sql, foo).Scan(&bar, &box)
	db.Query(cmd).PrepAll(&bar, &box).ScanPrep()
	// Result is closed after one scan due to the arity set to "One".

	res2 := db.Query(cmd)
	for {
		var eof bool
		eof = res2.ScanBuffer()
		if eof {
			break
		}

		bar := res2.Get("bar").(string)
		box := res2.Get("box").(string)

		_, _ = bar, box
	}

	res3 := db.Query(cmd)
	for {
		var bar, box string

		res2.Prep("bar", &bar)
		res2.Prep("box", &box)

		if res3.ScanPrep() {
			break
		}
	}

}
