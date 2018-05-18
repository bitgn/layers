package experiment

import (
	"github.com/apple/foundationdb/bindings/go/src/fdb"
	"github.com/apple/foundationdb/bindings/go/src/fdb/subspace"
	"github.com/apple/foundationdb/bindings/go/src/fdb/tuple"
	"github.com/bitgn/layers/go/benchmark/bench"
	"github.com/google/uuid"
)

type Custom1Bench struct {
	db    fdb.Database
	space subspace.Subspace
}

func NewCustom1Bench(db fdb.Database, pfx ...tuple.TupleElement) bench.Launcher {
	return &Custom1Bench{
		db:    db,
		space: subspace.Sub(pfx...),
	}
}

func (b *Custom1Bench) Describe() *bench.Description {
	return &bench.Description{
		Name:  "Custom Benchmark - OW workload",
		Setup: "tx: key lookup, 500b write, 2x index writes",
	}
}

func (b *Custom1Bench) randomID(table int) []byte {
	array := [16]byte(uuid.New())
	return b.space.Sub(table, array[:]).Bytes()
}

func (b *Custom1Bench) Exec(r uint64) error {

	id := b.randomID(1)
	index1 := b.randomID(2)
	index2 := b.randomID(3)

	_, err := b.db.Transact(func(tr fdb.Transaction) (interface{}, error) {
		key := fdb.Key(id)
		value := make([]byte, 500)
		val, err := tr.Get(key).Get()
		if err != nil {
			return nil, err
		}
		if val != nil {
			return nil, nil
		}

		tr.Set(key, value)
		tr.Set(fdb.Key(index1), id)
		tr.Set(fdb.Key(index2), id)

		return nil, nil

	})

	return err

}
