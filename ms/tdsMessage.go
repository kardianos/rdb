// Copyright 2014 Daniel Theophanes.
// Use of this source code is governed by a zlib-style
// license that can be found in the LICENSE file.

package ms

import (
	"bytes"
	"encoding/binary"
	"io"

	"bitbucket.org/kardianos/rsync/sbuffer"

	"encoding/hex"
	"fmt"
)

const (
	debug = false
)
const (
	// TDS Versions. There are older versions but anything before 7.2 is junk.
	// Many fields were expanded on 7.2 so it is simpler to ignore very old versions.
	version72  = 0x02000972 // SQL Server 2005
	version73A = 0x03000A73 // SQL Server 2008
	version73B = 0x03000B73 // SQL Server 2008 R2
	version74  = 0x04000074 // SQL Server 2012 & SQL Server 2014
)

type PacketType byte

const (
	packetSqlBatch      PacketType = 1
	packetOldLogin      PacketType = 2
	packetRpc           PacketType = 3
	packetTabularResult PacketType = 4 // Server response.
	packetTransaction   PacketType = 0x0E

	packetAttention PacketType = 6
	packetBulkLoad  PacketType = 7

	packetTransactionManagerRequest PacketType = 14

	packetTds7Login PacketType = 16
	packetSspi      PacketType = 17
	packetPreLogin  PacketType = 18
)

type MsgStatus byte

// Bit field values.
const (
	statusNormal                  MsgStatus = 0x00
	statusEOM                     MsgStatus = 0x01
	statusIgnore                  MsgStatus = 0x02
	statusResetConnection         MsgStatus = 0x08
	statusResetConnectionSkipTran MsgStatus = 0x10
)

const (
	maxPacketSize     = 1024 * 4
	maxPacketSizeBody = maxPacketSize - 8
)

type PacketWriter struct {
	w          io.Writer
	PacketType PacketType

	buffer *bytes.Buffer

	packetNumber uint8
}

func NewPacketWriter(w io.Writer) *PacketWriter {
	return &PacketWriter{
		w:      w,
		buffer: &bytes.Buffer{},
	}
}

func (tds *PacketWriter) WriteMessage(PacketType PacketType, bb []byte) (n int, err error) {
	tds.BeginMessage(PacketType)
	return tds.writeClose(bb, true)
}

func (tds *PacketWriter) BeginMessage(PacketType PacketType) error {
	tds.buffer.Reset()

	tds.PacketType = PacketType
	tds.packetNumber = 0
	return nil
}

func (tds *PacketWriter) Write(bb []byte) (n int, err error) {
	return tds.writeClose(bb, false)
}

func (tds *PacketWriter) WriteBuffer(v []byte) (n int) {
	n, _ = tds.buffer.Write(v)
	return
}
func (tds *PacketWriter) WriteByte(v byte) (n int) {
	tds.buffer.WriteByte(v)
	return 1
}
func (tds *PacketWriter) WriteUint16(v uint16) (n int) {
	bb := make([]byte, 2)
	binary.LittleEndian.PutUint16(bb, v)
	tds.buffer.Write(bb)
	return 2
}
func (tds *PacketWriter) WriteUint32(v uint32) (n int) {
	bb := make([]byte, 4)
	binary.LittleEndian.PutUint32(bb, v)
	tds.buffer.Write(bb)
	return 4
}
func (tds *PacketWriter) WriteUint64(v uint64) (n int) {
	bb := make([]byte, 8)
	binary.LittleEndian.PutUint64(bb, v)
	tds.buffer.Write(bb)
	return 8
}

func (tds *PacketWriter) EndMessage() error {
	_, err := tds.writeClose(nil, true)
	return err
}

func (tds *PacketWriter) writeClose(bb []byte, closeMessage bool) (int, error) {
	var SPID uint16

	var n, localN int
	var err error

	tds.buffer.Write(bb)

	for {
		status := statusNormal

		l := maxPacketSizeBody
		if tds.buffer.Len() <= maxPacketSizeBody {
			if !closeMessage {
				return n, err
			}
			l = tds.buffer.Len()
			status = statusEOM
		}

		length := l + 8 // Header is 8 bytes.

		buf := make([]byte, length)

		// Write packet to writer.
		// MsgType - uint8
		buf[0] = byte(tds.PacketType)

		// MsgStatus - uint8
		buf[1] = byte(status)

		// Length - uint16, include all headers and entire length.
		binary.BigEndian.PutUint16(buf[2:], uint16(length))

		// SPID - uint16, either send server ID or zero.
		binary.BigEndian.PutUint16(buf[4:], SPID)

		// PacketID - uint8, increment each time it is sent, allow overflow. for a given message.
		buf[6] = tds.packetNumber
		tds.packetNumber++

		// Window - uint8, should be zero. Ignored.
		buf[7] = 0

		// PacketData
		copy(buf[8:], tds.buffer.Next(l))

		if debug {
			fmt.Println("Client -> Server")
			fmt.Println(hex.Dump(buf))
		}
		localN, err = tds.w.Write(buf)
		if err != nil {
			return n, err
		}
		n += localN
		if status == statusEOM {
			return n, err
		}

	}
}

type PacketReader struct {
	buffer sbuffer.Buffer
}

func NewPacketReader(r io.Reader) *PacketReader {
	return &PacketReader{
		buffer: sbuffer.NewBuffer(r, maxPacketSize*12),
	}
}

func (tds *PacketReader) BeginMessage(expectType PacketType) *MessageReader {
	return &MessageReader{
		packet:  tds,
		msgType: expectType,
	}
}

type MessageReader struct {
	packet  *PacketReader
	msgType PacketType
	length  int

	// For fetch.
	current   []byte
	packetEOM bool
}

// Read another packet.
func (mr *MessageReader) Next() ([]byte, error) {
	buf := mr.packet.buffer
	if mr.length != 0 {
		buf.Used(mr.length)
		mr.length = 0
	}
	var err error
	var debugMessage []byte

	bb, err := buf.Next(8)
	if err != nil {
		return nil, err
	}

	if bb[0] != byte(mr.msgType) {
		buf.Used(8)
		return nil, UnexpectedMessage{
			Expected: mr.msgType,
			Recieved: PacketType(bb[0]),
		}
	}
	packetEOM := false
	if MsgStatus(bb[1]) == statusEOM {
		packetEOM = true
	}
	if debug {
		debugMessage = make([]byte, 8)
		copy(debugMessage, bb)
	}
	mr.length = int(binary.BigEndian.Uint16(bb[2:])) - 8
	buf.Used(8)

	bb, err = buf.Next(mr.length)
	if debug {
		fmt.Println("Server -> Client")
		debugMessage = append(debugMessage, bb...)
		fmt.Println(hex.Dump(debugMessage))
	}
	if err != nil {
		return nil, err
	}
	if packetEOM {
		err = io.EOF
	}
	return bb, err
}
func (mr *MessageReader) Close() error {
	buf := mr.packet.buffer
	if mr.length != 0 {
		buf.Used(mr.length)
		mr.length = 0
	}
	return nil
}

func (r *MessageReader) Fetch(n int) (ret []byte, err error) {
	if n == 0 {
		if r.packetEOM && len(r.current) == 0 {
			return nil, io.EOF
		}
		return nil, nil
	}
	if len(r.current) >= n {
		ret = r.current[:n:n]
		r.current = r.current[n:]
		return ret, nil
	}
	if r.packetEOM {
		return nil, io.EOF
	}
	var next []byte
	for len(r.current) < n {
		if r.packetEOM {
			return r.current, io.ErrUnexpectedEOF
		}
		next, err = r.Next()
		if err != nil {
			if err != io.EOF {
				return nil, err
			}
			r.packetEOM = true
		}
		if len(r.current) == 0 {
			r.current = next
		} else {
			r.current = append(r.current, next...)
		}
	}
	ret = r.current[:n:n]
	r.current = r.current[n:]
	return ret, nil
}
