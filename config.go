package rdb

import (
	"bitbucket.org/kardianos/rdb/semver"
	"net/url"
	"strconv"
	"strings"
)

type Config struct {
	DriverName string

	Username string
	Password string
	Hostname string
	Port     int
	Instance string
	Database string

	KV map[string]interface{}
}

// Provides a standard method to parse configuration options from a text.
// The instance field can also hold the filename in case of a file based connection.
//   driver://[username:password@][url[:port]]/[Instance]?db=mydatabase&opt1=valA&opt2=valB
//   sqlite:///C:/folder/file.sqlite3?opt1=valA&opt2=valB
//   sqlite:///srv/folder/file.sqlite3?opt1=valA&opt2=valB
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

	val := u.Query()
	db := val.Get("db")
	val.Del("db")
	instance := ""
	if len(u.Path) > 0 {
		instance = u.Path[1:]
	}

	// TODO: Now attempt to call specific driver and parse Key-Value options.

	return &Config{
		DriverName: u.Scheme,
		Username:   user,
		Password:   pass,
		Hostname:   host,
		Port:       port,
		Instance:   instance,
		Database:   db,
	}, nil
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
