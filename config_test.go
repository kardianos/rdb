package rdb

import (
	"reflect"
	"testing"
)

var configTestPass = map[string]*Config{
	"driver://username:password@localUrl:1234/ServerInstance?db=mydatabase&opt1=valA&opt2=valB": &Config{
		DriverName: "driver",
		Username:   "username",
		Password:   "password",
		Hostname:   "localUrl",
		Port:       1234,
		Instance:   "ServerInstance",
		Database:   "mydatabase",
	},
	"driver://username:password@localUrl:1234?db=mydatabase": &Config{
		DriverName: "driver",
		Username:   "username",
		Password:   "password",
		Hostname:   "localUrl",
		Port:       1234,
		Instance:   "",
		Database:   "mydatabase",
	},
	"driver://username@localUrl?db=mydatabase": &Config{
		DriverName: "driver",
		Username:   "username",
		Password:   "",
		Hostname:   "localUrl",
		Port:       0,
		Instance:   "",
		Database:   "mydatabase",
	},
	"driver://localUrl?db=mydatabase": &Config{
		DriverName: "driver",
		Username:   "",
		Password:   "",
		Hostname:   "localUrl",
		Port:       0,
		Instance:   "",
		Database:   "mydatabase",
	},
	"sqlite:///C:/folder/file.sqlite3?opt1=valA&opt2=valB": &Config{
		DriverName: "sqlite",
		Username:   "",
		Password:   "",
		Hostname:   "",
		Port:       0,
		Instance:   "C:/folder/file.sqlite3",
		Database:   "",
	},
	"sqlite:///srv/folder/file.sqlite3": &Config{
		DriverName: "sqlite",
		Username:   "",
		Password:   "",
		Hostname:   "",
		Port:       0,
		Instance:   "srv/folder/file.sqlite3",
		Database:   "",
	},
}

func TestConfig(t *testing.T) {
	for url, confExpect := range configTestPass {
		conf, err := ParseConfig(url)
		if err != nil {
			t.Errorf("Invalid connection string: %v", err)
		}
		if reflect.DeepEqual(confExpect, conf) == false {
			t.Errorf("Not as expected:\nurl: %s\ngot: %#v", url, conf)
		}
	}
}
