package main

import (
	"bytes"
	"fmt"
	"math/rand"

	es "github.com/bitgn/layers/go/eventstore"
)

type EventStoreBench struct {
	store     es.Store
	partition int
	total     int
	r         *rand.Rand
}

func NewEventStoreBench(store es.Store, partition, total int) *EventStoreBench {

	return &EventStoreBench{
		store:     store,
		partition: partition,
		total:     total,
		r:         rand.New(rand.NewSource(int64(partition))),
	}
}

func (this *EventStoreBench) Run() error {
	const streamsPerActor = 100

	size := 200

	aggID := r.Intn(streamsPerActor) + this.partition*streamsPerActor

	aggName := fmt.Sprintf("agg-%d", aggID)

	data := bytes.Repeat([]byte("Z"), size)
	pack := []es.Envelope{es.New("test", data)}
	err := this.store.AppendToAggregate(
		aggName,
		es.ExpectedVersionAny,
		pack,
	)

	return err
}
