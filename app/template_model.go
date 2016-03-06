package app

import (
	"fmt"
	"sort"

	"github.com/alecholmes/southwest_flight_watcher/model"
)

// Structs that map to templates
type Body struct {
	SearchGroups []*SearchGroup
}

type SearchGroup struct {
	Date    string
	MaxFare *string
	Note    *string
	Trips   []*Trip
}

type Trip struct {
	OriginAirport      string
	DestinationAirport string
	DepartureTime      string
	ArrivalTime        string
	Stops              int
	Cost               string
	Update             FlightStateChange
}

func NewBody(searches FlightSearchStates) *Body {
	searchGroups := make([]*SearchGroup, 0, len(searches))

	for _, search := range sortedFlightSearches(searches) {
		updates := searches[search]
		trips := make([]*Trip, 0, len(updates))

		for _, update := range sortedFlightStates(updates) {
			var cost string
			if fare := update.Flight.CheapestAvailableFare(); fare != nil {
				fareCents := update.Flight.CheapestAvailableFare().Cents
				cost = fmt.Sprintf("$ %d.%02d", fareCents/100, fareCents%100)
			}

			trips = append(trips, &Trip{
				OriginAirport:      update.Flight.OriginAirport,
				DestinationAirport: update.Flight.DestinationAirport,
				DepartureTime:      update.Flight.DepartureLocalTime.Format("3:04 PM"),
				ArrivalTime:        update.Flight.ArrivalLocalTime.Format("3:04 PM"),
				Stops:              len(update.Flight.Stops),
				Cost:               cost,
				Update:             update.Update,
			})
		}

		searchGroup := &SearchGroup{
			Date:  search.MinDepartureTime.Format("Mon Jan 2 2006"),
			Note:  search.Note,
			Trips: trips,
		}

		if search.MaxFareCents != nil {
			maxFare := fmt.Sprintf("$ %d.%02d", *search.MaxFareCents/100, *search.MaxFareCents%100)
			searchGroup.MaxFare = &maxFare
		}

		searchGroups = append(searchGroups, searchGroup)
	}
	return &Body{SearchGroups: searchGroups}
}

// Helper to sort FlightSearch by MinDepartureTime
type FlightSearchByMinDepartureTime []*FlightSearch

func (b FlightSearchByMinDepartureTime) Len() int      { return len(b) }
func (b FlightSearchByMinDepartureTime) Swap(i, j int) { b[i], b[j] = b[j], b[i] }
func (b FlightSearchByMinDepartureTime) Less(i, j int) bool {
	return b[i].MinDepartureTime.Before(b[j].MinDepartureTime)
}

func sortedFlightSearches(s FlightSearchStates) []*FlightSearch {
	searches := make([]*FlightSearch, 0, len(s))
	for search, _ := range s {
		searches = append(searches, search)
	}
	sort.Sort(FlightSearchByMinDepartureTime(searches))

	return searches
}

// Helper to sort FlightState by flight DepartureLocalTime
type FlightStateByDepartureLocalTime []FlightState

func (b FlightStateByDepartureLocalTime) Len() int      { return len(b) }
func (b FlightStateByDepartureLocalTime) Swap(i, j int) { b[i], b[j] = b[j], b[i] }
func (b FlightStateByDepartureLocalTime) Less(i, j int) bool {
	return b[i].Flight.DepartureLocalTime.Before(b[j].Flight.DepartureLocalTime)
}

func sortedFlightStates(u map[model.FlightId]FlightState) []FlightState {
	updates := make([]FlightState, 0, len(u))
	for _, update := range u {
		updates = append(updates, update)
	}
	sort.Sort(FlightStateByDepartureLocalTime(updates))

	return updates
}
