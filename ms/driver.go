// Copyright 2014 Daniel Theophanes.
// Use of this source code is governed by a zlib-style
// license that can be found in the LICENSE file.

package ms

import (
	"context"
	"net"
	"strconv"
	"time"

	"github.com/kardianos/rdb"
	"github.com/kardianos/rdb/ms/ssrp"
)

func init() {
	rdb.Register("ms", &Driver{})
}

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
