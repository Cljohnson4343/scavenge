package models

// A Coord is the representation of gps coordinates
//
// swagger:model Coord
type Coord struct {
	// the latitude of the coordinates
	//
	// required: true
	Latitude float32 `json:"latitude"`

	// the longitude of the coordinates
	//
	// required: true
	Longitude float32 `json:"longitude"`
}

// A Location is a representation of a Location
//
// swagger:model Location
type Location struct {

	// the name of the location
	//
	// maximum length: 80
	Name string `json:"name,omitempty"`

	// the gps coordinates for the location
	//
	// required: true
	Coords Coord `json:"coords"`
}
