package hunts

import (
	"github.com/cljohnson4343/scavenge/db"

	"github.com/cljohnson4343/scavenge/hunts/models"
	"github.com/cljohnson4343/scavenge/teams"
)

// A Hunt is the representation of a scavenger hunt.
//
// swagger:model Hunt
type Hunt struct {
	db.HuntDB

	// the teams for this hunt
	Teams []teams.Team `json:"teams"`

	// the items for this hunt
	//
	// min length: 1
	Items []models.Item `json:"items"`
}
