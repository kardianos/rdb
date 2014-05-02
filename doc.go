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
		(TODO) Column attributes (nullable, length, precision)
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

	TODO: Complete before release.
		hook up custom field output converters:
			AddOutputConvert(fieldType rdb.SqlType, func(SqlColumn, interface{}) (interface{}, error))
			Add convert callback to Field members
			tds: Hook up above.
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
*/
/*
Usage Example:
	import (
		"bitbucket.org/kardianos/rdb"
		_ "bitbucket.org/kardianos/tds"
	)

	func SimpleQuery() (ferr error) {
		defer func() {
			if re := recover(); re != nil {
				if localError, is := re.(rdb.MustError); is {
					ferr = localError
					return
				}
				panic(re)
			}
		}()
		config := rdb.ParseConfigMust("tds://TESTU@localhost/SqlExpress?db=master")

		cmd := &rdb.Command{
			Sql: `
				select
					cast('fox' as varchar(7)) as dock,
					box = cast(@animal as nvarchar(max))
				;
			`,
			Arity: rdb.OneOnly,
			Input: []rdb.Param{
				rdb.Param{
					N: "animal",
					T: rdb.TypeString,
				},
			},
		}

		db := rdb.OpenMust(config)
		defer db.Close()

		var dock string

		res := db.Query(cmd, rdb.Value{V: "Fish"})
		defer res.Close()

		// Prep all or some of the values.
		// Can also prep by name:
		// res.Prep("dock", &dock)
		res.PrepAll(&dock)

		res.Scan()

		// The other values in the row are buffered until the next call to Scan().
		box := string(res.Get("box").([]byte))
		_ = box
	}
*/
package rdb
