// Copyright 2014 Daniel Theophanes.
// Use of this source code is governed by a zlib-style
// license that can be found in the LICENSE file.

package ms

import (
	"bytes"
	"context"
	"encoding/binary"
	"io"
	"testing"
	"time"

	"github.com/kardianos/rdb"
	"github.com/kardianos/rdb/internal/uconv"
	"github.com/kardianos/rdb/semver"
)

// deadlineNop wraps an io.Reader/Writer with no-op deadline methods for benches.
type deadlineNop struct {
	r io.Reader
	w io.Writer
}

func (d *deadlineNop) Read(p []byte) (int, error) {
	return d.r.Read(p)
}
func (d *deadlineNop) Write(p []byte) (int, error) {
	return d.w.Write(p)
}
func (d *deadlineNop) SetReadDeadline(t time.Time) error  { return nil }
func (d *deadlineNop) SetWriteDeadline(t time.Time) error { return nil }

// buildTDSPacket builds a single EOM TDS packet of the given type and body.
func buildTDSPacket(pt PacketType, body []byte) []byte {
	length := len(body) + 8
	buf := make([]byte, length)
	buf[0] = byte(pt)
	buf[1] = byte(statusEOM)
	binary.BigEndian.PutUint16(buf[2:], uint16(length))
	copy(buf[8:], body)
	return buf
}

// buildMultiPacket splits body into maxPacketSizeBody-sized TDS packets.
func buildMultiPacket(pt PacketType, body []byte) []byte {
	var out []byte
	packetNum := byte(0)
	for len(body) > 0 {
		chunk := body
		eom := true
		if len(chunk) > maxPacketSizeBody {
			chunk = body[:maxPacketSizeBody]
			body = body[maxPacketSizeBody:]
			eom = false
		} else {
			body = nil
		}
		length := len(chunk) + 8
		buf := make([]byte, length)
		buf[0] = byte(pt)
		if eom {
			buf[1] = byte(statusEOM)
		}
		binary.BigEndian.PutUint16(buf[2:], uint16(length))
		buf[6] = packetNum
		packetNum++
		copy(buf[8:], chunk)
		out = append(out, buf...)
	}
	return out
}

// syntheticRowStream builds COLMETADATA (int + nvarchar) + N rows + DONE final.
func syntheticRowStream(rows int, text string) []byte {
	var body bytes.Buffer

	body.WriteByte(byte(tokenColumnMetaData))
	binary.Write(&body, binary.LittleEndian, uint16(2))

	// Column 0: intN(4)
	binary.Write(&body, binary.LittleEndian, uint32(0))
	body.Write([]byte{0x01, 0x00})
	body.WriteByte(byte(typeIntN))
	body.WriteByte(4)
	idName := uconv.Encode.FromString("id")
	body.WriteByte(byte(len(idName) / 2))
	body.Write(idName)

	// Column 1: nvarchar
	binary.Write(&body, binary.LittleEndian, uint32(0))
	body.Write([]byte{0x01, 0x00})
	body.WriteByte(byte(typeNVarChar))
	binary.Write(&body, binary.LittleEndian, uint16(100))
	body.Write([]byte{0x09, 0x04, 0xD0, 0x00, 0x34})
	nameName := uconv.Encode.FromString("name")
	body.WriteByte(byte(len(nameName) / 2))
	body.Write(nameName)

	textU16 := uconv.Encode.FromString(text)
	for i := 0; i < rows; i++ {
		body.WriteByte(byte(tokenRow))
		body.WriteByte(4)
		binary.Write(&body, binary.LittleEndian, int32(i))
		binary.Write(&body, binary.LittleEndian, uint16(len(textU16)))
		body.Write(textU16)
	}

	body.WriteByte(byte(tokenDone))
	binary.Write(&body, binary.LittleEndian, uint16(0)) // final
	binary.Write(&body, binary.LittleEndian, uint16(0))
	binary.Write(&body, binary.LittleEndian, uint64(uint64(rows)))

	return buildMultiPacket(packetTabularResult, body.Bytes())
}

func BenchmarkUconvEncodeString(b *testing.B) {
	s := "SELECT id, name FROM users WHERE active = 1 AND region = N'US-WEST'"
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = uconv.Encode.FromString(s)
	}
}

func BenchmarkUconvEncodeBytes(b *testing.B) {
	s := []byte("Hello, 世界 — mixed ASCII and CJK text for encode cost.")
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = uconv.Encode.FromBytes(s)
	}
}

func BenchmarkUconvDecodeToBytes(b *testing.B) {
	in := uconv.Encode.FromString("Hello, 世界 — mixed ASCII and CJK text for decode cost.")
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = uconv.Decode.ToBytes(in)
	}
}

func BenchmarkUconvDecodeToString(b *testing.B) {
	in := uconv.Encode.FromString("Hello, 世界 — mixed ASCII and CJK text for decode cost.")
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = uconv.Decode.ToString(in)
	}
}

func BenchmarkPacketWriterSmallMessage(b *testing.B) {
	ctx := context.Background()
	sink := &bytes.Buffer{}
	w := NewPacketWriter(&deadlineNop{w: sink})
	payload := make([]byte, 200)
	for i := range payload {
		payload[i] = byte(i)
	}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sink.Reset()
		if err := w.BeginMessage(ctx, packetSqlBatch, false); err != nil {
			b.Fatal(err)
		}
		w.WriteUint32(22)
		w.WriteUint16(2)
		w.WriteUint64(0)
		w.WriteUint32(1)
		w.WriteBuffer(payload)
		if err := w.EndMessage(ctx); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkPacketWriterMultiPacket(b *testing.B) {
	ctx := context.Background()
	sink := &bytes.Buffer{}
	w := NewPacketWriter(&deadlineNop{w: sink})
	payload := make([]byte, maxPacketSizeBody*3+100)
	for i := range payload {
		payload[i] = byte(i)
	}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sink.Reset()
		if err := w.BeginMessage(ctx, packetSqlBatch, false); err != nil {
			b.Fatal(err)
		}
		if _, err := w.Write(ctx, payload); err != nil {
			b.Fatal(err)
		}
		if err := w.EndMessage(ctx); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkMessageReaderFetch(b *testing.B) {
	body := make([]byte, 1024)
	for i := range body {
		body[i] = byte(i)
	}
	packet := buildTDSPacket(packetTabularResult, body)
	ctx := context.Background()

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r := bytes.NewReader(packet)
		pr := NewPacketReader(&deadlineNop{r: r})
		mr := pr.BeginMessage(ctx, packetTabularResult)
		for {
			bb, err := mr.Fetch(ctx, 64)
			if err != nil && err != io.EOF {
				if len(bb) == 0 {
					b.Fatal(err)
				}
			}
			if len(bb) == 0 {
				break
			}
		}
		mr.Close()
	}
}

func BenchmarkMessageReaderMultiPacket(b *testing.B) {
	body := make([]byte, maxPacketSizeBody*2+500)
	for i := range body {
		body[i] = byte(i)
	}
	stream := buildMultiPacket(packetTabularResult, body)
	ctx := context.Background()

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r := bytes.NewReader(stream)
		pr := NewPacketReader(&deadlineNop{r: r})
		mr := pr.BeginMessage(ctx, packetTabularResult)
		total := 0
		for total < len(body) {
			need := 128
			if need > len(body)-total {
				need = len(body) - total
			}
			bb, err := mr.Fetch(ctx, need)
			if err != nil && err != io.EOF {
				b.Fatal(err)
			}
			total += len(bb)
			if len(bb) == 0 {
				break
			}
		}
		mr.Close()
	}
}

// discardValuer counts rows/fields without retaining values.
type discardValuer struct {
	rows, fields int
}

func (v *discardValuer) Columns([]*rdb.Column) error { return nil }
func (v *discardValuer) Done() error                 { return nil }
func (v *discardValuer) RowScanned()                 { v.rows++ }
func (v *discardValuer) Message(*rdb.Message)        {}
func (v *discardValuer) WriteField(c *rdb.Column, value *rdb.DriverValue, assign rdb.Assigner) error {
	v.fields++
	return nil
}
func (v *discardValuer) RowsAffected(count uint64) {}

func BenchmarkTokenStreamDecodeRows(b *testing.B) {
	const nRows = 50
	stream := syntheticRowStream(nRows, "hello world row text")
	ctx := context.Background()

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r := bytes.NewReader(stream)
		pr := NewPacketReader(&deadlineNop{r: r})
		mr := pr.BeginMessage(ctx, packetTabularResult)
		conn := &Connection{pr: pr, mr: mr}
		val := &discardValuer{}
		conn.val = val

		for {
			res, err := conn.getSingleResponse(ctx, mr, true)
			if err != nil {
				b.Fatal(err)
			}
			switch res.(type) {
			case MsgEom, MsgFinalDone:
				goto done
			case MsgRow:
				val.RowScanned()
			}
		}
	done:
		mr.Close()
		if val.rows != nRows {
			b.Fatalf("rows=%d want %d", val.rows, nRows)
		}
	}
}

func BenchmarkEncodeParamString(b *testing.B) {
	ctx := context.Background()
	sink := &bytes.Buffer{}
	w := NewPacketWriter(&deadlineNop{w: sink})
	ver := &semver.Version{Major: 7, Minor: 4, Product: "TDS", InHex: true}
	param := &rdb.Param{Name: "Name", Type: rdb.TypeVarChar, Length: 100, Value: "Alice"}
	collation := [5]byte{0x09, 0x04, 0xD0, 0x00, 0x34}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sink.Reset()
		if err := w.BeginMessage(ctx, packetRPC, false); err != nil {
			b.Fatal(err)
		}
		if err := encodeParam(ctx, w, false, ver, param, param.Value, collation); err != nil {
			b.Fatal(err)
		}
		if err := w.EndMessage(ctx); err != nil {
			b.Fatal(err)
		}
	}
}
