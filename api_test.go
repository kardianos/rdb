package rdb

import (
	"testing"
)

func TestApi(t *testing.T) {
	conf, err := ParseConfig(`testmem://u:p@localhost/test`)
	// conf.PanicOnError = true

	db, err := Open(conf)
	_ = err

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
		Sql: sql,
		One: true,
		// Convert: GoString -> nvarchar Length=300
		Input: []Param{
			Param{N: "foo", T: SqlTypeString, V: foo},
			// input type is a go string, Length=10 (nvarchar mapping from command convert mapping).
		},
	}
	/*
		The output params need to be different:
			Don't need specific type information.
			Do need method to convert and how to handle null.
			Input needs server side typing and possible conversion helper.
			Output needs client side typing and possible conversion helper.
	*/

	var bar, box string

	// err := db.QueryRow(sql, foo).Scan(&bar, &box)
	res, _ := db.Query(cmd)
	if err != nil {
		panic(err)
	}
	_, err = res.PrepAll(&bar, &box).ScanPrep()
	if err != nil {
		panic(err)
	}

	res2, err := db.Query(cmd)
	if err != nil {
		panic(err)
	}
	for {
		var eof bool
		eof, err = res2.ScanBuffer()
		if err != nil {
			panic(err)
		}
		if eof {
			break
		}

		bar := res2.Get("bar").(string)
		box := res2.Get("box").(string)

		_, _ = bar, box
	}

	res3, err := db.Query(cmd)
	if err != nil {
		panic(err)
	}
	for {
		var bar, box string

		res2.Prep("bar", &bar)
		res2.Prep("box", &box)

		var eof bool
		eof, err = res3.ScanPrep()
		if err != nil {
			panic(err)
		}
		if eof {
			break
		}
	}

}
