// Copyright (c) 2011, The pg Authors. All Rights Reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package pg

import (
	"bitbucket.org/kardianos/rdb"
	"bufio"
	"errors"
	"fmt"
	"net/url"
	"os"
	"time"
)

type drv struct{}

func (d *drv) Open(conf *rdb.Config) (_ rdb.Conn, err error) {
	defer errRecover(&err)

	o := make(values)

	// A number of defaults are applied here, in this order:
	//
	// * Very low precedence defaults applied in every situation
	// * Environment variables
	// * Explicitly passed connection information
	o.Set("host", conf.Hostname)
	o.Set("port", fmt.Sprintf("%d", conf.Port))

	// N.B.: Extra float digits should be set to 3, but that breaks
	// Postgres 8.4 and older, where the max is 2.
	o.Set("extra_float_digits", "2")
	for k, v := range parseEnviron(os.Environ()) {
		o.Set(k, v)
	}

	// TODO: use the config KV or other options.
	// if err := parseOpts(name, o); err != nil {
	// 	return nil, err
	// }

	// Use the "fallback" application name if necessary
	if fallback := o.Get("fallback_application_name"); fallback != "" {
		if !o.Isset("application_name") {
			o.Set("application_name", fallback)
		}
	}
	o.Unset("fallback_application_name")

	// We can't work with any client_encoding other than UTF-8 currently.
	// However, we have historically allowed the user to set it to UTF-8
	// explicitly, and there's no reason to break such programs, so allow that.
	// Note that the "options" setting could also set client_encoding, but
	// parsing its value is not worth it.  Instead, we always explicitly send
	// client_encoding as a separate run-time parameter, which should override
	// anything set in options.
	if enc := o.Get("client_encoding"); enc != "" && !isUTF8(enc) {
		return nil, errors.New("client_encoding must be absent or 'UTF8'")
	}
	o.Set("client_encoding", "UTF8")
	// DateStyle needs a similar treatment.
	if datestyle := o.Get("datestyle"); datestyle != "" {
		if datestyle != "ISO, MDY" {
			panic(fmt.Sprintf("setting datestyle must be absent or %v; got %v", "ISO, MDY", datestyle))
		}
	} else {
		o.Set("datestyle", "ISO, MDY")
	}

	// If a user is not provided by any other means, the last
	// resort is to use the current operating system provided user
	// name.
	if o.Get("user") == "" {
		u, err := userCurrent()
		if err != nil {
			return nil, err
		} else {
			o.Set("user", u)
		}
	}

	c, err := dial(o)
	if err != nil {
		return nil, err
	}

	cn := &conn{
		c:      c,
		config: conf,
	}
	cn.ssl(o)
	cn.buf = bufio.NewReader(cn.c)
	cn.startup(o)
	// reset the deadline, in case one was set (see dial)
	err = cn.c.SetDeadline(time.Time{})
	return cn, err
}

func (d *drv) DriverMetaInfo() *rdb.DriverMeta {
	return &rdb.DriverMeta{}
}

var cmdPing = &rdb.Command{
	Sql:   "select 1 limit 0;",
	Arity: rdb.Zero,
}

func (d *drv) PingCommand() *rdb.Command {
	return cmdPing
}

func (d *drv) ParseOptions(KV map[string]interface{}, configOptions url.Values) error {
	return nil
}

func init() {
	rdb.Register("pg", &drv{})
}