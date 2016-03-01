package model

import (
	"strings"
	"time"
)

type Fare struct {
	Cents          uint32
	SeatsAvailable uint32
}

// Compares this fare to another.
// 0 indicates same fare,
// negative result indicates this fare is lower than o,
// positive result indicates this fare is greater than o.
// If both are nil, they are equal and 0 is returned.
// If one of the fares is nil, it is considered the greater fare.
func (f *Fare) Compare(o *Fare) int {
	if f == nil {
		if o == nil {
			return 0
		} else {
			return 1
		}
	} else {
		if f.Cents == o.Cents {
			return 0
		} else if f.Cents > o.Cents {
			return 1
		} else {
			return 0
		}
	}
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
