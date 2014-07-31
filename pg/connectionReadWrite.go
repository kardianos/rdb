// Copyright 2014 Daniel Theophanes.
// Use of this source code is governed by a zlib-style
// license that can be found in the rdb LICENSE file.

package pg

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
)

const debug = false

type panicError struct {
	err error
}

// Helper methods decrease message read length.

type reader struct {
	*bufio.Reader
	buf    []byte
	Length int32
}

func (pg *connection) reader() *reader {
	return &reader{
		Reader: pg.readBuffer,
		buf:    pg.scratch[:],
	}
}

func (r *reader) Msg() byte {
	msg := r.Byte()
	r.Length = r.Int32() - 4
	return msg
}

func (r *reader) Byte() byte {
	r.Length--
	val, err := r.ReadByte()
	if err != nil {
		panic(panicError{err})
	}
	return val
}
func (r *reader) Int8() int8 {
	r.Length--
	val, err := r.ReadByte()
	if err != nil {
		panic(panicError{err})
	}
	return int8(val)
}
func (r *reader) Int16() int16 {
	r.Length -= 2
	_, err := r.Read(r.buf[:2])
	if err != nil {
		panic(panicError{err})
	}
	return int16(binary.BigEndian.Uint16(r.buf[:2]))
}
func (r *reader) Int32() int32 {
	r.Length -= 4
	_, err := r.Read(r.buf[:4])
	if err != nil {
		panic(panicError{err})
	}
	return int32(binary.BigEndian.Uint32(r.buf[:4]))
}
func (r *reader) String() string {
	buf := &bytes.Buffer{}

	var value byte

	for r.Length > 0 {
		value = r.Byte()
		if value == 0 {
			break
		}
		buf.WriteByte(value)
	}
	return buf.String()
}

var errLengthTooLong = errors.New("Requested read is longer then message length.")

func (r *reader) Bytea(length int32) []byte {
	var buf []byte
	if r.Length < length {
		panic(panicError{errLengthTooLong})
	}
	if length <= int32(len(r.buf)) {
		buf = r.buf[:length]
	} else {
		buf = make([]byte, length)
	}
	n, err := io.ReadFull(r, buf)
	r.Length -= int32(n)
	if err != nil {
		panic(panicError{err})
	}
	return buf
}
func (r *reader) MsgDone() {
	if r.Length <= 0 {
		return
	}
	_, err := io.CopyN(ioutil.Discard, r.Reader, int64(r.Length))
	if err != nil {
		panic(panicError{err})
	}
	r.Length = 0
}

// Used to debug a server message read.
func (r *reader) HexDump() string {
	buf := r.Bytea(r.Length)
	return hex.Dump(buf)
}

type writer struct {
	io.Writer
	buf      []byte
	Length   int
	LengthAt int
	StartAt  int
}

func (pg *connection) writer() *writer {
	return &writer{
		Writer: pg.conn,
		buf:    pg.scratch[:],
	}
}

func (w *writer) Msg(value byte) {
	if value != 0 {
		w.buf[w.StartAt] = value
		w.Length = w.StartAt + 5
		w.LengthAt = w.StartAt + 1
		return
	}
	w.Length = w.StartAt + 4
	w.LengthAt = w.StartAt

}
func (w *writer) Byte(value byte) {
	w.buf[w.Length] = value
	w.Length++
}
func (w *writer) Int8(value int8) {
	w.Byte(byte(value))
}

func (w *writer) Int16(value int16) {
	binary.BigEndian.PutUint16(w.buf[w.Length:], uint16(value))
	w.Length += 2
}
func (w *writer) Int32(value int32) {
	binary.BigEndian.PutUint32(w.buf[w.Length:], uint32(value))
	w.Length += 4
}
func (w *writer) String(value string) {
	copy(w.buf[w.Length:], value)
	w.Length += len(value)
	w.Byte(0)
}
func (w *writer) Bytea(value []byte) {
	copy(w.buf[w.Length:], value)
	w.Length += len(value)
}
func (w *writer) MsgDone() {
	msgLen := uint32(int32(w.Length - w.LengthAt))
	binary.BigEndian.PutUint32(w.buf[w.LengthAt:], msgLen)

	if debug {
		fmt.Printf("Client -> Server (0x%X)\n%s\n", msgLen, hex.Dump(w.buf[w.StartAt:w.Length]))
	}
	w.StartAt = w.Length
}

func (w *writer) Send() error {
	_, err := w.Write(w.buf[:w.Length])
	w.Length = 0
	w.StartAt = 0
	return err
}
