package main

import (
	"fmt"
	"strings"
	"time"
)

// GVATime wraps time for manual parsing
type GVATime struct {
	time time.Time
}

func (me *GVATime) String() string {
	if me.time.IsZero() {
		return ""
	}
	return fmt.Sprint(me.time.Format(DisplayTimeFormat))
}

// UnmarshalJSON implements what is needed to turn JSON-like time in GVATime
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

// Flight contains relevant data extracted during JSON unmarshalling
type Flight struct {
	FlightIdentity               string       `json:"flight_identity"`
	DisplayedFlightIdentity      string       `json:"displayed_flight_identity"`
	Airport                      string       `json:"airport"`
	ViaAirport                   string       `json:"via_airport"`
	Carousel                     string       `json:"carousel"`
	Terminal                     string       `json:"terminal"`
	Aircraft                     string       `json:"aircraft"`
	FlightStatus                 FlightStatus `json:"flight_status"`
	DisplayedMasterFlightCodes   string       `json:"displayed_master_flight_codes"`
	Gate                         string       `json:"gate"`
	CheckinDesks                 string       `json:"checkin_desks"`
	DepartureBool                int          `json:"departure_bool"`
	Company                      string       `json:"company"`
	AirportCode                  string       `json:"airport_code"`
	OriginCountry                string       `json:"origin_country"`
	AirportCodeDestination       string       `json:"airport_code_destination"`
	FlightType                   string       `json:"flight_type"`
	FlightDurationMinutes        int          `json:"flight_duration_minuts"`
	ControleDouane               int          `json:"controle_douane"`
	GateWalkTime                 int          `json:"gate_walk_time"`
	DelayMinutes                 int          `json:"delay_minuts"`
	FlightID                     int          `json:"flight_id"`
	AircraftID                   int          `json:"aircraft_id"`
	NextPublicAdvice             GVATime      `json:"next_public_advice"`
	ScheduledDeparture           GVATime      `json:"scheduled_departure"`
	ScheduledArrival             GVATime      `json:"scheduled_arrival"`
	Departure                    GVATime      `json:"departure"`
	Arrival                      GVATime      `json:"arrival"`
	PublicArrival                GVATime      `json:"public_arrival"`
	PublicDeparture              GVATime      `json:"public_departure"`
	Airborn                      GVATime      `json:"airborn"`
	FirstPriorityBaggage         GVATime      `json:"first_priority_baggage"`
	DepartureFromPreviousAirport GVATime      `json:"departure_from_previous_airport"`
	EstimatedLanding             GVATime      `json:"estimated_landing"`
	EstimatedBoarding            GVATime      `json:"estimated_boarding"`
	ScheduledFlight              GVATime      `json:"scheduled_flight"`
	AircraftRegistration         string       `json:"aircraft_registration"`
	State                        string       `json:"state"`
	LastUpdate                   GVATime      `json:"last_update"`
	AirportCity                  string       `json:"airport_city"`
}
