package app

import (
	"github.com/alecholmes/southwest_flight_watcher/model"
)

type CurrentState struct {
	Flight *model.Flight
	Update UpdateResult
}

type FlightSearchStates map[*FlightSearch]map[model.FlightId]CurrentState

func (f FlightSearchStates) UpdateFlight(search *FlightSearch, flight *model.Flight, update UpdateResult) {
	flightIdUpdate, ok := f[search]
	if !ok {
		f[search] = make(map[model.FlightId]CurrentState)
	}
	flightIdUpdate[*flight.Id()] = CurrentState{flight, update}
}

func (f FlightSearchStates) OnlyAvailable() (available FlightSearchStates, added bool) {
	available = make(FlightSearchStates)
	for search, updatesByFlightId := range f {
		updates := make(map[model.FlightId]CurrentState)
		for flightId, update := range updatesByFlightId {
			if update.Update == Added {
				added = true
			}
			if update.Update != Removed {
				updates[flightId] = update
			}
		}

		if len(updates) > 0 {
			available[search] = updates
		}
	}
	return
}

// Function that takes all flights, and latest updates to those flights.
// Absence of value in the map indicates no change to flight.
// type FlightsNotifier func([]*model.Flight, map[model.FlightId]UpdateResult) error

// Container to keep and update flight state.
type CheapestFlightState struct {
	state    FlightSearchStates
	cache    *CheapestFlightCache
	fetcher  *FlightFetcher
	notifier SearchUpdateNotifier
}

func NewCheapestFlightState(searches []*FlightSearch, fetcher *FlightFetcher, notifier SearchUpdateNotifier) *CheapestFlightState {
	s := &CheapestFlightState{
		state:    make(FlightSearchStates),
		cache:    NewCheapestFlightCache(),
		fetcher:  fetcher,
		notifier: notifier,
	}

	for _, search := range searches {
		s.state[search] = make(map[model.FlightId]CurrentState)
	}

	return s
}

// Fetch latest flight info, apply it to cached values, and send updates to the notifier.
func (c *CheapestFlightState) Update() error {
	updates := make(map[model.FlightId]UpdateResult)
	for search, _ := range c.state {
		flights, err := c.fetcher.Fetch(search)
		if err != nil {
			return err
		}

		for _, flight := range flights {
			var result UpdateResult
			if _, ok := updates[*flight.Id()]; !ok {
				result = c.cache.Update(flight)
				updates[*flight.Id()] = result
			} else {
				result = updates[*flight.Id()]
			}

			c.state.UpdateFlight(search, flight, result)
		}
	}

	c.notifier.Notify(c.state)

	return nil
}
