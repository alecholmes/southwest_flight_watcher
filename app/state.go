package app

import (
	"github.com/alecholmes/southwest_flight_watcher/model"
)

// Function that returns flights.
type FlightsFetcher func() ([]*model.Flight, error)

// Function that takes all flights, and latest updates to those flights.
// Absence of value in the map indicates no change to flight.
type FlightsNotifier func([]*model.Flight, map[model.FlightId]UpdateResult) error

// Container to keep and update flight state.
type CheapestFlightState struct {
	cache    *CheapestFlightCache
	fetcher  FlightsFetcher
	notifier FlightsNotifier
}

func NewCheapestFlightState(fetcher FlightsFetcher, notifier FlightsNotifier) *CheapestFlightState {
	return &CheapestFlightState{
		cache:    NewCheapestFlightCache(),
		fetcher:  fetcher,
		notifier: notifier,
	}
}

// Fetch latest flight info, apply it to cached values, and send updates to the notifier.
func (c *CheapestFlightState) Update() error {
	flights, err := c.fetcher()
	if err != nil {
		return err
	}

	updates := make(map[model.FlightId]UpdateResult)
	for _, flight := range flights {
		result := c.cache.Update(flight)
		if result != Unchanged {
			updates[*flight.Id()] = result
		}
	}

	return c.notifier(c.cache.Values(), updates)
}
