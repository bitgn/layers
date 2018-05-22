package main

import (
	"flag"
	"fmt"

	"github.com/apple/foundationdb/bindings/go/src/fdb"
)

var (
	writes = flag.Uint("writes", 20, "percent of writes")
	hz     = flag.Int("hz", 1, "Througput to generate requests at")
)

var buildVersion string

func main() {
	fmt.Println("Version", buildVersion)

	flag.Parse()

	command := "help"
	if flag.NArg() == 1 {
		command = flag.Arg(0)
	}

	fdb.MustAPIVersion(510)
	db := fdb.MustOpenDefault()

	b := launch(command, db)
	if b != nil {
		ms := make(chan metrics, 200000)
		go runBenchmark(ms, *hz, b)
		stats(ms, db, *hz, b.Describe(), command)
		return
	}

	switch command {
	case "clear":
		clear(db)
	default:
		help()
	}
}

func help() {
	fmt.Println("FoundationDB benchmark tool")
	fmt.Println("Usage: benchmark [flags] command")
	fmt.Println("Flags:")
	flag.PrintDefaults()
}

func clear(db fdb.Database) {

	err, _ := db.Transact(func(tr fdb.Transaction) (interface{}, error) {
		r := fdb.KeyRange{
			Begin: fdb.Key([]byte{0}),
			End:   fdb.Key([]byte{255}),
		}

		tr.ClearRange(r)
		return nil, nil
	})

	if err != nil {
		panic(err)
	}
}
