package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"log"
	"net/smtp"
	"os"
	"os/signal"
	"time"

	"github.com/alecholmes/southwest_flight_watcher/app"
	"github.com/alecholmes/southwest_flight_watcher/client"
	"github.com/alecholmes/southwest_flight_watcher/model"
)

const (
	smtpPort            = 587
	searchesFlagStr     = "searchesFile"
	fromFlagStr         = "from"
	smtpFlagStr         = "smtp"
	smtpPasswordFileStr = "smtpPasswordFile"
)

func main() {
	searchesFlag := flag.String(searchesFlagStr, "", "filename of flight searches JSON")
	fromFlag := flag.String(fromFlagStr, "", "email address from which updates are sent")
	smtpFlag := flag.String(smtpFlagStr, "", fmt.Sprintf("SMTP host for mail delivery. Port %i is used.", smtpPort))
	smtpPasswordFileFlag := flag.String(smtpPasswordFileStr, "", "File containing SMTP password. Must have 0700 permissions.")

	// Check all the flags
	flag.Parse()

	if *searchesFlag == "" {
		fmt.Fprintf(os.Stderr, "%v flag must be set\n", searchesFlagStr)
		return
	}
	if *fromFlag == "" {
		fmt.Fprintf(os.Stderr, "%v flag must be set\n", fromFlagStr)
		return
	}
	if *smtpFlag == "" {
		fmt.Fprintf(os.Stderr, "%v flag must be set\n", smtpFlagStr)
		return
	}
	if *smtpPasswordFileFlag == "" {
		fmt.Fprintf(os.Stderr, "%v flag must be set\n", smtpPasswordFileStr)
		return
	}

	// Load SMTP password
	password, err := loadPassword(*smtpPasswordFileFlag)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to load password from %v: %v\n", *smtpPasswordFileFlag, err)
		return
	}

	searches, err := app.FlightSearchesFromFile(*searchesFlag)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Invalid searches file contents: %v\n", err)
		return
	}

	// Set up channel for OS signals, which are used to shutdown the app
	sigChannel := make(chan os.Signal, 1)
	signal.Notify(sigChannel, os.Interrupt, os.Kill)

	// Create an email notifier
	emailNotifier := app.EmailFlightsNotifier{
		SmtpAddress: fmt.Sprintf("%v:%v", *smtpFlag, smtpPort),
		Auth:        smtp.PlainAuth("", *fromFlag, password, *smtpFlag),
		From:        *fromFlag,
		To:          *fromFlag,
	}

	// Compose a notifier from stdoutFlightsNotifier and emailNotifier
	notifier := func(flights []*model.Flight, updates map[model.FlightId]app.UpdateResult) error {
		if err := stdoutFlightsNotifier(flights, updates); err != nil {
			return err
		}
		return emailNotifier.Notify(flights, updates)
	}

	// Create container for state with the function to update it
	state := app.NewCheapestFlightState(searches, app.NewFlightFetcher(client.NewClient(nil)), notifier)

	logger := log.New(os.Stdout, "", log.LstdFlags)
	updater := func() {
		logger.Print("Updating flights")
		if err := state.Update(); err != nil {
			logger.Printf("Error updating flights: %v\n", err)
		}
	}

	// Run immediately, and then every hour
	// TODO(alec): Determine if sharing state variable across goroutines is bad in this case
	ticker := time.NewTicker(1 * time.Hour)
	go func() {
		for _ = range ticker.C {
			updater()
		}
	}()
	updater()

	// Keep running until signal to shutdown
	<-sigChannel
	ticker.Stop()
}

func loadPassword(filename string) (string, error) {
	file, err := os.Open(filename)
	if err != nil {
		return "", err
	}
	defer file.Close()

	fileInfo, err := file.Stat()
	if err != nil {
		return "", err
	}

	if fileInfo.Mode() != 0600 {
		return "", fmt.Errorf("Password file mode must be 0600, not %v", fileInfo.Mode())
	}

	reader := bufio.NewReader(file)
	password, err := reader.ReadString('\n')
	if err == nil || err == io.EOF {
		return password, nil
	}
	return "", err
}

func stdoutFlightsNotifier(flights []*model.Flight, updates map[model.FlightId]app.UpdateResult) error {
	for _, flight := range flights {
		updateResult, ok := updates[*flight.Id()]
		if !ok {
			updateResult = app.Unchanged
		}

		fmt.Printf("(%v) %v %v -> %v %v: $%v\n",
			updateResult,
			flight.OriginAirport, flight.DepartureLocalTime,
			flight.DestinationAirport, flight.ArrivalLocalTime,
			flight.CheapestAvailableFare().Cents)
	}

	return nil
}
