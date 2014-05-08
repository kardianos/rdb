// Copyright 2014 Daniel Theophanes.
// Use of this source code is governed by a zlib-style
// license that can be found in the LICENSE file.

// SQL Relational Database Client.
/*
Features:
	(Done) Named Parameters
	(Done) Inspect driver support
	(Done) Support OUT parameters
	(TODO) Handle drivers with multiple result sets
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
A feature is done when there is at least one driver using it.
A section of the API is stable when at least two drivers implement it.

*/
/*
Usage examples at: https://bitbucket.org/kardianos/rdb/src/default/example/
*/
package rdb
