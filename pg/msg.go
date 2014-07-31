// Copyright 2014 Daniel Theophanes.
// Use of this source code is governed by a zlib-style
// license that can be found in the rdb LICENSE file.

package pg

import (
	"fmt"

	"time"

	"bitbucket.org/kardianos/rdb"
	"bitbucket.org/kardianos/rdb/semver"
)

func (pg *connection) getMessage() (_ interface{}, err error) {
	defer func() {
		if recovered := recover(); recovered != nil {
			if pe, is := recovered.(panicError); is {
				err = pe.err
				return
			}
			panic(recovered)
		}
	}()

	read := pg.reader()
	msgToken := read.Msg()

	var value interface{}

	if debug {
		defer func() {
			fmt.Printf("REC MSG: %T\n", value)
		}()
	}
	// Backend messages.
	switch msgToken {
	case tokenAuthenticationResponse:
		subType := read.Int32()
		switch subType {
		case 0:
			value = MsgAuthenticationOk{}
		case 2:
			value = MsgAuthenticationKerberosV5{}
		case 3:
			value = MsgAuthenticationCleartextPassword{}
		case 5:
			salt := read.Bytea(4)
			saltB := make([]byte, 4)
			copy(saltB, salt)
			value = MsgAuthenticationMD5Password{
				Salt: saltB,
			}
		case 6:
			value = MsgAuthenticationSCMCredential{}
		case 7:
			value = MsgAuthenticationGSS{}
		case 9:
			value = MsgAuthenticationSSPI{}
		case 8:
			value = MsgAuthenticationGSSContinue{
				Data: read.Bytea(read.Length),
			}
		}
	case tokenBackendKeyData:
		value = MsgBackendKeyData{
			ProcessID: read.Int32(),
			SecretKey: read.Int32(),
		}
	case tokenBindComplete:
		value = MsgBindComplete{}
	case tokenCloseComplete:
		value = MsgCloseComplete{}
	case tokenCommandComplete:
		value = MsgCommandComplete{
			Command: read.String(),
		}
	case tokenCopyData:
		value = MsgCopyData{
			Reader: &LimitedReader{
				r:     read,
				limit: int(read.Length),
			},
		}
		read.Length = 0
	case tokenCopyDone:
		value = MsgCopyDone{}
	case tokenCopyInResponse:
		{
			msg := MsgCopyInResponse{
				Format:      read.Int8(),
				ColumnCount: read.Int16(),
			}
			value = msg

			msg.ColumnFormat = make([]int16, int(msg.ColumnCount))
			for i := range msg.ColumnFormat {
				msg.ColumnFormat[i] = read.Int16()
			}
		}
	case tokenCopyOutResponse:
		{
			msg := MsgCopyOutResponse{
				Format:      read.Int8(),
				ColumnCount: read.Int16(),
			}
			value = msg

			msg.ColumnFormat = make([]int16, int(msg.ColumnCount))
			for i := range msg.ColumnFormat {
				msg.ColumnFormat[i] = read.Int16()
			}
		}
	case tokenCopyBothResponse:
		{
			msg := MsgCopyBothResponse{
				Format:      read.Int8(),
				ColumnCount: read.Int16(),
			}
			value = msg

			msg.ColumnFormat = make([]int16, int(msg.ColumnCount))
			for i := range msg.ColumnFormat {
				msg.ColumnFormat[i] = read.Int16()
			}
		}
	case tokenDataRow:
		value = MsgDataRow{
			ColumnCount: read.Int16(),
			FieldRead: &reader{
				Reader: read.Reader,
				buf:    read.buf,
				Length: 0,
			},
		}
		read.Length = 0
	case tokenEmptyQueryResponse:
		value = MsgEmptyQueryResponse{}
	case tokenErrorResponse:
		msg := MsgErrorResponse{}
		for read.Length > 0 {
			field := read.Byte()
			if field == 0 {
				break
			}
			msg.Messages = append(msg.Messages, &StatusMessage{
				Status:  status(field),
				Message: read.String(),
			})
		}

		value = msg
	case tokenFunctionCallResponse:
		msg := MsgFunctionCallResponse{
			Length: read.Int32(),
		}
		if msg.Length > 0 {
			// TODO: Decode fields.
		}
		value = msg
	case tokenNoData:
		value = MsgNoData{}
	case tokenNoticeResponse:
		value = MsgNoticeResponse{}
		// TODO: Decode body.
	case tokenNotificationResponse:
		value = MsgNotificationResponse{}
		// TODO: Decode body.
	case tokenParameterDescription:
		msg := MsgParameterDescription{
			Count: read.Int16(),
		}
		msg.ObjectID = make([]int32, msg.Count)
		for i := range msg.ObjectID {
			msg.ObjectID[i] = read.Int32()
		}
		value = msg
	case tokenParameterStatus:
		value = MsgParameterStatus{}
		parameterName := read.String()
		parameterValue := read.String()

		if debug {
			fmt.Printf("Server Param: %s = %s\n", parameterName, parameterValue)
		}
		pg.serverStatus[parameterName] = parameterValue

		switch parameterName {
		case "server_version":
			ver := &semver.Version{
				Product: "Postgres",
			}
			_, err = fmt.Sscanf(parameterValue, "%d.%d.%d", &ver.Major, &ver.Minor, &ver.Patch)
			if err != nil {
				_, err = fmt.Sscanf(parameterValue, "%d.%d%s", &ver.Major, &ver.Minor, &ver.PreRelease)
				if err != nil {
					return nil, fmt.Errorf("Failed to get product version number: %v", err)
				}
			}
			pg.serverVersion = ver
		case "TimeZone":
			pg.serverLocation, err = time.LoadLocation(parameterValue)
			if err != nil {
				return nil, err
			}
		}

	case tokenParseComplete:
		value = MsgParseComplete{}
	case tokenPortalSuspended:
		value = MsgPortalSuspended{}
	case tokenReadyForQuery:
		pg.inUse = false
		tranStatus := transactionStatus(read.Byte())
		value = MsgReadyForQuery{
			TransactionStatus: tranStatus,
		}
		pg.tranStatus = tranStatus
	case tokenRowDescription:
		value = MsgRowDescription{}

		//fmt.Printf("Row Description:\n%v", read.HexDump())
		//break

		count := read.Int16()
		schema := make([]Column, count)
		rdbSchema := make([]*rdb.Column, count)
		for i := range schema {
			name := read.String()
			tableId := read.Int32()
			columnId := read.Int16()
			objectId := read.Int32()
			colLength := read.Int16()
			typeMod := read.Int32()
			format := read.Int16()

			if debug {
				fmt.Printf("Column: %s (%d)\n", name, objectId)
			}

			columnType, found := typeLookup[Oid(objectId)]
			if !found {
				return nil, fmt.Errorf("Column type %v not found for column %v", objectId, name)
			}

			schema[i] = Column{
				Column: rdb.Column{
					Name:    name,
					Index:   i,
					Length:  int(colLength),
					Type:    columnType.Type,
					Generic: columnType.Generic,
				},
				Oid:     objectId,
				TypeMod: typeMod,
				Format:  format,

				TableID:  tableId,
				ColumnID: columnId,
			}
			rdbSchema[i] = &schema[i].Column
		}
		pg.columns = schema
		pg.valuer.Columns(rdbSchema)
	default:
		return nil, fmt.Errorf("Unhandled message type: %v", msgToken)
	}

	// Read message length.
	read.MsgDone()

	return value, nil
}
