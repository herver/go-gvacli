package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"sort"
	"time"

	"github.com/olekukonko/tablewriter"
)

// FlightInfos is just a container for Arrivals and Departures
type FlightInfos struct {
	Flights struct {
		Arrivals   []Flight `json:"arrivals"`
		Departures []Flight `json:"departures"`
	} `json:"flights"`
	cacheFilePath  string
	cacheAvailable bool
}

func NewFlightInfos() *FlightInfos {
	c := FlightInfos{}
	userCacheDir, err := os.UserCacheDir()
	if err != nil {
		log.Printf("Unable to determine local user cache directory")
		c.cacheAvailable = false
	}
	cacheDir := fmt.Sprintf("%s/gvacli", userCacheDir)

	if _, err := os.Stat(cacheDir); os.IsNotExist(err) {
		err := os.Mkdir(cacheDir, 0755)
		if err != nil {
			log.Printf("Unable to create cache directory")
			c.cacheAvailable = false
		}
	}
	c.cacheAvailable = true
	c.cacheFilePath = fmt.Sprintf("%s/flightinfos.json", cacheDir)
	return &c
}

func (me *FlightInfos) cacheCanBeConsumed() bool {
	if NoCache || !me.cacheAvailable {
		return false
	}
	if cacheInfo, err := os.Stat(me.cacheFilePath); err == nil {
		cacheAgeSeconds := time.Now().Sub(cacheInfo.ModTime()).Seconds()
		if cacheInfo.Size() > 0 && cacheAgeSeconds < CacheTTLSeconds {
			return true
		}
	}
	return false
}

func (me *FlightInfos) cacheRead() ([]byte, error) {
	return ioutil.ReadFile(me.cacheFilePath)
}

func (me *FlightInfos) cacheWrite(data []byte) {
	err := ioutil.WriteFile(me.cacheFilePath, data, 0644)
	if err != nil {
		log.Printf("Unable to write cache file (%s)", err)
	}
}

func (me *FlightInfos) getDataFromNetwork() ([]byte, error) {
	cli := http.Client{
		Timeout: time.Duration(APITimeout) * time.Second,
	}

	req, err := http.NewRequest(http.MethodPost, APIUrl, nil)
	if err != nil {
		fmt.Print(err)
		return nil, err
	}

	req.Header.Set("X-Requested-With", "XMLHttpRequest")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-type", "application/json")
	res, err := cli.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	if me.cacheAvailable {
		me.cacheWrite(body)
	}
	return body, nil
}

// GetData fetches flight data from the remote API or the local cache
func (me *FlightInfos) GetData() error {
	var body []byte
	var err error
	if me.cacheCanBeConsumed() {
		body, err = me.cacheRead()
		if err != nil {
			body, err = me.getDataFromNetwork()
		}
	} else {
		body, err = me.getDataFromNetwork()
	}

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
	sort.Slice((*me).Flights.Arrivals, func(i, j int) bool {
		return (*me).Flights.Arrivals[i].ScheduledArrival.time.Before((*me).Flights.Arrivals[j].ScheduledArrival.time)
	})
	sort.Slice((*me).Flights.Departures, func(i, j int) bool {
		return (*me).Flights.Departures[i].ScheduledDeparture.time.Before((*me).Flights.Departures[j].ScheduledDeparture.time)
	})
}

// PrepareDeparturesTable massages and formats the Departures table
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
			dep.PublicDeparture.StringDelay(dep.ScheduledDeparture),
			fmt.Sprintf("%s (%s)", dep.Airport, dep.AirportCodeDestination),
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

// PrepareArrivalsTable massages and formats the Arrivals table
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
			arr.PublicArrival.StringDelay(arr.ScheduledArrival),
			arr.DepartureFromPreviousAirport.String(),
			fmt.Sprintf("%s (%s)", arr.Airport, arr.AirportCode),
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
