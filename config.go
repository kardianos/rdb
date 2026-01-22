// Copyright 2014 Daniel Theophanes.
// Use of this source code is governed by a zlib-style
// license that can be found in the LICENSE file.

package rdb

import (
	"crypto/x509"
	"errors"
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

// Config Database connection configuration.
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

	// On rollback or query cancel, a special constructed context is used with this timeout.
	// If zero, a default of 30s is used.
	RollbackTimeout time.Duration

	// Max time for a connection to live.
	ConnectionMaxLifetime time.Duration

	// Time for an idle connection to be closed.
	// Zero if there should be no timeout.
	PoolIdleTimeout time.Duration

	// Time for a query to reset the connection to complete.
	// Zero if there should be no timeout.
	ResetConnectionTimeout time.Duration

	// Time to wait for a connection before expanding the connection pool.
	SoftWait time.Duration

	// How many connection should be created at startup.
	// Valid range is (0 < init, init <= max).
	PoolInitCapacity int

	// Max number of connections to create.
	// Valid range is (0 < max).
	PoolMaxCapacity int

	// Number of connections to add when expanding the pool.
	// If zero, defaults to 6.
	ExpandPoolBy int

	// Require the driver to establish a secure connection.
	Secure bool

	// Disable encryption on the connection.
	InsecureDisableEncryption bool

	// Do not require the secure connection to verify the remote host name.
	// Ignored if Secure is false.
	InsecureSkipVerify bool

	// Root Certificate Authorities for server.
	RootCAs *x509.CertPool

	// ResetQuery is executed after the connection is reset.
	ResetQuery string

	KV map[string]interface{}
}

const optPrefix = "opt_"

// ParseConfigURL provides a standard method to parse configuration options from a text.
// The instance field can also hold the filename in case of a file based connection.
//
//	driver://[username:password@][url[:port]]/[Instance]?db=mydatabase&opt1=valA&opt2=valB
//	sqlite:///C:/folder/file.sqlite3?opt1=valA&opt2=valB
//	sqlite:///srv/folder/file.sqlite3?opt1=valA&opt2=valB
//	ms://TESTU@localhost/SqlExpress?db=master&dial_timeout=3s
//
// This will attempt to find the driver to load additional parameters.
//
//	Additional field options:
//	   db=<string>:                      Database
//	   dial_timeout=<time.Duration>:     Dial Timeout
//	   max_lifetime=<time.Duration>:     Max Connection Lifetime
//	   init_cap=<int>:                   Pool Init Capacity
//	   max_cap=<int>:                    Pool Max Capacity
//	   expand_by=<int>:                  Number of connections to add when expanding pool (default 6)
//	   idle_timeout=<time.Duration>:     Pool Idle Timeout
//	   reset_timeout=<time.Duration>:    Reset Connection Timeout
//	   soft_wait=<time.Duration>:        Time to wait for connection in pool before expanding pool. Default 20ms.
//	   rollback_timeout=<time.Duration>: Rollback or cancel connection Timeout.
//	   require_encryption=<bool>:        Require Connection Encryption
//	   disable_encryption=<bool>:        Disable Connection Encryption
//	   cert=<string>:                    Load the cert file as root CA, repeatable.
//	                                     SQL Server doens't send intermediate certificates.
//	                                     May be required even if root CA is known and trusted.
//	   insecure_skip_verify=<bool>:      INSECURE. Skip  encryption certificate verification.
//	   opt_<any>=<any>:                  include values, unchecked here, into KV. "opt_" prefix is stripped.
func ParseConfigURL(connectionString string) (*Config, error) {
	if len(connectionString) == 0 {
		return nil, errors.New("empty DSN")
	}
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
		KV:         map[string]interface{}{},
	}

	val := u.Query()
	for key, vv := range u.Query() {
		if len(vv) == 0 {
			return nil, fmt.Errorf("invalid setting: %v", key)
		}
		allowMultiple := false
		v0 := vv[0]
		switch key {
		default:
			if !strings.HasPrefix(key, optPrefix) {
				return nil, fmt.Errorf("unknown setting: %v", key)
			}
			key := strings.TrimPrefix(key, optPrefix)
			conf.KV[key] = v0
		case "db":
			conf.Database = v0
		case "dial_timeout":
			conf.DialTimeout, err = time.ParseDuration(v0)
			if err != nil {
				return nil, fmt.Errorf("DSN property %q: %w", key, err)
			}
		case "max_lifetime":
			conf.ConnectionMaxLifetime, err = time.ParseDuration(v0)
			if err != nil {
				return nil, fmt.Errorf("DSN property %q: %w", key, err)
			}
		case "idle_timeout":
			conf.PoolIdleTimeout, err = time.ParseDuration(v0)
			if err != nil {
				return nil, fmt.Errorf("DSN property %q: %w", key, err)
			}
		case "reset_timeout":
			conf.ResetConnectionTimeout, err = time.ParseDuration(v0)
			if err != nil {
				return nil, fmt.Errorf("DSN property %q: %w", key, err)
			}
		case "rollback_timeout":
			conf.RollbackTimeout, err = time.ParseDuration(v0)
			if err != nil {
				return nil, fmt.Errorf("DSN property %q: %w", key, err)
			}
		case "soft_wait":
			conf.SoftWait, err = time.ParseDuration(v0)
			if err != nil {
				return nil, fmt.Errorf("DSN property %q: %w", key, err)
			}
		case "query_timeout":
			// Ignore this.
			// All query timeouts controlled from context.
			continue
		case "init_cap":
			conf.PoolInitCapacity, err = strconv.Atoi(v0)
			if err != nil {
				return nil, fmt.Errorf("DSN property %q: %w", key, err)
			}
		case "max_cap":
			conf.PoolMaxCapacity, err = strconv.Atoi(v0)
			if err != nil {
				return nil, fmt.Errorf("DSN property %q: %w", key, err)
			}
		case "expand_by":
			conf.ExpandPoolBy, err = strconv.Atoi(v0)
			if err != nil {
				return nil, fmt.Errorf("DSN property %q: %w", key, err)
			}
		case "insecure_skip_verify":
			conf.InsecureSkipVerify, err = strconv.ParseBool(v0)
			if err != nil {
				return nil, fmt.Errorf("DSN property %q: %w", key, err)
			}
		case "require_encryption":
			conf.Secure, err = strconv.ParseBool(v0)
			if err != nil {
				return nil, fmt.Errorf("DSN property %q: %w", key, err)
			}
		case "disable_encryption":
			conf.InsecureDisableEncryption, err = strconv.ParseBool(v0)
			if err != nil {
				return nil, fmt.Errorf("DSN property %q: %w", key, err)
			}
		case "cert":
			allowMultiple = true
			certs := x509.NewCertPool()
			for index, v := range vv {
				b, err := os.ReadFile(v)
				if err != nil {
					return nil, fmt.Errorf("DSN property %q[%d], cert %q: %w", key, index, v, err)
				}
				ok := certs.AppendCertsFromPEM(b)
				if !ok {
					return nil, fmt.Errorf("DSN property %q[%d], cert %q: failed to append cert", key, index, v)
				}
			}
			conf.RootCAs = certs
		}
		if !allowMultiple && len(vv) > 1 {
			return nil, fmt.Errorf("DSN property %q must not be repeated", key)
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
	meta := dr.DriverInfo()
	for _, op := range meta.Options {
		if op.Parse == nil {
			continue
		}
		v, err := op.Parse(val.Get(op.Name))
		if err != nil {
			return nil, err
		}
		conf.KV[op.Name] = v
	}

	return conf, nil
}
