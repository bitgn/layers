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
	streams   int
	eventSize int
	space     subspace.Subspace
	db        fdb.Database
}

func NewEventStoreBench(db fdb.Database, pfx ...tuple.TupleElement) bench.Launcher {

	return &EventStore{
		streams:   1000000,
		eventSize: 200,
		space:     subspace.Sub(pfx...),
		db:        db,
	}
}

func (e *EventStore) Describe() *bench.Description {
	return &bench.Description{
		Name: "EventStore experimental v2 (go) - es-v2-append",
		Setup: fmt.Sprintf("event size: %d, streams %d",
			e.eventSize, e.streams),
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

	contract := tuple.Tuple{"test"}.Pack()

	var gk, sk, offset []byte

	gk = b.space.Sub(globalTable).Bytes()
	gk, offset = verstamp(gk, 1)
	// add contract
	gk = append(gk, contract...)
	gk = append(gk, offset...)

	sk = b.space.Sub(streamTable, aggName).Bytes()
	sk, offset = verstamp(sk, 1)
	sk = append(sk, contract...)
	sk = append(sk, offset...)

	// TODO: check for the concurrent change

	vs, err := b.db.Transact(func(tr fdb.Transaction) (interface{}, error) {

		vs := tr.GetVersionstamp()
		tr.SetVersionstampedKey(fdb.Key(gk), data)
		tr.SetVersionstampedKey(fdb.Key(sk), data)
		return vs, nil
	})

	if err != nil {
		return err
	}

	_, err = vs.(fdb.FutureKey).Get()

	return err

}
