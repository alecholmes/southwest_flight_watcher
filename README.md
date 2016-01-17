# Southwest Flight Watcher

## Introduction

Southwest Price Watcher is an application that constantly searches for Southwest flights that meet certain criteria, emailing the search results when they change. It is useful for monitoring for drops in flight prices.

The application runs locally and requires SMTP email credentials for sending notifications.

## Installing and Running

### Install

To install, run:

```
go install github.com/alecholmes/southwest_flight_watcher
```

### Running

To run, arguments with search criteria and email credentials must be passed via command line args.

```
$GOPATH/bin/southwest_flight_watcher \
  -searchesFile example_searches.json \
  -from your_email@gmail.com \
  -smtp smtp.gmail.com \
  -smtpPasswordFile FILE_WITH_YOUR_EMAIL_PASSWORD
```

This will start the app, using the configuration in the searchesFile (detailed below) to determine which flights to check. It will immediately run a search, then run searches every hour. When the results of a search change, a report will be emailed to the address specified in the `-from` flag.

To stop the app, hit `Ctrl-C`.

While the example above uses a GMail account, any email account will work.

### Command Line Arguments

#### -searchesFile

This is the path to a JSON file containing one or more search criteria. An `example_searches.json` is included in this project.

The file should contain an array of searches.

```
[
  {
  	// One or more original airport codes.
    "origin_airports": ["SFO", "OAK"],
    
    // One ore more destination airport codes.
    "destination_airports": ["BUR", "LAX", "ONT"],
    
    // The earliest time, inclusive, a flight may depart.
    // The time zone is ignored; the search treats this as
    // the local time at the departure airport.
    // The search will only include flights that leave on this date,
    // regardless of the value of max_arrival_time.
    "min_departure_time": "2016-01-31T17:00:00Z",
    
    // The latest time, inclusive, a flight may arrive.
    // The time zone is ignored; the search treats this as
    // the local time at the arrival airport.
    "max_arrival_time": "2016-01-31T21:00:00Z",
    
    // Optional. The maximum price of the flight in USD cents.
    "max_fare_cents": 7200,
    
    // Optional. The maximum number of stops allowed. 0 is a direct flight.
    "max_number_stops": 0
  }
]
```

#### -smtpPasswordFile

A password is required to authenticate with the SMTP server sending email. This argument is the path to a file containing only the password.

The permissions of this file *must* be `0600` (`-rw-------`) in order for this file to be used.


## Known Issues

This code is poorly tested. There are no unit tests and few ad hoc tests have been run.

## License

This project uses the [GPLv3](http://www.gnu.org/licenses/gpl-3.0.html) license.