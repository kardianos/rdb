// Copyright (c) 2011, The pg Authors. All Rights Reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package pg

import (
	"bufio"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"net"
	"time"

	"bitbucket.org/kardianos/rdb"
)

var ErrSecureNotSupported = fmt.Errorf("Secure connection is not supported.")

func errUnhandledMessage(at string, msg interface{}) error {
	return fmt.Errorf("At %s, unhandled response code: %T", at, msg)
}

type driver struct{}

func (d *driver) Open(conf *rdb.Config) (rdb.DriverConn, error) {
	port := 5432
	if conf.Port != 0 {
		port = conf.Port
	}
	hostname := "localhost"
	if len(conf.Hostname) != 0 && conf.Hostname != "." {
		hostname = conf.Hostname
	}

	var conn net.Conn
	var err error

	addr := fmt.Sprintf("%s:%d", hostname, port)
	if conf.DialTimeout == 0 {
		conn, err = net.Dial("tcp", addr)
	} else {
		conn, err = net.DialTimeout("tcp", addr, conf.DialTimeout)
	}
	if err != nil {
		return nil, err
	}

	pg := &connection{
		conn: conn,

		serverStatus: make(map[string]string, 5),
	}

	options := map[string]string{
		"client_encoding": "UTF8",
		"datestyle":       "ISO, MDY",

		"user":     conf.Username,
		"database": conf.Database,
	}

	pg.readBuffer = bufio.NewReader(pg.conn)
	pg.writeBuffer = bufio.NewWriter(pg.conn)

	/*
		if conf.Secure {
			tlsConf := tls.Config{
				InsecureSkipVerify: conf.InsecureSkipVerify,
			}

			// TODO: Handle these errors.
			pg.writeByte(0)
			pg.writeInt32(80877103)
			pg.writeFlush()

			any, err := pg.getMessage()
			if err != nil {
				return nil, err
			}
			switch any.(type) {
			case MsgParameterStatus:
			default:
				return nil, ErrSecureNotSupported
			}

			pg.conn = tls.Client(pg.conn, &tlsConf)
		}
	*/

	// Start startup.
	write := pg.writer()
	write.Msg(0)
	write.Int32(0x30000)

	// Send startup parameters.
	for k, v := range options {
		write.String(k)
		write.String(v)
	}
	write.String("")
	write.MsgDone()
	err = write.Send()
	if err != nil {
		return nil, err
	}

loop:
	for {
		any, err := pg.getMessage()
		if err != nil {
			return nil, err
		}
		switch msg := any.(type) {
		case MsgBackendKeyData:
		case MsgParameterStatus:
		case MsgReadyForQuery:
			pg.open = true
			pg.tranStatus = transactionStatus(msg.TransactionStatus)
			break loop
		case MsgAuthenticationOk:
			// TODO: Note in connection status authentication status perhaps.
		case MsgAuthenticationCleartextPassword:
			write.Msg(tokenPasswordMessage)
			write.String(conf.Password)
			write.MsgDone()
			err = write.Send()
			if err != nil {
				return nil, err
			}
		case MsgAuthenticationMD5Password:
			write.Msg(tokenPasswordMessage)
			write.String(md5AuthDigest(conf.Username, conf.Password, msg.Salt))
			write.MsgDone()
			err = write.Send()
			if err != nil {
				return nil, err
			}

		case MsgErrorResponse:
			return nil, msg
		default:
			return nil, errUnhandledMessage("driver.Open", msg)
		}
	}
	// End startup.

	// Reset any deadline.
	err = pg.conn.SetDeadline(time.Time{})
	return pg, err
}
func md5AuthDigest(username, password string, salt []byte) string {
	hasher := md5.New()
	hasher.Write([]byte(password))
	hasher.Write([]byte(username))
	inner := hex.EncodeToString(hasher.Sum(nil))
	hasher.Reset()

	hasher.Write([]byte(inner))
	hasher.Write(salt)
	return "md5" + hex.EncodeToString(hasher.Sum(nil))
}

func (d *driver) DriverInfo() *rdb.DriverInfo {
	return &rdb.DriverInfo{
		DriverSupport: rdb.DriverSupport{
			PreparePerConn: true,

			NamedParameter:   true,
			FluidType:        false,
			MultipleResult:   false,
			SecureConnection: true,
			BulkInsert:       false,
			Notification:     false,
			UserDataTypes:    false,
		},
	}
}

var cmdPing2 = &rdb.Command{
	Sql:   "select 1 limit 0;",
	Arity: rdb.Zero,
}

func (d *driver) PingCommand() *rdb.Command {
	return cmdPing2
}

func init() {
	rdb.Register("pg", &driver{})
}
