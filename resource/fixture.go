package resource

import (
	"math/rand"
	"time"

	"github.com/google/uuid"
)

const (
	// MaxTravelTimeRange is a maximum range for travel time
	// [3-15] == [3-16) == [0-13)
	MaxTravelTimeRange = 13
	// MinTravelTime is the minimum travel time
	MinTravelTime = 3
)

// Order represents an object of order dispatched
type Order struct {
	// ID is an identifier of the courier
	ID string `json:"id"`
	// Name is the name of the order
	Name string `json:"name"`
	// PrepTime is the preparation time in seconds
	PrepTime int `json:"prepTime"`
}

// Courier represents a courier to pick-up an order
type Courier struct {
	// ID is an identifier of the courier
	ID string `json:"id"`
	// OrderID is an optional identifier (for picking-up a specific order only)
	OrderID string `json:"order_id"`
	// TravelTime is the time for courier to travel
	TravelTime int `json:"travelTime"`
}

// NewCourier constructs a new courier structure
func NewCourier(orderID string, travelTime int) *Courier {
	return &Courier{
		ID:         uuid.NewString(),
		OrderID:    orderID,
		TravelTime: travelTime,
	}
}

// GetFixedSeedRandomNumberGenerator gets a fixed seed random number generator
// so that the numbers generated are pseudo-random, but the order is deterministic
func GetFixedSeedRandomNumberGenerator() *rand.Rand {
	return rand.New(rand.NewSource(1))
}

// GetTimeBasedSeedRandomNumberGenerator gets a time-based seed random number generator
// so that the numbers generated are pseudo-random, and the order is determined
// by time of generation (hence non-deterministic)
func GetTimeBasedSeedRandomNumberGenerator() *rand.Rand {
	return rand.New(rand.NewSource(time.Now().UnixNano()))
}

// GetCourierTravelTime gets the courier travel time in between 3 and 15 seconds, inclusively
func GetCourierTravelTime(r *rand.Rand) int {
	if r == nil {
		return rand.Intn(MaxTravelTimeRange) + MinTravelTime
	}
	return r.Intn(MaxTravelTimeRange) + MinTravelTime
}
