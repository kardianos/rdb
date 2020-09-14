// Copyright 2014 Daniel Theophanes.
// Use of this source code is governed by a zlib-style
// license that can be found in the LICENSE file.

// +build ignore

// Use with: go run exec.go -conn="ms://" -sql="batch.sql"
package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/kardianos/rdb"
	_ "github.com/kardianos/rdb/ms"
	"github.com/kardianos/rdb/ms/batch"
)

var cs = flag.String("conn", "", "Connection String URL")
var sqlFile = flag.String("sql", "", "SQL File")

func main() {

	flag.Parse()

	if len(*cs) == 0 || len(*sqlFile) == 0 {
		flag.PrintDefaults()
		os.Exit(1)
		return
	}
	err := exec()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v", err)
		os.Exit(2)
	}
}

func exec() error {
	conifg, err := rdb.ParseConfigURL(*cs)
	if err != nil {
		return err
	}
	cp, err := rdb.Open(conifg)
	if err != nil {
		return err
	}

	sql, err := ioutil.ReadFile(*sqlFile)
	if err != nil {
		return err
	}

	err = batch.ExecuteBatchSql(cp, string(sql), "go")
	if err != nil {
		return err
	}
	return nil
}
