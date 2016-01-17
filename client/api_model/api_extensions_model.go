package api_model

import (
	"encoding/json"
	"time"
)

type LocalDateTime struct {
	time.Time
}

var _ json.Marshaler = (*LocalDateTime)(nil)

func (l *LocalDateTime) UnmarshalJSON(value []byte) error {
	// Remove quotes
	if value[0] == '"' && value[len(value)-1] == '"' {
		value = value[1 : len(value)-1]
	}

	parsed, err := time.Parse("2006-01-02T15:04", string(value))
	if err != nil {
		return err
	}
	l.Time = parsed
	return nil
}

type CurrencyPrice struct {
	TotalFareCents uint32 `json:"totalFareCents"`
}

type FareProduct struct {
	FareType       string         `json:"fareType"`
	CurrencyPrice  *CurrencyPrice `json:"currencyPrice"`
	SeatsAvailable string         `json:"seatsAvailable"`
}

type Segment struct {
	OriginationAirportCode string        `json:"originationAirportCode"`
	DestinationAirportCode string        `json:"destinationAirportCode"`
	DepartureDateTime      LocalDateTime `json:"departureDateTime"`
	ArrivalDateTime        LocalDateTime `json:"arrivalDateTime"`
}

type AirProducts struct {
	FareProducts []*FareProduct `json:"fareProducts"`
	Segments     []*Segment     `json:"segments"`
}

type Trip struct {
	AirProducts []*AirProducts `json:"airProducts"`
}

type ListFlightsResponse struct {
	Trips []*Trip `json:"trips"`
}
