package compat

import (
	"fmt"
	"testing"

	"github.com/apple/foundationdb/bindings/go/src/fdb"
	"github.com/apple/foundationdb/bindings/go/src/fdb/subspace"
)

func TestStampRewrite(t *testing.T) {
	pre := subspace.Sub("prefix", VersionStampTemplate, "contract")

	fmt.Println("pre", pre.Bytes())

	after := InjectVersionStamp(pre, 2)

	fmt.Println("aft", after)

	decoded, err := Unpack(after[0 : len(after)-2])
	if err != nil {
		t.Fatal(err)
	}

	fmt.Println(decoded)
}

func TestVersionstampWrite(t *testing.T) {
	fdb.MustAPIVersion(510)
	db := fdb.MustOpenDefault()

	future, err := db.Transact(func(tr fdb.Transaction) (interface{}, error) {
		pre := subspace.Sub("prefix", VersionStampTemplate, "contract")
		after := InjectVersionStamp(pre, 0)
		tr.SetVersionstampedKey(after, make([]byte, 0))
		return tr.GetVersionstamp(), nil
	})

	if err != nil {
		t.Fatal(err)
	}
	key := future.(fdb.FutureKey).MustGet()

	if len(key) != 10 {
		t.Fatal("versionstamp should be 10 bytes")
	}

	fmt.Println(key)

}
