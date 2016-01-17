package app

import (
	"github.com/alecholmes/southwest_flight_watcher/model"
)

type UpdateResult int

const (
	Unchanged UpdateResult = iota
	Added
	Removed
)

type CheapestFlightCache struct {
	cache map[model.FlightId]*model.Flight
}

func NewCheapestFlightCache() *CheapestFlightCache {
	return &CheapestFlightCache{
		cache: make(map[model.FlightId]*model.Flight),
	}
}

func (c *CheapestFlightCache) Values() []*model.Flight {
	flights := make([]*model.Flight, 0, len(c.cache))
	for _, flight := range c.cache {
		flights = append(flights, flight)
	}
	return flights
}

func (c *CheapestFlightCache) Update(flight *model.Flight) UpdateResult {
	id := *flight.Id()
	newFare := flight.CheapestAvailableFare()

	oldFlight, existed := c.cache[id]
	if !existed {
		if newFare == nil {
			return Unchanged
		} else {
			c.cache[id] = flight
			return Added
		}
	}

	if oldFlight.CheapestAvailableFare().Cents > newFare.Cents {
		c.cache[id] = flight
		return Added
	} else {
		return Unchanged
	}
}
