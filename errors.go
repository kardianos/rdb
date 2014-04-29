package rdb

import (
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
