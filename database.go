// Copyright 2014 Daniel Theophanes.
// Use of this source code is governed by a zlib-style
// license that can be found in the LICENSE file.

package rdb

import (
	"fmt"
)

var drivers = map[string]Driver{}

// Panics if called twice with the same name.
// Make the driver instance available clients.
func Register(name string, dr Driver) {
	_, found := drivers[name]
	if found {
		panic(fmt.Sprintf("Driver already present: %s", name))
	}
	drivers[name] = dr
}

func getDriver(name string) (Driver, error) {
	dr, found := drivers[name]
	if !found {
		return nil, DriverNotFound{name: name}
	}
	return dr, nil
}

func Open(config *Config) (Database, error) {
	dr, err := getDriver(config.DriverName)
	if err != nil {
		return nil, err
	}
	return dr.Open(config)
}
