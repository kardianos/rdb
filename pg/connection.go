// Copyright 2014 Daniel Theophanes.
// Use of this source code is governed by a zlib-style
// license that can be found in the rdb LICENSE file.

package pg

import (
	"bufio"
	"fmt"
	"net"
	"time"

	"bitbucket.org/kardianos/rdb"
	"bitbucket.org/kardianos/rdb/semver"
)

type connection struct {
	conn   net.Conn
	config *rdb.Config
	open   bool
	inUse  bool

	serverVersion  *semver.Version
	serverLocation *time.Location

	valuer  rdb.DriverValuer
	columns []Column

	scratch     [512]byte
	readBuffer  *bufio.Reader
	writeBuffer *bufio.Writer

	tranStatus   transactionStatus
	serverStatus map[string]string
}

// Return version information regarding the currently connected server.
func (pg *connection) ConnectionInfo() *rdb.ConnectionInfo { return nil }

func (pg *connection) done() error {
	pg.inUse = false
	if pg.valuer == nil {
		return nil
	}
	return pg.valuer.Done()
}

func (pg *connection) NextQuery() error {
	if pg.inUse == false {
		return nil
	}
	var value interface{}
	var err error

	for {
		value, err = pg.getMessage()
		if err != nil {
			return err
		}
		switch msg := value.(type) {
		case MsgCommandComplete:
		case MsgReadyForQuery:
			return pg.done()
		case MsgErrorResponse:
			return msg
		case MsgDataRow:
			columnCount := int(msg.ColumnCount)
			for i := 0; i < columnCount; i++ {
				msg.NextField()
				msg.FieldRead.MsgDone()
			}
			pg.valuer.RowScanned()
		default:
			return errUnhandledMessage("NextQuery", msg)
		}
	}
}

func (pg *connection) NextResult() (bool, error) {
	return false, pg.NextQuery()
}

// Read the next row from the connection. For each field in the row
// call the Valuer.WriteField(...) method. Propagate the reportRow field.
func (pg *connection) Scan() error {
	var value interface{}
	var err error

	for {
		value, err = pg.getMessage()
		if err != nil {
			return err
		}
		switch msg := value.(type) {
		case MsgCommandComplete:
		case MsgReadyForQuery:
			return pg.done()
		case MsgErrorResponse:
			return msg
		case MsgDataRow:
			columnCount := int(msg.ColumnCount)
			for i := 0; i < columnCount; i++ {
				rCol := &pg.columns[i].Column
				col := &pg.columns[i]
				if isNull := msg.NextField(); isNull {
					pg.valuer.WriteField(rCol, &rdb.DriverValue{Null: true}, nil)
					continue
				}
				// Read from msg.FieldRead each field.
				// Decode field from field bytes.
				val, err := decodeField(col, msg.FieldRead)
				if err != nil {
					return err
				}
				pg.valuer.WriteField(rCol, val, nil)
			}
			msg.FieldRead.MsgDone()
			pg.valuer.RowScanned()
		default:
			return errUnhandledMessage("Scan", msg)
		}
	}
}

func (pg *connection) SavePoint(name string) error { return nil }
func (pg *connection) Status() rdb.DriverConnStatus {
	if pg.open == false {
		return rdb.StatusDisconnected
	}
	if pg.inUse == false {
		return rdb.StatusReady
	}
	return rdb.StatusQuery
}

func (pg *connection) Begin(level rdb.IsolationLevel) (err error) {
	return nil
}

func (pg *connection) Commit() (err error) {
	return nil
}

func (pg *connection) Rollback(savepoint string) (err error) {
	return nil
}

func (pg *connection) Prepare(*rdb.Command) (preparedStatementToken interface{}, err error) {
	return nil, nil
}
func (pg *connection) Unprepare(preparedStatementToken interface{}) (err error) {
	return nil
}

func (pg *connection) Close() {
}

func (pg *connection) Query(cmd *rdb.Command, params []rdb.Param, preparedToken interface{}, val rdb.DriverValuer) error {
	pg.valuer = val
	pg.inUse = true

	// 	write := pg.writer()

	if len(params) == 0 {
		return pg.textOnlyQuery(cmd.Sql)
	}
	return pg.paramQuery(cmd.Sql, cmd.Arity, params)
}

func (pg *connection) paramQuery(sql string, arity rdb.Arity, params []rdb.Param) error {
	var err error
	var value interface{}

	write := pg.writer()
	write.Msg(tokenParse)
	write.String("") // Prepared statement name.
	write.String(sql)
	write.Int16(int16(len(params)))
	for i := range params {
		oidType, found := rdbTypeLookup[params[i].Type]
		if !found {
			return fmt.Errorf("Unhandled rdb data type for parameter %s: %v", params[i].Name, params[i].Type)
		}
		write.Int32(int32(oidType.Oid))
	}
	write.MsgDone()

	write.Msg(tokenDescribe)
	write.Byte(byte('S'))
	write.String("") // Prepared statement name.
	write.MsgDone()

	write.Msg(tokenSync)
	write.MsgDone()

	err = write.Send()
	if err != nil {
		return err
	}
loop:
	for {
		value, err = pg.getMessage()
		if err != nil {
			return err
		}
		switch msg := value.(type) {
		case MsgParameterDescription:
		case MsgRowDescription:
		case MsgParseComplete:
		case MsgReadyForQuery:
			break loop
		case MsgErrorResponse:
			return msg
		default:
			return errUnhandledMessage("textOnlyQuery", msg)
		}
	}

	write.Msg(tokenBind)
	write.String("") // Portal name.
	write.String("") // Statement name.
	write.Int16(1)
	write.Int16(0) // Parameter format codes: zero text, one binary.
	write.Int16(int16(len(params)))
	for i := range params {
		// TODO: Handle null (set param data length to -1.
		// int32 param data length.
		// param data.
		err = encodeField(&params[i], write)
		if err != nil {
			return err
		}
	}
	write.Int16(1)
	write.Int16(0) // Result format codes: zero text, one binary.
	write.MsgDone()

	write.Msg(tokenExecute)
	write.String("")
	maxReturn := int32(0)
	switch {
	case arity&rdb.Zero != 0:
		maxReturn = 1
	case arity&rdb.One != 0:
		maxReturn = 2
	}
	write.Int32(maxReturn)
	write.MsgDone()

	write.Msg(tokenSync)
	write.MsgDone()

	err = write.Send()
	if err != nil {
		return err
	}

	for {
		value, err = pg.getMessage()
		if err != nil {
			return err
		}
		switch msg := value.(type) {
		case MsgCommandComplete:
		case MsgParseComplete:
		case MsgParameterDescription:
		case MsgBindComplete:
			return nil
		case MsgReadyForQuery:
			return pg.done()
		case MsgErrorResponse:
			return msg
		default:
			return errUnhandledMessage("textOnlyQuery", msg)
		}
	}
}

func (pg *connection) textOnlyQuery(sql string) error {
	write := pg.writer()
	write.Msg(tokenQuery)
	write.String(sql)
	write.MsgDone()

	err := write.Send()
	if err != nil {
		return err
	}
	var value interface{}

	for {
		value, err = pg.getMessage()
		if err != nil {
			return err
		}
		switch msg := value.(type) {
		case MsgCommandComplete:
		case MsgReadyForQuery:
			return pg.done()
		case MsgErrorResponse:
			return msg
		case MsgRowDescription:
			return nil
		default:
			return errUnhandledMessage("textOnlyQuery", msg)
		}
	}
}
