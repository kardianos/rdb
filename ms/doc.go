// Copyright 2014 Daniel Theophanes.
// Use of this source code is governed by a zlib-style
// license that can be found in the LICENSE file.

// Microsoft SQL Server (MS SQL Server) TDS Protocol database client.
//
// For use with sql client interface:
//   github.com/kardianos/rdb
//
// This driver doesn't use cgo or any c libraries and is self contained.
//
/*

Supported Data Types
	rdb.
		TypeNull

		TypeString     :: Maps to nvarchar
		TypeAnsiString :: Maps to varchar
		TypeBinary     :: Maps to varbinary

		TypeText
		TypeAnsiText
		TypeVarChar
		TypeAnsiVarChar
		TypeChar
		TypeAnsiChar

		TypeBool
		TypeInt8
		TypeInt16
		TypeInt32
		TypeInt64

		TypeFloat32
		TypeFloat64

		TypeDecimal

		TypeTDZ
		TypeTime
		TypeDate
		TypeTD   :: Maps to DateTime2

	tds.
		TypeOldTD :: Maps to DateTime

The following types support io.Writer for output fields, and io.Reader for
input parameters:
	TypeString
	TypeAnsiString
	TypeBinary
	TypeVarChar
	TypeAnsiVarChar

Transactions are not yet supported.

Out parameters are not yet supported.

Parameter names are not optional. They must be supplied.

# TEXTSIZE Limit

SQL Server has a server-side TEXTSIZE setting that controls the maximum number of
bytes returned for varchar(max), nvarchar(max), varbinary(max), text, ntext, and
image columns. The default TEXTSIZE is 4096 bytes, which may truncate large data.

To prevent truncation, set TEXTSIZE in your connection's ResetQuery configuration:

	config := &rdb.Config{
		// ... other settings ...
		ResetQuery: "SET TEXTSIZE 2147483647",
	}

Reference: https://learn.microsoft.com/en-us/sql/t-sql/statements/set-textsize-transact-sql

# Always Encrypted

Give the column encryption keys in the DSN, one column_key per key, in the
text form of package ms/aecrypt; there is no key store. extract_cek.ps1 in that
package prints them on the machine holding the column master key certificate.

	ms://user:pass@host/EVOLUTION?db=MatchX&column_key=<fingerprint>.<key>

The connection then decrypts encrypted result columns, reporting their
plaintext type. Before each parameterized query it asks the server
(sp_describe_parameter_encryption) which parameters meet encrypted columns, and
sends those encrypted. A parameter stored into an encrypted column must be
declared with the column's exact type, Length included: an nvarchar(60) column
needs Length 60, or the server reports an operand type clash. A parameter only
compared with the column may be declared longer. Comparisons on a deterministic
column are exact (binary collation).

Decrypted columns may be char, varchar, nchar, nvarchar (each including max),
binary, varbinary, or an integer type; encrypted parameters may be nchar,
nvarchar, binary, varbinary, or an integer type. Other types, encrypted output
parameters, and encrypted parameters of a stored procedure call (rather than a
batch) are not supported. With keys set, nvarchar parameters are never sent as
varchar (the utf8 option).
*/
package ms
