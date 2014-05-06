// Copyright 2014 Daniel Theophanes.
// Use of this source code is governed by a zlib-style
// license that can be found in the LICENSE file.

package ms

import (
	"bitbucket.org/kardianos/rdb"
	"bitbucket.org/kardianos/rdb/ms/uconv"
	"bitbucket.org/kardianos/rdb/semver"
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
)

type Connection struct {
	pw *PacketWriter
	pr *PacketReader

	wc io.ReadWriteCloser

	open  bool
	inUse bool

	ProductVersion  *semver.Version
	ProtocolVersion *semver.Version
}

func NewConnection(c io.ReadWriteCloser) *Connection {
	return &Connection{
		pw: NewPacketWriter(c),
		pr: NewPacketReader(c),
		wc: c,
	}
}

func (tds *Connection) Open(config *rdb.Config) (*ServerInfo, error) {
	if tds.open {
		return nil, connectionOpenError
	}
	var err error

	err = tds.pw.PreLogin(config.Instance)
	if err != nil {
		return nil, err
	}

	_, err = tds.pr.Prelogin()
	if err != nil {
		return nil, err
	}

	// Write LOGIN7 message.
	err = tds.pw.Login(config)
	if err != nil {
		return nil, err
	}

	si, err := tds.pr.LoginAck()
	if err != nil {
		return nil, err
	}
	tds.ProductVersion = &semver.Version{
		Major:   uint16(si.MajorVersion),
		Minor:   uint16(si.MinorVersion),
		Patch:   si.BuildNumber,
		Product: si.ProgramName,
	}
	tds.ProtocolVersion = &semver.Version{
		Major:   uint16(si.TdsVersion[3]),
		Minor:   uint16(si.TdsVersion[0]),
		Patch:   uint16(si.TdsVersion[1]),
		Product: "TDS",
		InHex:   true,
	}

	tds.open = true

	return si, err
}

func (tds *Connection) Close() error {
	err := tds.wc.Close()
	tds.open = false
	return err
}

func (tds *Connection) Execute(sql string, truncValue bool, arity rdb.Arity, params []*rdb.Param, values []rdb.Value, fields []*rdb.Field) (*Result, error) {
	if !tds.open {
		return nil, connectionNotOpenError
	}
	if tds.inUse {
		return nil, connectionInUseError
	}
	tds.inUse = true

	if len(values) != 0 {
		pm := make(map[string]*rdb.Param, len(params))
		for _, param := range params {
			pm[param.N] = param
		}
		for i := range values {
			value := &values[i]
			var ok bool
			if len(value.N) == 0 {
				if i >= len(params) {
					return nil, rdb.ErrorColumnNotFound{At: "Map values to parameters", Index: i}
				}
				value.Param = params[i]
			} else {
				value.Param, ok = pm[value.N]
				if !ok {
					return nil, rdb.ErrorColumnNotFound{At: "Map values to parameters", Name: value.N}
				}
			}
		}
	}

	err := tds.sendRpc(sql, truncValue, params, values)
	if err != nil {
		return nil, err
	}
	result := &Result{
		initFields: fields,
		arity:      arity,
	}

	r := tds.pr
	result.tds = tds
	result.mr = r.BeginMessage(packetTabularResult)

	return result, result.Process(false)
}

const (
	sp_ExecuteSql = 10
)

var rpcHeaderParam = &rdb.Param{
	T: rdb.TypeString,
	L: 0,
}

func (tds *Connection) sendRpc(sql string, truncValue bool, params []*rdb.Param, values []rdb.Value) error {
	// To make a SQL Query with params:
	// * RPC Param 1 = {Name: "", Type: NText, Field: SqlQuery}
	// * RPC Param 2 = {Name: "", Type: NText, Field: "@MySqlParam1 int,@Foo varchar(400)"}
	// * RPC Param 3 = {Name: "@MySqlParam1", Type: Int, Field: value}
	// * RPC Param 4 = {Name: "@Foo", Type: VarChar, Field: value}
	// Simple! Once figured out.

	var procID uint16 = sp_ExecuteSql
	withRecomp := false
	// collation := []byte{0x09, 0x04, 0xD0, 0x00, 0x34}

	w := tds.pw
	err := w.BeginMessage(packetRpc)
	if err != nil {
		return err
	}

	var options uint16 = 0
	if withRecomp {
		options = 1
	}
	/*
		ParameterData is repeated once for each parameter in the request.

		Stream Definition:

		RPCRequest =
			ALL_HEADERS
			(
				(
					US_VARCHAR ProcName
					OR (
						%xFF %xFF
						USHORT ProcID
					)
				) NameLenProcID
				(
					BIT fWithRecomp
					BIT fNoMetaData
					BIT fReuseMetaData
					13 BIT 13FRESERVEDBIT
				) OptionFlags
				*(
					(
						B_VARCHAR
						(
							BIT fByRefValue
							BIT fDefaultValue
							6 BIT 6FRESERVEDBIT
						)StatusFlags
						TYPE_INFO
					) ParamMetaData
					TYPE_VARBYTE ParamLenData
				)ParameterData
			) RPCReqBatch
			*(
				%xFF BatchFlag
				RPCReqBatch
			)
			[%xFF BatchFlag]
	*/

	w.WriteBuffer(sqlRequestHeader)
	w.WriteUint16(0xffff) // ProcIDSwitch
	w.WriteUint16(procID)
	w.WriteUint16(options) // 16 bits (2 bytes) - Options: fWithRecomp, fNoMetaData, fReuseMetaData, 13FRESERVEDBIT

	paramNames := &bytes.Buffer{}
	for i, param := range params {
		if i != 0 {
			paramNames.WriteRune(',')
		}
		if len(param.N) == 0 {
			return fmt.Errorf("Missing parameter name at index: %d", i)
		}

		st, found := sqlTypeLookup[param.T]
		if !found {
			panic(fmt.Sprintf("SqlType not found: %d", param.T))
		}
		fmt.Fprintf(paramNames, "@%s %s", param.N, st.TypeString(param))
	}
	err = encodeParam(w, truncValue, tds.ProtocolVersion, rpcHeaderParam, []byte(sql))
	if err != nil {
		return err
	}
	err = encodeParam(w, truncValue, tds.ProtocolVersion, rpcHeaderParam, paramNames.Bytes())
	if err != nil {
		return err
	}

	// Other parameters.
	if len(values) == 0 {
		for _, param := range params {
			err = encodeParam(w, truncValue, tds.ProtocolVersion, param, param.V)
			if err != nil {
				return err
			}
		}
	} else {
		for _, value := range values {
			err = encodeParam(w, truncValue, tds.ProtocolVersion, value.Param, value.V)
			if err != nil {
				return err
			}
		}
	}
	w.WriteByte(0xFF)

	err = w.EndMessage()
	if err != nil {
		return err
	}
	return nil
}

func (tds *Connection) getSingleResponse(m *MessageReader, result *Result) (response interface{}, err error) {
	var bb []byte

	defer func() {
		if recovered := recover(); recovered != nil {
			if re, is := recovered.(recoverError); is {
				if re.err == io.EOF {
					return
				}
				err = re.err
				return
			}
			panic(recovered)
		}
	}()
	read := func(n int) []byte {
		var readErr error
		bb, readErr = m.Fetch(n)
		if readErr != nil {
			panic(recoverError{err: readErr})
		}
		return bb
	}

	switch read(1)[0] {
	case tokenError:
		sqlErr := &rdb.SqlError{}
		_ = binary.LittleEndian.Uint16(read(2)) // length
		sqlErr.Number = int32(binary.LittleEndian.Uint32(read(4)))
		state := read(1)[0]
		class := read(1)[0]

		_, msg := uconv.Decode.Prefix2(read)
		sqlErr.Message = fmt.Sprintf("%s (%d, %d)", msg, state, class)
		_, sqlErr.ServerName = uconv.Decode.Prefix1(read)
		_, sqlErr.ProcName = uconv.Decode.Prefix1(read)
		sqlErr.LineNumber = int32(binary.LittleEndian.Uint32(read(4)))

		return sqlErr, nil
	case tokenColumnMetaData:
		bb = read(2)
		if bb[0] == 0xff && bb[1] == 0xff {
			return []*SqlColumn{}, nil
		}
		{
			var columns []*SqlColumn
			count := int(binary.LittleEndian.Uint16(bb))
			for i := 0; i < count; i++ {
				column := decodeColumnInfo(read)
				column.Index = i
				columns = append(columns, column)
			}

			return columns, nil
		}
	case tokenReturnStatus:
		return SqlRpcResult(binary.LittleEndian.Uint32(read(4))), nil
	case tokenDoneProc:
		fallthrough
	case tokenDoneInProc:
		fallthrough
	case tokenDone:
		return &SqlDone{
			StatusCode: binary.LittleEndian.Uint16(read(2)),
			CurrentCmd: binary.LittleEndian.Uint16(read(2)),
			Rows:       binary.LittleEndian.Uint64(read(8)),
		}, nil
	case tokenRow:
		for _, column := range result.Columns {
			decodeFieldValue(read, column, result)
		}

		return &SqlRow{}, nil
	default:
		return nil, fmt.Errorf("Unknown response code: 0x%X", bb[0])
	}
}

var sqlRequestHeader = func() []byte {
	/*
		type ALL_HEADER struct {
			TotalLength uint32 // Includes length.
			Headers     []struct {
				Length uint32 // Includes length.
				Type   uint16
				Data   []byte
			}
		}
		Transaction Description: {
			Type = 0x0002
			Data = struct {
				TransactionDescriptor   uint64 // =0
				OutstandingRequestCount uint32 // =1
			}
		}

	*/
	length := 4 + (4 + 2 + (4 + 8))
	bb := make([]byte, length)

	at := 0
	binary.LittleEndian.PutUint32(bb[at:], uint32(length))
	at += 4

	binary.LittleEndian.PutUint32(bb[at:], uint32(length)-4)
	at += 4

	binary.LittleEndian.PutUint16(bb[at:], 0x0002)
	at += 2

	binary.LittleEndian.PutUint64(bb[at:], 0)
	at += 8

	binary.LittleEndian.PutUint32(bb[at:], 1)
	at += 4

	return bb
}()
