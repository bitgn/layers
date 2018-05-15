package compat

import (
	"bytes"

	"github.com/apple/foundationdb/bindings/go/src/fdb"
	"github.com/apple/foundationdb/bindings/go/src/fdb/tuple"
	patch "github.com/bitgn/foundationdb/bindings/go/src/fdb/tuple"
)

var (
	// place this 11-byte string
	VersionStampTemplate = "ver_stamp__"
	// and it will expand to \x02 + 11-bytes + \x00 on the byte level
	// then we can replace it with \x33 + 10-bytes + 2-bytes that
	// occupy the same space. Once we are on 520 we could add switch
	versionStampSearch = []byte("\x02" + VersionStampTemplate + "\x00")
	// correct 1+12-byte placeholder for the versionstamp
	versionStampReplace = []byte("\x33UUUUUUUUUUxx")
)

// versionstamp code is 0x33
// https://github.com/apple/foundationdb/blob/master/bindings/python/fdb/tuple.py#L47

// InjectVersionStamp searches the string for the entry of
// \x02ver_stamp\x00" and replaces it with
// \x33ver_stamp_ while also appending two byte little-endian pointer at the end
func InjectVersionStamp(key fdb.KeyConvertible, ver uint) fdb.Key {
	// https://github.com/apple/foundationdb/blob/master/bindings/python/fdb/tuple.py#L388

	src := []byte(key.FDBKey())
	bs := make([]byte, len(src))
	copy(bs, src)

	idx := bytes.Index(bs, versionStampSearch)
	copy(bs[idx:], versionStampReplace)
	bs[idx+11] = byte(ver)
	bs[idx+12] = byte(ver >> 8)

	// append the pointer
	b := append(bs, littleEndian(idx+1)...)
	return fdb.Key(b)
}

// Read Version stamp
func Unpack(key []byte) (tuple.Tuple, error) {
	t1, err := patch.Unpack(key)
	if err != nil {
		return nil, err
	}

	t2 := make(tuple.Tuple, len(t1))

	for i, te := range t1 {
		t2[i] = te
	}
	return t2, err
}

func littleEndian(i int) []byte {
	b := make([]byte, 2)
	b[0] = byte(i)
	b[1] = byte(i >> 8)
	return b
}
