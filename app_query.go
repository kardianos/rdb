// +build ignore

package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/kardianos/rdb"
	_ "github.com/kardianos/rdb/ms"
	"github.com/kardianos/rdb/table"
)

func main() {
	err := run()
	if err != nil {
		log.Fatal(err)
	}
}

func run() error {
	cs := flag.String("connect", "", "Connection string")
	queryFile := flag.String("queryfile", "", "Query filename")
	flag.Parse()

	if len(*cs) == 0 || len(*queryFile) == 0 {
		flag.PrintDefaults()
		os.Exit(1)
	}

	query, err := ioutil.ReadFile(*queryFile)
	if err != nil {
		return err
	}

	fmt.Println("START")

	config, err := rdb.ParseConfigURL(*cs)
	if err != nil {
		return err
	}
	db, err := rdb.Open(config)
	if err != nil {
		return err
	}
	defer db.Close()

	t, err := table.FillCommand(db, &rdb.Command{
		Sql: string(query),
	})

	if err != nil {
		return err
	}

	fmt.Println("Rows: ", len(t.Row))

	return nil
}
