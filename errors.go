package rdb

import (
	"bytes"
	"fmt"
)

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

// Type panic'ed with after calling a Must method.
type MustError struct {
	Err error
}

func (err MustError) Error() string {
	return err.Err.Error()
}

type SqlErrors []*SqlError

func (errs SqlErrors) Error() string {
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

type SqlError struct {
	Message    string
	ServerName string
	ProcName   string
	LineNumber int32
	SqlState   string
	Number     int32
}

func (err *SqlError) Error() string {
	return fmt.Sprintf("(%s %s: %d) L%d: %s)", err.ServerName, err.ProcName, err.Number, err.LineNumber, err.Message)
}
