package client

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/alecholmes/southwest_flight_watcher/client/api_model"
	"github.com/alecholmes/southwest_flight_watcher/model"
)

const (
	apiKeyHeader = "X-API-Key"
	apiKeyValue  = "l7xx12ebcbc825eb480faa276e7f192d98d1"

	userAgentHeader = "User-Agent"
	userAgentValue  = "Southwest/3.0.26 (iPhone; iOS 9.1; Scale/2.00)"
)

var (
	apiExtensionsBaseUrl = parseUrlOrPanic("https://mobile.southwest.com")
)

type Client struct {
	httpClient *http.Client
}

// NewClient returns a new Southwest API client.
// If a nil httpClient is provided, http.DefaultClient will be used.
func NewClient(httpClient *http.Client) *Client {
	if httpClient == nil {
		httpClient = http.DefaultClient
	}

	return &Client{
		httpClient: httpClient,
	}
}

func (c *Client) ListFlights(departureDate time.Time, originAirport string, destinationAirport string) ([]*model.Flight, error) {
	originAirport, err := normalizeAirportCode(originAirport)
	if err != nil {
		return nil, err
	}
	destinationAirport, err = normalizeAirportCode(destinationAirport)
	if err != nil {
		return nil, err
	}

	url, err := url.Parse("/api/extensions/v1/mobile/flights/products")
	if err != nil {
		return nil, err
	}

	url = apiExtensionsBaseUrl.ResolveReference(url)

	queryValues := url.Query()
	queryValues.Set("currency-type", "Dollars")
	queryValues.Set("number-adult-passengers", "1")
	queryValues.Set("number-senior-passengers", "0")
	queryValues.Set("promo-code", "")
	queryValues.Set("origination-airport", originAirport)
	queryValues.Set("destination-airport", destinationAirport)
	queryValues.Set("departure-date", departureDate.Format("2006-01-02"))
	url.RawQuery = queryValues.Encode()

	httpReq, err := http.NewRequest("GET", url.String(), nil)
	if err != nil {
		return nil, err
	}

	httpReq.Header.Add(apiKeyHeader, apiKeyValue)
	httpReq.Header.Add(userAgentHeader, userAgentValue)

	httpResp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode != 200 {

		body, _ := ioutil.ReadAll(httpResp.Body)
		return nil, fmt.Errorf("Unexpected response code listing flights. req=%s, statusCode=%d, body=%s", httpReq, httpResp.StatusCode, string(body))
	} else if httpResp.Body == nil {
		return nil, fmt.Errorf("No body listing flights")
	}

	body, err := ioutil.ReadAll(httpResp.Body)
	if err != nil {
		return nil, err
	}

	swResponse := api_model.ListFlightsResponse{}
	if err := json.Unmarshal(body, &swResponse); err != nil {
		return nil, err
	}

	return swListFlightsResponseToFlights(&swResponse)
}

type FlightFilter interface {
	Matches(*model.Flight) bool
}

type MaxStopsFilter struct {
	Stops int
}

func (m *MaxStopsFilter) Matches(flight *model.Flight) bool {
	return len(flight.Stops) <= m.Stops
}

type MaxAvailableFareFilter struct {
	Cents uint32
}

func (m *MaxAvailableFareFilter) Matches(flight *model.Flight) bool {
	cheapestAvailableFare := flight.CheapestAvailableFare()
	return cheapestAvailableFare != nil && flight.CheapestAvailableFare().Cents <= m.Cents
}

// Filter for flights leaving after a time, inclusive.
type departAfterFilter struct {
	time.Time
}

func (d *departAfterFilter) Matches(flight *model.Flight) bool {
	return !flight.DepartureLocalTime.Before(d.Time)
}

// Filter for flights arriving before a time, inclusive.
type arriveBeforeFilter struct {
	time.Time
}

func (a *arriveBeforeFilter) Matches(flight *model.Flight) bool {
	return !flight.ArrivalLocalTime.After(a.Time)
}

func (c *Client) SearchFlights(
	originAirports []string,
	destinationAirports []string,
	minLocalDepartureTime time.Time,
	maxLocalArrivalTime time.Time,
	filters []FlightFilter) ([]*model.Flight, error) {

	departureTime := minLocalDepartureTime.In(time.UTC)
	arrivalTime := maxLocalArrivalTime.In(time.UTC)

	flights := make([]*model.Flight, 0)
	for _, originAirport := range originAirports {
		for _, destinationAirport := range destinationAirports {
			someFlights, err := c.ListFlights(departureTime, originAirport, destinationAirport)
			if err != nil {
				return nil, err
			}
			flights = append(flights, someFlights...)
		}
	}

	filters = append(filters,
		&departAfterFilter{departureTime},
		&arriveBeforeFilter{arrivalTime})

	filteredFlights := make([]*model.Flight, 0)
	for _, flight := range flights {
		matches := true
		for _, filter := range filters {
			if !filter.Matches(flight) {
				matches = false
			}
		}
		if matches {
			filteredFlights = append(filteredFlights, flight)
		}
	}
	return filteredFlights, nil
}

func swListFlightsResponseToFlights(swResponse *api_model.ListFlightsResponse) ([]*model.Flight, error) {
	if len(swResponse.Trips) > 1 {
		return nil, fmt.Errorf("Unexpected number of trips: %d", len(swResponse.Trips))
	}

	flights := make([]*model.Flight, 0)
	for _, trip := range swResponse.Trips {
		// An airProduct will result in a flight with one or more fares
		for _, airProduct := range trip.AirProducts {
			fares := make([]*model.Fare, len(airProduct.FareProducts))
			for i, fareProduct := range airProduct.FareProducts {
				seatsAvailable, err := strconv.ParseUint(fareProduct.SeatsAvailable, 10, 32)
				if err != nil {
					return nil, fmt.Errorf("Unexpected seats available value: %s", fareProduct.SeatsAvailable)
				}

				fares[i] = &model.Fare{
					Cents:          fareProduct.CurrencyPrice.TotalFareCents,
					SeatsAvailable: uint32(seatsAvailable),
				}
			}

			stops := make([]string, len(airProduct.Segments)-1)
			for i := 0; i < len(stops); i++ {
				stops[i] = airProduct.Segments[i].DestinationAirportCode
			}

			firstSegment := airProduct.Segments[0]
			lastSegment := airProduct.Segments[len(airProduct.Segments)-1]
			flight := model.Flight{
				OriginAirport:      firstSegment.OriginationAirportCode,
				DestinationAirport: lastSegment.DestinationAirportCode,
				DepartureLocalTime: firstSegment.DepartureDateTime.Time,
				ArrivalLocalTime:   lastSegment.ArrivalDateTime.Time,
				Stops:              stops,
				Fares:              fares,
			}
			flights = append(flights, &flight)
		}
	}
	return flights, nil
}

func parseUrlOrPanic(urlStr string) *url.URL {
	url, err := url.Parse(urlStr)
	if err != nil {
		panic(fmt.Sprintf("Invalid URL %v", url))
	}
	return url
}
