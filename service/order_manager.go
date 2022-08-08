package service

import (
	"container/list"
	"log"
	"math/rand"
	"sync"
	"time"

	"wonsoh.private/cloudkitchens/resource"
)

var (
	matchedOrderManagerInstance OrderManager
	fifoOrderManagerInstance    OrderManager
)

// OrderManagerStatistics represents a statistics object
type OrderManagerStatistics struct {
	TotalOrderCount      int
	TotalFoodWaitTime    int
	TotalCourierWaitTime int

	mutex *sync.Mutex
}

func (o *OrderManagerStatistics) GetAverageStatistics() (
	avgFoodWaitTime float64,
	avgCourierWaitTime float64,
) {
	if o != nil && o.TotalOrderCount > 0 {
		avgFoodWaitTime = float64(o.TotalFoodWaitTime) / float64(o.TotalOrderCount)
		avgCourierWaitTime = float64(o.TotalCourierWaitTime) / float64(o.TotalOrderCount)
	}
	return
}

func (o *OrderManagerStatistics) IncrementTotalOrderCount() {
	o.mutex.Lock()
	defer o.mutex.Unlock()
	o.TotalOrderCount++
}

func (o *OrderManagerStatistics) IncrementTotalFoodWaitTime(byMs int) {
	o.mutex.Lock()
	defer o.mutex.Unlock()
	o.TotalFoodWaitTime += byMs
}

func (o *OrderManagerStatistics) IncrementTotalCourierWaitTime(byMs int) {
	o.mutex.Lock()
	defer o.mutex.Unlock()
	o.TotalCourierWaitTime += byMs
}

func (o *OrderManagerStatistics) ReportStatistics() {
	if o == nil || o.TotalOrderCount == 0 {
		log.Printf(
			`
			NO ORDERS HAVE BEEN PROCESSED. NO STATISTICS TO REPORT.
			`,
		)
	} else {
		avgFoodWaitTime, avgCourierWaitTime := o.GetAverageStatistics()
		log.Printf(
			`
		***************************************************************
		[ALL ORDERS HAVE BEEN PROCESSED]
		Total Order Count: %d order(s)
		Average Food Wait Time: %.4f ms
		Average Courier Wait Time: %.4f ms
		***************************************************************
		`,
			o.TotalOrderCount,
			avgFoodWaitTime,
			avgCourierWaitTime,
		)

	}
}

// OrderManager is a generic interface that performs dispatching of an order
type OrderManager interface {
	Init(random *rand.Rand)
	DispatchOrder(order *resource.Order) error
	Wait()
	ReportStatistics()
	GetStatistics() *OrderManagerStatistics

	// private functions
	finishOrder(d *dispatchedOrder) error
	finishPickUp(d *dispatchedCourier) error
}

type orderManagerBase struct {
	mutex  *sync.RWMutex
	wg     *sync.WaitGroup
	random *rand.Rand

	stats *OrderManagerStatistics
}

type matchedOrderManager struct {
	*orderManagerBase
	finishedOrderMap *sync.Map
	courierMap       *sync.Map
}

type fifoOrderManager struct {
	*orderManagerBase
	finishedOrderQueue *list.List
	courierQueue       *list.List
}

// Init initializes the order manager instance
func (o *orderManagerBase) Init(random *rand.Rand) {
	o.random = random
	o.stats = &OrderManagerStatistics{
		mutex: &sync.Mutex{},
	}
}

func (o *orderManagerBase) lock() {
	o.mutex.Lock()
}

func (o *orderManagerBase) unlock() {
	o.mutex.Unlock()
}

func (o *orderManagerBase) wgAdd() {
	o.wg.Add(1)
}

func (o *orderManagerBase) wgDone() {
	o.wg.Done()
}

// finishing order by decrementing waitgroup count and increasing order count
func (o *orderManagerBase) completeOrder() {
	o.stats.IncrementTotalOrderCount()
	o.wgDone()
}

func (o *orderManagerBase) incrementTotalFoodWaitTime(byMs int) {
	o.stats.IncrementTotalFoodWaitTime(byMs)
}

func (o *orderManagerBase) incrementTotalCourierWaitTime(byMs int) {
	o.stats.IncrementTotalCourierWaitTime(byMs)
}

// Wait waits for order manager to be done
func (o *orderManagerBase) Wait() {
	o.wg.Wait()
}

func (o *orderManagerBase) ReportStatistics() {
	o.stats.ReportStatistics()
}

func (o *orderManagerBase) GetStatistics() *OrderManagerStatistics {
	return o.stats
}

// Init initializes matched order manager instance
func (m *matchedOrderManager) Init(random *rand.Rand) {
	m.orderManagerBase.Init(random)
	m.finishedOrderMap = &sync.Map{}
	m.courierMap = &sync.Map{}
}

// Init initializes FIFO order manager instance
func (f *fifoOrderManager) Init(random *rand.Rand) {
	f.orderManagerBase.Init(random)
	f.finishedOrderQueue.Init()
	f.courierQueue.Init()
}

// DispatchOrder dispatches order to the order manager (using matched strategy)
func (m *matchedOrderManager) DispatchOrder(order *resource.Order) error {
	m.wgAdd()
	log.Printf(
		`
		===============================================================
		[ORDER DISPATCHED] ID: %s 
		Order Name:		%s	Preparation Time (s):	%d
		===============================================================
		`,
		order.ID,
		order.Name,
		order.PrepTime,
	)
	dispatchedOrder := getDispatchedOrder(m, order)
	dispatchedCourier := getDispatchedCourier(
		m,
		resource.NewCourier(
			order.ID,
			resource.GetCourierTravelTime(m.random),
		),
	)
	go dispatchedOrder.processOrder()  // non-blocking
	go dispatchedCourier.pickUpOrder() // non-blocking
	return nil
}

// DispatchOrder dispatches order to the order manager (using FIFO strategy)
func (f *fifoOrderManager) DispatchOrder(order *resource.Order) error {
	f.wgAdd()
	log.Printf(
		`
		===============================================================
		[ORDER DISPATCHED] ID: %s 
		Order Name:		%s	Preparation Time (s):	%d
		===============================================================
		`,
		order.ID,
		order.Name,
		order.PrepTime,
	)
	dispatchedOrder := getDispatchedOrder(f, order)
	dispatchedCourier := getDispatchedCourier(
		f,
		resource.NewCourier(
			order.ID,
			resource.GetCourierTravelTime(f.random),
		),
	)
	go dispatchedOrder.processOrder()  // non-blocking
	go dispatchedCourier.pickUpOrder() // non-blocking
	return nil
}

// finishOrder <private> finish order (food) for matched strategy
func (m *matchedOrderManager) finishOrder(order *dispatchedOrder) error {
	m.lock() // global lock to prevent deadlock for channel
	m.finishedOrderMap.Store(order.Order.ID, order)
	courier, ok := m.courierMap.Load(order.Order.ID)
	m.unlock()
	if ok { // finished, and waiting courier found (order GETS PICKED UP by courier)
		order.PickedUpTime = time.Now()
		courier.(*dispatchedCourier).notification <- order
		defer m.completeOrder() // one order is processed, so decrement the event wait group by one
	} else { // since courier is not found, wait in line
		<-order.notification            // wait for courier to be ready
		order.PickedUpTime = time.Now() // picked up
	}
	m.incrementTotalFoodWaitTime(order.getWaitTimeInMs())
	return nil
}

// evictFromFinishedOrderQueue evicts an element from courier queue using FIFO strategy
func (f *fifoOrderManager) evictFromFinishedOrderQueue(elem *list.Element) {
	f.lock()
	defer f.unlock()
	f.finishedOrderQueue.Remove(elem) // evict self from the queue since order has ben picked up
}

// finishOrder <private> finish order (food) for FIFO strategy
func (f *fifoOrderManager) finishOrder(order *dispatchedOrder) error {
	var courier *dispatchedCourier
	f.lock() // global lock to prevent deadlock for channel
	elem := f.finishedOrderQueue.PushBack(order)
	ok := f.courierQueue.Len() > 0
	if ok {
		courier = f.courierQueue.Remove(f.courierQueue.Front()).(*dispatchedCourier)
	}
	f.unlock()
	if ok { // finished, and waiting courier found (order GETS PICKED UP by courier)
		order.PickedUpTime = time.Now()
		f.evictFromFinishedOrderQueue(elem) // picked up; evict
		courier.notification <- order
		defer f.completeOrder() // one order is processed, so decrement the event wait group by one
	} else { // since courier is not found, wait in line
		<-order.notification                // wait for courier to be ready
		f.evictFromFinishedOrderQueue(elem) // picked up; evict
		order.PickedUpTime = time.Now()     // picked up
	}
	f.incrementTotalFoodWaitTime(order.getWaitTimeInMs())
	return nil
}

// finishPickUp <private> finish pick-up (courier) for matched strategy
func (m *matchedOrderManager) finishPickUp(courier *dispatchedCourier) error {
	m.lock() // global lock to prevent deadlock for channel
	m.courierMap.Store(courier.Courier.OrderID, courier)
	order, ok := m.finishedOrderMap.Load(courier.Courier.OrderID)
	m.unlock()
	if ok { // arrived, and order found (courier PICKS UP the order)
		courier.PickedUpTime = time.Now()
		order.(*dispatchedOrder).notification <- courier
		defer m.completeOrder() // one order is processed, so decrement the event wait group by one
	} else {
		order = <-courier.notification // wait for order to be ready
		courier.PickedUpTime = time.Now()
	}
	dOrder := order.(*dispatchedOrder)
	logPickUpEvent(dOrder, courier)
	m.incrementTotalCourierWaitTime(courier.getWaitTimeInMs())
	return nil
}

// evictFromCourierQueue evicts an element from courier queue using FIFO strategy
func (f *fifoOrderManager) evictFromCourierQueue(elem *list.Element) {
	f.lock()
	defer f.unlock()
	f.courierQueue.Remove(elem)
}

// finishPickUp <private> finish pick-up (courier) for FIFO strategy
func (f *fifoOrderManager) finishPickUp(courier *dispatchedCourier) error {
	var order *dispatchedOrder
	f.lock() // global lock to prevent deadlock for channel
	elem := f.courierQueue.PushBack(courier)
	ok := f.finishedOrderQueue.Len() > 0
	if ok {
		order = f.finishedOrderQueue.Remove(f.finishedOrderQueue.Front()).(*dispatchedOrder)
	}
	f.unlock()
	if ok { // arrived, and order found (courier PICKS UP the order)
		courier.PickedUpTime = time.Now()
		f.evictFromCourierQueue(elem) // picked up; evict
		order.notification <- courier
		defer f.completeOrder() // one order is processed, so decrement the event wait group by one
	} else {
		order = <-courier.notification // wait for order to be ready
		f.evictFromCourierQueue(elem)  // picked up; evict
		courier.PickedUpTime = time.Now()
	}
	logPickUpEvent(order, courier)
	f.incrementTotalCourierWaitTime(courier.getWaitTimeInMs())
	return nil
}

func getOrderManagerBaseClass(random *rand.Rand) *orderManagerBase {
	return &orderManagerBase{
		random: random,
		mutex:  &sync.RWMutex{},
		wg:     &sync.WaitGroup{},
		stats: &OrderManagerStatistics{
			mutex: &sync.Mutex{},
		},
	}
}

// GetMatchedOrderManager gets the singleton instance of order manager that uses
// assigned order strategy
func GetMatchedOrderManager(random *rand.Rand) OrderManager {
	if matchedOrderManagerInstance == nil {
		matchedOrderManagerInstance = &matchedOrderManager{
			orderManagerBase: getOrderManagerBaseClass(random),
			finishedOrderMap: &sync.Map{},
			courierMap:       &sync.Map{},
		}
	}
	return matchedOrderManagerInstance
}

// GetFIFOOrderManager gets the singleton instance of order manager that uses
// FIFO order strategy
func GetFIFOOrderManager(random *rand.Rand) OrderManager {
	if fifoOrderManagerInstance == nil {
		fifoOrderManagerInstance = &fifoOrderManager{
			orderManagerBase:   getOrderManagerBaseClass(random),
			finishedOrderQueue: list.New(),
			courierQueue:       list.New(),
		}
	}
	return fifoOrderManagerInstance
}
