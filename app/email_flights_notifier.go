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

type EmailFlightsNotifier struct {
	SmtpAddress string
	Auth        smtp.Auth
	From        string
	To          string
}

var (
	htmlTemplate = template.Must(template.New("EmailFlightsNotifier").Parse(htmlTemplateDef))
	mime         = "MIME-version: 1.0;\nContent-Type: text/html; charset=\"UTF-8\";\n\n"
)

func (e *EmailFlightsNotifier) Notify(flights []*model.Flight, updates map[model.FlightId]UpdateResult) error {
	flightAdded := false
	filteredFlights := make([]*model.Flight, 0, len(flights))
	for _, flight := range flights {
		// Remove any flights that are no longer available
		if result, ok := updates[*flight.Id()]; !ok || result != Removed {
			filteredFlights = append(filteredFlights, flight)
		}

		if result, ok := updates[*flight.Id()]; ok && result == Added {
			flightAdded = true
		}
	}

	// Only send notification if at least one flight's cached value was updated
	if !flightAdded {
		return nil
	}

	flightsByDate := groupByDate(filteredFlights)

	// Create body from template
	var bodyBuffer bytes.Buffer
	err := htmlTemplate.Execute(&bodyBuffer, newBody(flightsByDate))
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

// Helper to sort Flights by time
type FlightsByDepartureTime []*model.Flight

func (b FlightsByDepartureTime) Len() int      { return len(b) }
func (b FlightsByDepartureTime) Swap(i, j int) { b[i], b[j] = b[j], b[i] }
func (b FlightsByDepartureTime) Less(i, j int) bool {
	return b[i].DepartureLocalTime.Before(b[j].DepartureLocalTime)
}

// Helper to sort Times
type ByTime []time.Time

func (b ByTime) Len() int           { return len(b) }
func (b ByTime) Swap(i, j int)      { b[i], b[j] = b[j], b[i] }
func (b ByTime) Less(i, j int) bool { return b[i].Before(b[j]) }

func sortedDates(groups map[time.Time][]*model.Flight) []time.Time {
	dates := make([]time.Time, 0, len(groups))
	for date, _ := range groups {
		dates = append(dates, date)
	}
	sort.Sort(ByTime(dates))
	return dates
}

// Structs that map to HTML template
type Body struct {
	DateGroups []*DateGroup
}

type DateGroup struct {
	Date  string
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

func newBody(flightsByDate map[time.Time][]*model.Flight) *Body {
	dateGroups := make([]*DateGroup, 0, len(flightsByDate))
	dates := sortedDates(flightsByDate)
	for _, date := range dates {
		trips := make([]*Trip, 0, len(flightsByDate[date]))
		for _, flight := range flightsByDate[date] {
			fareCents := flight.CheapestAvailableFare().Cents
			trips = append(trips, &Trip{
				OriginAirport:      flight.OriginAirport,
				DestinationAirport: flight.DestinationAirport,
				DepartureTime:      flight.DepartureLocalTime.Format("3:04 PM"),
				ArrivalTime:        flight.ArrivalLocalTime.Format("3:04 PM"),
				Stops:              len(flight.Stops),
				Cost:               fmt.Sprintf("$ %d.%02d", fareCents/100, fareCents%100),
			})
		}
		dateGroups = append(dateGroups, &DateGroup{
			Date:  date.Format("Mon Jan 2 2006"),
			Trips: trips,
		})
	}
	return &Body{DateGroups: dateGroups}
}

var htmlTemplateDef = `
<html>
  <head></head>
  <body style="font-family: Arial, Helvetica, sans-serif">
    {{range .DateGroups}}
      <div>
        <h3 style="margin-bottom: 3px;">{{.Date}}</h3>
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
