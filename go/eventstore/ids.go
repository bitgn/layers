package events

import (
	"bytes"

	"github.com/apple/foundationdb/bindings/go/src/fdb"
	"github.com/apple/foundationdb/bindings/go/src/fdb/subspace"
)

type KeyReader interface {
	GetKey(sel fdb.Selectable) fdb.FutureKey
}

type LastKeyFuture struct {
	key   fdb.FutureKey
	space subspace.Subspace
}

func GetLastKeyFuture(tr KeyReader, space subspace.Subspace) *LastKeyFuture {
	_, end := space.FDBRangeKeys()
	key := tr.GetKey(fdb.LastLessThan(end))

	return &LastKeyFuture{key, space}
}

func (r *LastKeyFuture) MustGetNextIndex(position int) int {
	key := r.key.MustGet()

	start, _ := r.space.FDBRangeKeys()

	if i := bytes.Compare(key, []byte(start.FDBKey())); i < 0 {
		return 0
	}

	if t, err := r.space.Unpack(key); err != nil {
		panic("Failed to unpack key")
	} else {
		return int(t[0].(int64)) + 1
	}
}
