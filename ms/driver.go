// Copyright 2014 Daniel Theophanes.
// Use of this source code is governed by a zlib-style
// license that can be found in the LICENSE file.

package ms

import (
	"context"
	"crypto/tls"
	"errors"
	"io"
	"net"
	"strconv"
	"sync"
	"syscall"
	"time"

	"github.com/kardianos/rdb"
	"github.com/kardianos/rdb/ms/ssrp"
)

func init() {
	rdb.Register("ms", &Driver{})
}

// tds8Support tracks which servers support TDS 8.0.
// Key is "host:port", value is true if TDS 8.0 is NOT supported (failed before).
var (
	tds8Unsupported   = make(map[string]bool)
	tds8UnsupportedMu sync.RWMutex
)

type Driver struct{}

func (dr *Driver) Open(ctx context.Context, c *rdb.Config) (rdb.DriverConn, error) {
	hostname := c.Hostname
	if len(c.Hostname) == 0 || c.Hostname == "." {
		hostname = "localhost"
	}

	port := c.Port
	if c.Port == 0 {
		ii, err := ssrp.FetchInstanceInfo(hostname, c.Instance)
		if err == nil {
			port = ii.Tcp
		} else {
			port = 1433
		}
	}
	d := net.Dialer{
		Timeout: c.DialTimeout,
		KeepAliveConfig: net.KeepAliveConfig{
			Enable: true,
			// Suggested defaults be TDS protocol.
			Idle:     30 * time.Second,
			Interval: 1 * time.Second,

			Count: 3,
		},
	}
	addr := net.JoinHostPort(hostname, strconv.FormatInt(int64(port), 10))

	// Check for TDS8-specific config options.
	tds8Only := false
	tds8Disable := false
	if v, ok := c.KV["tds8"]; ok {
		switch v := v.(type) {
		case bool:
			tds8Only = v
		case string:
			switch v {
			case "only":
				tds8Only = true
			case "disable":
				tds8Disable = true
			}
		}
	}

	// Check if we already know this server doesn't support TDS 8.0.
	// This cache is only used for auto-detection, not for explicit tds8=only mode.
	tds8UnsupportedMu.RLock()
	serverNoTDS8 := tds8Unsupported[addr]
	tds8UnsupportedMu.RUnlock()

	// Try TDS 8.0 first if:
	// - tds8=only is set (always try, ignore cache), OR
	// - Secure mode is requested AND server not known to lack TDS 8.0 support
	// Skip TDS 8.0 if:
	// - Encryption is disabled, OR
	// - tds8=disable is set
	tryTDS8 := (tds8Only || (c.Secure && !serverNoTDS8)) && !c.InsecureDisableEncryption && !tds8Disable

	if tryTDS8 {
		conn, err := d.DialContext(ctx, "tcp", addr)
		if err != nil {
			return nil, err
		}

		tds := NewConnection(conn, c.ResetConnectionTimeout, c.RollbackTimeout)
		_, err = tds.OpenTDS8(ctx, c)
		if err == nil {
			return tds, nil
		}
		tds.Close()

		// If tds8=only, don't fall back to TDS 7.x.
		if tds8Only {
			return nil, err
		}

		// Check if this is a TLS protocol error (server doesn't support TDS 8.0).
		// If so, remember this and fall back to TDS 7.x. Otherwise, return the error.
		if !isTLSProtocolError(err) {
			return nil, err
		}

		// Remember that this server doesn't support TDS 8.0.
		tds8UnsupportedMu.Lock()
		tds8Unsupported[addr] = true
		tds8UnsupportedMu.Unlock()

		// Fall through to TDS 7.x.
	}

	conn, err := d.DialContext(ctx, "tcp", addr)
	if err != nil {
		return nil, err
	}

	tds := NewConnection(conn, c.ResetConnectionTimeout, c.RollbackTimeout)

	_, err = tds.Open(ctx, c)
	if err != nil {
		tds.Close()
		return nil, err
	}

	return tds, nil
}

// isTLSProtocolError checks if the error indicates the server doesn't support TDS 8.0.
// This happens when the server expects PRELOGIN instead of TLS ClientHello.
func isTLSProtocolError(err error) bool {
	if err == nil {
		return false
	}

	// Check for io.EOF - server closed connection because it didn't expect TLS.
	if errors.Is(err, io.EOF) {
		return true
	}
	if errors.Is(err, io.ErrUnexpectedEOF) {
		return true
	}

	// Check for network errors (connection reset, refused, etc.)
	var netErr *net.OpError
	if errors.As(err, &netErr) {
		// Check for syscall errors like ECONNRESET, ECONNREFUSED
		var syscallErr syscall.Errno
		if errors.As(netErr.Err, &syscallErr) {
			switch syscallErr {
			case syscall.ECONNRESET, syscall.ECONNREFUSED, syscall.ECONNABORTED:
				return true
			}
		}
		return true
	}

	// Check for TLS-specific record layer errors
	var tlsRecordErr tls.RecordHeaderError
	if errors.As(err, &tlsRecordErr) {
		return true
	}

	return false
}
func (dr *Driver) DriverInfo() *rdb.DriverInfo {
	return &rdb.DriverInfo{
		DriverSupport: rdb.DriverSupport{
			PreparePerConn: false,

			NamedParameter:   true,
			FluidType:        false,
			MultipleResult:   true,
			SecureConnection: true,
			BulkInsert:       false,
			Notification:     false,
			UserDataTypes:    false,
		},
	}
}

var pingCommand = &rdb.Command{
	SQL:   "select top 0 1;",
	Arity: rdb.ZeroMust,
}

func (db *Driver) PingCommand() *rdb.Command {
	return pingCommand
}
