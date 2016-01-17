package app

import (
	"encoding/json"
	"io/ioutil"
	"time"
)

// Config JSON representation of flight search
type FlightSearch struct {
	OriginAirports      []string  `json:"origin_airports"`
	DestinationAirports []string  `json:"destination_airports"`
	MinDepartureTime    time.Time `json:"min_departure_time"`
	MaxArrivalTime      time.Time `json:"max_arrival_time"`
	MaxFareCents        *uint32   `json:"max_fare_cents"`
	MaxNumberStops      *uint8    `json:"max_number_stops"`
}

func FlightSearchesFromJson(data []byte) ([]*FlightSearch, error) {
	results := make([]*FlightSearch, 0)
	if err := json.Unmarshal(data, &results); err != nil {
		return nil, err
	}
	return results, nil
}

func FlightSearchesFromFile(filename string) ([]*FlightSearch, error) {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	return FlightSearchesFromJson(data)
}
