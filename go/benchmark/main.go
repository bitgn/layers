package main

import (
	"flag"
	"fmt"

	"github.com/apple/foundationdb/bindings/go/src/fdb"
	"github.com/apple/foundationdb/bindings/go/src/fdb/tuple"
	"github.com/bitgn/layers/go/benchmark/experiment"
	esbench "github.com/bitgn/layers/go/eventstore/benchmark"
)

var (
	writes = flag.Uint("writes", 20, "percent of writes")
	hz     = flag.Int("hz", 1, "Througput to generate requests at")
)

var ()

func main() {

	flag.Parse()

	command := "help"
	if flag.NArg() == 1 {
		command = flag.Arg(0)
	}

	fdb.MustAPIVersion(510)
	db := fdb.MustOpenDefault()

	switch command {

	case "clear":
		clear(db)

	case "simple":
		ms := make(chan metrics, 200000)
		b := NewSimpleBench(db, *writes, tuple.Tuple{BitgnPrefix})
		go runBenchmark(ms, *hz, b)
		stats(ms, db, *hz, b.Describe())
	case "es-append":

		ms := make(chan metrics, 200000)
		b := esbench.NewAppendBench(db, tuple.Tuple{BitgnPrefix})
		go runBenchmark(ms, *hz, b)
		stats(ms, db, *hz, b.Describe())
	case "es-v2-append":

		ms := make(chan metrics, 200000)
		b := experiment.NewEventStoreBench(db, tuple.Tuple{BitgnPrefix})
		go runBenchmark(ms, *hz, b)
		stats(ms, db, *hz, b.Describe())
	default:
		help()
		return
	}
}

var BitgnPrefix = "bgn"

func help() {
	fmt.Println("FoundationDB benchmark tool")
	fmt.Println("Usage: benchmark [flags] command")
	fmt.Println("Flags:")
	flag.PrintDefaults()
}

func clear(db fdb.Database) {

	err, _ := db.Transact(func(tr fdb.Transaction) (interface{}, error) {

		t := tuple.Tuple{BitgnPrefix}
		r, _ := fdb.PrefixRange(t.Pack())

		tr.ClearRange(r)
		return nil, nil
	})

	if err != nil {
		panic(err)
	}
}
