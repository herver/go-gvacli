package main

import (
	"log"

	flag "github.com/spf13/pflag"
)

// Some useful timestamps and values
var (
	DisplayTimeFormat = "02/01 15:04"
	JSONTimeFormat    = "2006-01-02 15:04:05"
	APIUrl            string
	APITimeout        int
	Departures        bool
	Arrivals          bool
	ShowCodeShare     bool
	ShowAllFlights    bool
)

func init() {
	flag.StringVar(&APIUrl, "api-url", "https://app4airport.com/api/flights", "API URL of remote webservice")
	flag.IntVar(&APITimeout, "api-timeout", 10, "API reply timeout (in seconds)")
	flag.BoolVar(&ShowCodeShare, "code-shares", false, "Show code shares")
	flag.BoolVar(&Departures, "departures", false, "Show departures")
	flag.BoolVar(&Arrivals, "arrivals", false, "Show arrivals")
	flag.BoolVar(&ShowAllFlights, "all-flights", false, "Show all flights, despite of the status")

	flag.Parse()
}

func main() {

	// If we hide everything, show arrivals by default
	if !Departures && !Arrivals {
		Arrivals = true
	}

	info := FlightInfos{}
	if err := info.GetData(); err != nil {
		log.Fatalf("Unable to fetch data from remote API: %s", err)
	}
	if Departures {
		depTable := info.PrepareDeparturesTable(info.Flights.Departures)
		info.PrintTable(
			"Departures",
			[]string{"Scheduled", "Expected", "Dest", "Flight", "Airline", "Gate", "Aircraft", "Reg", "Status"},
			depTable,
		)
	}

	if Arrivals {
		arrTable := info.PrepareArrivalsTable(info.Flights.Arrivals)
		info.PrintTable(
			"Arrivals",
			[]string{"Scheduled", "Expected", "Departed", "Source", "Flight", "Airline", "Belt", "Aircraft", "Reg", "Status"},
			arrTable,
		)
	}
}
