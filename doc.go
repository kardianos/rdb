// By Daniel Theophanes

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
*/
package rdb
