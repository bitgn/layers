package compat

import (
	"bytes"
	"errors"

	"github.com/apple/foundationdb/bindings/go/src/fdb"
	"github.com/apple/foundationdb/bindings/go/src/fdb/subspace"
	"github.com/apple/foundationdb/bindings/go/src/fdb/tuple"
)

func UnpackSubspace(s subspace.Subspace, k fdb.KeyConvertible) (tuple.Tuple, error) {
	key := k.FDBKey()
	b := s.Bytes()
	if !bytes.HasPrefix(key, b) {
		return nil, errors.New("key is not in subspace")
	}
	return Unpack(key[len(b):])
}
