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

	seed := rand.Int63() + int64(partition)

	return &EventStoreBench{
		store:     store,
		partition: partition,
		total:     total,
		r:         rand.New(rand.NewSource(seed)),
	}
}

func (this *EventStoreBench) Run() error {
	const streamsPerActor = 100

	size := 200

	aggID := this.r.Intn(streamsPerActor) + this.partition*streamsPerActor

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
