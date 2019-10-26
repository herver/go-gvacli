package main

import (
	"sort"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/olekukonko/tablewriter"
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

type GVATime struct {
	time time.Time
}

func (me *GVATime) String() string {
	if me.time.IsZero() {
		return ""
	}
	return fmt.Sprintf(me.time.Format(DisplayTimeFormat))
}

func (me *GVATime) UnmarshalJSON(input []byte) error {
    strInput := string(input)
    strInput = strings.Trim(strInput, `"`)
    if strInput == "null" {
	return nil
    }
    newTime, err := time.Parse(JSONTimeFormat, strInput)
    if err != nil {
        return err
    }

    me.time = newTime
    return nil
}

// FlightStatus is a dummy struct to allow Stringer() redef
type FlightStatus struct {
	status string
}

// "Took off", "Cancelled, "Departed", "Boarding", "Go to gate"
func (me *FlightStatus) String() string {
	switch status := me.status; status {
	case "Boarding":
		c := color.New(color.FgGreen).Add(color.BlinkSlow)
		return c.Sprintf(status)
	case "Go to gate":
		c := color.New(color.FgGreen).Add(color.BlinkSlow)
		return c.Sprintf(status)
	case "Arrived":
		c := color.New(color.FgGreen).Add(color.BlinkSlow)
		return c.Sprintf(status)
	case "Departed":
		c := color.New(color.FgYellow).Add(color.BlinkSlow)
		return c.Sprintf(status)
	case "Delayed":
		c := color.New(color.FgYellow).Add(color.BlinkSlow)
		return c.Sprintf(status)
	case "Cancelled":
		c := color.New(color.BgRed).Add(color.BlinkSlow)
		return c.Sprintf(status)
	default:
		c := color.New(color.FgWhite)
		return c.Sprintf(status)
	}
}

// UnmarshalJSON is a custom parser for flight status
func (me *FlightStatus) UnmarshalJSON(b []byte) (err error) {
	s := strings.Trim(string(b), "\"")
	if s == "null" {
		me.status = ""
		return
	}
	me.status = s
	return
}

// Flight contains relevant data extracted during JSON unmarshalling
type Flight struct {
	FlightIdentity			string	`json:"flight_identity"`
	DisplayedFlightIdentity		string	`json:"displayed_flight_identity"`
	Airport				string	`json:"airport"`
	ViaAirport			string	`json:"via_airport"`
	Carousel			string	`json:"carousel"`
	Terminal			string	`json:"terminal"`
	Aircraft			string	`json:"aircraft"`
	FlightStatus			FlightStatus	`json:"flight_status"`
	DisplayedMasterFlightCodes	string	`json:"displayed_master_flight_codes"`
	Gate				string	`json:"gate"`
	CheckinDesks			string	`json:"checkin_desks"`
	DepartureBool			int	`json:"departure_bool"`
	Company				string	`json:"company"`
	AirportCode			string	`json:"airport_code"`
	OriginCountry			string	`json:"origin_country"`
	AirportCodeDestination		string	`json:"airport_code_destination"`
	FlightType			string	`json:"flight_type"`
	FlightDurationMinutes		int	`json:"flight_duration_minuts"`
	ControleDouane			int	`json:"controle_douane"`
	GateWalkTime			int	`json:"gate_walk_time"`
	DelayMinutes			int	`json:"delay_minuts"`
	FlightId			int	`json:"flight_id"`
	AircraftId			int	`json:"aircraft_id"`
	NextPublicAdvice		GVATime	`json:"next_public_advice"`
	ScheduledDeparture		GVATime	`json:"scheduled_departure"`
	ScheduledArrival		GVATime	`json:"scheduled_arrival"`
	Departure			GVATime	`json:"departure"`
	Arrival				GVATime	`json:"arrival"`
	PublicArrival			GVATime	`json:"public_arrival"`
	PublicDeparture			GVATime	`json:"public_departure"`
	Airborn				GVATime	`json:"airborn"`
	FirstPriorityBaggage		GVATime	`json:"first_priority_baggage"`
	DepartureFromPreviousAirport	GVATime	`json:"departure_from_previous_airport"`
	EstimatedLanding		GVATime	`json:"estimated_landing"`
	EstimatedBoarding		GVATime	`json:"estimated_boarding"`
	ScheduledFlight			GVATime	`json:"scheduled_flight"`
	AircraftRegistration		string	`json:"aircraft_registration"`
	State				string	`json:"state"`
	LastUpdate			GVATime	`json:"last_update"`
	AirportCity			string	`json:"airport_city"`
}

// FlightInfos is just a container for Arrivals and Departures
type FlightInfos struct {
	Flights struct {
		Arrivals   []Flight  `json:"arrivals"`
		Departures []Flight  `json:"departures"`
	} `json:"flights"`
}

// GetData fetches flight data from the remote API
func (me *FlightInfos) GetData() error {
	cli := http.Client{
		Timeout: time.Duration(APITimeout) * time.Second,
	}

	req, err := http.NewRequest(http.MethodPost, APIUrl, nil)
	if err != nil {
		fmt.Print(err)
		return err
	}

	req.Header.Set("X-Requested-With", "XMLHttpRequest")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-type", "application/json")
	res, err := cli.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return err
	}

	jsonErr := json.Unmarshal(body, &me)
	if jsonErr != nil {
		return jsonErr
	}

	me.sortByScheduledDate()

	return nil
}

func (me *FlightInfos) sortByScheduledDate()  {
	sort.Slice((*me).Flights.Arrivals, func(i, j int) bool {
		return (*me).Flights.Arrivals[i].ScheduledArrival.time.Before((*me).Flights.Arrivals[j].ScheduledArrival.time)
	})
	sort.Slice((*me).Flights.Departures, func(i, j int) bool {
		return (*me).Flights.Departures[i].ScheduledDeparture.time.Before((*me).Flights.Departures[j].ScheduledDeparture.time)
	})
}

func (me *FlightInfos) PrepareDeparturesTable(f []Flight) [][]string {
	dataDep := [][]string{}
	for _, dep := range f {

		// Only show flights assigned to a gate
		if !ShowAllFlights && len(dep.Gate) < 2 {
			continue
		}

		// Show only today's flights
		if !ShowAllFlights && dep.ScheduledDeparture.time.Day() != time.Now().Day() {
			continue
		}

		var flightIds = dep.FlightIdentity
		if len(dep.DisplayedMasterFlightCodes) > 1 && ShowCodeShare {
			flightIds = fmt.Sprintf("%s (%s)", dep.FlightIdentity, dep.DisplayedMasterFlightCodes)
		}

		dataDep = append(dataDep, []string{
			dep.ScheduledDeparture.String(),
			dep.PublicDeparture.String(),
			dep.Airport,
			flightIds,
			dep.Company,
			dep.Gate,
			dep.Aircraft,
			dep.AircraftRegistration,
			dep.FlightStatus.String(),
		})
	}

	return dataDep
}

func (me *FlightInfos) PrepareArrivalsTable(f []Flight) [][]string {
	dataArr := [][]string{}
	for _, arr := range f {

		// Hide not-expected flights or those without a status
		if !ShowAllFlights && arr.PublicArrival.time.IsZero() && len(arr.FlightStatus.String()) < 10 {
			continue
		}

		// Show only today's flights
		if !ShowAllFlights && arr.ScheduledArrival.time.Day() != time.Now().Day() {
			continue
		}

		var flightIds = arr.FlightIdentity
		if len(arr.DisplayedMasterFlightCodes) > 0 && ShowCodeShare {
			flightIds = fmt.Sprintf("%s (%s)", arr.FlightIdentity, arr.DisplayedMasterFlightCodes)
		}

		dataArr = append(dataArr, []string{
			arr.ScheduledArrival.String(),
			arr.PublicArrival.String(),
			arr.DepartureFromPreviousAirport.String(),
			arr.Airport,
			flightIds,
			arr.Company,
			arr.Carousel,
			arr.Aircraft,
			arr.AircraftRegistration,
			arr.FlightStatus.String(),
		})
	}

	return dataArr
}

// PrintTable does the heavy lifting to print a nice table
func (me *FlightInfos) PrintTable(title string, headers []string, data [][]string) {
	tab := tablewriter.NewWriter(os.Stdout)
	tab.SetAutoWrapText(false)
	tab.SetHeader(headers)
	tab.SetCaption(true, title)
	tab.AppendBulk(data)
	tab.Render()
}

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
