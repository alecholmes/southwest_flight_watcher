package app

import (
	"bytes"
	"fmt"
	"html/template"
	"net/smtp"
	"sort"
	"time"

	"github.com/alecholmes/southwest_flight_watcher/model"
)

var (
	htmlTemplate = template.Must(template.New("EmailFlightsNotifier").Parse(htmlTemplateDef))
	mime         = "MIME-version: 1.0;\nContent-Type: text/html; charset=\"UTF-8\";\n\n"
)

type EmailFlightsNotifier struct {
	SmtpAddress string
	Auth        smtp.Auth
	From        string
	To          string
}

var _ SearchUpdateNotifier = &EmailFlightsNotifier{}

func (e *EmailFlightsNotifier) Notify(searchStates FlightSearchStates) error {
	available, added := searchStates.OnlyAvailable()

	// Only send notification if there are flights and at least one flight was updated
	if !added || len(available) == 0 {
		return nil
	}

	// Create body from template
	var bodyBuffer bytes.Buffer
	err := htmlTemplate.Execute(&bodyBuffer, newBody(available))
	if err != nil {
		return err
	}

	// Send email
	msg := "From: " + e.From + "\n" +
		"To: " + e.To + "\n" +
		"Subject: Southwest Flight Price Update\n" +
		mime + "\n\n" +
		bodyBuffer.String() + "\n"
	return smtp.SendMail(e.SmtpAddress, e.Auth, e.From, []string{e.To}, []byte(msg))
}

func groupByDate(flights []*model.Flight) map[time.Time][]*model.Flight {
	grouped := make(map[time.Time][]*model.Flight)
	for _, flight := range flights {
		flightDate := date(flight.DepartureLocalTime)
		if flights, ok := grouped[flightDate]; ok {
			grouped[flightDate] = append(flights, flight)
		} else {
			grouped[flightDate] = []*model.Flight{flight}
		}
	}

	// Sort each group by departure time
	for _, group := range grouped {
		sort.Sort(FlightsByDepartureTime(group))
	}

	return grouped
}

func date(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC)
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

// Helper to sort CurrentState by flight DepartureLocalTime
type CurrentStateByDepartureLocalTime []CurrentState

func (b CurrentStateByDepartureLocalTime) Len() int      { return len(b) }
func (b CurrentStateByDepartureLocalTime) Swap(i, j int) { b[i], b[j] = b[j], b[i] }
func (b CurrentStateByDepartureLocalTime) Less(i, j int) bool {
	return b[i].Flight.DepartureLocalTime.Before(b[j].Flight.DepartureLocalTime)
}

func sortedCurrentStates(u map[model.FlightId]CurrentState) []CurrentState {
	updates := make([]CurrentState, 0, len(u))
	for _, update := range u {
		updates = append(updates, update)
	}
	sort.Sort(CurrentStateByDepartureLocalTime(updates))

	return updates
}

// Helper to sort Flights by time
type FlightsByDepartureTime []*model.Flight

func (b FlightsByDepartureTime) Len() int      { return len(b) }
func (b FlightsByDepartureTime) Swap(i, j int) { b[i], b[j] = b[j], b[i] }
func (b FlightsByDepartureTime) Less(i, j int) bool {
	return b[i].DepartureLocalTime.Before(b[j].DepartureLocalTime)
}

// Structs that map to HTML template
type Body struct {
	SearchGroups []*SearchGroup
}

type SearchGroup struct {
	Date  string
	Note  *string
	Trips []*Trip
}

type Trip struct {
	OriginAirport      string
	DestinationAirport string
	DepartureTime      string
	ArrivalTime        string
	Stops              int
	Cost               string
}

func newBody(searches FlightSearchStates) *Body {
	searchGroups := make([]*SearchGroup, 0, len(searches))

	for _, search := range sortedFlightSearches(searches) {
		updates := searches[search]
		trips := make([]*Trip, 0, len(updates))

		for _, update := range sortedCurrentStates(updates) {
			fareCents := update.Flight.CheapestAvailableFare().Cents
			trips = append(trips, &Trip{
				OriginAirport:      update.Flight.OriginAirport,
				DestinationAirport: update.Flight.DestinationAirport,
				DepartureTime:      update.Flight.DepartureLocalTime.Format("3:04 PM"),
				ArrivalTime:        update.Flight.ArrivalLocalTime.Format("3:04 PM"),
				Stops:              len(update.Flight.Stops),
				Cost:               fmt.Sprintf("$ %d.%02d", fareCents/100, fareCents%100),
			})
		}

		searchGroups = append(searchGroups, &SearchGroup{
			Date:  search.MinDepartureTime.Format("Mon Jan 2 2006"),
			Note:  search.Note,
			Trips: trips,
		})
	}
	return &Body{SearchGroups: searchGroups}
}

var htmlTemplateDef = `
<html>
  <head></head>
  <body style="font-family: Arial, Helvetica, sans-serif">
    {{range .SearchGroups}}
      <div>
        <h3 style="margin-bottom: 3px;">{{.Date}}</h3>
        {{.Note}}
        <table style="border-collapse: collapse">
          <thead>
            <th colspan="2" style="text-align: left; padding: 0px 15px 0px 0px">From</th>
            <th colspan="2" style="text-align: left; padding: 0px 15px 0px 0px">To</th>
            <th style="text-align: left; padding: 0px 15px 0px 0px">Stops</th>
            <th style="text-align: left; padding: 0px 15px 0px 0px">Price</th>
          </thead>
          <tbody>
            {{range .Trips}}
              <tr>
                <td style="padding: 0px 15px 0px 0px">{{.OriginAirport}}</td>
                <td style="padding: 0px 15px 0px 0px">{{.DepartureTime}}</td>
                <td style="padding: 0px 15px 0px 0px">{{.DestinationAirport}}</td>
                <td style="padding: 0px 15px 0px 0px">{{.ArrivalTime}}</td>
                <td style="padding: 0px 15px 0px 0px">{{.Stops}}</td>
                <td style="padding: 0px 15px 0px 0px">{{.Cost}}</td>
              </tr>
            {{end}}
          </tbody>
      </div>
    {{end}}
  </body>
</html>
`
