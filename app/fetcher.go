package app

import (
	"github.com/alecholmes/southwest_flight_watcher/client"
	"github.com/alecholmes/southwest_flight_watcher/model"
)

type FlightFetcher struct {
	swClient *client.Client
}

func NewFlightFetcher(swClient *client.Client) *FlightFetcher {
	return &FlightFetcher{swClient}
}

// Fetch all flights that match a given search
func (f *FlightFetcher) Fetch(search *FlightSearch) ([]*model.Flight, error) {
	filters := make([]client.FlightFilter, 0)
	if search.MaxFareCents != nil {
		filters = append(filters, &client.MaxAvailableFareFilter{*search.MaxFareCents})
	}
	if search.MaxNumberStops != nil {
		filters = append(filters, &client.MaxStopsFilter{int(*search.MaxNumberStops)})
	}

	flights, err := f.swClient.SearchFlights(
		search.OriginAirports,
		search.DestinationAirports,
		search.MinDepartureTime,
		search.MaxArrivalTime,
		filters)
	if err != nil {
		return nil, err
	}
	return flights, nil
}
