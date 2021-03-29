package main

import (
	"strings"

	"github.com/jwalton/gchalk"
)

// FlightStatus is a dummy struct to allow Stringer() redef
type FlightStatus struct {
	status string
}

// "Took off", "Cancelled, "Departed", "Boarding", "Go to gate"
func (me *FlightStatus) String() string {
	switch status := me.status; status {
	case "Boarding":
		return gchalk.WithBrightMagenta().Bold(status)
	case "Go to gate":
		return gchalk.WithBrightCyan().Bold(status)
	case "Arrived":
		return gchalk.Green(status)
	case "Departed":
		return gchalk.Blue(status)
	case "Delayed":
		return gchalk.Yellow(status)
	case "Next Info":
		return gchalk.WithWhite().BgYellow("Delayed")
	case "Cancelled":
		return gchalk.WithWhite().WithBgRed().Bold(status)
	default:
		return gchalk.WithGreen().Bold(status)
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

type FlightType uint

const (
	FlightTypeSchengen = iota
	FlightTypeInternational
	FlightTypeFrance
	FlightTypeOther
	FlightTypeUnknown
)

func (me *FlightType) String() string {
	switch *me {
	case FlightTypeSchengen:
		return "Schengen"
	case FlightTypeInternational:
		return "International"
	case FlightTypeFrance:
		return "France"
	case FlightTypeOther:
		return "Other" // Not clear
	default:
		return "Unknown"
	}
}

func (me *FlightType) UnmarshalJSON(b []byte) (err error) {
	switch strings.Trim(string(b), "\"") {
	case "S":
		*me = FlightTypeSchengen
	case "I":
		*me = FlightTypeInternational
	case "F":
		*me = FlightTypeFrance
	case "O":
		*me = FlightTypeOther
	case "null":
		*me = FlightTypeUnknown
	}
	return
}
