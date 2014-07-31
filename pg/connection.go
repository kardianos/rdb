// Copyright 2014 Daniel Theophanes.
// Use of this source code is governed by a zlib-style
// license that can be found in the rdb LICENSE file.

package pg

import (
	"bufio"
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

// Read the next row from the connection. For each field in the row
// call the Valuer.WriteField(...) method. Propagate the reportRow field.
func (pg *connection) Scan(reportRow bool) error {
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
			return pg.valuer.Done()
		case MsgErrorResponse:
			return msg
		case MsgDataRow:
			columnCount := int(msg.ColumnCount)
			for i := 0; i < columnCount; i++ {
				rCol := &pg.columns[i].Column
				col := &pg.columns[i]
				if isNull := msg.NextField(); isNull {
					pg.valuer.WriteField(rCol, reportRow, &rdb.DriverValue{Null: true}, nil)
					continue
				}
				// Read from msg.FieldRead each field.
				// Decode field from field bytes.
				val, err := decodeField(col, msg.FieldRead)
				if err != nil {
					return err
				}
				pg.valuer.WriteField(rCol, reportRow, val, nil)
			}
			msg.FieldRead.MsgDone()
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

	write := pg.writer()

	if len(params) == 0 {
		write.Msg(tokenQuery)
		write.String(cmd.Sql)
		write.MsgDone()
		err := write.Send()
		if err != nil {
			return err
		}
	} else {
		// TODO: Prepare and add parameters.
		/*
			err = pg.prepareToSimpleStmt(cmd.Sql, "")
			if err != nil {
				panic(err)
			}
			pg.exec("", params, val)
			return nil
		*/
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
			return pg.valuer.Done()
		case MsgErrorResponse:
			return msg
		case MsgRowDescription:
			return nil
		default:
			return errUnhandledMessage("textOnlyQuery", msg)
		}
	}
}

func (pg *connection) textOnlyQuery(cmd *rdb.Command, val rdb.DriverValuer) error {
	write := pg.writer()
	write.Msg(tokenQuery)
	write.String(cmd.Sql)
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
			return pg.valuer.Done()
		case MsgErrorResponse:
			return msg
		case MsgRowDescription:
			return nil
		default:
			return errUnhandledMessage("textOnlyQuery", msg)
		}
	}
}
