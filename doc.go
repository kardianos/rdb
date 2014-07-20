// Copyright 2014 Daniel Theophanes.
// Use of this source code is governed by a zlib-style
// license that can be found in the LICENSE file.

// SQL Relational Database Client.
/*

A query is defined through a *Command. Do not create a new *Command structure every
time a query is run, but create a *Command for each query once. Once created a
*Command structure shouldn't be modified while it might be used.

The Arity field in *Command controls how the number of rows are handled.
If the Arity is zero, the result is automatically closed after execution.
If the Arity is one, the result is closed after reading the first row.

Depending on the driver, the name fields "N" may be optional and the order of
of the parameters or values used. Refer to each driver's documentation for more
information. To pass in a NULL parameter, use "rdb.TypeNull" as the value.

Result.Prep should be called before Scan on each row. To prepare multiple values
Result.Scan(...) can be used to scan into value by index all at once.
Some drivers will support io.Writer for some data types.
If a value is not prep'ed, the value will be stored in a row buffer until
the next Result.Scan().
*/
/*
Simple example:
	cmd := &rdb.Command{
		Sql: `select * from Account where ID = :ID;`,
		Arity: rdb.OneMust,
	}
	res, err := db.Query(cmd, rdb.Param{Name: "ID", Value: 2})
	if err != nil {
		return err
	}
	// Next() is not required.
	var id int
	err = res.Scan(&id)
	if err != nil {
		return err
	}
*/
/*
Usage examples at: https://bitbucket.org/kardianos/rdb/src/default/example/
*/
package rdb
