package reader

import (
	"encoding/json"
	"io/ioutil"
	"os"

	"wonsoh.private/cloudkitchens/resource"
)

// OrderReader is a reader that reads order
type OrderReader interface {
	ReadOrders() ([]*resource.Order, error)
}

type orderReaderImpl struct{}

func (o *orderReaderImpl) ReadOrders() ([]*resource.Order, error) {
	ordersFile, err := os.Open("resource/dispatch_orders.json")
	if err != nil {
		return nil, err
	}
	defer ordersFile.Close()
	orderBytes, err := ioutil.ReadAll(ordersFile)
	if err != nil {
		return nil, err
	}
	orders := make([]*resource.Order, 10)
	if e := json.Unmarshal(orderBytes, &orders); e != nil {
		return nil, e
	}
	return orders, nil
}

// GetOrderReader constructs a new OrderReader instance
func GetOrderReader() OrderReader {
	return &orderReaderImpl{}
}
