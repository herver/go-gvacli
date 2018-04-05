package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/olekukonko/tablewriter"
	flag "github.com/spf13/pflag"
)

// Some useful timestamps and values
var (
	ReqTimeFormat     = "2006-01-02 15:04:05"
	JSONTimeFormat    = "02-01-2006 15:04:05"
	DisplayTimeFormat = "02-01-2006 15:04"
	APIUrl            string
	APITimeout        int
	AllFlights        bool
)

// GVATime is just used to convert the "custom" date
type GVATime struct {
	time.Time
}

// UnmarshalJSON is a Custom parser for date format
func (me *GVATime) UnmarshalJSON(b []byte) (err error) {
	s := strings.Trim(string(b), "\"")
	if s == "null" {
		me.Time = time.Time{}
		return
	}
	me.Time, err = time.Parse(JSONTimeFormat, s)
	return
}

func (me *GVATime) String() string {
	if me.IsZero() {
		return ""
	}
	return fmt.Sprintf(me.Format(DisplayTimeFormat))
}

// FlightStatus is a dummy struct to allow Stringer() redef
type FlightStatus struct {
	status string
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

func (me *FlightStatus) String() string {
	switch status := me.status; status {
	case "Boarding":
		c := color.New(color.FgGreen).Add(color.BlinkSlow)
		return c.Sprintf(status)
	case "Delayed":
		c := color.New(color.FgYellow).Add(color.BlinkSlow)
		return c.Sprintf(status)
	case "Cancelled":
		c := color.New(color.BgRed).Add(color.BlinkSlow)
		return c.Sprintf(status)
	default:
		return color.GreenString(status)
	}
}

// Arrival contains relevant data extracted during JSON unmarshalling
type Arrival struct {
	SAircraftRegistration         string       `json:"sAircraftRegistration"`
	SFlightIdentity               string       `json:"sFlightIdentity"`
	SDelay                        string       `json:"sDelay"`
	SAirport                      string       `json:"sAirport"`
	STerminal                     string       `json:"sTerminal"`
	SAircraft                     string       `json:"sAircraft"`
	SFlightStatus                 FlightStatus `json:"sFlightStatus"`
	SCompany                      string       `json:"sCompany"`
	STrueFlightStatus             string       `json:"sTrueFlightStatus"`
	SFieldLabel                   string       `json:"sFieldLabel"`
	SAirportCode                  string       `json:"sAirportCode"`
	SFlightType                   string       `json:"sFlightType"`
	SFlightDuration               string       `json:"sFlightDuration"`
	IFlightDurationMinuts         string       `json:"iFlightDurationMinuts"`
	SOriginCountry                string       `json:"sOriginCountry"`
	SDestCountry                  string       `json:"sDestCountry"`
	SDestCountryVia               string       `json:"sDestCountryVia"`
	BControleDouane               string       `json:"bControleDouane"`
	IDelayMinuts                  string       `json:"iDelayMinuts"`
	State                         string       `json:"state"`
	SCarousel                     string       `json:"sCarousel"`
	DScheduledArrival             GVATime      `json:"dScheduledArrival"`
	DPublicArrival                GVATime      `json:"dPublicArrival"`
	DLastUpdate                   GVATime      `json:"dLastUpdate"`
	DAirbornFromPreviousAirport   GVATime      `json:"dAirbornFromPreviousAirport"`
	DDepartureFromPreviousAirport GVATime      `json:"dDepartureFromPreviousAirport"`
	DExpectedBaggageDelivery      GVATime      `json:"dExpectedBaggageDelivery"`
	DScheduledFlight              GVATime      `json:"dScheduledFlight"`
}

// Departure contains relevant data extracted during JSON unmarshalling
type Departure struct {
	SAircraftRegistration string       `json:"sAircraftRegistration"`
	SFlightIdentity       string       `json:"sFlightIdentity"`
	SDelay                string       `json:"sDelay"`
	SAirport              string       `json:"sAirport"`
	STerminal             string       `json:"sTerminal"`
	SAircraft             string       `json:"sAircraft"`
	SFlightStatus         FlightStatus `json:"sFlightStatus"`
	SGate                 string       `json:"sGate"`
	SGateStatus           string       `json:"sGateStatus"`
	SCheckinDesks         string       `json:"sCheckinDesks"`
	SCompany              string       `json:"sCompany"`
	SFieldLabel           string       `json:"sFieldLabel"`
	SFlightType           string       `json:"sFlightType"`
	BControleDouane       string       `json:"bControleDouane"`
	IDelayMinuts          string       `json:"iDelayMinuts"`
	DAirborn              GVATime      `json:"dAirborn"`
	DEstimatedBoarding    GVATime      `json:"dEstimatedBoarding"`
	DScheduledDeparture   GVATime      `json:"dScheduledDeparture"`
	DPublicDeparture      GVATime      `json:"dPublicDeparture"`
}

// FlightInfos is just a container for Arrivals and Departures
type FlightInfos struct {
	Arrivals   []Arrival   `json:"arrivals"`
	Departures []Departure `json:"departures"`
	LastUpdate string      `json:"lastUpdate"`
}

// GetData fetches flight data newer than lastSync
// from the remote API
func (me *FlightInfos) GetData(lastSync time.Time) error {
	cli := http.Client{
		Timeout: time.Duration(APITimeout) * time.Second,
	}

	req, err := http.NewRequest(http.MethodGet, APIUrl, nil)
	if err != nil {
		fmt.Print(err)
		return err
	}
	q := req.URL.Query()
	q.Add("lastSync", lastSync.Format(ReqTimeFormat))
	req.URL.RawQuery = q.Encode()
	//fmt.Println(req.URL.String())

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

func (me *FlightInfos) sortByScheduledDate() {
	sort.Slice(me.Departures, func(i, j int) bool {
		return me.Departures[i].DScheduledDeparture.Before(me.Departures[j].DScheduledDeparture.Time)
	})
	sort.Slice(me.Arrivals, func(i, j int) bool {
		return me.Arrivals[i].DScheduledArrival.Before(me.Arrivals[j].DScheduledArrival.Time)
	})
}

// PrintTable does the heavy lifting to print a nice table
func (me *FlightInfos) PrintTable(headers []string, data [][]string) {
	tab := tablewriter.NewWriter(os.Stdout)
	tab.SetAutoWrapText(false)
	tab.SetHeader(headers)
	tab.AppendBulk(data)
	tab.Render()
}

func init() {
	flag.StringVar(&APIUrl, "api-url", "http://gva.atipik.ch/api2/flights", "API URL of remote webservice")
	flag.IntVar(&APITimeout, "api-timeout", 10, "API reply timeout (in seconds)")
	flag.BoolVar(&AllFlights, "all-flights", false, "Show all flights")

	flag.Parse()
}
func main() {

	var lastSync time.Time
	if !AllFlights {
		lastSync = time.Now().Add(time.Minute * -15)
	}

	f := FlightInfos{}
	f.GetData(lastSync)

	fmt.Println("Departures:")
	dataDep := [][]string{}
	for _, dep := range f.Departures {
		var airport = dep.SAirport
		if dep.BControleDouane == "1" {
			airport = fmt.Sprintf("%s %s", dep.SAirport, "[P]")
		}
		dataDep = append(dataDep, []string{
			dep.DScheduledDeparture.String(),
			dep.DPublicDeparture.String(),
			airport,
			dep.SFlightIdentity,
			dep.SCompany,
			dep.SGate,
			dep.SAircraft,
			dep.SAircraftRegistration,
			dep.SFlightStatus.String(),
		})
	}
	f.PrintTable(
		[]string{"Scheduled", "Expected", "Dest", "Flight", "Airline", "Gate", "Aircraft", "Reg", "Status"},
		dataDep,
	)

	fmt.Println("Arrivals:")
	dataArr := [][]string{}
	for _, arr := range f.Arrivals {
		dataArr = append(dataArr, []string{
			arr.DScheduledArrival.String(),
			arr.DPublicArrival.String(),
			arr.SAirport,
			arr.SFlightIdentity,
			arr.SCompany,
			arr.SCarousel,
			arr.SAircraft,
			arr.SAircraftRegistration,
			arr.SFlightStatus.String(),
		})
	}
	f.PrintTable(
		[]string{"Scheduled", "Expected", "Origin", "Flight", "Airline", "Belt", "Aircraft", "Reg", "Status"},
		dataArr,
	)
}
