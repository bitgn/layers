package experiment

import (
	"bytes"
	"encoding/binary"
	"fmt"

	"github.com/apple/foundationdb/bindings/go/src/fdb"
	"github.com/apple/foundationdb/bindings/go/src/fdb/subspace"
	"github.com/apple/foundationdb/bindings/go/src/fdb/tuple"
	"github.com/bitgn/layers/go/benchmark/bench"
)

type EventStore struct {
	streams     int
	eventSize   int
	space       subspace.Subspace
	db          fdb.Database
	denormalize bool
}

func NewEventStoreBench(db fdb.Database, denormalize bool, pfx ...tuple.TupleElement) bench.Launcher {

	return &EventStore{
		streams:     1000000,
		eventSize:   200,
		space:       subspace.Sub(pfx...),
		db:          db,
		denormalize: denormalize,
	}
}

func (e *EventStore) Describe() *bench.Description {
	return &bench.Description{
		Name: "EventStore experimental v2 (go) - es-v2-append",
		Setup: fmt.Sprintf("event size: %d, streams: %d, denormalize: %t",
			e.eventSize, e.streams, e.denormalize),
		Explanation: `
This benchmark simulates appends in an experimental event store. It writes a copy of event into global event stream and a named event stream. Both entries are versionstamped.

If 'denormalize' it true, then we write only an event pointer to the named event stream.

Stream names are in form 'agg-%d', where the number is randomly generated (even distribution). Event size is fixed.

At the end of the transaction we also retrieve current transaction versionstamp (to be used by the application logic).
`,
	}
}

const (
	globalTable = 0
	streamTable = 1
)

func verstamp(key []byte, user int32) (value []byte, trailer []byte) {

	// versiontsamp is:

	// | 8 bytes db version | 2 bytes tr order | user version | .....
	// plus 2 bytes versionstamp offset (which will be trimmed by the DB)
	pos := len(key)

	offset := []byte{
		byte(pos),
		byte(pos >> 8),
	}

	buf := new(bytes.Buffer)
	// key
	buf.Write(key)
	// DB portion
	buf.Write(make([]byte, 8+2))
	// user portion
	binary.Write(buf, binary.BigEndian, user)

	return buf.Bytes(), offset

}

func (b *EventStore) Exec(r uint64) error {

	aggID := int(r) % b.streams
	aggName := fmt.Sprintf("agg-%d", aggID)

	data := bytes.Repeat([]byte("Z"), b.eventSize)

	// schema:
	// |global table|versiontsamp|IDX | contract -> value
	// |stream table|name|versionstamp|IDX|contract -> value
	// we don't have the versionstamp support yet, so we do it manually

	// add versiontsamp

	var gk, sk, offset []byte

	gk = b.space.Sub(globalTable).Bytes()
	gk, offset = verstamp(gk, 1)
	// add contract
	gk = append(gk, offset...)

	sk = b.space.Sub(streamTable, aggName).Bytes()
	sk, offset = verstamp(sk, 1)
	sk = append(sk, offset...)

	// TODO: check for the concurrent change

	vs, err := b.db.Transact(func(tr fdb.Transaction) (interface{}, error) {

		vs := tr.GetVersionstamp()
		tr.SetVersionstampedKey(fdb.Key(gk), data)

		if b.denormalize {
			tr.SetVersionstampedKey(fdb.Key(sk), data)
		} else {
			tr.SetVersionstampedKey(fdb.Key(sk), nil)
		}
		return vs, nil
	})

	if err != nil {
		return err
	}

	_, err = vs.(fdb.FutureKey).Get()

	return err

}
