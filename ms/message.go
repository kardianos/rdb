// Copyright 2014 Daniel Theophanes.
// Use of this source code is governed by a zlib-style
// license that can be found in the LICENSE file.

package ms

import (
	"context"
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"os"
	"time"

	"github.com/kardianos/rdb"
	"github.com/kardianos/rdb/internal/uconv"

	"errors"
)

const (
	preloginVersion    = 0x00
	preloginEncryption = 0x01
	preloginInstance   = 0x02
	preloginMars       = 0x04
	preloginTerminator = 0xff
)

type tdsToken byte

//go:generate stringer -type tdsToken -trimprefix token

const (
	tokenLoginAck tdsToken = 0xAD
	tokenError    tdsToken = 0xAA
	tokenInfo     tdsToken = 0xAB
	tokenDone     tdsToken = 0xFD

	tokenReturnStatus   tdsToken = 0x79
	tokenReturnValue    tdsToken = 0xAC
	tokenDoneProc       tdsToken = 0xFE
	tokenDoneInProc     tdsToken = 0xFF
	tokenColumnMetaData tdsToken = 0x81
	tokenRow            tdsToken = 0xD1
	tokenNBCRow         tdsToken = 0xD2
	tokenEnvChange      tdsToken = 0xE3

	tokenOrder tdsToken = 0xA9
)

// Document the highest version this driver can handle.
const protoVersionMax = version74

type EncryptAvailable byte

const (
	encryptOff          EncryptAvailable = 0 // Encryption is available but off.
	encryptOn           EncryptAvailable = 1 // Encryption is available and on.
	encryptNotSupported EncryptAvailable = 2 // Encryption is not available.
	encryptRequired     EncryptAvailable = 3 // Encryption is required.
)

// Pre-Login
func (tds *PacketWriter) PreLogin(ctx context.Context, instance string, encrypt EncryptAvailable) error {
	var err error
	type option struct {
		t byte
		d []byte
	}

	opts := make([]option, 0)

	addToken := func(tokenType byte, data []byte) {
		opts = append(opts, option{
			t: tokenType,
			d: data,
		})
	}

	version := make([]byte, 6)
	binary.BigEndian.PutUint32(version, protoVersionMax)

	addToken(preloginVersion, version)
	addToken(preloginMars, []byte{0x00})                // MARS OFF (0x01 is ON).
	addToken(preloginEncryption, []byte{byte(encrypt)}) // Encription not available. Pg 65.
	if len(instance) > 0 {
		addToken(preloginInstance, uconv.Encode.FromString(instance))
	}

	tds.BeginMessage(ctx, packetPreLogin, false)

	tokenListLen := uint16((5 * len(opts)) + 1)
	payload := make([]byte, 0, 20)

	token := make([]byte, 5)
	for _, option := range opts {
		token[0] = option.t                                                      // Type.
		binary.BigEndian.PutUint16(token[1:], tokenListLen+uint16(len(payload))) // Offset.
		binary.BigEndian.PutUint16(token[3:], uint16(len(option.d)))             // Length.
		payload = append(payload, option.d...)
		_, err = tds.Write(ctx, token)
		if err != nil {
			return err
		}
	}
	_, err = tds.Write(ctx, []byte{0xff})
	if err != nil {
		return err
	}

	_, err = tds.Write(ctx, payload)
	if err != nil {
		return err
	}

	err = tds.EndMessage(ctx)
	if err != nil {
		return err
	}

	return nil
}

// Rturned from Pre-Login.
type ServerConnection struct {
	Version    [6]byte
	Encryption EncryptAvailable
	Instance   string
	MARS       bool
}

// Returned from Login.
type ServerInfo struct {
	AcceptTSql  bool
	TdsVersion  [4]byte
	ProgramName string

	MajorVersion byte
	MinorVersion byte
	BuildNumber  uint16
}

func (si *ServerInfo) String() string {
	return fmt.Sprintf("%s %d.%d.%d", si.ProgramName, si.MajorVersion, si.MinorVersion, si.BuildNumber)
}

func (tds *PacketReader) Prelogin(ctx context.Context) (*ServerConnection, error) {
	read := tds.BeginMessage(ctx, packetTabularResult)

	bb, err := read.Next(ctx)
	if err != nil && err != io.EOF {
		return nil, err
	}
	defer read.Close()

	type option struct {
		t      byte
		offset uint16
		length uint16
		d      []byte
	}
	var ops = make([]*option, 0, 2)

	at := 0

	for {
		if at >= len(bb) {
			break
		}
		t := bb[at]
		if t == preloginTerminator {
			break
		}
		if at+4 >= len(bb) {
			break
		}
		ops = append(ops, &option{
			t:      bb[at],
			offset: binary.BigEndian.Uint16(bb[at+1:]),
			length: binary.BigEndian.Uint16(bb[at+3:]),
		})
		at += 5
	}
	si := &ServerConnection{}
	for _, o := range ops {
		o.d = make([]byte, o.length)
		copy(o.d, bb[o.offset:])
		switch o.t {
		case 0x00:
			copy(si.Version[:], o.d[:6])
		case 0x01:
			si.Encryption = EncryptAvailable(o.d[0])
		case 0x02:
			si.Instance = uconv.Decode.ToString(o.d)
		case 0x03:
			// Thread ID.
		case 0x04:
			if o.d[0] != 0 {
				si.MARS = true
			}
		default:
			// Ignore.
		}
	}

	return si, nil
}

// Write LOGIN7. Page 53.
func (tds *PacketWriter) Login(ctx context.Context, config *rdb.Config) error {
	var err error
	/*
		Versions:
		Length uint32
		TDSVersion uint32
		PacketSize uint32
		ClientProgVer uint32
		ClientPID uint32
		ConnectionID uint32

		OptionFlags1 byte
		OptionFlags2 byte
		TypeFlags byte
		(FRESERVEDBYTE / OptionFlags3) byte
		ClientTimZone int32
		ClientLCID [4]byte
		OffsetLength
		Data
		[FeatureExt]

		OffsetLength is a list of [{MessageOffset, ValueLength uint16}] with a few exceptions.
			0 HostName
			1 UserName
			2 Password
			3 AppName
			4 ServerName
			5 Unused
			6 Extension
			7 CltIntName - Interface Library Name
			8 Language
			9 Database
			10 ClientID : [6]byte
			11 SSPI
			12 AtchDBFile
			13 ChangePassword
			14 SSPILong : uint32, will replace SSPI Length if SSPI == 0xffff.
	*/

	SSPI := []byte{}
	ClientID := [6]byte{}

	iface, err := net.Interfaces()
	if err != nil && len(iface) > 0 {
		copy(ClientID[:], []byte(iface[0].HardwareAddr))
	}

	partALen := 9 * 4        // Message length up to OffsetLength section.
	partBLen := 12*4 + 4 + 6 // OffsetLength section.

	at := partALen + partBLen

	type token struct {
		raw    bool
		offset uint16
		length uint16
		data   []byte
	}

	tt := make([]token, 14)

	writeToken := func(index int, data []byte, str bool) {
		l := uint16(len(data))
		if str {
			l = l / 2
		}
		tt[index].offset = uint16(at)
		tt[index].length = l
		tt[index].data = data
		at += len(data)
	}

	// TODO: Check max lengths, truncate if too long.
	writeToken(0, uconv.Encode.FromString(config.Hostname), true)
	writeToken(1, uconv.Encode.FromString(config.Username), true)

	passwordBytes := uconv.Encode.FromString(config.Password)
	for i, b := range passwordBytes {
		passwordBytes[i] = ((b << 4) | (b >> 4)) ^ 0xA5
	}
	writeToken(2, passwordBytes, true) // The password is obfuscated here.

	writeToken(3, uconv.Encode.FromString(""), true) // AppName - Name of the client application.
	writeToken(4, uconv.Encode.FromString(config.Instance), true)
	// 5 - Unused.
	// 6 - Library Name.
	// 7 - Language.
	writeToken(8, uconv.Encode.FromString(config.Database), true)

	tt[9].raw = true
	tt[9].data = ClientID[:]

	// Make sure SSPI tokens are encoded last.
	if len(SSPI) > 0 {
		tt[10].length = 0xffff
		tt[10].offset = uint16(at)
		tt[10].data = SSPI
	}
	// 11 - Attach DB.
	// 12 - Change Password.

	tt[13].raw = true
	tt[13].data = make([]byte, 4)
	binary.LittleEndian.PutUint32(tt[13].data, uint32(len(SSPI)))
	at += len(SSPI)

	buf := make([]byte, at)

	binary.LittleEndian.PutUint32(buf[0:], uint32(at))           // Total length.
	binary.BigEndian.PutUint32(buf[4:], protoVersionMax)         // TDSVersion.
	binary.LittleEndian.PutUint32(buf[8:], maxPacketSize)        // PacketSize.
	binary.LittleEndian.PutUint32(buf[12:], 4176642822)          // ClientProgVer.
	binary.LittleEndian.PutUint32(buf[16:], uint32(os.Getpid())) // ClientPID.
	binary.LittleEndian.PutUint32(buf[20:], 0)                   // ConnectionID.

	buf[24] = 0 // OptionFlags1.
	buf[25] = 0 // OptionFlags2.
	buf[26] = 1 // TypeFlags. Flip first bit to use TSQL.
	buf[27] = 0 // OptionFlags3.

	_, zone := time.Now().Zone()
	binary.LittleEndian.PutUint32(buf[28:], uint32(zone/3600)) // ClientTimZone.
	binary.LittleEndian.PutUint32(buf[32:], 1033)              // ClientLCID - Language code identifier.

	at = partALen

	prevOffset := tt[0].offset
	for _, t := range tt {
		if t.raw {
			copy(buf[at:], t.data)
			at += len(t.data)
			continue
		}
		if t.offset == 0 {
			t.offset = prevOffset
		}
		binary.LittleEndian.PutUint16(buf[at:], t.offset)
		at += 2
		binary.LittleEndian.PutUint16(buf[at:], t.length)
		at += 2
		if t.length > 0 {
			copy(buf[t.offset:], t.data)
		}

		prevOffset = t.offset
	}

	tds.BeginMessage(ctx, packetTDS7Login, false)

	_, err = tds.Write(ctx, buf)
	if err != nil {
		return err
	}

	err = tds.EndMessage(ctx)
	if err != nil {
		return err
	}

	return nil
}

func (tds *PacketReader) LoginAck(ctx context.Context) (*ServerInfo, error) {
	// Page 95.
	read := tds.BeginMessage(ctx, packetTabularResult)

	bb, err := read.Next(ctx)
	if err != nil && err != io.EOF {
		return nil, fmt.Errorf("login ack next: %w", err)
	}
	defer read.Close()
	if len(bb) == 0 {
		return nil, errors.New("unable to authenticate to server or database")
	}

	at := 0
	token := tdsToken(bb[at])
	at++
	if token != tokenLoginAck {
		if token == tokenError {
			tp := rdb.SqlError
			sqlMsg := &rdb.Message{
				Type: tp,
			}
			// tokenLen := int(binary.LittleEndian.Uint16(bb[at:])) // length
			at += 2
			sqlMsg.Number = int32(binary.LittleEndian.Uint32(bb[at:]))
			at += 4
			state := bb[at]
			at++
			class := bb[at]
			at++

			msgLen := int(binary.LittleEndian.Uint16(bb[at:])) * 2
			at += 2
			msg := uconv.Decode.ToString(bb[at : at+msgLen])
			at += msgLen
			sqlMsg.Message = fmt.Sprintf("%s (%d, %d)", msg, state, class)

			strLen := int(bb[at]) * 2
			at++
			sqlMsg.ServerName = uconv.Decode.ToString(bb[at : at+strLen])
			at += strLen

			strLen = int(bb[at]) * 2
			at++
			sqlMsg.ProcName = uconv.Decode.ToString(bb[at : at+strLen])
			at += strLen

			sqlMsg.LineNumber = int32(binary.LittleEndian.Uint32(bb[at:]))
			at += 4
			return nil, rdb.Errors{sqlMsg}
		}
		return nil, fmt.Errorf("expected type %X but got %X", tokenLoginAck, bb[at])
	}

	si := &ServerInfo{}

	// The little endian uint16 length of the following fields. Ignore.
	at += 2

	si.AcceptTSql = false
	if bb[at] == 1 {
		si.AcceptTSql = true
	}
	at += 1

	copy(si.TdsVersion[:], bb[at:])
	at += 4

	// Byte length prefix string.
	programNameLen := int(bb[at]) * 2
	at += 1
	si.ProgramName = uconv.Decode.ToString(bb[at : at+programNameLen-4]) // Remove trailing nulls.
	at += programNameLen

	si.MajorVersion = bb[at]
	at += 1

	si.MinorVersion = bb[at]
	at += 1

	si.BuildNumber = binary.BigEndian.Uint16(bb[at:])

	return si, nil
}
