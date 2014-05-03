// Copyright 2014 Daniel Theophanes.
// Use of this source code is governed by a zlib-style
// license that can be found in the LICENSE file.

// SQL Relational Database Client.
/*
Features:
	(Done) Named Parameters
	(Done) Inspect driver support
	(TODO) Driver callbacks - events
	(TODO) Manage data type mapping
		(TODO) Handle custom data types
	(Done) For result fields, can write to io.Writer
	(Done) For input parameters, can read from io.Reader
	(Done) View schema
		(Done) Column Names
		(TODO) Mapped data types
		(Done) Driver data types
		(Done) Column attributes (nullable, length, precision)
	(TODO) Bulk Insert
	(TODO) Handle cases where idle DB connection is reset and must be reconneted to.
		(TODO) Automatically re-preparing any prepared statements.
	(TODO) Handle Transactions
		(TODO) Commit
		(TODO) Rollback
		(TODO) Save point
	(TODO) Handle different isolation levels
	(TODO) Provide a unified method of logging:
		(TODO) Full executed sql statements with all parameter values.
			(Parameters map optionally be opmitted from log).
		(TODO) Any errors that occur.
		(TODO) Time taken for execution. Useful for ongoing QOS. Associate with query.
		(Done) (Can name Commands)
	(Done) Provide a standard SQL syntax error structure that can be inspected:
		Sql text
		Line number
		Error text
		Can have multiple SqlErrors for a given query.
	(TODO) Set active collation.

	(TODO) Should be able to infer many parameter types.
	(TODO) Should be able to set default types based on native types for inputs.
	(TODO) Should be able to set default for data type outputs.
	(TODO) Custom marshal hooks.

	TODO:
		hook up custom field output converters:
			AddOutputConvert(fieldType rdb.SqlType, func(SqlColumn, interface{}) (interface{}, error))
			Add convert callback to Field members
		hook up Custom types for input parameters:
			AddInputConvert(func(in interface{}) (handled bool, tp SqlType, out interface{})
*/
/*
A query is defined through a *Command. If the input values are constant or if
the values are defined at the time of query with []Value, then the *Command
structure can be defined once and used concurrently for all queries. The Arity
field in *Command controls how the number of rows are handled. This is used in
place of an "Exequte" or "QueryRow" API.

Depending on the driver, the name fields "N" may be optional and the order of
of the parameters or values used. Refer to each driver's documentation for more
information. To pass in a NULL parameter, use "rdb.TypeNull" as the value.

Result.Prep and Result.PrepAll(...) should be called before Scan on each row.
The values will get stored directly into the preprared value.
Some drivers will support io.Writer for some data types.
If a value is not prep'ed, the value will be stored in a row buffer until
the next Result.Scan(). Until then, they may be accessed with
Result.{Get, Getx, GetN, GetxN}

The API is not yet final.

*/
/*
Usage Example:
	import (
		"bitbucket.org/kardianos/rdb"
		_ "bitbucket.org/kardianos/tds"

		"fmt"
	)

	const testConnectionString = "tds://TESTU@localhost/SqlExpress?db=master"

	func TestSimpleQuery() {
		err := QueryTest()
		if err != nil {
			t.Error(err)
		}
	}

	func QueryTest() (ferr error) {
		defer func() {
			if re := recover(); re != nil {
				if localError, is := re.(rdb.MustError); is {
					ferr = localError
					return
				}
				panic(re)
			}
		}()
		config := rdb.ParseConfigMust(testConnectionString)

		db := rdb.OpenMust(config)
		defer db.Close()

		SimpleQuery(db)
		RowsQuery(db)
		LargerQuery(db)
		return nil
	}

	func SimpleQuery(db rdb.DatabaseMust) {
		var myFav string
		db.Query(&rdb.Command{
			Sql: `
				select @animal as 'MyAnimal';`,
			Arity: rdb.OneMust,
			Input: []rdb.Param{
				rdb.Param{
					N: "animal",
					T: rdb.TypeString,
					L: 8,
					V: "DogIsFriend",
				},
			},
			TruncLongText: true,
		}).Prep("MyAnimal", &myFav).Scan()
	}
	func RowsQuery(db rdb.DatabaseMust) {
		var myFav string
		res := db.Query(&rdb.Command{
			Sql: `
				select @animal as 'MyAnimal'
				union all
				select N'Hello again!'
			;`,
			Arity: rdb.Any,
			Input: []rdb.Param{
				rdb.Param{
					N: "animal",
					T: rdb.TypeString,
					V: "Dreaming boats.",
				},
			},
			TruncLongText: true,
		})
		defer res.Close()
		for {
			res.Prep("MyAnimal", &myFav)
			if !res.Scan() {
				break
			}
			fmt.Printf("Animal: %s\n", myFav)
		}
	}
	func LargerQuery(db rdb.DatabaseMust) {
		cmd := &rdb.Command{
			Sql: `
				select
					432 as ID,
					987.654 as Val,
					cast('fox' as varchar(7)) as dock,
					box = cast(@animal as nvarchar(max))
				;
				`,
			Arity: rdb.OneMust,
			Input: []rdb.Param{
				rdb.Param{
					N: "animal",
					T: rdb.TypeString,
				},
			},
		}

		var dock string
		var id int
		var val float64

		res := db.Query(cmd, rdb.Value{V: "Fish"})
		defer res.Close()

		res.PrepAll(&id, &val, &dock)

		res.Scan()

		// The other values in the row are buffered until the next call to Scan().
		box := string(res.Get("box").([]byte))
		_ = box
		fmt.Printf("ID: %d\n", id)
		fmt.Printf("Val: %f\n", val)
	}

*/
package rdb
