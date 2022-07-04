// Copyright 2014 Daniel Theophanes.
// Use of this source code is governed by a zlib-style
// license that can be found in the LICENSE file.

package rdb

import (
	"bytes"
	"encoding/json"
	"testing"
)

var configTestPass = map[string]*Config{
	"driver://username:password@localUrl:1234/ServerInstance?db=mydatabase&opt_1=valA&opt_2=valB": {
		DriverName: "driver",
		Username:   "username",
		Password:   "password",
		Hostname:   "localUrl",
		Port:       1234,
		Instance:   "ServerInstance",
		Database:   "mydatabase",
		KV: map[string]interface{}{
			"1": "valA",
			"2": "valB",
		},
	},
	"driver://username:password@localUrl:1234?db=mydatabase": {
		DriverName: "driver",
		Username:   "username",
		Password:   "password",
		Hostname:   "localUrl",
		Port:       1234,
		Instance:   "",
		Database:   "mydatabase",
		KV:         make(map[string]interface{}),
	},
	"driver://username@localUrl?db=mydatabase": {
		DriverName: "driver",
		Username:   "username",
		Password:   "",
		Hostname:   "localUrl",
		Port:       0,
		Instance:   "",
		Database:   "mydatabase",
		KV:         make(map[string]interface{}),
	},
	"driver://localUrl?db=mydatabase": {
		DriverName: "driver",
		Username:   "",
		Password:   "",
		Hostname:   "localUrl",
		Port:       0,
		Instance:   "",
		Database:   "mydatabase",
		KV:         make(map[string]interface{}),
	},
	"sqlite:///C:/folder/file.sqlite3?opt_1=valA&opt_2=valB": {
		DriverName: "sqlite",
		Username:   "",
		Password:   "",
		Hostname:   "",
		Port:       0,
		Instance:   "C:/folder/file.sqlite3",
		Database:   "",
		KV: map[string]interface{}{
			"1": "valA",
			"2": "valB",
		},
	},
	"sqlite:///srv/folder/file.sqlite3": {
		DriverName: "sqlite",
		Username:   "",
		Password:   "",
		Hostname:   "",
		Port:       0,
		Instance:   "srv/folder/file.sqlite3",
		Database:   "",
		KV:         make(map[string]interface{}),
	},
}

func TestConfigURL(t *testing.T) {
	for url, confExpect := range configTestPass {
		conf, err := ParseConfigURL(url)
		if err != nil {
			if _, is := err.(DriverNotFound); !is {
				t.Fatalf("Invalid connection string: %v", err)
			}
		}
		got, _ := json.MarshalIndent(conf, "", "\t")
		want, _ := json.MarshalIndent(confExpect, "", "\t")
		if !bytes.Equal(got, want) {
			t.Errorf("Not as expected:\nurl: %s\ngot: %s\nwant: %s", url, got, want)
		}
	}
}
