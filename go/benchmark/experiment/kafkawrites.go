package experiment

import (
	"bytes"
	"fmt"

	"github.com/apple/foundationdb/bindings/go/src/fdb"
	"github.com/apple/foundationdb/bindings/go/src/fdb/subspace"
	"github.com/apple/foundationdb/bindings/go/src/fdb/tuple"
	"github.com/bitgn/layers/go/benchmark/bench"
)

type KafkaBench struct {
	topics      int
	messageSize int
	space       subspace.Subspace
	db          fdb.Database
}

func NewKafkaBench(db fdb.Database, pfx ...tuple.TupleElement) bench.Launcher {

	return &KafkaBench{
		topics:      10,
		messageSize: 50,
		space:       subspace.Sub(pfx...),
		db:          db,
	}
}

func (e *KafkaBench) Describe() *bench.Description {
	return &bench.Description{
		Name: "Kafka-like Publish Benchmark (go)",
		Setup: fmt.Sprintf("message: %d, bytes topics: %d",
			e.messageSize, e.topics),
		Explanation: fmt.Sprintf(`
This benchmark does publishing to a kafka-like layer. Writes are evenly distributed between %d topics. Each message is %d bytes.
`, e.topics, e.messageSize),
	}
}

func (b *KafkaBench) Exec(r uint64) error {

	id := int(r) % b.topics
	topic := fmt.Sprintf("topic-%d", id)

	data := bytes.Repeat([]byte("Z"), b.messageSize)

	// schema:
	// |global table|versiontsamp|IDX | contract -> value
	// |stream table|name|versionstamp|IDX|contract -> value
	// we don't have the versionstamp support yet, so we do it manually

	// add versiontsamp

	var gk, offset []byte

	gk = b.space.Sub(topic).Bytes()
	gk, offset = verstamp(gk, 1)
	// add contract
	gk = append(gk, offset...)

	_, err := b.db.Transact(func(tr fdb.Transaction) (interface{}, error) {

		tr.SetVersionstampedKey(fdb.Key(gk), data)

		return nil, nil
	})
	return err
}
