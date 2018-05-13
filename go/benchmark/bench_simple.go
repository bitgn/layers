package main

import (
	"math/rand"

	"github.com/apple/foundationdb/bindings/go/src/fdb"
)

// rand seed

type SimpleBench struct {
	db fdb.Database
	r  *rand.Rand
}

func NewSimpleBench(db fdb.Database, seed int64) *SimpleBench {
	r := rand.New(rand.NewSource(seed))
	return &SimpleBench{
		r:  r,
		db: db,
	}
}

func (self *SimpleBench) Run() error {

	write := self.r.Intn(100) < *writes

	if write {
		_, err := self.db.Transact(func(tr fdb.Transaction) (interface{}, error) {
			key := newKey(1)
			value := make([]byte, 200)
			tr.Set(key, value)
			return nil, nil
		})
		return err
	}

	_, err := self.db.ReadTransact(func(tr fdb.ReadTransaction) (interface{}, error) {
		key := newKey(1)
		_, err := tr.Get(key).Get()
		return nil, err
	})

	return err
}
