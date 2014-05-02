// By Daniel Theophanes

// SQL Relational Database Client.
/*
	Named Parameters
	Inspect driver support
	Driver callbacks - events
	Manage data type mapping
		Handle custom data types
	For result fields, can write to io.Writer
	For input parameters, can read from io.Reader
	View schema
		Column Names
		Mapped data types
		Driver data types
		Column attributes (nullable, length, precision)
	Bulk Insert
	Handle cases where idle DB connection is reset and must be reconneted to.
		Automatically re-preparing any prepared statements.
	Handle Transactions
		Commit
		Rollback
		Save point
	Handle different isolation levels
	Provide a unified method of logging:
		Full executed sql statements with all parameter values.
			(Parameters map optionally be opmitted from log).
		Any errors that occur.
		Time taken for execution. Useful for ongoing QOS. Associate with query.
		(Can name Commands)
	Provide a standard SQL syntax error structure that can be inspected:
		Sql text
		Line number
		Error text
		Can have multiple SqlErrors for a given query.
	Set active collation.

*/
package rdb
