package rdb

import (
	"testing"
)

func TestApi(t *testing.T) {
	conn := ParseConfigM(`testmem://u:p@localhost/test`)

	db, err := Open(conn)
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
			Param{N: "foo", T: Type{}, V: foo},
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
	db.QueryM(cmd).PrepAll(&bar, &box).ScanPrepM()
}
