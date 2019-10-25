package main

import (
	"bytes"
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
	DisplayTimeFormat = "02-01-2006 15:04"
	JSONTimeFormat    = "02.01.2006 15:04"
	APIUrl            string
	APITimeout        int
	Departures        bool
	Arrivals          bool
	ShowCodeShare     bool
)

// GVATime is just used to convert the "custom" date
type GVATime struct {
	time.Time
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
	ID                     string `json:"ID"`
	Type                   string `json:"Type"`
	DepartureScheduled     GVATime
	DepartureScheduledTime string `json:"DepartureScheduledTime"`
	DepartureScheduledDate string `json:"DepartureScheduledDate"`
	DepartureExpected      GVATime
	DepartureExpectedTime  string `json:"DepartureExpectedTime"`
	DepartureExpectedDate  string `json:"DepartureExpectedDate"`
	ArrivalScheduled       GVATime
	ArrivalScheduledTime   string `json:"ArrivalScheduledTime"`
	ArrivalScheduledDate   string `json:"ArrivalScheduledDate"`
	ArrivalExpected        GVATime
	ArrivalExpectedTime    string       `json:"ArrivalExpectedTime"`
	ArrivalExpectedDate    string       `json:"ArrivalExpectedDate"`
	Delay                  int          `json:"Delay"`
	Destination            string       `json:"Destination"`
	DestinationShort       string       `json:"DestinationShort"`
	Departure              string       `json:"Departure"`
	DepartureShort         string       `json:"DepartureShort"`
	Airline                string       `json:"Airline"`
	Aircraft               string       `json:"Aircraft"`
	Name                   string       `json:"Name"`
	Status                 FlightStatus `json:"Status"`
	StatusClass            string       `json:"StatusClass"`
	StatusDetails          string       `json:"StatusDetails"`
	MasterFlightID         string       `json:"MasterFlightId"`
	FlightIds              string       `json:"FlightIds"`
	GateRef                string       `json:"GateRef"`
	RegistrationRef        string       `json:"RegistrationRef"`
	ConveyorBeltRef        string       `json:"ConveyorBeltRef"`
	LastKenticoUpdate      string       `json:"LastKenticoUpdate"`
	PlanePicto             string       `json:"PlanePicto"`
	Via                    string       `json:"Via"`
	ViaShort               string       `json:"ViaShort"`
	IsLate                 bool         `json:"IsLate"`
	//DScheduledFlight       GVATime `json:"dScheduledFlight"`
}

// FlightInfos is just a container for Arrivals and Departures
type FlightInfos []Flight

// GetData fetches flight data newer than lastSync
// from the remote API
func (me *FlightInfos) GetData(dataType string) error {
	cli := http.Client{
		Timeout: time.Duration(APITimeout) * time.Second,
	}

	jsonQry := fmt.Sprintf(`{"datas":"{\"Type\":\"%s\", \"Culture\":\"en-GB\"}"}`, dataType)

	req, err := http.NewRequest(http.MethodPost, APIUrl, bytes.NewBuffer([]byte(jsonQry)))
	if err != nil {
		fmt.Print(err)
		return err
	}
	//fmt.Println(req.URL.String())

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

// Parses time and dates in GVATime and sorts slice entries accordingly
func (me *FlightInfos) sortByScheduledDate() {
	for i, f := range *me {
		(*me)[i].DepartureExpected.Time, _ = time.Parse(JSONTimeFormat, fmt.Sprintf("%s %s", f.DepartureExpectedDate, f.DepartureExpectedTime))
		(*me)[i].DepartureScheduled.Time, _ = time.Parse(JSONTimeFormat, fmt.Sprintf("%s %s", f.DepartureScheduledDate, f.DepartureScheduledTime))

		(*me)[i].ArrivalExpected.Time, _ = time.Parse(JSONTimeFormat, fmt.Sprintf("%s %s", f.ArrivalExpectedDate, f.ArrivalExpectedTime))
		(*me)[i].ArrivalScheduled.Time, _ = time.Parse(JSONTimeFormat, fmt.Sprintf("%s %s", f.ArrivalScheduledDate, f.ArrivalScheduledTime))
	}
}

func (me *FlightInfos) PrepareDeparturesTable(f []Flight) [][]string {
	dataDep := [][]string{}
	for _, dep := range f {

		// Only show flights assigned to a gate
		if len(dep.GateRef) < 2 {
			continue
		}

		var flightIds = dep.MasterFlightID
		if len(dep.FlightIds) > 1 && ShowCodeShare {
			flightIds = fmt.Sprintf("%s (%s)", dep.MasterFlightID, dep.FlightIds)
		}

		dataDep = append(dataDep, []string{
			dep.DepartureScheduled.String(),
			dep.DepartureExpected.String(),
			dep.Destination,
			flightIds,
			dep.Airline,
			dep.GateRef,
			dep.Aircraft,
			dep.Status.String(),
			dep.StatusDetails,
		})
	}

	return dataDep
}

func (me *FlightInfos) PrepareArrivalsTable(f []Flight) [][]string {
	dataArr := [][]string{}
	for _, arr := range f {

		// Hide not-expected flights or those without a status
		if arr.ArrivalExpected.IsZero() && len(arr.Status.String()) < 10 {
			continue
		}

		var flightIds = arr.MasterFlightID
		if len(arr.FlightIds) > 0 && ShowCodeShare {
			flightIds = fmt.Sprintf("%s (%s)", arr.MasterFlightID, arr.FlightIds)
		}

		dataArr = append(dataArr, []string{
			arr.ArrivalScheduled.String(),
			arr.ArrivalExpected.String(),
			arr.DepartureScheduled.String(),
			arr.Departure,
			flightIds,
			arr.Airline,
			arr.ConveyorBeltRef,
			arr.Aircraft,
			arr.Status.String(),
			arr.StatusDetails,
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
	flag.StringVar(&APIUrl, "api-url", "https://www.gva.ch/api/v1/data/flights", "API URL of remote webservice")
	flag.IntVar(&APITimeout, "api-timeout", 10, "API reply timeout (in seconds)")
	flag.BoolVar(&ShowCodeShare, "code-shares", false, "Show code shares")
	flag.BoolVar(&Departures, "departures", false, "Show departures")
	flag.BoolVar(&Arrivals, "arrivals", false, "Show arrivals")

	flag.Parse()
}

func main() {

	// If we hide everything, show arrivals by default
	if !Departures && !Arrivals {
		Arrivals = true
	}

	if Departures {
		depFlights := FlightInfos{}
		if err := depFlights.GetData("DEPARTURE"); err != nil {
			log.Fatalf("Unable to fetch data from remote API: %s", err)
		}
		depFlights.sortByScheduledDate()
		depTable := depFlights.PrepareDeparturesTable(depFlights)
		depFlights.PrintTable(
			"Departures",
			[]string{"Scheduled", "Expected", "Dest", "Flight", "Airline", "Gate", "Aircraft", "Status", "Status detail"},
			depTable,
		)
	}

	if Arrivals {
		arrFlights := FlightInfos{}
		if err := arrFlights.GetData("ARRIVAL"); err != nil {
			log.Fatalf("Unable to fetch data from remote API: %s", err)
		}
		arrFlights.sortByScheduledDate()
		arrTable := arrFlights.PrepareArrivalsTable(arrFlights)
		arrFlights.PrintTable(
			"Arrivals",
			[]string{"Scheduled", "Expected", "Departed", "Source", "Flight", "Airline", "Belt", "Aircraft", "Status", "Status detail"},
			arrTable,
		)
	}
}
