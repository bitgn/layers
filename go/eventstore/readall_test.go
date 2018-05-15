package events

import (
	"testing"

	"github.com/apple/foundationdb/bindings/go/src/fdb"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

// declare our unit test suite

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestGivenFilledStore(t *testing.T) {
	suite.Run(t, new(given_filled_store))
}

type given_filled_store struct {
	suite.Suite
	store Store
}

func (s *given_filled_store) SetupTest() {
	fdb.MustAPIVersion(510)
	db := fdb.MustOpenDefault()
	s.store = NewFdbStore(db, "es")

	r1 := New("Test", []byte("One"))
	r2 := New("Test", []byte("Two"))
	s.store.AppendToAggregate("test", ExpectedVersionAny, []Envelope{r1, r2})
}
func (s *given_filled_store) TearDownTest() {
	s.store.Clear()
}

func (s *given_filled_store) Test_when_we_read_records_by_one() {
	slice1 := s.store.ReadAll(nil, 1)

	a := assert.New(s.T())

	a.Equal(len(slice1.Items), 1)
	a.NotNil(slice1.Last)

	slice2 := s.store.ReadAll(slice1.Last, 1)

	a.Equal(len(slice2.Items), 1)
	a.NotNil(slice2.Last)

	slice3 := s.store.ReadAll(slice2.Last, 1)

	a.Equal(len(slice3.Items), 0)
	a.Equal(slice3.Last, slice2.Last)
}
