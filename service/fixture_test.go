package service

import (
	"errors"
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
	"wonsoh.private/cloudkitchens/resource"
)

type FixtureTestSuite struct {
	suite.Suite

	mockOrderManager *mockOrderManager
}

type mockOrderManager struct {
	dispatchOrderError bool
	finishOrderError   bool
	finishPickUpError  bool
}

func (m *mockOrderManager) reset() {
	m.dispatchOrderError = false
	m.finishOrderError = false
	m.finishPickUpError = false
}
func (m *mockOrderManager) Init(random *rand.Rand) {}
func (m *mockOrderManager) DispatchOrder(order *resource.Order) error {
	if m.dispatchOrderError {
		return errors.New("DispatchOrder errror")
	}
	return nil
}
func (m *mockOrderManager) Wait()             {}
func (m *mockOrderManager) ReportStatistics() {}
func (m *mockOrderManager) GetStatistics() *OrderManagerStatistics {
	return nil
}
func (m *mockOrderManager) finishOrder(d *dispatchedOrder) error {
	if m.finishOrderError {
		return errors.New("finishOrder error")
	}
	return nil
}
func (m *mockOrderManager) finishPickUp(d *dispatchedCourier) error {
	if m.finishPickUpError {
		return errors.New("finishPickUp error")
	}
	return nil
}

func (f *FixtureTestSuite) SetupTest() {
	f.mockOrderManager = &mockOrderManager{}
}

func (f *FixtureTestSuite) BeforeTest() {
	f.mockOrderManager.reset()
}

func (f *FixtureTestSuite) TestDispatchedOrder() {
	order := getDispatchedOrder(f.mockOrderManager, &resource.Order{
		ID:       "1",
		Name:     "Test Food",
		PrepTime: 1,
	})

	start := time.Now()
	order.processOrder()
	f.GreaterOrEqual(time.Now().Sub(start).Seconds(), float64(1))
	f.mockOrderManager.finishOrderError = true
	f.NotPanics(func() {
		order.processOrder()
	})
	order.PickedUpTime = order.FinishTime.Add(time.Hour)
	f.EqualValues(time.Hour.Milliseconds(), order.getWaitTimeInMs())

}

func (f *FixtureTestSuite) TestDispatchedCourier() {
	courier := getDispatchedCourier(f.mockOrderManager, resource.NewCourier("1", 1))
	start := time.Now()
	courier.pickUpOrder()
	f.GreaterOrEqual(time.Now().Sub(start).Seconds(), float64(1))
	f.mockOrderManager.finishPickUpError = true
	f.NotPanics(func() {
		courier.pickUpOrder()
	})
	courier.PickedUpTime = courier.ArrivedTime.Add(time.Minute)
	f.EqualValues(time.Minute.Milliseconds(), courier.getWaitTimeInMs())
}

func TestFixtureTestSuite(t *testing.T) {
	suite.Run(t, new(FixtureTestSuite))
}
