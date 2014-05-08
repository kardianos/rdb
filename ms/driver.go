// Copyright 2014 Daniel Theophanes.
// Use of this source code is governed by a zlib-style
// license that can be found in the LICENSE file.

package ms

import (
	"bitbucket.org/kardianos/rdb"
	"fmt"
	"net"
	"net/url"
)

func init() {
	rdb.Register("ms", &Driver{})
}

type Driver struct{}

func (dr *Driver) Open(c *rdb.Config) (rdb.Conn, error) {
	port := 1433
	if c.Port != 0 {
		port = c.Port
	}
	hostname := "localhost"
	if len(c.Hostname) != 0 && c.Hostname != "." {
		hostname = c.Hostname
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
func (dr *Driver) DriverMetaInfo() *rdb.DriverMeta {
	return &rdb.DriverMeta{
		DriverSupport: rdb.DriverSupport{
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

func (dr *Driver) ParseOptions(KV map[string]interface{}, configOptions url.Values) error {
	return nil
}

var pingCommand = &rdb.Command{
	Sql:   "select top 0 1;",
	Arity: rdb.ZeroMust,
}

func (db *Driver) PingCommand() *rdb.Command {
	return pingCommand
}
