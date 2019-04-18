package models

import "github.com/cljohnson4343/scavenge/db"

// A Location is a representation of a Location
//
// swagger:model Location
type Location struct {
	db.LocationDB
}
