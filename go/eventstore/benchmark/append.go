package benchmark

import (
	"bytes"
	"fmt"

	"github.com/apple/foundationdb/bindings/go/src/fdb"
	"github.com/apple/foundationdb/bindings/go/src/fdb/tuple"
	"github.com/bitgn/layers/go/benchmark/bench"
	events "github.com/bitgn/layers/go/eventstore"
)

type AppendBench struct {
	store     events.Store
	streams   int
	eventSize int
}

func NewAppendBench(db fdb.Database, pfx ...tuple.TupleElement) *AppendBench {
	return &AppendBench{
		streams:   100000,
		eventSize: 200,
		store:     events.NewFdbStore(db, pfx...),
	}
}

func (b *AppendBench) Describe() *bench.Description {
	return &bench.Description{
		Name:  "es-bench",
		Setup: fmt.Sprintf("event size: %d, streams: %d", b.eventSize, b.streams),
	}

}

func (b *AppendBench) Exec(r uint64) error {
	aggID := int(r) % b.streams
	aggName := fmt.Sprintf("agg-%d", aggID)

	data := bytes.Repeat([]byte("Z"), b.eventSize)
	pack := []events.Envelope{events.New("test", data)}
	err := b.store.AppendToAggregate(
		aggName,
		events.ExpectedVersionAny,
		pack,
	)

	return err
}
