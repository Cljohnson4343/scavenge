package hunts

import (
	"net/http"

	"github.com/cljohnson4343/scavenge/db"
	"github.com/cljohnson4343/scavenge/response"

	"github.com/cljohnson4343/scavenge/hunts/models"
	"github.com/cljohnson4343/scavenge/teams"
)

// A Hunt is the representation of a scavenger hunt.
//
// swagger:model Hunt
type Hunt struct {
	db.HuntDB

	// the teams for this hunt
	Teams []*teams.Team `json:"teams"`

	// the items for this hunt
	//
	// min length: 1
	Items []*models.Item `json:"items"`
}

// Validate will validate the Hunt
func (h *Hunt) Validate(r *http.Request) *response.Error {
	e := response.NewNilError()

	huntErr := h.HuntDB.Validate(r)
	if huntErr != nil {
		e.AddError(huntErr)
	}

	for _, t := range h.Teams {
		teamErr := t.ValidateWithoutHuntID(r)
		if teamErr != nil {
			e.AddError(teamErr)
		}
	}

	for _, i := range h.Items {
		itemErr := i.Validate(r)
		if itemErr != nil {
			e.AddError(itemErr)
		}
	}

	return e.GetError()
}
