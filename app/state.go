package app

import (
	"github.com/alecholmes/southwest_flight_watcher/model"
)

// Function that takes all flights, and latest updates to those flights.
// Absence of value in the map indicates no change to flight.
type FlightsNotifier func([]*model.Flight, map[model.FlightId]UpdateResult) error

// Container to keep and update flight state.
type CheapestFlightState struct {
	searches []*FlightSearch
	cache    *CheapestFlightCache
	fetcher  *FlightFetcher
	notifier FlightsNotifier
}

func NewCheapestFlightState(searches []*FlightSearch, fetcher *FlightFetcher, notifier FlightsNotifier) *CheapestFlightState {
	return &CheapestFlightState{
		searches: searches,
		cache:    NewCheapestFlightCache(),
		fetcher:  fetcher,
		notifier: notifier,
	}
}

// Fetch latest flight info, apply it to cached values, and send updates to the notifier.
func (c *CheapestFlightState) Update() error {
	updates := make(map[model.FlightId]UpdateResult)
	for _, search := range c.searches {
		flights, err := c.fetcher.Fetch(search)
		if err != nil {
			return err
		}

		for _, flight := range flights {
			result := c.cache.Update(flight)
			if result != Unchanged {
				updates[*flight.Id()] = result
			}
		}
	}

	return c.notifier(c.cache.Values(), updates)
}
