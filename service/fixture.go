package service

import (
	"log"
	"time"

	"wonsoh.private/cloudkitchens/resource"
)

// dispatchedOrder represents an event with a dispatched order
type dispatchedOrder struct {
	manager      OrderManager
	Order        *resource.Order
	StartTime    time.Time
	FinishTime   time.Time
	PickedUpTime time.Time
	notification chan *dispatchedCourier
}

// dispatchedCourier represents an event with a dispatched courier
type dispatchedCourier struct {
	manager        OrderManager
	Courier        *resource.Courier
	DispatchedTime time.Time
	ArrivedTime    time.Time
	PickedUpTime   time.Time
	notification   chan *dispatchedOrder
}

func (d *dispatchedOrder) processOrder() {
	log.Printf(
		"[ORDER RECEIVED] ID: %s	Name: %s	Prep time: %d second(s)",
		d.Order.ID,
		d.Order.Name,
		d.Order.PrepTime,
	)
	time.Sleep(time.Duration(d.Order.PrepTime) * time.Second)
	d.FinishTime = time.Now()
	log.Printf(
		"[ORDER PREPARED] ID: %s	Name: %s",
		d.Order.ID,
		d.Order.Name,
	)
	if e := d.manager.finishOrder(d); e != nil {
		log.Printf(
			"[ERROR] Error happenned while finishing order for order ID %s (msg: %v)",
			d.Order.ID,
			e,
		)
	}
}

func (d *dispatchedOrder) getWaitTimeInMs() int {
	return int(d.PickedUpTime.Sub(d.FinishTime).Milliseconds())
}

func (d *dispatchedCourier) pickUpOrder() {
	log.Printf(
		"[COURIER DISPATCHED] ID: %s	Travel time: %d second(s)",
		d.Courier.ID,
		d.Courier.TravelTime,
	)
	time.Sleep(time.Duration(d.Courier.TravelTime) * time.Second)
	d.ArrivedTime = time.Now()
	log.Printf(
		"[COURIER ARRIVED] ID: %s",
		d.Courier.ID,
	)
	if e := d.manager.finishPickUp(d); e != nil {
		log.Printf(
			"[ERROR] Error happenned while picking up order for courier ID %s (msg: %v)",
			d.Courier.ID,
			e,
		)
	}
}

func (d *dispatchedCourier) getWaitTimeInMs() int {
	return int(d.PickedUpTime.Sub(d.ArrivedTime).Milliseconds())
}

func logPickUpEvent(
	order *dispatchedOrder,
	courier *dispatchedCourier,
) {
	orderPickUpTime := order.PickedUpTime
	courierPickUpTime := courier.PickedUpTime
	actualPickUpTime := orderPickUpTime // choose the maximum between the two
	if courierPickUpTime.After(orderPickUpTime) {
		actualPickUpTime = courierPickUpTime
	}

	log.Printf(
		`
		===============================================================
		[COURIER PICKED UP FOOD]
		---------------------------------------------------------------
		Courier ID:	%s
		Courier Dispatched: %s	Arrived: %s	Picked Up At:	%s
		---------------------------------------------------------------
		Order Name:	%s
		Order ID:	%s
		Food Cooking Started: %s	Finished: %s	Picked Up At: %s
		---------------------------------------------------------------
		Courier has been waiting for:	%d ms
		Food has been waiting for: %d ms
		===============================================================
		`,
		courier.Courier.ID,
		courier.DispatchedTime.Format(time.StampMilli),
		courier.ArrivedTime.Format(time.StampMilli),
		courier.PickedUpTime.Format(time.StampMilli),
		order.Order.Name,
		order.Order.ID,
		order.StartTime.Format(time.StampMilli),
		order.FinishTime.Format(time.StampMilli),
		order.StartTime.Format(time.StampMilli),
		actualPickUpTime.Sub(courier.ArrivedTime).Milliseconds(),
		actualPickUpTime.Sub(order.FinishTime).Milliseconds(),
	)
}

func getDispatchedOrder(
	m OrderManager,
	order *resource.Order,
) *dispatchedOrder {
	return &dispatchedOrder{
		manager:      m,
		Order:        order,
		StartTime:    time.Now(),
		notification: make(chan *dispatchedCourier),
	}
}

func getDispatchedCourier(
	m OrderManager,
	courier *resource.Courier,
) *dispatchedCourier {
	return &dispatchedCourier{
		manager:        m,
		Courier:        courier,
		DispatchedTime: time.Now(),
		notification:   make(chan *dispatchedOrder),
	}
}
