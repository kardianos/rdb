// Copyright 2014 Daniel Theophanes.
// Use of this source code is governed by a zlib-style
// license that can be found in the LICENSE file.

package rdb

import (
	"reflect"
	"testing"
)

var configTestPass = map[string]*Config{
	"driver://username:password@localUrl:1234/ServerInstance?db=mydatabase&opt1=valA&opt2=valB": {
		DriverName: "driver",
		Username:   "username",
		Password:   "password",
		Hostname:   "localUrl",
		Port:       1234,
		Instance:   "ServerInstance",
		Database:   "mydatabase",
	},
	"driver://username:password@localUrl:1234?db=mydatabase": {
		DriverName: "driver",
		Username:   "username",
		Password:   "password",
		Hostname:   "localUrl",
		Port:       1234,
		Instance:   "",
		Database:   "mydatabase",
	},
	"driver://username@localUrl?db=mydatabase": {
		DriverName: "driver",
		Username:   "username",
		Password:   "",
		Hostname:   "localUrl",
		Port:       0,
		Instance:   "",
		Database:   "mydatabase",
	},
	"driver://localUrl?db=mydatabase": {
		DriverName: "driver",
		Username:   "",
		Password:   "",
		Hostname:   "localUrl",
		Port:       0,
		Instance:   "",
		Database:   "mydatabase",
	},
	"sqlite:///C:/folder/file.sqlite3?opt1=valA&opt2=valB": {
		DriverName: "sqlite",
		Username:   "",
		Password:   "",
		Hostname:   "",
		Port:       0,
		Instance:   "C:/folder/file.sqlite3",
		Database:   "",
	},
	"sqlite:///srv/folder/file.sqlite3": {
		DriverName: "sqlite",
		Username:   "",
		Password:   "",
		Hostname:   "",
		Port:       0,
		Instance:   "srv/folder/file.sqlite3",
		Database:   "",
	},
}

func TestConfigURL(t *testing.T) {
	for url, confExpect := range configTestPass {
		conf, err := ParseConfigURL(url)
		if err != nil {
			if _, is := err.(DriverNotFound); !is {
				t.Errorf("Invalid connection string: %v", err)
			}
		}
		if reflect.DeepEqual(confExpect, conf) == false {
			t.Errorf("Not as expected:\nurl: %s\ngot: %#v", url, conf)
		}
	}
}
