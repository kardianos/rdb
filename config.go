// Copyright 2014 Daniel Theophanes.
// Use of this source code is governed by a zlib-style
// license that can be found in the LICENSE file.

package rdb

import (
	"bitbucket.org/kardianos/rdb/semver"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// Database connection configuration.
// Drivers may have additional properties held in KV.
// If a driver is file based, the file name should be in the "Instance" field.
type Config struct {
	DriverName string

	Username string
	Password string
	Hostname string
	Port     int
	Instance string
	Database string // Initial database to connect to.

	// Timeout time for connection dial.
	// Zero for no timeout.
	DialTimeout time.Duration

	// Time for an idle connection to be closed.
	// Zero if there should be no timeout.
	PoolIdleTimeout time.Duration

	// How many connection should be created at startup.
	// Valid range is (0 < init, init <= max).
	PoolInitCapacity int

	// Max number of connections to create.
	// Valid range is (0 < max).
	PoolMaxCapacity int

	KV map[string]interface{}

	// An ordered list of input converters.
	InputConverters []ConvertInput

	// A per-SqlType convertion for result ouput fields.
	OuputConverters map[SqlType]ConvertOutput
}

// Provides a standard method to parse configuration options from a text.
// The instance field can also hold the filename in case of a file based connection.
//   driver://[username:password@][url[:port]]/[Instance]?db=mydatabase&opt1=valA&opt2=valB
//   sqlite:///C:/folder/file.sqlite3?opt1=valA&opt2=valB
//   sqlite:///srv/folder/file.sqlite3?opt1=valA&opt2=valB
//   ms://TESTU@localhost/SqlExpress?db=master&dial_timeout=3s
// This will attempt to find the driver to load additional parameters.
//   Additional field options:
//      db=<string>:                  Database
//      dial_timeout=<time.Duration>: DialTimeout
//      init_cap=<int>:               PoolInitCapacity
//      max_cap=<int>:                PoolMaxCapacity
//      idle_timeout=<time.Duration>: PoolIdleTimeout
func ParseConfig(connectionString string) (*Config, error) {
	u, err := url.Parse(connectionString)
	if err != nil {
		return nil, err
	}
	var user, pass string
	if u.User != nil {
		user = u.User.Username()
		pass, _ = u.User.Password()
	}
	port := 0
	host := ""

	if len(u.Host) > 0 {
		hostPort := strings.Split(u.Host, ":")
		host = hostPort[0]
		if len(hostPort) > 1 {
			parsedPort, err := strconv.ParseUint(hostPort[1], 10, 16)
			if err != nil {
				return nil, err
			}
			port = int(parsedPort)
		}
	}

	conf := &Config{
		DriverName: u.Scheme,
		Username:   user,
		Password:   pass,
		Hostname:   host,
		Port:       port,
	}

	val := u.Query()

	conf.Database = val.Get("db")

	if st := val.Get("dial_timeout"); len(st) != 0 {
		conf.DialTimeout, err = time.ParseDuration(st)
		if err != nil {
			return nil, err
		}
	}
	if st := val.Get("idle_timeout"); len(st) != 0 {
		conf.PoolIdleTimeout, err = time.ParseDuration(st)
		if err != nil {
			return nil, err
		}
	}
	if st := val.Get("init_cap"); len(st) != 0 {
		conf.PoolInitCapacity, err = strconv.Atoi(st)
		if err != nil {
			return nil, err
		}
	}
	if st := val.Get("max_cap"); len(st) != 0 {
		conf.PoolMaxCapacity, err = strconv.Atoi(st)
		if err != nil {
			return nil, err
		}
	}

	if len(u.Path) > 0 {
		conf.Instance = u.Path[1:]
	}

	// Now attempt to call specific driver and parse Key-Value options.
	dr, err := getDriver(conf.DriverName)
	if err != nil {
		return conf, err
	}
	dr.ParseOptions(conf.KV, val)

	return conf, nil
}

type DriverOption struct {
	Name string

	Description  string
	DefaultValue interface{}
}

type DriverSupport struct {
	NamedParameter   bool // Supports named parameters.
	FluidType        bool // Like SQLite.
	MultipleResult   bool // Supports returning multiple result sets.
	SecureConnection bool // Supports a secure connection.
	BulkInsert       bool // Supports a fast bulk insert method.
	Notification     bool // Supports driver notifications.
	UserDataTypes    bool // Handles user supplied data types.
}

type DriverMeta struct {
	Options []*DriverOption
	DriverSupport
}

type ConnectionInfo struct {
	Server, Protocol *semver.Version
}
