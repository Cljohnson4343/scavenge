package db

import "time"

// A LocationDB is a representation of a row in the locations table
//
// swagger:model LocationDB
type LocationDB struct {

	// The id of the team
	//
	// required: true
	TeamID int `json:"team_id"`

	// The id of the location
	//
	// required: true
	ID int `json:"id"`

	// the latitude
	//
	// required: true
	Latitude float32 `json:"latitude"`

	// the longitude
	//
	// required: true
	Longitude float32 `json:"longitude"`

	// the time stamp for this location
	//
	// required: true
	// swagger:strfmt date
	TimeStamp time.Time `json:"time_stamp"`
}
