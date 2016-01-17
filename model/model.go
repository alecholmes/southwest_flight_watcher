package model

import (
	"strings"
	"time"
)

type Fare struct {
	Cents          uint32
	SeatsAvailable uint32
}

type FlightId struct {
	originAirport      string
	destinationAirport string
	departureLocalTime time.Time
	arrivalLocalTime   time.Time
	stops              string
}

type Flight struct {
	OriginAirport      string
	DestinationAirport string
	DepartureLocalTime time.Time
	ArrivalLocalTime   time.Time
	Stops              []string
	Fares              []*Fare
}

func (f *Flight) Id() *FlightId {
	return &FlightId{
		originAirport:      f.OriginAirport,
		destinationAirport: f.DestinationAirport,
		departureLocalTime: f.DepartureLocalTime,
		arrivalLocalTime:   f.ArrivalLocalTime,
		stops:              strings.Join(f.Stops, ":"),
	}
}

func (f *Flight) CheapestAvailableFare() *Fare {
	var cheapestFare *Fare
	for _, fare := range f.Fares {
		if fare.SeatsAvailable > 0 {
			if cheapestFare == nil || fare.Cents < cheapestFare.Cents {
				cheapestFare = fare
			}
		}
	}
	return cheapestFare
}
