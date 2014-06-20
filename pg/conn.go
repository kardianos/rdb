// Copyright (c) 2011, The pg Authors. All Rights Reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package pg

import (
	"bufio"
	"crypto/tls"
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"strconv"

	"bitbucket.org/kardianos/rdb"
	"bitbucket.org/kardianos/rdb/pg/oid"
	"bitbucket.org/kardianos/rdb/semver"
	// "strings"
	"time"
)

type conn struct {
	c         net.Conn
	config    *rdb.Config
	buf       *bufio.Reader
	namei     int
	scratch   [512]byte
	txnStatus transactionStatus

	parameterStatus parameterStatus

	open  bool
	inUse bool

	val rdb.DriverValuer
	col []*rdb.SqlColumn
}

// Return version information regarding the currently connected server.
func (c *conn) ConnectionInfo() *rdb.ConnectionInfo { return nil }

// Read the next row from the connection. For each field in the row
// call the Valuer.WriteField(...) method. Propagate the reportRow field.
func (conn *conn) Scan(reportRow bool) (err error) {
	if conn.inUse == false {
		return nil
	}

	defer errRecover(&err)

	for {
		t, r := conn.recv1()
		switch t {
		case 'E':
			err = parseError(r)
		case 'C':
			continue
		case 'Z':
			conn.processReadyForQuery(r)
			conn.inUse = false
			if err != nil {
				return err
			}
			return nil
		case 'D':
			n := r.int16()

			for i := 0; i < n; i++ {
				col := conn.col[i]
				l := r.int32()
				if l == -1 {
					conn.val.WriteField(col, reportRow, &rdb.DriverValue{
						Null: true,
					}, nil)
					continue
				}
				conn.val.WriteField(col, reportRow, &rdb.DriverValue{
					Value: decode(&conn.parameterStatus, r.next(l), oid.Oid(col.SqlType-rdb.TypeDriverThresh)),
				}, nil)
			}
			conn.val.RowScanned()
			return
		default:
			errorf("unexpected message after execute: %q", t)
		}
	}

	panic("not reached")
}

func (c *conn) SavePoint(name string) error { return nil }
func (c *conn) Status() rdb.DriverConnStatus {
	if c.open == false {
		return rdb.StatusDisconnected
	}
	if c.inUse == false {
		return rdb.StatusReady
	}
	return rdb.StatusQuery
}

func (c *conn) writeBuf(b byte) *writeBuf {
	c.scratch[0] = b
	w := writeBuf(c.scratch[:5])
	return &w
}

func (cn *conn) isInTransaction() bool {
	return cn.txnStatus == txnStatusIdleInTransaction ||
		cn.txnStatus == txnStatusInFailedTransaction
}

func (cn *conn) checkIsInTransaction(intxn bool) {
	if cn.isInTransaction() != intxn {
		errorf("unexpected transaction status %v", cn.txnStatus)
	}
}

func (cn *conn) Begin() (err error) {
	defer errRecover(&err)

	/*cn.checkIsInTransaction(false)
	_, commandTag, err := cn.simpleExec("BEGIN")
	if err != nil {
		return err
	}
	if commandTag != "BEGIN" {
		return fmt.Errorf("unexpected command tag %s", commandTag)
	}
	if cn.txnStatus != txnStatusIdleInTransaction {
		return fmt.Errorf("unexpected transaction status %v", cn.txnStatus)
	}
	*/
	return nil
}

func (cn *conn) Commit() (err error) {
	defer errRecover(&err)

	/*cn.checkIsInTransaction(true)
	// We don't want the client to think that everything is okay if it tries
	// to commit a failed transaction.  However, no matter what we return,
	// database/sql will release this connection back into the free connection
	// pool so we have to abort the current transaction here.  Note that you
	// would get the same behaviour if you issued a COMMIT in a failed
	// transaction, so it's also the least surprising thing to do here.
	if cn.txnStatus == txnStatusInFailedTransaction {
		if err := cn.Rollback(""); err != nil {
			return err
		}
		return ErrInFailedTransaction
	}

	_, commandTag, err := cn.simpleExec("COMMIT")
	if err != nil {
		return err
	}
	if commandTag != "COMMIT" {
		return fmt.Errorf("unexpected command tag %s", commandTag)
	}
	cn.checkIsInTransaction(false)
	*/
	return nil
}

func (cn *conn) Rollback(savepoint string) (err error) {
	defer errRecover(&err)

	/*cn.checkIsInTransaction(true)
	_, commandTag, err := cn.simpleExec("ROLLBACK")
	if err != nil {
		return err
	}
	if commandTag != "ROLLBACK" {
		return fmt.Errorf("unexpected command tag %s", commandTag)
	}
	cn.checkIsInTransaction(false)
	*/
	return nil
}

func (cn *conn) gname() string {
	cn.namei++
	return strconv.FormatInt(int64(cn.namei), 10)
}

// Very similar to simpleQuery, but doesn't interact with valuer or
// kick out to an external scan.
func (cn *conn) simpleExec(q string) (commandTag string, err error) {
	defer errRecover(&err)
	cn.inUse = true

	b := cn.writeBuf('Q')
	b.string(q)
	cn.send(b)

	for {
		t, r := cn.recv1()
		switch t {
		case 'C':
			_, commandTag = parseComplete(r.string())
		case 'Z':
			cn.processReadyForQuery(r)
			cn.inUse = false
			return
		case 'E':
			err = parseError(r)
		case 'T', 'D':
			// ignore any results
		default:
			errorf("unknown response for simple query: %q", t)
		}
	}
	panic("not reached")
}

func (cn *conn) simpleQuery(cmd *rdb.Command, val rdb.DriverValuer) (err error) {
	defer errRecover(&err)
	cn.inUse = true

	b := cn.writeBuf('Q')
	b.string(cmd.Sql)
	cn.send(b)

	for {
		t, r := cn.recv1()
		switch t {
		case 'C':
			return val.Done()
		case 'Z':
			cn.processReadyForQuery(r)
			return val.Done()
		case 'E':
			err = parseError(r)
		case 'T':
			cn.parseMeta(r)
			return
		default:
			errorf("unknown response for simple query: %q", t)
		}
	}
}

func (cn *conn) prepareToSimpleStmt(q, stmtName string) (err error) {
	defer errRecover(&err)
	cn.inUse = true

	b := cn.writeBuf('P')
	b.string(stmtName)
	b.string(q)
	b.int16(0)
	cn.send(b)

	b = cn.writeBuf('D')
	b.byte('S')
	b.string(stmtName)
	cn.send(b)

	cn.send(cn.writeBuf('S'))

	for {
		t, r := cn.recv1()
		switch t {
		case '1':
		case 't':
			// TODO: What to do with these...?
			nparams := int(r.int16())
			cols := make([]*rdb.SqlColumn, nparams)

			for _ = range cols {
				// st.paramTyps[i] = r.oid()
				_ = r.oid()
			}
		case 'T':
			cn.parseMeta(r)
		case 'n':
			// no data
		case 'Z':
			cn.processReadyForQuery(r)
			return err
		case 'E':
			err = parseError(r)
		default:
			errorf("unexpected describe rows response: %q", t)
		}
	}

	panic("not reached")
}

func (c *conn) exec(statementName string, v []rdb.Param, val rdb.DriverValuer) {
	// TODO: Add this check back in.
	// if len(v) != len(st.paramTyps) {
	// 	errorf("got %d parameters but the statement requires %d", len(v), len(st.paramTyps))
	// }
	c.inUse = true

	w := c.writeBuf('B')
	w.string("")
	w.string(statementName)
	w.int16(0)
	w.int16(len(v))
	for i, x := range v {
		if x.Null || x.V == nil {
			w.int32(-1)
		} else {
			// TODO: Send in SqlType.
			tp := oid.Oid(c.col[i].SqlType - rdb.TypeDriverThresh)
			b := encode(&c.parameterStatus, x.V, tp)
			w.int32(len(b))
			w.bytes(b)
		}
	}
	w.int16(0)
	c.send(w)

	w = c.writeBuf('E')
	w.string("")
	w.int32(0)
	c.send(w)

	c.send(c.writeBuf('S'))

	var err error
	for {
		t, r := c.recv1()
		switch t {
		case 'E':
			err = parseError(r)
		case '2':
			if err != nil {
				panic(err)
			}
			return
		case 'Z':
			c.processReadyForQuery(r)
			if err != nil {
				panic(err)
			}
			return
		default:
			errorf("unexpected bind response: %q", t)
		}
	}
}

func (c *conn) Prepare(*rdb.Command) (preparedStatementToken interface{}, err error) {
	/*Prepare(q string) (driver.Stmt, error)
		return cn.prepareCopyIn(q)
	}
	return cn.prepareTo(q, cn.gname())
	*/
	return nil, nil
}
func (c *conn) Unprepare(preparedStatementToken interface{}) (err error) {
	return nil
}

func (cn *conn) Close() {
	cn.open = false
	var err error
	defer errRecover(&err)

	// Don't go through send(); ListenerConn relies on us not scribbling on the
	// scratch buffer of this connection.
	err = cn.sendSimpleMessage('X')
	// TODO: Determine if the error value should get set on the conn so it doesn't get reused.
	if err != nil {
		return
	}
	cn.c.Close()
	return
}

func (c *conn) Query(cmd *rdb.Command, params []rdb.Param, preparedToken interface{}, val rdb.DriverValuer) (err error) {
	defer errRecover(&err)
	c.val = val

	if len(params) == 0 {
		return c.simpleQuery(cmd, val)
	}

	err = c.prepareToSimpleStmt(cmd.Sql, "")
	if err != nil {
		panic(err)
	}
	c.exec("", params, val)
	return nil
}

// Assumes len(*m) is > 5
func (cn *conn) send(m *writeBuf) {
	b := (*m)[1:]
	binary.BigEndian.PutUint32(b, uint32(len(b)))

	if (*m)[0] == 0 {
		*m = b
	}

	_, err := cn.c.Write(*m)
	if err != nil {
		panic(err)
	}
}

// Send a message of type typ to the server on the other end of cn.  The
// message should have no payload.  This method does not use the scratch
// buffer.
func (cn *conn) sendSimpleMessage(typ byte) (err error) {
	_, err = cn.c.Write([]byte{typ, '\x00', '\x00', '\x00', '\x04'})
	return err
}

// recvMessage receives any message from the backend, or returns an error if
// a problem occurred while reading the message.
func (cn *conn) recvMessage() (byte, *readBuf, error) {
	x := cn.scratch[:5]
	_, err := io.ReadFull(cn.buf, x)
	if err != nil {
		return 0, nil, err
	}
	t := x[0]

	b := readBuf(x[1:])
	n := b.int32() - 4
	var y []byte
	if n <= len(cn.scratch) {
		y = cn.scratch[:n]
	} else {
		y = make([]byte, n)
	}
	_, err = io.ReadFull(cn.buf, y)
	if err != nil {
		return 0, nil, err
	}

	return t, (*readBuf)(&y), nil
}

// recv receives a message from the backend, but if an error happened while
// reading the message or the received message was an ErrorResponse, it panics.
// NoticeResponses are ignored.  This function should generally be used only
// during the startup sequence.
func (cn *conn) recv() (t byte, r *readBuf) {
	for {
		var err error
		t, r, err = cn.recvMessage()
		if err != nil {
			panic(err)
		}

		switch t {
		case 'E':
			panic(parseError(r))
		case 'N':
			// ignore
		default:
			return
		}
	}

	panic("not reached")
}

// recv1 receives a message from the backend, panicking if an error occurs
// while attempting to read it.  All asynchronous messages are ignored, with
// the exception of ErrorResponse.
func (cn *conn) recv1() (t byte, r *readBuf) {
	for {
		var err error
		t, r, err = cn.recvMessage()
		if err != nil {
			panic(err)
		}

		switch t {
		case 'A', 'N':
			// ignore
		case 'S':
			cn.processParameterStatus(r)
		default:
			return
		}
	}

	panic("not reached")
}

func (cn *conn) ssl(o values) {
	tlsConf := tls.Config{}
	switch mode := o.Get("sslmode"); mode {
	case "require", "":
		tlsConf.InsecureSkipVerify = true
	case "verify-full":
		// fall out
	case "disable":
		return
	default:
		errorf(`unsupported sslmode %q; only "require" (default), "verify-full", and "disable" supported`, mode)
	}

	w := cn.writeBuf(0)
	w.int32(80877103)
	cn.send(w)

	b := cn.scratch[:1]
	_, err := io.ReadFull(cn.c, b)
	if err != nil {
		panic(err)
	}

	if b[0] != 'S' {
		panic(ErrSSLNotSupported)
	}

	cn.c = tls.Client(cn.c, &tlsConf)
}

func (cn *conn) startup(o values) {
	w := cn.writeBuf(0)
	w.int32(196608)
	// Send the backend the name of the database we want to connect to, and the
	// user we want to connect as.  Additionally, we send over any run-time
	// parameters potentially included in the connection string.  If the server
	// doesn't recognize any of them, it will reply with an error.
	for k, v := range o {
		// skip options which can't be run-time parameters
		if k == "password" || k == "host" ||
			k == "port" || k == "sslmode" {
			continue
		}
		// The protocol requires us to supply the database name as "database"
		// instead of "dbname".
		if k == "dbname" {
			k = "database"
		}
		w.string(k)
		w.string(v)
	}
	w.string("")
	cn.send(w)

	for {
		t, r := cn.recv()
		switch t {
		case 'K':
		case 'S':
			cn.processParameterStatus(r)
		case 'R':
			cn.auth(r, o)
		case 'Z':
			cn.processReadyForQuery(r)
			cn.open = true
			return
		default:
			errorf("unknown response for startup: %q", t)
		}
	}
}

func (cn *conn) auth(r *readBuf, o values) {
	switch code := r.int32(); code {
	case 0:
		// OK
	case 3:
		w := cn.writeBuf('p')
		w.string(o.Get("password"))
		cn.send(w)

		t, r := cn.recv()
		if t != 'R' {
			errorf("unexpected password response: %q", t)
		}

		if r.int32() != 0 {
			errorf("unexpected authentication response: %q", t)
		}
	case 5:
		s := string(r.next(4))
		w := cn.writeBuf('p')
		w.string("md5" + md5s(md5s(o.Get("password")+o.Get("user"))+s))
		cn.send(w)

		t, r := cn.recv()
		if t != 'R' {
			errorf("unexpected password response: %q", t)
		}

		if r.int32() != 0 {
			errorf("unexpected authentication response: %q", t)
		}
	default:
		errorf("unknown authentication response: %d", code)
	}
}

func (c *conn) processParameterStatus(r *readBuf) {
	var err error

	param := r.string()
	switch param {
	case "server_version":
		ver := &semver.Version{
			Product: "Postgres",
		}
		_, err = fmt.Sscanf(r.string(), "%d.%d.%d", &ver.Major, &ver.Minor, &ver.Patch)
		if err == nil {
			c.parameterStatus.serverVersion = ver
		}

	case "TimeZone":
		c.parameterStatus.currentLocation, err = time.LoadLocation(r.string())
		if err != nil {
			c.parameterStatus.currentLocation = nil
		}

	default:
		// ignore
	}
}

func (c *conn) processReadyForQuery(r *readBuf) {
	c.txnStatus = transactionStatus(r.byte())
}
