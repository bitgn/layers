package main

import (
	"flag"
	"fmt"
	"time"

	"github.com/apple/foundationdb/bindings/go/src/fdb"
	"github.com/apple/foundationdb/bindings/go/src/fdb/tuple"
	es "github.com/bitgn/layers/go/eventstore"
)

var (
	actors = flag.Int("actors", 1000, "number of actors to run")
	writes = flag.Int("writes", 20, "percent of writes")
)

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

		for i := 0; i < *actors; i++ {
			b := NewSimpleBench(db, int64(i))
			go benchmark(ms, b)
		}

		stats(ms, db)
	case "es-append":

		ms := make(chan metrics, 200000)
		store := es.NewFdbStore(db, tuple.Tuple{BitgnPrefix})

		for i := 0; i < *actors; i++ {
			b := NewEventStoreBench(store, i, *actors)
			go benchmark(ms, b)
		}
		stats(ms, db)
	default:
		help()
		return
	}
}

func help() {
	fmt.Println("FoundationDB benchmark tool")
	fmt.Println("Usage: benchmark [flags] command")
	fmt.Println("Flags:")
	flag.PrintDefaults()
}

type metrics struct {
	nanoseconds int64
	error       bool
}

type action func(db fdb.Database) error

var BitgnPrefix = "bgn"

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

var (
	throughput = flag.Int("hz", 1, "Througput to generate requests at")
)

func benchmark(out chan metrics, b Bench) {

	period := time.Second / time.Duration(*throughput)

	for range time.Tick(period) {

		begin := time.Now()
		go func() {
			err := b.Run()
			total := time.Since(begin)

			result := metrics{
				error:       err != nil,
				nanoseconds: total.Nanoseconds(),
			}
			out <- result
		}()

	}
}

type Bench interface {
	Run() error
}
