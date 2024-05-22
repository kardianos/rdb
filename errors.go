// Copyright 2014 Daniel Theophanes.
// Use of this source code is governed by a zlib-style
// license that can be found in the LICENSE file.

package rdb

import (
	"bytes"
	"errors"
	"fmt"
)

var ErrScanNull = errors.New("can only scan NULL value into a Nullable type")
var ErrPreparedTokenNotValid = errors.New("the prepared token is not valid")

// Should be returned by a driver that doesn't implement a feature.
var ErrNotImplemented = errors.New("the feature has not been implemented")

// Used when a column lookup fails, either with a name or index.
type ErrorColumnNotFound struct {
	At    string
	Name  string
	Index int
}

func (err ErrorColumnNotFound) Error() string {
	if len(err.Name) == 0 {
		return fmt.Sprintf("At <%s>, Column index not valid: %d", err.At, err.Index)
	}
	return fmt.Sprintf("At <%s>, Column name not valid: %s", err.At, err.Name)
}

// List of SQL errors returned by the server.
type Errors []*Message

func (errs Errors) Error() string {
	bb := &bytes.Buffer{}
	if errs == nil {
		return ""
	}
	for i, err := range errs {
		if i != 0 {
			bb.WriteString("\n")
		}
		bb.WriteString(fmt.Sprintf("%v", err))
	}
	return bb.String()
}

type MessageType byte

const (
	_                    = iota
	SqlError MessageType = iota
	SqlInfo
)

func (mt MessageType) String() string {
	switch mt {
	default:
		return "?"
	case SqlError:
		return "ERROR"
	case SqlInfo:
		return "INFO"
	}
}

// SQL errors reported by the server.
// Must always be wrapped by Errors.
// This is why it doesn't satisfy the error interface.
type Message struct {
	Type       MessageType
	Message    string
	ServerName string
	ProcName   string
	LineNumber int32
	SqlState   string
	Number     int32
	State      byte
	Class      byte
}

func (err *Message) String() string {
	return fmt.Sprintf("(%s %s: %d) L%d: %s (%d, %d)", err.ServerName, err.ProcName, err.Number, err.LineNumber, err.Message, err.State, err.Class)
}

// Exposed to help isolate error paths when starting a client.
type DriverNotFound struct {
	name string
}

func (dr DriverNotFound) Error() string {
	return fmt.Sprintf("Driver name not found: %s", dr.name)
}

var ErrArity = errors.New("result row count does not match desired arity")

var ErrCancel = errors.New("Query Cancelled")
