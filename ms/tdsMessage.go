// Copyright 2014 Daniel Theophanes.
// Use of this source code is governed by a zlib-style
// license that can be found in the LICENSE file.

package ms

import (
	"bytes"
	"context"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/kardianos/rdb/internal/pools/sync2"
	"github.com/kardianos/rdb/internal/sbuffer"
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
	packetRPC           PacketType = 3
	packetTabularResult PacketType = 4 // Server response.
	packetTransaction   PacketType = 0x0E

	packetAttention PacketType = 6
	packetBulkLoad  PacketType = 7

	packetTransactionManagerRequest PacketType = 14

	packetTDS7Login PacketType = 16
	packetSSPI      PacketType = 17
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

type connWriteDeadline interface {
	io.Writer
	SetWriteDeadline(t time.Time) error
}

type PacketWriter struct {
	w          connWriteDeadline
	PacketType PacketType

	buffer *bytes.Buffer

	packetNumber uint8
	resetPacket  bool
	open         bool

	single *sync2.Semaphore
}

func NewPacketWriter(w connWriteDeadline) *PacketWriter {
	return &PacketWriter{
		w:      w,
		buffer: &bytes.Buffer{},
		single: sync2.NewSemaphore(1),
	}
}

func (tds *PacketWriter) BeginMessage(ctx context.Context, PacketType PacketType, reset bool) error {
	err := tds.single.Acquire(ctx)
	if err != nil {
		return err
	}
	tds.buffer.Reset()

	tds.resetPacket = reset
	tds.PacketType = PacketType
	tds.packetNumber = 0
	tds.open = true
	return nil
}

func (tds *PacketWriter) Write(ctx context.Context, bb []byte) (n int, err error) {
	return tds.writeClose(ctx, bb, false)
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

func (tds *PacketWriter) EndMessage(ctx context.Context) error {
	if !tds.open {
		tds.single.Release()
		return nil
	}
	_, err := tds.writeClose(ctx, nil, true)
	return err
}

const writeContextCheckPeriod = time.Millisecond * 120

func (tds *PacketWriter) writeClose(ctx context.Context, bb []byte, closeMessage bool) (int, error) {
	var SPID uint16

	var n, localN int
	var err error

	bufN, _ := tds.buffer.Write(bb)

	for {
		status := statusNormal
		if tds.resetPacket {
			tds.resetPacket = false
			status |= statusResetConnection
		}

		l := maxPacketSizeBody
		if tds.buffer.Len() <= maxPacketSizeBody {
			if !closeMessage {
				return bufN, err
			}
			l = tds.buffer.Len()
			status |= statusEOM
			tds.open = false
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

		if debugProto {
			fmt.Println("Client -> Server")
			fmt.Println(hex.Dump(buf))
		}

		for {
			err = tds.w.SetWriteDeadline(time.Now().Add(writeContextCheckPeriod))
			if err != nil {
				tds.single.Release()
				return bufN, err
			}
			localN, err = tds.w.Write(buf)
			buf = buf[localN:]
			n += localN
			if err != nil {
				if errors.Is(err, os.ErrDeadlineExceeded) {
					err = nil
					if len(buf) == 0 {
						break
					}
					continue
				}
				tds.single.Release()
				return bufN, err
			}
			if len(buf) == 0 {
				break
			}
		}
		if statusEOM&status != 0 {
			tds.single.Release()
			return bufN, err
		}

	}
}

type PacketReader struct {
	buffer sbuffer.Buffer
}

func NewPacketReader(r sbuffer.ConnReadDeadline) *PacketReader {
	return &PacketReader{
		buffer: sbuffer.NewBuffer(r, maxPacketSize),
	}
}

func (tds *PacketReader) BeginMessage(_ context.Context, expectType PacketType) *MessageReader {
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
func (mr *MessageReader) Next(ctx context.Context) ([]byte, error) {
	buf := mr.packet.buffer
	if mr.length != 0 {
		buf.Used(mr.length)
		mr.length = 0
	}
	var err error
	var debugMessage []byte

	bb, err := buf.Next(ctx, 8)
	if err != nil {
		return nil, err
	}

	if bb[0] != byte(mr.msgType) {
		buf.Used(8)
		if debugProto {
			fmt.Println("Server -> Client (header)")
			debugMessage = append(debugMessage, bb...)
			fmt.Println(hex.Dump(debugMessage))
		}
		return nil, UnexpectedMessage{
			Expected: mr.msgType,
			Received: PacketType(bb[0]),
		}
	}
	packetEOM := false
	if MsgStatus(bb[1]) == statusEOM {
		packetEOM = true
	}
	if debugProto {
		debugMessage = make([]byte, 8)
		copy(debugMessage, bb)
	}
	mr.length = int(binary.BigEndian.Uint16(bb[2:])) - 8
	buf.Used(8)

	if mr.length > maxPacketSize {
		panic("packet length too large")
	}
	bb, err = buf.Next(ctx, mr.length)
	if debugProto {
		fmt.Println("Server -> Client")
		debugMessage = append(debugMessage, bb...)
		fmt.Println(hex.Dump(debugMessage))
	}
	if err != nil {
		return nil, fmt.Errorf("buf.Next: %w", err)
	}
	if packetEOM {
		err = io.EOF
	}
	return bb, err
}
func (mr *MessageReader) Close() error {
	if mr == nil || mr.packet == nil {
		return nil
	}
	buf := mr.packet.buffer
	if mr.length != 0 {
		buf.Used(mr.length)
		mr.length = 0
	}
	return nil
}

func (r *MessageReader) FetchAll(ctx context.Context) (ret []byte, err error) {
	_, err = r.Fetch(ctx, 0)
	if err != nil {
		return ret, err
	}
	return r.Fetch(ctx, r.length)
}

func (r *MessageReader) PeekByte(ctx context.Context) (out byte, err error) {
	if r == nil {
		return 0, io.EOF
	}

	const n = 1
	if len(r.current) >= n {
		return r.current[0], nil
	}
	if r.packetEOM {
		return 0, io.EOF
	}
	for len(r.current) < n {
		err = r.fill(ctx)
		if err != nil {
			return 0, err
		}
	}
	return r.current[0], nil
}

func (r *MessageReader) fill(ctx context.Context) error {
	if r.packetEOM {
		return io.ErrUnexpectedEOF
	}
	next, err := r.Next(ctx)
	if err != nil {
		if err != io.EOF {
			return err
		}
		r.packetEOM = true
	}
	// TODO(kardianos): find a way to make the bytes immutable, and normally avoid the copy.
	x := next
	next = make([]byte, len(x))
	copy(next, x)
	if len(r.current) == 0 {
		r.current = next
	} else {
		r.current = append(r.current, next...)
	}
	return nil
}
func (r *MessageReader) Fetch(ctx context.Context, n int) (ret []byte, err error) {
	if r == nil {
		return nil, io.EOF
	}
	if n == 0 {
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
	for len(r.current) < n {
		err = r.fill(ctx)
		if err != nil {
			return nil, err
		}
	}
	ret = r.current[:n:n]
	r.current = r.current[n:]
	return ret, nil
}
