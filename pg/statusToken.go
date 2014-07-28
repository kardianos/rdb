// Copyright 2014 Daniel Theophanes.
// Use of this source code is governed by a zlib-style
// license that can be found in the rdb LICENSE file.

package pg

type transactionStatus byte

const (
	tranStatusIdle                transactionStatus = 'I'
	tranStatusIdleInTransaction   transactionStatus = 'T'
	tranStatusInFailedTransaction transactionStatus = 'E'
)

type status byte

type StatusMessage struct {
	Status  status
	Message string
}

var statusString = map[status]string{
	statusSeverity:         "Severity",
	statusCode:             "Code",
	statusMessage:          "Message",
	statusDetail:           "Detail",
	statusHint:             "Hint",
	statusPosition:         "Position",
	statusInternalPosition: "Internal position",
	statusInternalQuery:    "Internal query",
	statusWhere:            "Where",
	statusSchema:           "Schema",
	statusTable:            "Table",
	statusColumn:           "Column",
	statusDataType:         "Data type",
	statusConstraint:       "Constraint",
	statusFile:             "File",
	statusLine:             "Line",
	statusRoutine:          "Routine",
}

func (msg *StatusMessage) String() string {
	code, found := statusString[msg.Status]
	if !found {
		code = "???"
	}
	return code + ": " + msg.Message
}

const (
	// Severity: the field contents are ERROR, FATAL, or PANIC (in an error message), or WARNING, NOTICE, DEBUG, INFO, or LOG (in a notice message), or a localized translation of one of these. Always present.
	statusSeverity status = 'S'

	// Code: the SQLSTATE code for the error (see Appendix A). Not localizable. Always present.
	statusCode status = 'C'

	// Message: the primary human-readable error message. This should be accurate but terse (typically one line). Always present.
	statusMessage status = 'M'

	// Detail: an optional secondary error message carrying more detail about the problem. Might run to multiple lines.
	statusDetail status = 'D'

	// Hint: an optional suggestion what to do about the problem. This is intended to differ from Detail in that it offers advice (potentially inappropriate) rather than hard facts. Might run to multiple lines.
	statusHint status = 'H'

	// Position: the field value is a decimal ASCII integer, indicating an error cursor position as an index into the original query string. The first character has index 1, and positions are measured in characters not bytes.
	statusPosition status = 'P'

	// Internal position: this is defined the same as the P field, but it is used when the cursor position refers to an internally generated command rather than the one submitted by the client. The q field will always appear when this field appears.
	statusInternalPosition status = 'p'

	// Internal query: the text of a failed internally-generated command. This could be, for example, a SQL query issued by a PL/pgSQL function.
	statusInternalQuery status = 'q'

	// Where: an indication of the context in which the error occurred. Presently this includes a call stack traceback of active procedural language functions and internally-generated queries. The trace is one entry per line, most recent first.
	statusWhere status = 'W'

	// Schema name: if the error was associated with a specific database object, the name of the schema containing that object, if any.
	statusSchema status = 's'

	// Table name: if the error was associated with a specific table, the name of the table. (Refer to the schema name field for the name of the table's schema.)
	statusTable status = 't'

	// Column name: if the error was associated with a specific table column, the name of the column. (Refer to the schema and table name fields to identify the table.)
	statusColumn status = 'c'

	// Data type name: if the error was associated with a specific data type, the name of the data type. (Refer to the schema name field for the name of the data type's schema.)
	statusDataType status = 'd'

	// Constraint name: if the error was associated with a specific constraint, the name of the constraint. Refer to fields listed above for the associated table or domain. (For this purpose, indexes are treated as constraints, even if they weren't created with constraint syntax.)
	statusConstraint status = 'n'

	// File: the file name of the source-code location where the error was reported.
	statusFile status = 'F'

	// Line: the line number of the source-code location where the error was reported.
	statusLine status = 'L'

	// Routine: the name of the source-code routine reporting the error.
	statusRoutine status = 'R'
)
