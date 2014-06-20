// Copyright 2014 Daniel Theophanes.
// Use of this source code is governed by a zlib-style
// license that can be found in the LICENSE file.

package ms

import (
	"fmt"
	"strings"

	"bitbucket.org/kardianos/rdb"
)

type SqlDone struct {
	StatusCode uint16
	CurrentCmd uint16
	Rows       uint64
}

func (done *SqlDone) Status() string {
	if done.StatusCode == 0 {
		return "Final"
	}
	codes := []string{}

	if 0x01&done.StatusCode != 0 {
		codes = append(codes, "More")
	}
	if 0x02&done.StatusCode != 0 {
		codes = append(codes, "Error")
	}
	if 0x04&done.StatusCode != 0 {
		codes = append(codes, "Transaction in progress")
	}
	if 0x10&done.StatusCode != 0 {
		codes = append(codes, fmt.Sprintf("Rows: %d", done.Rows))
	}
	if 0x20&done.StatusCode != 0 {
		codes = append(codes, "Attention")
	}
	if 0x100&done.StatusCode != 0 {
		codes = append(codes, "Server Error. Discard results.")
	}
	if len(codes) == 0 {
		panic(fmt.Sprintf("Unknown code: %d", done.StatusCode))
	}
	return strings.Join(codes, " & ")
}
func (done *SqlDone) String() string {
	return fmt.Sprintf("Done Cmd=%d Status=%s", done.CurrentCmd, done.Status())
}
func (done *SqlDone) Error() string {
	return done.Status()
}

type SqlColumn struct {
	rdb.SqlColumn

	Collation [5]byte

	code driverType
	info typeInfo
}

type SqlRow struct{}

type SqlRpcResult int32

type recoverError struct {
	err error
}
