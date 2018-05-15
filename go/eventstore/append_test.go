package events

import (
	"testing"

	"github.com/apple/foundationdb/bindings/go/src/fdb"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

// declare our unit test suite
type given_empty_store struct {
	store Store
	suite.Suite
}

func TestGivenEmptyStore(t *testing.T) {
	suite.Run(t, new(given_empty_store))
}

func (s *given_empty_store) SetupTest() {
	fdb.MustAPIVersion(510)

	db := fdb.MustOpenDefault()
	s.store = NewFdbStore(db, "es")
}
func (s *given_empty_store) TearDownTest() {
	s.store.Clear()
}

func (s *given_empty_store) Test_when_we_append_one_record() {
	// when
	evt := []Envelope{New("Test", []byte("Hi"))}
	err := s.store.AppendToAggregate("test1", ExpectedVersionAny, evt)

	a := assert.New(s.T())
	a.NoError(err)

	recs := s.store.ReadAll(nil, 0).Items
	expectedGlobal := []GlobalRecord{
		GlobalRecord{
			Contract: "Test",
			Data:     []byte("Hi"),
		},
	}

	globalEquals(a, expectedGlobal, recs)

	expectedAggregate := []AggregateEvent{
		AggregateEvent{
			Contract: "Test",
			Data:     []byte("Hi"),
			Index:    0,
		},
	}

	aggregateEquals(a, expectedAggregate, s.store.ReadAllFromAggregate("test1"))
}

func aggregateEquals(a *assert.Assertions, exp, act []AggregateEvent) {
	if !a.Equal(len(exp), len(act)) {
		return
	}

	for i, e := range exp {
		a.Equal(e.Contract, act[i].Contract)
		a.Equal(e.Data, act[i].Data)
		a.Equal(e.Index, act[i].Index)
	}
}

func globalEquals(a *assert.Assertions, exp, act []GlobalRecord) {
	if a.Equal(len(exp), len(act), "global records") {

		for i, e := range exp {

			a.Equal(e.Contract, act[i].Contract)
			a.Equal(e.Data, act[i].Data)
		}
	}
}

func (s *given_empty_store) Test_when_we_append_two_records_at_once() {
	r1 := New("Test", []byte("One"))
	r2 := New("Test", []byte("Two"))
	err := s.store.AppendToAggregate("test1", ExpectedVersionAny, []Envelope{r1, r2})

	a := assert.New(s.T())
	a.NoError(err)

	recs := s.store.ReadAll(nil, 0).Items
	expectedGlobal := []GlobalRecord{
		GlobalRecord{
			Contract: "Test",
			Data:     []byte("One"),
		},
		GlobalRecord{
			Contract: "Test",
			Data:     []byte("Two"),
		},
	}

	globalEquals(a, expectedGlobal, recs)

	expectedAggregate := []AggregateEvent{
		AggregateEvent{
			Contract: "Test",
			Data:     []byte("One"),
			Index:    0,
		},
		AggregateEvent{
			Contract: "Test",
			Data:     []byte("Two"),
			Index:    1,
		},
	}

	aggregateEquals(a, expectedAggregate, s.store.ReadAllFromAggregate("test1"))
}

func (s *given_empty_store) Test_when_we_append_expecting_some_version() {

	evt := New("Test", []byte("Hi"))
	err := s.store.AppendToAggregate("test1", 1, []Envelope{evt})

	expect := &ErrConcurrencyViolation{
		AggregateId:     "test1",
		ExpectedVersion: 1,
		ActualVersion:   -1,
	}

	assert.Equal(s.T(), expect, err)

}

func (s *given_empty_store) Test_when_we_append_expecting_0_version() {
	evt := New("Test", []byte("Hi"))
	err := s.store.AppendToAggregate("test1", 0, []Envelope{evt})

	expect := &ErrConcurrencyViolation{
		AggregateId:     "test1",
		ExpectedVersion: 0,
		ActualVersion:   -1,
	}

	assert.Equal(s.T(), expect, err)
}

func (s *given_empty_store) Test_when_we_append_expecting_no_aggregate() {
	evt := New("Test", []byte("Hi"))

	err := s.store.AppendToAggregate("test1", ExpectedVersionNone, []Envelope{evt})

	assert.NoError(s.T(), err)
}

func (s *given_empty_store) Test_when_we_read_records_from_start() {
	slice := s.store.ReadAll(nil, 10)

	a := assert.New(s.T())
	a.Equal(len(slice.Items), 0)
	a.Nil(slice.Last)
}
