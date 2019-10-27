package main

import (
	"strings"

	"gopkg.in/gookit/color.v1"
)

// FlightStatus is a dummy struct to allow Stringer() redef
type FlightStatus struct {
	status string
}

// "Took off", "Cancelled, "Departed", "Boarding", "Go to gate"
func (me *FlightStatus) String() string {
	switch status := me.status; status {
	case "Boarding":
		return color.Question.Sprintf(status)
	case "Go to gate":
		return color.Note.Sprintf(status)
	case "Arrived":
		return color.Info.Sprintf(status)
	case "Departed":
		return color.Primary.Sprintf(status)
	case "Delayed":
		return color.Danger.Sprintf(status)
	case "Cancelled":
		return color.Error.Sprintf(status)
	default:
		return color.Success.Sprintf(status)
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
