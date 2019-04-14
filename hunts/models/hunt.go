package models

import (
	"time"
)

// A Hunt is the representation of a scavenger hunt.
//
// swagger:model Hunt
type Hunt struct {

	// The name of the Hunt
	//
	// required: true
	// maximum length: 255
	Title string `json:"title"`

	// The maximum number of teams that can participate in the Hunt.
	//
	// minimum: 1
	// required: true
	MaxTeams int `json:"max_teams"`

	// The id of the Hunt
	//
	// required: true
	ID int `json:"id"`

	// The start time for the Hunt
	//
	// required: true
	// swagger:strfmt date
	Start time.Time `json:"start"`

	// The end time for the Hunt
	//
	// required: true
	// swagger:strfmt date
	End time.Time `json:"end"`

	// the teams for this hunt
	Teams []Team `json:"teams"`

	// the items for this hunt
	//
	// min length: 1
	Items []Item `json:"items"`

	// the location information for this hunt
	Location Location `json:"location"`
}
