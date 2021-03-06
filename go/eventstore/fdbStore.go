package events

import (
	"crypto/rand"

	"github.com/apple/foundationdb/bindings/go/src/fdb"
	"github.com/apple/foundationdb/bindings/go/src/fdb/subspace"
	"github.com/apple/foundationdb/bindings/go/src/fdb/tuple"
)

// fdbStore maintains two subspaces:
// Global / [versionstamp] / contract / <- vs pointer
// Aggregate / id / version / contract /
type fdbStore struct {
	space         subspace.Subspace
	db            fdb.Database
	reportMetrics bool
}

func NewFdbStore(db fdb.Database, el ...tuple.TupleElement) Store {
	space := subspace.Sub(el...)
	return &fdbStore{
		space,
		db,
		false,
	}
}

const (
	globalPrefix = 0
	aggregPrefix = 1
)

var (
	Start = make([]tuple.TupleElement, 0)
)

func nextRandom() []byte {
	b := make([]byte, 20)
	if _, err := rand.Read(b); err == nil {
		return b
	} else {
		panic(err)
	}
}

// ReportMetrics enables FSD metrics reporting. It is disabled by default
// to avoid polluting unit tests
//func (es *fdbStore) ReportMetrics() {
//	es.reportMetrics = true
//}

func (es *fdbStore) Clear() {
	es.db.Transact(func(tr fdb.Transaction) (interface{}, error) {
		tr.ClearRange(es.space)
		return nil, nil
	})
}

func (es *fdbStore) Append(records []Envelope) (err error) {

	globalSpace := es.space.Sub(globalPrefix)

	uuid := NewSequentialUUID()

	_, err = es.db.Transact(func(tr fdb.Transaction) (interface{}, error) {

		for i, evt := range records {
			contract, data := evt.Payload()
			tr.Set(globalSpace.Sub(uuid, i, contract), data)
		}

		return nil, nil
	})

	return
}

func (es *fdbStore) AppendToAggregate(stream string, expectedVersion int, records []Envelope) (err error) {

	globalSpace := es.space.Sub(globalPrefix)
	aggregSpace := es.space.Sub(aggregPrefix, stream)
	uuid := NewSequentialUUID()
	// TODO add random key to reduce contention

	_, err = es.db.Transact(func(tr fdb.Transaction) (interface{}, error) {
		// we are getting them in parallel
		aggregRecord := GetLastKeyFuture(tr, aggregSpace)

		//globalRecord := GetLastKeyFuture(tr.Snapshot(), globalSpace)

		nextAggregIndex := aggregRecord.MustGetNextIndex(0)

		switch expectedVersion {
		case ExpectedVersionAny:
			break
		case ExpectedVersionNone:
			if nextAggregIndex != 0 {
				return nil, &ErrConcurrencyViolation{
					stream,
					expectedVersion,
					nextAggregIndex - 1,
				}
			}
		default:
			if (nextAggregIndex - 1) != expectedVersion {
				return nil, &ErrConcurrencyViolation{
					stream,
					expectedVersion,
					nextAggregIndex - 1,
				}
			}
		}

		for i, evt := range records {
			aggregIndex := nextAggregIndex + i

			contract, data := evt.Payload()

			tr.Set(globalSpace.Sub(uuid, i, contract), data)
			tr.Set(aggregSpace.Sub(aggregIndex, contract), data)
		}

		return nil, nil
	})

	return
}

func (es *fdbStore) ReadAll(last []byte, limit int) *GlobalSlice {
	globalSpace := es.space.Sub(globalPrefix)
	start, end := globalSpace.FDBRangeKeys()

	r, err := es.db.ReadTransact(func(tr fdb.ReadTransaction) (interface{}, error) {

		var scan fdb.KeyRange
		if nil == last {
			scan = fdb.KeyRange{Begin: start, End: end}
		} else {
			next := tr.Snapshot().GetKey(fdb.FirstGreaterThan(fdb.Key(last))).MustGet()
			scan = fdb.KeyRange{Begin: next, End: end}
		}

		rr := tr.Snapshot().GetRange(scan, fdb.RangeOptions{Limit: limit})

		return rr.GetSliceOrPanic(), nil

	})

	if err != nil {
		panic("Failed to read all events")
	}

	kvs := r.([]fdb.KeyValue)

	result := make([]GlobalRecord, len(kvs))

	for i, kv := range kvs {

		if t, err := globalSpace.Unpack(kv.Key); err != nil {
			panic("Failed to unpack key")
		} else {
			result[i].Contract = t[2].(string)
			result[i].Data = kv.Value
			last = []byte(kv.Key)
		}
	}
	return &GlobalSlice{result, last}
}

func (es *fdbStore) ReadAllFromAggregate(stream string) []AggregateEvent {
	streamSpace := es.space.Sub(aggregPrefix, stream)
	r, err := es.db.ReadTransact(func(tr fdb.ReadTransaction) (interface{}, error) {

		r := fdb.RangeOptions{
			Limit:   0,
			Mode:    fdb.StreamingModeWantAll,
			Reverse: false,
		}
		rr := tr.Snapshot().GetRange(streamSpace, r)
		return rr.GetSliceOrPanic(), nil

	})

	if err != nil {
		panic("Failed to read all from aggregate")
	}

	kvs := r.([]fdb.KeyValue)

	result := make([]AggregateEvent, len(kvs))

	for i, kv := range kvs {

		if t, err := streamSpace.Unpack(kv.Key); err != nil {
			panic("Failed to unpack key")
		} else {
			result[i].Index = int(t[0].(int64))
			result[i].Contract = t[1].(string)
			result[i].Data = kv.Value
		}
	}
	return result
}
