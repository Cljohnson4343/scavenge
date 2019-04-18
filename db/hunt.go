package db

import (
	"time"
)

// A HuntDB is the representation of a hunt from the database's
// hunts table.
//
// swagger:model Hunt
type HuntDB struct {

	// The name of the Hunt
	//
	// required: true
	// maximum length: 255
	Name string `json:"name"`

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
	StartTime time.Time `json:"start_time"`

	// The end time for the Hunt
	//
	// required: true
	// swagger:strfmt date
	EndTime time.Time `json:"end_time"`

	// The creation time for the Hunt
	//
	// required: true
	// swagger:strfmt date
	CreatedAt time.Time `json:"created_at"`

	// The name of the location of the Hunt
	//
	// required: true
	// maximum length: 80
	LocationName string `json:"location_name"`

	// The latitude for the Hunt
	//
	// required: true
	Latitude float32 `json:"latitude"`

	// The longitude for the Hunt
	//
	// required: true
	Longitude float32 `json:"longitude"`
}
