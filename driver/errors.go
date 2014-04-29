package driver

import (
	"fmt"
)

type ErrorColumnNotFound struct {
	Name  string
	Index int
}

func (err ErrorColumnNotFound) Error() string {
	if len(err.Name) == 0 {
		return fmt.Sprintf("Column index not valid: %d", err.Index)
	}
	return fmt.Sprintf("Column name not valid: %s", err.Name)
}

// Type panic'ed with after calling a Must method.
type MustError struct {
	Err error
}

func (err MustError) Error() string {
	return err.Err.Error()
}
