// Copyright 2014 Daniel Theophanes.
// Use of this source code is governed by a zlib-style
// license that can be found in the LICENSE file.

package ms

import (
	"fmt"
	"net"

	"bitbucket.org/kardianos/rdb"
	"bitbucket.org/kardianos/rdb/ms/ssrp"
)

func init() {
	rdb.Register("ms", &Driver{})
}

type Driver struct{}

func (dr *Driver) Open(c *rdb.Config) (rdb.DriverConn, error) {
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
	var conn net.Conn
	var err error

	addr := fmt.Sprintf("%s:%d", hostname, port)
	if c.DialTimeout == 0 {
		conn, err = net.Dial("tcp", addr)
	} else {
		conn, err = net.DialTimeout("tcp", addr, c.DialTimeout)
	}
	if err != nil {
		return nil, err
	}

	tds := NewConnection(conn)

	_, err = tds.Open(c)
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
			MultipleResult:   false,
			SecureConnection: false,
			BulkInsert:       false,
			Notification:     false,
			UserDataTypes:    false,
		},
	}
}

var pingCommand = &rdb.Command{
	Sql:   "select top 0 1;",
	Arity: rdb.ZeroMust,
}

func (db *Driver) PingCommand() *rdb.Command {
	return pingCommand
}
