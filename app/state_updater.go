package app

import (
	"time"
)

// Container to keep and update flight state.
type SearchStateUpdater struct {
	states   FlightSearchStates
	searches []*FlightSearch
	fetcher  *FlightFetcher
	notifier SearchUpdateNotifier
}

func NewSearchStateUpdater(searches []*FlightSearch, fetcher *FlightFetcher, notifier SearchUpdateNotifier) *SearchStateUpdater {
	s := &SearchStateUpdater{
		states:   NewFlightSearchStates(),
		searches: searches,
		fetcher:  fetcher,
		notifier: notifier,
	}

	return s
}

// Fetch latest flight info, update state, and send updates to the notifier.
func (c *SearchStateUpdater) Update() error {
	today := date(time.Now())
	for _, search := range c.searches {
		if !today.After(date(search.MinDepartureTime)) {
			flights, err := c.fetcher.Fetch(search)
			if err != nil {
				return err
			}
			c.states.Update(search, flights)
		}
	}

	if err := c.notifier.Notify(c.states); err != nil {
		return err
	}

	return nil
}
