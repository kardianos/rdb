// Copyright 2014 Daniel Theophanes.
// Use of this source code is governed by a zlib-style
// license that can be found in the LICENSE file.

package ms

import (
	"errors"
	"fmt"
)

var connectionOpenError = errors.New("Connection already open")
var connectionNotOpenError = errors.New("Connection not open")
var connectionInUseError = errors.New("Connection already in use")

type UnexpectedMessage struct {
	Expected PacketType
	Recieved PacketType
}

func (msg UnexpectedMessage) Error() string {
	return fmt.Sprintf("Expected message type %X, recieved type %X.", msg.Expected, msg.Recieved)
}

type InputToolong struct {
	DataLen, TypeLen uint32
}

func (err InputToolong) Error() string {
	return fmt.Sprintf("Value too long: data length is %d bytes, type lenth is %d bytes", err.DataLen, err.TypeLen)
}
