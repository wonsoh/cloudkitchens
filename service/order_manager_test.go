package service

import (
	"math/rand"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/suite"
	"wonsoh.private/cloudkitchens/mocks"
	"wonsoh.private/cloudkitchens/resource"
)

var (
	testOrders = []*resource.Order{
		{
			ID:       "1",
			Name:     "Food 1",
			PrepTime: 2,
		},
		{
			ID:       "2",
			Name:     "Food 2",
			PrepTime: 10,
		},
		{
			ID:       "3",
			Name:     "Food 3",
			PrepTime: 4,
		},
		{
			ID:       "4",
			Name:     "Food 4",
			PrepTime: 6,
		},
	}

	/**
	 	For matched strategy:
		Courier 1 arrives AFTER (food waits 2 second)
		Courier 2 arrives BEFORE (courier waits 5 seconds)
		Courier 3 arrives BEFORE (courier waits 1 second)
		Courier 4 arrives AFTER (food waits 2 seconds)

		For FIFO strategy:
		Food 1 gets queued in [2s]
		Courier 3 picks up Food 1 (Food 1 waits 1 second) [3s]
		Food 3 gets queued in + Courier 1 arrives (neither waits) [4s]
		Courier 2 gets queued in [5s]
		Food 4 gets picked up by Courier 2 (Courier 2 waits 1 second) [6]
		Courier 4 gets queued in [8s]
		Food 2 gets picked up by Courier 4 (Courier 4 waits 2 seconds) [10s]
	*/
	testCourierTravelTimes = []int{
		4,
		5,
		3,
		8,
	}
)

type OrderManagerTestSuite struct {
	suite.Suite
	ctrl *gomock.Controller
}

func (o *OrderManagerTestSuite) SetupTest() {
	o.ctrl = gomock.NewController(o.T())
}

func (o *OrderManagerTestSuite) getMockRand() *rand.Rand {
	mockRandSrc := mocks.NewMockSource(o.ctrl)
	pointer := 0
	mockRandSrc.EXPECT().Int63().AnyTimes().DoAndReturn(func() int64 {
		n := len(testCourierTravelTimes)
		value := testCourierTravelTimes[pointer%n]
		pointer++
		return int64((value - resource.MinTravelTime) << 32)
	}).AnyTimes()
	return rand.New(mockRandSrc)
}

func (o *OrderManagerTestSuite) TestTravelTimeGeneration() {
	random := o.getMockRand()
	for _, travelTime := range testCourierTravelTimes {
		o.Equal(travelTime, resource.GetCourierTravelTime(random))
	}
}

func (o *OrderManagerTestSuite) TestOrderManagerBase() {
	random := o.getMockRand()
	base := getOrderManagerBaseClass(random)
	base.wgAdd()
	base.wgAdd()
	go func(b *orderManagerBase) {
		defer b.completeOrder()
		b.incrementTotalFoodWaitTime(14)
	}(base)
	go (func(b *orderManagerBase) {
		defer b.completeOrder()
		b.incrementTotalCourierWaitTime(6)
	})(base)
	base.Wait()
	stats := base.GetStatistics()
	avgFoodWaitTime, avgCourierWaitTime := stats.GetAverageStatistics()
	o.Equal(2, stats.TotalOrderCount)
	o.Equal(14, stats.TotalFoodWaitTime)
	o.Equal(6, stats.TotalCourierWaitTime)
	o.EqualValues(7.0, avgFoodWaitTime)
	o.EqualValues(3.0, avgCourierWaitTime)
	base.Init(random)
	o.NotPanics(func() {
		base.GetStatistics().GetAverageStatistics()
		base.ReportStatistics()
	})
}

func (o *OrderManagerTestSuite) TestMatchedOrderManager() {
	// For matched strategy:
	// Courier 1 arrives AFTER (food waits 2 second)
	// Courier 2 arrives BEFORE (courier waits 5 seconds)
	// Courier 3 arrives BEFORE (courier waits 1 second)
	// Courier 4 arrives AFTER (food waits 2 seconds)
	// Food waits total of 4 seconds (avg 1000 ms)
	// Courier waits total of 6 seconds (avg 750 ms)
	random := o.getMockRand()
	manager := GetMatchedOrderManager(random)
	o.Equal(manager, GetMatchedOrderManager(random)) // test singleton
	for _, order := range testOrders {
		o.NoError(manager.DispatchOrder(order))
	}
	manager.Wait()
	time.Sleep(1 * time.Second)
	o.NotPanics(func() {
		manager.ReportStatistics()
	})
	stats := manager.GetStatistics()
	totalFoodWaitTimeRounded := (time.Duration(stats.TotalFoodWaitTime) * time.Millisecond).Round(time.Second)
	totalCourierWaitTimeRounded := (time.Duration(stats.TotalCourierWaitTime) * time.Millisecond).Round(time.Second)
	o.EqualValues(4, stats.TotalOrderCount)
	o.EqualValues(
		time.Duration(4)*time.Second,
		totalFoodWaitTimeRounded,
	)
	o.EqualValues(
		time.Duration(6)*time.Second,
		totalCourierWaitTimeRounded,
	)

	manager.Init(random)
	o.NotPanics(func() {
		manager.GetStatistics().GetAverageStatistics()
		manager.ReportStatistics()
	})
}

func (o *OrderManagerTestSuite) TestFIFOOrderManager() {
	// For FIFO strategy:
	// Food 1 gets queued in [2s]
	// Courier 3 picks up Food 1 (Food 1 waits 1 second) [3s]
	// Food 3 gets queued in + Courier 1 arrives (neither waits) [4s]
	// Courier 2 gets queued in [5s]
	// Food 4 gets picked up by Courier 2 (Courier 2 waits 1 second) [6]
	// Courier 4 gets queued in [8s]
	// Food 2 gets picked up by Courier 4 (Courier 4 waits 2 seconds) [10s]
	// Food waits total of 1 second (avg 250 ms)
	// Courier waits total of 3 seconds (avg 750 ms)
	random := o.getMockRand()
	manager := GetFIFOOrderManager(random)
	o.Equal(manager, GetFIFOOrderManager(random)) // test singleton
	for _, order := range testOrders {
		o.NoError(manager.DispatchOrder(order))
	}
	manager.Wait()
	time.Sleep(1 * time.Second)
	o.NotPanics(func() {
		manager.ReportStatistics()
	})
	stats := manager.GetStatistics()
	totalFoodWaitTimeRounded := (time.Duration(stats.TotalFoodWaitTime) * time.Millisecond).Round(time.Second)
	totalCourierWaitTimeRounded := (time.Duration(stats.TotalCourierWaitTime) * time.Millisecond).Round(time.Second)
	o.EqualValues(4, stats.TotalOrderCount)
	o.EqualValues(
		time.Duration(1)*time.Second,
		totalFoodWaitTimeRounded,
	)
	o.EqualValues(
		time.Duration(3)*time.Second,
		totalCourierWaitTimeRounded,
	)

	manager.Init(random)
	o.NotPanics(func() {
		manager.GetStatistics().GetAverageStatistics()
		manager.ReportStatistics()
	})
}

func TestOrderManagerTestSuite(t *testing.T) {
	suite.Run(t, new(OrderManagerTestSuite))
}
