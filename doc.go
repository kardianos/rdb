// Copyright 2014 Daniel Theophanes.
// Use of this source code is governed by a zlib-style
// license that can be found in the LICENSE file.

// SQL Relational Database Client.
/*

A query is defined through a *Command. Do not create a new *Command structure every
time a query is run, but create a *Command for each query once. Once created a
*Command structure shouldn't be modified while it might be used. Input parameters
are not included in the *Command and are a separate parameter in the Query methods.
Input parameters use a Value array.

The Arity field in *Command controls how the number of rows are handled. This is
used in place of an "Execute" or "QueryRow" API. If the Arity is zero, the
result is automatically closed after execution. If the Arity is one, the result
is closed after reading the first row.

Depending on the driver, the name fields "N" may be optional and the order of
of the parameters or values used. Refer to each driver's documentation for more
information. To pass in a NULL parameter, use "rdb.TypeNull" as the value.

Result.Prep should be called before Scan on each row. To prepare multiple values
Result.Scan(...) can be used to scan into value by index all at once.
The values will get stored directly into the preprared value.
Some drivers will support io.Writer for some data types.
If a value is not prep'ed, the value will be stored in a row buffer until
the next Result.Scan(). Until then, they may be accessed with
Result.{Get, Getx, GetN, GetxN}


Design Rationale:

This API was designed to work with small schemas, but also scale well to database
schemas with tables containing tens or hundreds of columns which are designed by
an external vendor. For this support named parameters and named fields are
required. Alternate values for NULL fields are also provided.

The ConnPool is not called "Database" because it doesn't represent a database, but
a connection pool (except for embedded solutions like SQLite).

Preparing a SQL statment is not an explicit option, but Command field. This
allows the interface to hide auto-preparing a statement and keeps more
connections in a connection pool available for general use. In some databases
the prepared statement is per connection, while others a prepared statement is
global. Even when a prepared statement is global, database servers can be
restarted and all existing prepared statements lost. When this happens the
database interface should simply re-recreate the prepared statement as needed.

SQL data types vary from vendor to vendor, but most at least tip their hat to a
common standard. Common SQL data types can be defined, and vendor specific data
types can be allowed.

When building tools to work with databases automatically, discovering the table
schema at query time is needed. This can help properly serialize arbitrary queries
or database mappings.

Transaction support is required. Nested transactions are not supported but
transaction save points are. Most things that could be expressed as nested
transactions can also be expressed as a single transaction with save points.

Data access operations are often very granular (looping through rows, scanning
results, verifying data), but their failure modes tend to be bulkier. Once a
failure occures, cleanup is standard and often the same for every type of failure.
Thus much of the operational API is mirrored: one that returns errors and one that
panics. All panics are of the same type and a helper function "Run" is defined
to translate panics of that type into error values. While not ideal from many
perspectives, it is very practical.

*/
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
	(Done) Handle cases where idle DB connection is reset and must be reconneted to.
		(Done) Automatically re-preparing any prepared statements.
	(Done) Handle Transactions
		(Done) Commit
		(Done) Rollback
		(Done) Save point
	(Done) Handle different isolation levels
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
Usage examples at: https://bitbucket.org/kardianos/rdb/src/default/example/
*/
package rdb
