package main

import (
	"github.com/apple/foundationdb/bindings/go/src/fdb/tuple"
	"github.com/google/uuid"
)

func newKey(prefix int) tuple.Tuple {
	id := uuid.New()
	buf := [16]byte(id)
	return tuple.Tuple{BitgnPrefix, prefix, buf[:]}
}
