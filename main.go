package main

import (
	"flag"
	"fmt"
	"log"

	"wonsoh.private/cloudkitchens/reader"
	"wonsoh.private/cloudkitchens/resource"
	"wonsoh.private/cloudkitchens/service"
)

func main() {
	strategy := flag.Int("s", 0, "strategy value to use. 0 for matched; 1 for FIFO. [default is 0--matched]")
	flag.Parse()
	reader := reader.GetOrderReader()
	orders, _ := reader.ReadOrders()
	random := resource.GetFixedSeedRandomNumberGenerator()
	var manager service.OrderManager
	switch *strategy {
	case 1:
		manager = service.GetFIFOOrderManager(
			random,
		)
	default:
		manager = service.GetMatchedOrderManager(
			random,
		)
	}
	for _, order := range orders {
		if e := manager.DispatchOrder(order); e != nil {
			log.Panic(e)
		}
	}
	manager.Wait()
	manager.ReportStatistics()
	fmt.Println("DONE") // this line should appear after all orders have been processed
}
