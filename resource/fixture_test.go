package resource

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"
)

type FixtureTestSuite struct {
	suite.Suite
}

func (t *FixtureTestSuite) TestNewCourier() {
	c := NewCourier("testID", 10)
	t.EqualValues("testID", c.OrderID)
	_, err := uuid.Parse(c.ID)
	t.NoError(err)
	t.EqualValues(10, c.TravelTime)
}

func (t *FixtureTestSuite) TestFixedSeedRandomNumberGenerator() {
	r1, r2 := GetFixedSeedRandomNumberGenerator(), GetFixedSeedRandomNumberGenerator()
	for i := 0; i < 100; i++ {
		t.EqualValues(r1.Intn(100), r2.Intn(100)) // both should be identical
	}
}

func (t *FixtureTestSuite) TestTimeBasedSeedRandomNumberGenerator() {
	r := GetTimeBasedSeedRandomNumberGenerator()
	for i := 0; i < 100; i++ {
		t.Condition(func() bool {
			v := r.Intn(100)
			return 0 <= v && v < 100
		})
	}
}

func (t *FixtureTestSuite) TestGetCourierTravelTime() {
	r := GetTimeBasedSeedRandomNumberGenerator()
	for i := 0; i < 100; i++ {
		t.Condition(func() bool {
			v := GetCourierTravelTime(r)
			return 0 <= 3 && v <= 15
		})
		t.Condition(func() bool {
			v := GetCourierTravelTime(nil)
			return 0 <= 3 && v <= 15
		})
	}
}

func TestFixtureTestSuite(t *testing.T) {
	suite.Run(t, new(FixtureTestSuite))
}
