package app

import (
	"fmt"
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
			fmt.Printf("(%v) %v %v -> %v %v: $%v\n",
				update.Update,
				update.Flight.OriginAirport, update.Flight.DepartureLocalTime,
				update.Flight.DestinationAirport, update.Flight.ArrivalLocalTime,
				update.Flight.CheapestAvailableFare().Cents)
		}
	}
	return nil
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
