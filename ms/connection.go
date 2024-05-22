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
	"runtime/debug"
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
	col    []*SQLColumn
	params []rdb.Param

	allHeaders            []byte
	allHeaderNumberOffset int

	currentTransaction uint64

	opened              time.Time
	defaultResetTimeout time.Duration

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
	if debugToken {
		fmt.Printf("\tOPEN\n")
	}
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
	return tds.Query(ctx, &rdb.Command{SQL: c.ResetQuery}, nil, nil, nil)
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
	return nil, rdb.ErrNotImplemented
}
func (tds *Connection) Unprepare(preparedStatementToken interface{}) (err error) {
	return rdb.ErrNotImplemented
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

	if tds.mr != nil && !tds.mr.packetEOM {
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
	case rdb.LevelReadUncommitted:
		level = levelReadUncommitted
	case rdb.LevelReadCommitted:
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

	if tds.mr != nil && !tds.mr.packetEOM {
		return fmt.Errorf("connection not ready to be re-used yet for query")
	}

	go tds.asyncWaitCancel(ctx, cmd.Name)
top:
	select {
	default:
		// Nothing
	case <-ctx.Done():
		return ctx.Err()
	}

	moreExec, err := tds.execute(ctx, cmd, params)
	if err != nil {
		return err
	}
	tds.syncClose.Lock()
	doNext := tds.status == rdb.StatusQuery && err == nil
	tds.syncClose.Unlock()

	if doNext {
		_, err = tds.nextResult()
	}

	if moreExec {
		goto top
	}

	return err
}

func (tds *Connection) asyncWaitCancel(ctx context.Context, sqlName string) {
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

	var lastMessage *rdb.Message
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
			lastMessage = v
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
			if v.IsAttention {
				return rdb.ErrCancel
			}
			if lastMessage != nil {
				return rdb.Errors{lastMessage}
			}
			if v.IsServerError {
				return fmt.Errorf("unknown server error, check messages")
			}
			return fmt.Errorf("unknown error, check messages")
		}
		if tds.col == nil {
			continue
		}
		pb, err := tds.mr.PeekByte()
		switch tdsToken(pb) {
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

func (tds *Connection) execute(ctx context.Context, cmd *rdb.Command, params []rdb.Param) (more bool, err error) {
	tds.syncClose.Lock()

	if tds.status == rdb.StatusDisconnected {
		tds.syncClose.Unlock()
		return false, connectionNotOpenError
	}
	if tds.status != rdb.StatusReady {
		tds.syncClose.Unlock()
		return false, connectionInUseError
	}
	tds.status = rdb.StatusQuery
	tds.syncClose.Unlock()

	if debugToken {
		if cmd.Bulk != nil {
			fmt.Printf("BULK\n")
		} else {
			fmt.Printf("SQL: %q\n", cmd.SQL)
		}
	}

	tds.mr = tds.pr.BeginMessage(packetTabularResult)

	switch {
	default:
		err = tds.sendSimpleQuery(ctx, cmd.SQL, tds.resetNext)
	case cmd.Bulk != nil:
		bulk := cmd.Bulk
		prefixSQL := cmd.SQL
		if len(prefixSQL) == 0 || len(params) == 0 {
			startSQL, startParams, err := bulk.Start()
			if err != nil {
				return more, err
			}
			if len(prefixSQL) == 0 {
				prefixSQL = startSQL
			}
			if len(params) == 0 {
				params = startParams
			}
		}
		if len(prefixSQL) > 0 {
			err = tds.sendSimpleQuery(ctx, prefixSQL, tds.resetNext)
			if err != nil {
				return more, err
			}
			var resp interface{}
			withLock(&tds.syncClose, func() {
				resp, err = tds.getSingleResponse(tds.mr, false)
				if _, ok := resp.(MsgFinalDone); ok && !more {
					tds.mr.packetEOM = false
				}
			})
			if err != nil {
				return more, err
			}
		}
		if len(params) == 0 {
			return more, fmt.Errorf("missing params for bulk insert")
		}
		more, err = tds.sendBulk(ctx, cmd.Bulk, cmd.TruncLongText, params, false)
	case len(params) > 0:
		err = tds.sendRPC(ctx, cmd.SQL, cmd.TruncLongText, params, tds.resetNext)
	}
	tds.resetNext = false
	if err != nil {
		return more, err
	}

	return more, tds.scan()
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

func (tds *Connection) sendRPC(ctx context.Context, sql string, truncValue bool, params []rdb.Param, reset bool) error {
	// To make a SQL Query with params:
	// * RPC Param 1 = {Name: "", Type: NText, Field: SqlQuery}
	// * RPC Param 2 = {Name: "", Type: NText, Field: "@MySqlParam1 int,@Foo varchar(400)"}
	// * RPC Param 3 = {Name: "@MySqlParam1", Type: Int, Field: value}
	// * RPC Param 4 = {Name: "@Foo", Type: VarChar, Field: value}
	// Simple! Once figured out.

	tds.params = params
	isProc := !strings.ContainsAny(sql, " \t\r\n")
	withRecomp := false

	var procID uint16 = sp_ExecuteSql

	w := tds.pw
	err := w.BeginMessage(ctx, packetRPC, reset)
	if err != nil {
		return err
	}
	w.WriteBuffer(tds.getAllHeaders())

	var options uint16 = 0
	if withRecomp {
		options = 1
	}

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
	w.WriteByte(byte(tokenDoneInProc))

	return w.EndMessage()
}

func (tds *Connection) sendBulk(ctx context.Context, bulk rdb.Bulk, truncValue bool, params []rdb.Param, reset bool) (more bool, err error) {
	w := tds.pw
	err = w.BeginMessage(ctx, packetBulkLoad, reset)
	if err != nil {
		return false, err
	}
	var complete bool
	defer func() {
		if complete {
			return
		}
		w.EndMessage()
	}()

	w.WriteByte(byte(tokenColumnMetaData))
	w.WriteUint16(uint16(len(params)))
	tdsVer := tds.ProtocolVersion
	// Write column metadata.
	meta := make([]paramTypeInfo, len(params))
	for i, p := range params {
		// UserType ULONG
		// Flags
		// TYPE_INFO
		// ColName (B_VARCHAR)

		var userType uint32
		w.WriteUint32(userType)
		flags := colFlags{}
		w.Write(colFlagsToSlice(flags))

		ti, err := getParamTypeInfo(tdsVer, p.Type)
		if err != nil {
			return false, err
		}
		meta[i] = ti
		err = encodeType(w, ti, &p)
		if err != nil {
			return false, err
		}
		nameU16 := uconv.Encode.FromString(p.Name)
		l := len(nameU16) / 2
		if l > 0xff {
			return false, fmt.Errorf("parameter name too long %q", p.Name)
		}
		w.WriteByte(byte(l))
		w.WriteBuffer(nameU16)
	}

	var ct int
loop:
	for {
		err = ctx.Err()
		if err != nil {
			return false, err
		}
		err = bulk.Next(ct, params)
		if err != nil {
			switch err {
			default:
				return false, err
			case io.EOF:
				break loop
			case rdb.ErrBulkSkip:
				continue loop
			case rdb.ErrBulkBatchDone:
				more = true
				break loop
			}
		}
		// Write column data.
		ct++
		w.WriteByte(byte(tokenRow))
		for i, p := range params {
			ti := meta[i]
			err = encodeValue(w, ti, &p, truncValue, p.Value)
			if err != nil {
				return false, err
			}
		}
	}
	complete = true
	w.WriteByte(byte(tokenDone))
	w.WriteUint16(0x10)       // Status.
	w.WriteUint16(0)          // Current Command.
	w.WriteUint64(uint64(ct)) // Row Count.

	return more, w.EndMessage()
}

func (tds *Connection) getSingleResponse(m *MessageReader, reportRow bool) (response interface{}, err error) {
	if debugToken {
		fmt.Printf("getSingleResponse\n")
		defer func() {
			fmt.Printf("<getSingleResponse MSG %[1]T : %[1]v\n", response)
		}()
	}

	defer func() {
		if recovered := recover(); recovered != nil {
			if re, is := recovered.(recoverError); is {
				err = re.err
				return
			}
			panic(fmt.Errorf("getSingleResponse panic: %v\n%s", recovered, debug.Stack()))
		}
	}()

	var bb []byte
	read := func(n int) []byte {
		var readErr error
		bb, readErr = m.Fetch(n)
		if len(bb) > 0 {
			return bb
		}
		if readErr != nil {
			panic(recoverError{err: readErr})
		}
		return bb
	}
	tokenBuf, err := m.Fetch(1)
	if err != nil {
		if len(tokenBuf) != 1 && err == io.EOF {
			return MsgEom{}, nil
		}
		return nil, err
	}
	token := tdsToken(tokenBuf[0])
	if token == 0 {
		return nil, errors.New("bad token, is zero")
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

		return sqlMsg, nil
	case tokenColumnMetaData:
		var columns []*SQLColumn
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
			return MsgCancel{
				IsAttention:   msg.StatusCode&doneAttn != 0,
				IsServerError: msg.StatusCode&doneServerError != 0,
				IsError:       msg.StatusCode&doneError != 0,
			}, nil
		}
		if msg.StatusCode&0x10 != 0 {
			return MsgRowCount{Count: msg.Rows}, nil
		}
		return msg, nil
	case tokenRow:
		for _, column := range tds.col {
			tds.decodeFieldValue(read, column, tds.val.WriteField, reportRow)
		}

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

		return MsgRow{}, nil
	case tokenOrder:
		// Just read the token.
		length := binary.LittleEndian.Uint16(read(2)) / 2
		var order MsgOrder = make([]uint16, length)
		for i := uint16(0); i < length; i++ {
			order[i] = binary.LittleEndian.Uint16(read(2))
		}
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
				return nil, fmt.Errorf("unknown length: %d", buf[0])
			}
		case 15:
			// Type 15 doesn't obey the length.
			return nil, fmt.Errorf("un-handled env-change type: %d", tokenType)
		case 18:
			if debugToken {
				fmt.Printf("\tRESETCONNECTION\n")
			}
			read(length)
		default:
			read(length)
		}
		// Currently ignore all the data.

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
			panic(recoverError{fmt.Errorf("unknown status value: 0x%X", status)})
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

		if len(tds.params) <= col.Index {
			return nil, fmt.Errorf("INDEX OUT OF RANGE (params=%#v, col=%#v)", tds.params, *col)
		}

		err := rdb.AssignValue(&col.Column, outValue, tds.params[col.Index].Value, nil)
		if err != nil {
			return nil, err
		}

		return MsgParamValue{}, nil
	default:
		return nil, fmt.Errorf("unknown response code: 0x%X", token)
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
