// Copyright 2014 Daniel Theophanes.
// Use of this source code is governed by a zlib-style
// license that can be found in the LICENSE file.

package ms

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/kardianos/rdb"
	"github.com/kardianos/rdb/internal/uconv"
	"github.com/kardianos/rdb/semver"

	"errors"
)

const (
	debugToken = false
	debugAPI   = false
	debugProto = false
)

type Connection struct {
	pw *PacketWriter
	pr *PacketReader

	wc      net.Conn
	onDone  chan struct{} // Write to when message is done.
	onClose chan struct{} // Close to when connection closes

	status    rdb.DriverConnStatus
	available bool
	resetNext bool
	syncClose sync.Mutex

	ProductVersion  *semver.Version
	ProtocolVersion *semver.Version
	Encrypted       bool

	mr     *MessageReader
	val    rdb.DriverValuer
	col    []*SqlColumn
	params []rdb.Param

	allHeaders            []byte
	allHeaderNumberOffset int

	currentTransaction uint64

	opened              time.Time
	defaultResetTimeout time.Duration

	// Next token type.
	peek tdsToken

	// The next byte of ucs2 if split between packets.
	ucs2Next []byte
}

func NewConnection(c net.Conn, defaultResetTimeout time.Duration) *Connection {
	return &Connection{
		pw:      NewPacketWriter(c),
		pr:      NewPacketReader(c),
		wc:      c,
		opened:  time.Now(),
		onDone:  make(chan struct{}),
		onClose: make(chan struct{}),

		defaultResetTimeout: defaultResetTimeout,
	}
}

func (tds *Connection) SetAvailable(available bool) {
	tds.available = available
}
func (tds *Connection) Available() bool {
	return tds.available
}

func (tds *Connection) getAllHeaders() []byte {
	binary.LittleEndian.PutUint64(tds.allHeaders[tds.allHeaderNumberOffset:], tds.currentTransaction)
	return tds.allHeaders
}

func (tds *Connection) Open(config *rdb.Config) (*ServerInfo, error) {
	if tds.Status() != rdb.StatusDisconnected {
		return nil, connectionOpenError
	}
	var err error

	tds.allHeaders, tds.allHeaderNumberOffset = getHeaderTemplate()

	encrypt := encryptOn
	if config.InsecureDisableEncryption {
		encrypt = encryptNotSupported
	}
	if config.Secure {
		encrypt = encryptRequired
	}

	err = tds.pw.PreLogin(config.Instance, encrypt)
	if err != nil {
		return nil, err
	}

	sc, err := tds.pr.Prelogin()
	if err != nil {
		return nil, err
	}

	switch sc.Encryption {
	default:
		if config.Secure {
			return nil, fmt.Errorf("encryption required but server does not support encryption")
		}
	case encryptOn, encryptRequired:
		tlsConfig := &tls.Config{
			DynamicRecordSizingDisabled: true,
			InsecureSkipVerify:          config.InsecureSkipVerify,
			ServerName:                  config.Hostname,
			MinVersion:                  tls.VersionTLS12,
			RootCAs:                     config.RootCAs,
		}

		handshakeConn := &tlsHandshakeConn{
			conn: tds,
		}
		connSwitch := &passthroughConn{c: handshakeConn}
		tlsConn := tls.Client(connSwitch, tlsConfig)
		err = tlsConn.Handshake()
		if err != nil {
			return nil, fmt.Errorf("TLS Handshake error: %w", err)
		}

		connSwitch.c = tds.wc
		tds.pw = NewPacketWriter(tlsConn)
		tds.pr = NewPacketReader(tlsConn)
		tds.Encrypted = true
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

	tds.syncClose.Lock()
	tds.status = rdb.StatusReady
	tds.syncClose.Unlock()

	return si, tds.NextQuery()
}

// this connection is used during TLS Handshake
// TDS protocol requires TLS handshake messages to be sent inside TDS packets
type tlsHandshakeConn struct {
	conn          *Connection
	mr            *MessageReader
	readBuffer    []byte
	packetPending bool
	continueRead  bool
}

func (c *tlsHandshakeConn) Read(b []byte) (n int, err error) {
	if c.packetPending {
		c.packetPending = false
		_, err = c.conn.pw.writeClose([]byte{}, true)
		if err != nil {
			return 0, fmt.Errorf("cannot send handshake packet: %s", err.Error())
		}
		c.continueRead = false
	}
	if !c.continueRead || len(c.readBuffer) == 0 {
		if c.mr == nil {
			c.mr = c.conn.pr.BeginMessage(packetPreLogin)
		}
		c.readBuffer, err = c.mr.Next()
		if err == io.EOF && n > 0 {
			err = nil
		}
		c.continueRead = true
	}
	n = copy(b, c.readBuffer)
	c.readBuffer = c.readBuffer[n:]
	return n, err
}

func (c *tlsHandshakeConn) Write(b []byte) (n int, err error) {
	if !c.packetPending {
		c.conn.pw.BeginMessage(context.Background(), packetPreLogin, false)
		c.packetPending = true
	}
	_, err = c.conn.pw.Write(b)
	return len(b), err
}

func (c *tlsHandshakeConn) Close() error {
	return c.conn.wc.Close()
}

func (c *tlsHandshakeConn) LocalAddr() net.Addr {
	return nil
}

func (c *tlsHandshakeConn) RemoteAddr() net.Addr {
	return nil
}

func (c *tlsHandshakeConn) SetDeadline(_ time.Time) error {
	return nil
}

func (c *tlsHandshakeConn) SetReadDeadline(_ time.Time) error {
	return nil
}

func (c *tlsHandshakeConn) SetWriteDeadline(_ time.Time) error {
	return nil
}

type passthroughConn struct {
	c net.Conn
}

func (c passthroughConn) Read(b []byte) (n int, err error) {
	return c.c.Read(b)
}

func (c passthroughConn) Write(b []byte) (n int, err error) {
	return c.c.Write(b)
}

func (c passthroughConn) Close() error {
	return c.c.Close()
}

func (c passthroughConn) LocalAddr() net.Addr {
	return c.c.LocalAddr()
}

func (c passthroughConn) RemoteAddr() net.Addr {
	return c.c.RemoteAddr()
}

func (c passthroughConn) SetDeadline(t time.Time) error {
	return c.c.SetDeadline(t)
}

func (c passthroughConn) SetReadDeadline(t time.Time) error {
	return c.c.SetReadDeadline(t)
}

func (c passthroughConn) SetWriteDeadline(t time.Time) error {
	return c.c.SetWriteDeadline(t)
}

func (tds *Connection) Reset(c *rdb.Config) error {
	tds.resetNext = true
	if len(c.ResetQuery) == 0 {
		return nil
	}
	ctx := context.Background()
	if c.ResetConnectionTimeout > 0 {
		var cancel func()
		ctx, cancel = context.WithTimeout(ctx, c.ResetConnectionTimeout)
		defer cancel()
	}
	return tds.Query(ctx, &rdb.Command{Sql: c.ResetQuery}, nil, nil, nil)
}

func (tds *Connection) ConnectionInfo() *rdb.ConnectionInfo {
	return &rdb.ConnectionInfo{
		Server:   tds.ProductVersion,
		Protocol: tds.ProtocolVersion,
	}
}

func (tds *Connection) Opened() time.Time {
	return tds.opened
}

func (tds *Connection) Close() {
	tds.syncClose.Lock()
	if tds.status == rdb.StatusDisconnected {
		tds.syncClose.Unlock()
		return
	}
	tds.val = nil
	tds.mr = nil
	tds.status = rdb.StatusDisconnected
	close(tds.onClose)
	tds.syncClose.Unlock()

	tds.done()
	tds.wc.Close()
}

func (tds *Connection) Status() rdb.DriverConnStatus {
	tds.syncClose.Lock()
	status := tds.status
	tds.syncClose.Unlock()
	return status
}

func (tds *Connection) Prepare(*rdb.Command) (preparedStatementToken interface{}, err error) {
	return nil, rdb.NotImplemented
}
func (tds *Connection) Unprepare(preparedStatementToken interface{}) (err error) {
	return rdb.NotImplemented
}

/*
0 = TM_GET_DTC_ADDRESS. Returns DTC network address as a result set with a single-column, single-row binary value.
1 = TM_PROPAGATE_XACT. Imports DTC transaction into the server and returns a local transaction descriptor as a varbinary result set.
5 = TM_BEGIN_XACT. Begins a transaction and returns the descriptor in an ENVCHANGE type 8.
6 = TM_PROMOTE_XACT. Converts an active local transaction into a distributed transaction and returns an opaque buffer in an ENVCHANGE type 15.
7 = TM_COMMIT_XACT. Commits a transaction. Depending on the payload of the request, it can additionally request that another local transaction be started.
8 = TM_ROLLBACK_XACT. Rolls back a transaction. Depending on the payload of the request, it can indicate that after the rollback, a local transaction is to be started.
9 = TM_SAVE_XACT. Sets a savepoint within the active transaction. This request MUST specify a nonempty name for the savepoint.
The request types 5 - 9 were introduced in TDS 7.2.
*/
const (
	tranBegin     = 5
	tranCommit    = 7
	tranRollback  = 8
	tranSavepoint = 9
)

const (
	levelDefault         = 0x00
	levelReadUncommitted = 0x01
	levelReadCommited    = 0x02
	levelRepeatableRead  = 0x03
	levelSerializable    = 0x04
	levelSnapshot        = 0x05
)

func (tds *Connection) transaction(tran uint16, label string, iso rdb.IsolationLevel) error {
	ctx := context.Background()

	tds.syncClose.Lock()
	if tds.status == rdb.StatusDisconnected {
		tds.syncClose.Unlock()
		return connectionNotOpenError
	}
	if tds.status != rdb.StatusReady {
		tds.syncClose.Unlock()
		return connectionInUseError
	}
	tds.status = rdb.StatusQuery
	tds.syncClose.Unlock()

	if tds.mr != nil && tds.mr.packetEOM == false {
		panic("Connection not ready to be re-used yet for transaction.")
	}

	tds.mr = tds.pr.BeginMessage(packetTabularResult)
	err := tds.pw.BeginMessage(ctx, packetTransaction, false)
	if err != nil {
		return err
	}

	var level byte
	switch iso {
	case rdb.LevelDefault:
		level = levelDefault
	case rdb.LevelReadUncommited:
		level = levelReadUncommitted
	case rdb.LevelReadCommited:
		level = levelReadCommited
	case rdb.LevelRepeatableRead:
		level = levelRepeatableRead
	case rdb.LevelSerializable:
		level = levelSerializable
	case rdb.LevelSnapshot:
		level = levelSnapshot
	}

	if len(label) > 254 {
		label = label[:254]
	}
	labelLen := byte(len(label))
	tds.pw.WriteBuffer(tds.getAllHeaders())
	tds.pw.WriteUint16(tran)
	switch tran {
	case tranBegin:
		tds.pw.WriteByte(level)
		tds.pw.WriteByte(labelLen)
		if labelLen != 0 {
			tds.pw.Write([]byte(label))
		}
	case tranCommit:
		tds.pw.WriteByte(labelLen)
		if labelLen != 0 {
			tds.pw.Write([]byte(label))
		}
		tds.pw.WriteByte(0) // Don't start another transaction.
	case tranRollback:
		tds.pw.WriteByte(labelLen)
		if labelLen != 0 {
			tds.pw.Write([]byte(label))
		}
		tds.pw.WriteByte(0) // Don't start another transaction.
	case tranSavepoint:
		tds.pw.WriteByte(labelLen)
		if labelLen != 0 {
			tds.pw.Write([]byte(label))
		}
	default:
		panic("Unknown transaction request.")
	}

	err = tds.pw.EndMessage()
	if err != nil {
		return err
	}
	return tds.NextQuery()
}
func (tds *Connection) Begin(iso rdb.IsolationLevel) error {
	return tds.transaction(tranBegin, "", iso)
}
func (tds *Connection) Rollback(savepoint string) error {
	return tds.transaction(tranRollback, savepoint, rdb.LevelDefault)
}
func (tds *Connection) Commit() error {
	return tds.transaction(tranCommit, "", rdb.LevelDefault)
}
func (tds *Connection) SavePoint(name string) error {
	return tds.transaction(tranSavepoint, name, rdb.LevelDefault)
}

type noopValuer struct {
}

func (noopValuer) Columns([]*rdb.Column) error { return nil }
func (noopValuer) Done() error                 { return nil }
func (noopValuer) RowScanned()                 {}
func (noopValuer) Message(*rdb.Message)        {}
func (noopValuer) WriteField(c *rdb.Column, value *rdb.DriverValue, assign rdb.Assigner) error {
	return nil
}
func (noopValuer) RowsAffected(count uint64) {}

func (tds *Connection) Query(ctx context.Context, cmd *rdb.Command, params []rdb.Param, preparedToken interface{}, valuer rdb.DriverValuer) error {
	if debugAPI {
		fmt.Printf("API Query\n")
	}
	tds.syncClose.Lock()
	if tds.status != rdb.StatusReady {
		tds.syncClose.Unlock()
		return connectionInUseError
	}
	tds.syncClose.Unlock()

	if valuer == nil {
		valuer = noopValuer{}
	}
	tds.val = valuer

	if tds.mr != nil && tds.mr.packetEOM == false {
		return fmt.Errorf("Connection not ready to be re-used yet for query.")
	}
	go tds.asyncWaitCancel(ctx, cmd.Sql)
	tds.mr = tds.pr.BeginMessage(packetTabularResult)

	err := tds.execute(ctx, cmd.Sql, cmd.TruncLongText, cmd.Arity, params)
	if err != nil {
		return err
	}
	tds.syncClose.Lock()
	doNext := tds.status == rdb.StatusQuery && err == nil
	tds.syncClose.Unlock()

	if doNext {
		_, err = tds.nextResult()
	}
	return err
}

func (tds *Connection) asyncWaitCancel(ctx context.Context, sql string) {
	select {
	case <-ctx.Done():
		// Wait until message is done.
		err := tds.pw.BeginMessage(context.TODO(), packetAttention, false)
		if err != nil {
			// TODO: Determine a better error path.
			log.Printf("Cancel begin message: %v\n", err)
			select {
			case <-tds.onDone:
			case <-tds.onClose:
			}
			return
		}
		err = tds.pw.EndMessage()
		if err != nil {
			// TODO: Determine a better error path.
			log.Printf("Cancel end message: %v\n", err)
		}
		select {
		case <-tds.onDone:
		case <-tds.onClose:
		}
	case <-tds.onDone:
		// Nothing.
	case <-tds.onClose:
		// Nothing.
	}
}

func (tds *Connection) NextResult() (more bool, err error) {
	if debugAPI {
		fmt.Printf("API NextResult\n")
	}
	return tds.nextResult()
}

func (tds *Connection) nextResult() (more bool, err error) {
	tds.syncClose.Lock()

	more = (tds.status == rdb.StatusResultDone)
	if debugAPI {
		fmt.Printf("API nextResult more=%t, tds.status=%d\n", more, tds.status)
	}
	if more {
		tds.status = rdb.StatusQuery
		tds.syncClose.Unlock()

		err = tds.scan()

		tds.syncClose.Lock()
		more = tds.status == rdb.StatusResultDone || tds.status == rdb.StatusQuery
		tds.syncClose.Unlock()
	} else {
		more = tds.status == rdb.StatusResultDone || tds.status == rdb.StatusQuery
		tds.syncClose.Unlock()
	}
	return more, err
}

func (tds *Connection) NextQuery() (err error) {
	if debugAPI {
		fmt.Printf("API NextQuery\n")
		defer fmt.Printf("<API NextQuery\n")
	}
	run := true
	for run {
		var res interface{}
		var err error
		withLock(&tds.syncClose, func() {
			run = tds.status != rdb.StatusReady && tds.status != rdb.StatusDisconnected
			if !run {
				return
			}
			res, err = tds.getSingleResponse(tds.mr, false)
		})
		if !run {
			break
		}
		if err != nil {
			tds.done()
			return err
		}
		switch res.(type) {
		case MsgEom:
			// END OF (TDS) MESSAGE.
			return tds.done()
		case MsgFinalDone:
			return tds.done()
		}
	}
	return nil
}

func (tds *Connection) done() error {
	if tds == nil {
		return nil
	}
	select {
	case tds.onDone <- struct{}{}:
	default:
	}

	mrCloseErr := tds.mr.Close()
	tds.params = nil

	tds.syncClose.Lock()
	tds.col = nil
	if tds.status != rdb.StatusDisconnected {
		tds.status = rdb.StatusReady
	}
	tds.syncClose.Unlock()

	var err error
	if tds.val != nil {
		err = tds.val.Done()
		if err == nil {
			err = mrCloseErr
		}
	}
	return err
}

func (tds *Connection) Scan() error {
	if debugAPI {
		fmt.Printf("API Scan\n")
	}
	return tds.scan()
}

func (tds *Connection) scan() error {
	if debugAPI {
		fmt.Printf("api scan\n")
		defer fmt.Printf("<api scan\n")
	}
	tds.syncClose.Lock()
	if tds.status == rdb.StatusResultDone {
		tds.syncClose.Unlock()
		return io.EOF
	}
	if tds.status != rdb.StatusQuery {
		tds.syncClose.Unlock()
		return nil
	}
	tds.syncClose.Unlock()

	hasCol := false
	for {
		var res interface{}
		var err error
		withLock(&tds.syncClose, func() {
			res, err = tds.getSingleResponse(tds.mr, true)
		})
		if err != nil {
			tds.done()
			return err
		}
		switch v := res.(type) {
		case MsgEom:
			// END OF (TDS) MESSAGE.
			err = tds.done()
			if hasCol {
				tds.syncClose.Lock()
				if tds.status != rdb.StatusDisconnected {
					tds.status = rdb.StatusResultDone
				}
				tds.syncClose.Unlock()
			}
			return err
		case *rdb.Message:
			tds.val.Message(v)
		case MsgColumn:
			hasCol = true
		case MsgRow:
			// Sent after the row is scanned.
			// Prep values must be cleared after the initial fill.
			// The prior prep values are no longer valid as they are filled
			// during the row scan.
			tds.val.RowScanned()
		case MsgRowCount:
			tds.val.RowsAffected(v.Count)
		case MsgOrder:
		case MsgDone:
		case MsgFinalDone:
			err = tds.done()
			if hasCol {
				tds.syncClose.Lock()
				if tds.status != rdb.StatusDisconnected {
					tds.status = rdb.StatusResultDone
				}
				tds.syncClose.Unlock()
			}
			return err
		case MsgCancel:
			err = tds.done()
			if hasCol {
				tds.syncClose.Lock()
				if tds.status != rdb.StatusDisconnected {
					tds.status = rdb.StatusResultDone
				}
				tds.syncClose.Unlock()
			}
			if err != nil {
				return err
			}
			return rdb.ErrCancel
		}
		if tds.col == nil {
			continue
		}
		switch tds.peek {
		case tokenColumnMetaData:
			tds.status = rdb.StatusResultDone
			return nil
		case tokenRow:
			return nil
		case tokenNBCRow:
			return nil
		}
	}
}

func (tds *Connection) execute(ctx context.Context, sql string, truncValue bool, arity rdb.Arity, params []rdb.Param) error {
	tds.syncClose.Lock()

	if tds.status == rdb.StatusDisconnected {
		tds.syncClose.Unlock()
		return connectionNotOpenError
	}
	if tds.status != rdb.StatusReady {
		tds.syncClose.Unlock()
		return connectionInUseError
	}
	tds.status = rdb.StatusQuery
	tds.syncClose.Unlock()

	if debugToken {
		fmt.Printf("SQL: %q\n", sql)
	}

	var err error
	if len(params) == 0 {
		err = tds.sendSimpleQuery(ctx, sql, tds.resetNext)
	} else {
		err = tds.sendRpc(ctx, sql, truncValue, params, tds.resetNext)
	}
	tds.resetNext = false
	if err != nil {
		return err
	}

	return tds.scan()
}

const (
	sp_ExecuteSql = 10
	sp_Execute    = 12
)

var rpcHeaderParam = &rdb.Param{
	Type:   rdb.Text,
	Length: 0,
}

func (tds *Connection) sendSimpleQuery(ctx context.Context, sql string, reset bool) error {
	w := tds.pw
	err := w.BeginMessage(ctx, packetSqlBatch, reset)
	if err != nil {
		return err
	}

	w.WriteBuffer(tds.getAllHeaders())
	w.WriteBuffer(uconv.Encode.FromString(sql))
	return w.EndMessage()
}

func (tds *Connection) sendRpc(ctx context.Context, sql string, truncValue bool, params []rdb.Param, reset bool) error {
	// To make a SQL Query with params:
	// * RPC Param 1 = {Name: "", Type: NText, Field: SqlQuery}
	// * RPC Param 2 = {Name: "", Type: NText, Field: "@MySqlParam1 int,@Foo varchar(400)"}
	// * RPC Param 3 = {Name: "@MySqlParam1", Type: Int, Field: value}
	// * RPC Param 4 = {Name: "@Foo", Type: VarChar, Field: value}
	// Simple! Once figured out.

	tds.params = params
	isProc := strings.IndexAny(sql, " \t\r\n") < 0
	withRecomp := false

	// collation := []byte{0x09, 0x04, 0xD0, 0x00, 0x34}

	var procID uint16 = sp_ExecuteSql

	w := tds.pw
	err := w.BeginMessage(ctx, packetRpc, reset)
	if err != nil {
		return err
	}

	var options uint16 = 0
	if withRecomp {
		options = 1
	}

	w.WriteBuffer(tds.getAllHeaders())

	if !isProc {
		w.WriteUint16(0xffff) // ProcIDSwitch
		w.WriteUint16(procID)
		w.WriteUint16(options) // 16 bits (2 bytes) - Options: fWithRecomp, fNoMetaData, fReuseMetaData, 13FRESERVEDBIT

		paramNames := &bytes.Buffer{}
		for i := range params {
			param := &params[i]
			if i != 0 {
				paramNames.WriteRune(',')
			}
			if len(param.Name) == 0 {
				return fmt.Errorf("missing parameter name at index: %d", i)
			}

			st, found := sqlTypeLookup[param.Type]
			if !found {
				return fmt.Errorf("param %q type not found: %d", param.Name, param.Type)
			}
			fmt.Fprintf(paramNames, "@%s %s", param.Name, st.TypeString(param))
		}
		err = encodeParam(w, truncValue, tds.ProtocolVersion, rpcHeaderParam, []byte(sql))
		if err != nil {
			return err
		}
		err = encodeParam(w, truncValue, tds.ProtocolVersion, rpcHeaderParam, paramNames.Bytes())
		if err != nil {
			return err
		}
	} else {
		w.WriteUint16(uint16(len(sql))) // ProcIDSwitch
		w.WriteBuffer(uconv.Encode.FromString(sql))
		w.WriteUint16(options)
	}

	// Other parameters.
	for i := range params {
		param := &params[i]
		err = encodeParam(w, truncValue, tds.ProtocolVersion, param, param.Value)
		if err != nil {
			return err
		}
	}
	w.WriteByte(0xFF)

	return w.EndMessage()
}

func (tds *Connection) getSingleResponse(m *MessageReader, reportRow bool) (response interface{}, err error) {
	var bb []byte

	if debugToken {
		fmt.Printf("getSingleResponse\n")
		defer func() {
			fmt.Printf("<getSingleResponse MSG %[1]T : %[1]v (peek: 0x%[2]X)\n", response, tds.peek)
		}()
	}

	defer func() {
		if recovered := recover(); recovered != nil {
			if re, is := recovered.(recoverError); is {
				if re.err == io.EOF {
					response = MsgEom{}
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
	assignPeek := func() {
		tds.peek = tdsToken(read(1)[0])
		if tds.peek == 0 {
			panic("bad peek")
			// panic(recoverError{err: errors.New("bad peek")})
		}
	}
	var token tdsToken
	if tds.peek == 0 {
		token = tdsToken(read(1)[0])
		if token == 0 {
			return nil, errors.New("bad token, is zero")
		}
	} else {
		token = tds.peek
		tds.peek = 0
	}

	switch token {
	case tokenInfo:
		fallthrough
	case tokenError:
		tp := rdb.SqlError
		if token == tokenInfo {
			tp = rdb.SqlInfo
		}
		sqlMsg := &rdb.Message{
			Type: tp,
		}
		_ = binary.LittleEndian.Uint16(read(2)) // length
		sqlMsg.Number = int32(binary.LittleEndian.Uint32(read(4)))
		state := read(1)[0]
		class := read(1)[0]

		_, msg := uconv.Decode.Prefix2(read)
		sqlMsg.Message = msg
		_, sqlMsg.ServerName = uconv.Decode.Prefix1(read)
		_, sqlMsg.ProcName = uconv.Decode.Prefix1(read)
		sqlMsg.LineNumber = int32(binary.LittleEndian.Uint32(read(4)))
		sqlMsg.State = state
		sqlMsg.Class = class

		assignPeek()
		return sqlMsg, nil
	case tokenColumnMetaData:
		var columns []*SqlColumn
		count := int(binary.LittleEndian.Uint16(read(2)))
		if count == 0xffff {
			count = 0
		}
		for i := 0; i < count; i++ {
			column := decodeColumnInfo(read)
			if column.info.Table {
				parts := read(1)[0]
				for pi := byte(0); pi < parts; pi++ {
					uconv.Decode.Prefix2(read)
				}
			}
			_, column.Name = uconv.Decode.Prefix1(read)
			column.Index = i
			columns = append(columns, column)
		}

		assignPeek()

		tds.col = columns
		cc := make([]*rdb.Column, len(tds.col))
		for i, dsc := range tds.col {
			cc[i] = &dsc.Column
		}
		tds.val.Columns(cc)

		return MsgColumn{}, nil
	case tokenReturnStatus:
		return MsgRpcResult(binary.LittleEndian.Uint32(read(4))), nil
	case tokenDoneProc:
		fallthrough
	case tokenDoneInProc:
		fallthrough
	case tokenDone:
		msg := MsgDone{
			StatusCode: binary.LittleEndian.Uint16(read(2)),
			CurrentCmd: binary.LittleEndian.Uint16(read(2)),
			Rows:       binary.LittleEndian.Uint64(read(8)),
		}
		if msg.StatusCode == 0 {
			return MsgFinalDone{}, nil
		}
		const (
			doneError       = 0x2
			doneAttn        = 0x20
			doneServerError = 0x100
		)
		if msg.StatusCode&doneAttn != 0 || msg.StatusCode&doneServerError != 0 || msg.StatusCode&doneError != 0 {
			return MsgCancel{}, nil
		}
		assignPeek()
		if msg.StatusCode&0x10 != 0 {
			return MsgRowCount{Count: msg.Rows}, nil
		}
		return msg, nil
	case tokenRow:
		for _, column := range tds.col {
			tds.decodeFieldValue(read, column, tds.val.WriteField, reportRow)
		}

		assignPeek()
		return MsgRow{}, nil
	case tokenNBCRow:
		bitlen := (len(tds.col) + 7) / 8
		nulls := read(bitlen)
		for i, column := range tds.col {
			if nulls[i/8]&(1<<(uint(i)%8)) != 0 {
				err = tds.val.WriteField(&column.Column, &rdb.DriverValue{
					Null: true,
				}, nil)
				if err != nil {
					panic(recoverError{err: err})
				}
				continue
			}
			tds.decodeFieldValue(read, column, tds.val.WriteField, reportRow)
		}

		assignPeek()
		return MsgRow{}, nil
	case tokenOrder:
		// Just read the token.
		length := binary.LittleEndian.Uint16(read(2)) / 2
		var order MsgOrder = make([]uint16, length)
		for i := uint16(0); i < length; i++ {
			order[i] = binary.LittleEndian.Uint16(read(2))
		}
		assignPeek()
		return order, nil
	case tokenEnvChange:
		length := int(binary.LittleEndian.Uint16(read(2)) - 1)
		tokenType := read(1)[0] // Token Type
		switch tokenType {
		case 8, 9, 10: // 8: begin, 9: commit, 10: rollback.
			buf := read(length)
			switch buf[0] {
			case 0:
				tds.currentTransaction = 0
			case 8:
				tds.currentTransaction = binary.LittleEndian.Uint64(buf[1:])
			default:
				return nil, fmt.Errorf("Unknown length: %d", buf[0])
			}
		case 15:
			// Type 15 doesn't obey the length.
			return nil, fmt.Errorf("Un-handled env-change type: %d", tokenType)
		case 18:
			if debugToken {
				fmt.Printf("\tRESETCONNECTION\n")
			}
			read(length)
		default:
			read(length)
		}
		// Currently ignore all the data.

		assignPeek()
		return MsgEnvChange{}, nil
	case tokenReturnValue:
		//ParamOrdinal ushort
		//ParamName B_VARCHAR
		//Status BYTE
		//UserType ULONG
		//Flags 2 BYTES
		//TypeInfo TYPE_INFO
		//Value TYPE_VARBYTE

		paramIndex := binary.LittleEndian.Uint16(read(2))
		_, paramName := uconv.Decode.Prefix1(read)
		status := read(1)[0]
		switch status {
		case 0x01:
		// Output param.
		case 0x02:
		// User defined function.
		default:
			panic(recoverError{fmt.Errorf("Unknown status value: 0x%X", status)})
		}

		col := decodeColumnInfo(read)
		col.Name = paramName
		col.Index = int(paramIndex)

		outValue := rdb.Nullable{}

		wf := func(col *rdb.Column, value *rdb.DriverValue, assign rdb.Assigner) error {
			outValue.Value = value.Value
			outValue.Null = value.Null
			return nil
		}
		tds.decodeFieldValue(read, col, wf, true)

		err := rdb.AssignValue(&col.Column, outValue, tds.params[col.Index].Value, nil)
		if err != nil {
			return nil, err
		}

		assignPeek()

		return MsgParamValue{}, nil
	default:
		return nil, fmt.Errorf("Unknown response code: 0x%X", token)
	}
}

func getHeaderTemplate() ([]byte, int) {
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

	tranNumberOffset := at
	binary.LittleEndian.PutUint64(bb[at:], 0)
	at += 8

	binary.LittleEndian.PutUint32(bb[at:], 1)
	at += 4

	return bb, tranNumberOffset
}

func withLock(lk sync.Locker, f func()) {
	lk.Lock()
	defer lk.Unlock() // For panics.

	f()
}
