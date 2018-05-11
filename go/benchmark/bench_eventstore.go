package main

import (
	"bytes"
	"fmt"
	"math"
	"sync"

	"github.com/apple/foundationdb/bindings/go/src/fdb"
	es "github.com/bitgn/layers/go/eventstore"
)

var (
	mux sync.Mutex
)

func normID(deviation, min, max int) int {
	mux.Lock()
	f := int(math.Abs(r.NormFloat64())*float64(deviation)) + min
	mux.Unlock()

	if f > max {
		f = max
	}
	return f
}

func benchEventStoreAppends(db fdb.Database) error {

	store := es.NewFdbStore(db, BitgnPrefix)

	// split between 10000 aggregates

	aggID := normID(10000, 0, 100000)
	aggName := fmt.Sprintf("agg-%d", aggID)

	size := normID(100, 10, 500)

	data := bytes.Repeat([]byte("Z"), size)
	pack := []es.Envelope{es.New("test", data)}
	err := store.AppendToAggregate(
		aggName,
		es.ExpectedVersionAny,
		pack,
	)

	return err
}
