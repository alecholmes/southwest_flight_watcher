package app

import (
	"github.com/alecholmes/southwest_flight_watcher/model"
)

type FlightStateChange int

const (
	Unchanged FlightStateChange = iota
	FareIncrease
	FareDecrease
	Added
	Removed
)

type FlightState struct {
	Flight *model.Flight
	Update FlightStateChange
}

type FlightStates map[model.FlightId]FlightState

func (s FlightStates) Update(flights []*model.Flight) {
	for _, f := range flights {
		change := Unchanged
		existing, ok := s[*f.Id()]
		if !ok {
			change = Added
		} else {
			fareChange := f.CheapestAvailableFare().Compare(existing.Flight.CheapestAvailableFare())
			if fareChange > 0 {
				change = FareIncrease
			} else if fareChange < 0 {
				change = FareDecrease
			}
		}

		s[*f.Id()] = FlightState{f, change}
	}
}

func (s FlightStates) OnlyAvailable() (available FlightStates, improved bool) {
	available = make(FlightStates)
	for id, state := range s {
		if state.Update != Removed {
			available[id] = state
		}
		if state.Update == Added || state.Update == FareDecrease {
			improved = true
		}
	}
	return
}

type FlightSearchStates map[*FlightSearch]FlightStates

func NewFlightSearchStates() FlightSearchStates {
	return make(FlightSearchStates)
}

func (f FlightSearchStates) Update(search *FlightSearch, flights []*model.Flight) {
	f.getFlightStates(search).Update(flights)
}

func (f FlightSearchStates) OnlyAvailable() (available FlightSearchStates, improved bool) {
	available = make(FlightSearchStates)

	for search, states := range f {
		availableStates, statesImproved := states.OnlyAvailable()
		if len(availableStates) > 0 {
			available[search] = availableStates
			if statesImproved {
				improved = true
			}
		}
	}
	return
}

func (f FlightSearchStates) getFlightStates(search *FlightSearch) FlightStates {
	states, ok := f[search]
	if !ok {
		states = make(FlightStates)
		f[search] = states
	}
	return states
}
