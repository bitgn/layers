package main

import (
	"flag"

	"github.com/apple/foundationdb/bindings/go/src/fdb"
	"github.com/apple/foundationdb/bindings/go/src/fdb/tuple"
	"github.com/bitgn/layers/go/benchmark/bench"
	"github.com/bitgn/layers/go/benchmark/experiment"
	esbench "github.com/bitgn/layers/go/eventstore/benchmark"
)

type Factory func() bench.Launcher

var (
	catalogue   = make(map[string]Factory)
	BitgnPrefix = "bgn"
	BitgnTuple  = tuple.Tuple{BitgnPrefix}

	denormalize = flag.Bool("denormalize", false, "Denormalize option")
)

func launch(name string, db fdb.Database) bench.Launcher {
	switch name {
	case "kafka":
		return experiment.NewKafkaBench(db, BitgnTuple)
	case "simple":
		return NewSimpleBench(db, *writes, BitgnTuple)
	case "es-append":
		return esbench.NewAppendBench(db, BitgnTuple)
	case "custom1":
		return experiment.NewCustom1Bench(db, BitgnTuple)
	case "es-v2-append":
		return experiment.NewEventStoreBench(db, *denormalize, BitgnTuple)
	}

	return nil
}
