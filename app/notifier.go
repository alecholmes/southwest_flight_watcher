package app

import (
	"fmt"
	"time"
)

type SearchUpdateNotifier interface {
	Notify(searchStates FlightSearchStates) error
}

type StdoutNotifier struct{}

var _ SearchUpdateNotifier = &StdoutNotifier{}

func (s *StdoutNotifier) Notify(searchStates FlightSearchStates) error {
	for search, updates := range searchStates {
		fmt.Println(search)
		for _, update := range updates {
			fmt.Printf("  (%s) %v %v -> %v %v: $%v\n",
				s.flightStateChangeString(update.Update),
				update.Flight.OriginAirport, s.timeString(update.Flight.DepartureLocalTime),
				update.Flight.DestinationAirport, s.timeString(update.Flight.ArrivalLocalTime),
				float64(update.Flight.CheapestAvailableFare().Cents)/100)
		}
	}
	return nil
}

func (s *StdoutNotifier) flightStateChangeString(f FlightStateChange) string {
	switch f {
	case Unchanged:
		return "-"
	case FareIncrease:
		return "▲"
	case FareDecrease:
		return "▼"
	case Added:
		return "+"
	case Removed:
		return "x"
	default:
		return "?"
	}
}

func (s *StdoutNotifier) timeString(t time.Time) string {
	return t.Format("2006-01-02 15:04")
}

type SearchUpdateNotifierChain []SearchUpdateNotifier

func (s SearchUpdateNotifierChain) Notify(searchStates FlightSearchStates) error {
	for _, notifier := range s {
		if err := notifier.Notify(searchStates); err != nil {
			return err
		}
	}
	return nil
}
