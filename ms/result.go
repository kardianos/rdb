// Copyright 2014 Daniel Theophanes.
// Use of this source code is governed by a zlib-style
// license that can be found in the LICENSE file.

package ms

import (
	"fmt"
	"strings"

	"github.com/kardianos/rdb"
)

type MsgEom struct{}
type MsgRow struct{}
type MsgColumn struct{}
type MsgFinalDone struct{}
type MsgCancel struct{}
type MsgRowCount struct {
	Count uint64
}
type MsgOther struct {
	Other interface{}
}

type MsgDone struct {
	StatusCode uint16
	CurrentCmd uint16
	Rows       uint64
}

func (done MsgDone) Status() string {
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
func (done MsgDone) String() string {
	return fmt.Sprintf("Done Cmd=%d Status=%s", done.CurrentCmd, done.Status())
}
func (done MsgDone) Error() string {
	return done.Status()
}

type SQLColumn struct {
	rdb.Column

	Collation [5]byte

	code driverType
	info typeInfo
}

type MsgEnvChange struct{}

type MsgParamValue struct{}

type MsgRpcResult int32

type MsgOrder []uint16

type recoverError struct {
	err error
}
