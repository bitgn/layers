package main

import (
	"fmt"

	"github.com/apple/foundationdb/bindings/go/src/fdb"
	"github.com/apple/foundationdb/bindings/go/src/fdb/tuple"
	"github.com/bitgn/layers/go/benchmark/bench"
	"github.com/google/uuid"
)

// rand seed

type SimpleBench struct {
	db         fdb.Database
	tpl        []tuple.TupleElement
	writeRatio uint
	valueSize  int
}

func NewSimpleBench(db fdb.Database, writeRatio uint, tpl ...tuple.TupleElement) *SimpleBench {
	return &SimpleBench{
		tpl:        tpl,
		db:         db,
		writeRatio: writeRatio,
		valueSize:  200,
	}
}

func (b *SimpleBench) Describe() *bench.Description {
	return &bench.Description{
		Name:  "fdb-simple",
		Setup: fmt.Sprintf("writes: %d, value: %d bytes", b.writeRatio, b.valueSize),
	}
}

func newKey(tpl []tuple.TupleElement, prefix int) tuple.Tuple {
	id := uuid.New()
	buf := [16]byte(id)
	return append(tpl, prefix, buf[:])
}

func (self *SimpleBench) Exec(r uint64) error {

	var write bool

	switch self.writeRatio {
	case 0:
		write = false
	case 1:
		write = true
	default:
		write = uint(r%100) < self.writeRatio
	}

	if write {
		_, err := self.db.Transact(func(tr fdb.Transaction) (interface{}, error) {
			key := newKey(self.tpl, 1)
			value := make([]byte, self.valueSize)
			tr.Set(key, value)
			return nil, nil
		})
		return err
	}

	_, err := self.db.ReadTransact(func(tr fdb.ReadTransaction) (interface{}, error) {
		key := newKey(self.tpl, 1)
		_, err := tr.Get(key).Get()
		return nil, err
	})

	return err
}
