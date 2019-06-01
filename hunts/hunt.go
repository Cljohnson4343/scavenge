package hunts

import (
	"net/http"

	"github.com/cljohnson4343/scavenge/db"
	"github.com/cljohnson4343/scavenge/hunts/models"
	"github.com/cljohnson4343/scavenge/pgsql"
	"github.com/cljohnson4343/scavenge/response"
	"github.com/cljohnson4343/scavenge/teams"
)

// A Hunt is the representation of a scavenger hunt.
//
// swagger:model Hunt
type Hunt struct {
	db.HuntDB `valid:"-"`

	// the teams for this hunt
	Teams []*teams.Team `json:"teams" valid:"-"`

	// the items for this hunt
	//
	// min length: 1
	Items []*models.Item `json:"items" valid:"-"`

	// the players for this hunt
	Players []*db.PlayerDB `json:"players" valid:"-"`

	// the invites for this hunt
	Invites []*db.HuntInvitationDB `json:"invites" valid:"-"`
}

// Validate will validate the Hunt
func (h *Hunt) Validate(r *http.Request) *response.Error {
	e := response.NewNilError()

	huntErr := h.HuntDB.Validate(r)
	if huntErr != nil {
		e.AddError(huntErr)
	}

	for _, t := range h.Teams {
		// make sure each team has the same hunt_id as the hunt being validated or
		// doesn't specify a hunt_id, i.e zero valued
		if h.ID != t.HuntID && t.HuntID != 0 {
			e.Add(http.StatusBadRequest, "error: teams can not specify a hunt_id that differs from their enclosing hunt")
			break
		}
		teamErr := t.Validate(r)
		if teamErr != nil {
			e.AddError(teamErr)
		}
	}

	for _, i := range h.Items {
		// make sure each item has the same hunt_id as the hunt being validated or
		// doesn't specify a hunt_id, i.e zero valued
		if h.ID != i.HuntID && i.HuntID != 0 {
			e.Add(http.StatusBadRequest, "error: items can not specify a hunt_id that differs from their enclosing hunt")
			break
		}
		itemErr := i.Validate(r)
		if itemErr != nil {
			e.AddError(itemErr)
		}
	}

	return e.GetError()
}

// GetTableColumnMaps returns mappings for each non-zero value field and
// entity that h contains. These mappings associate an entity with its
// table, column name, and value
func (h *Hunt) GetTableColumnMaps() []pgsql.TableColumnMap {
	numMaps := 1 + len(h.Teams) + len(h.Items)
	tblColMaps := make([]pgsql.TableColumnMap, 0, numMaps)

	tblColMaps = append(tblColMaps, h.HuntDB.GetTableColumnMap())

	for _, v := range h.Teams {
		tblColMaps = append(tblColMaps, v.TeamDB.GetTableColumnMap())
	}

	for _, v := range h.Items {
		tblColMaps = append(tblColMaps, v.ItemDB.GetTableColumnMap())
	}

	return tblColMaps
}

// PatchValidate will validate only the non-zero fields of the Hunt
// TODO think about implementing govalidator's customtagtype validators
// 		for embedded fields and slice fields of Hunt, Item, and Team types
func (h *Hunt) PatchValidate(r *http.Request, huntID int) *response.Error {
	e := response.NewNilError()
	huntDBErr := h.HuntDB.PatchValidate(r, huntID)
	if huntDBErr != nil {
		e.AddError(huntDBErr)
	}

	for _, team := range h.Teams {
		// make sure each team has the same hunt_id as the hunt being validated or
		// doesn't specify a hunt_id, i.e zero valued
		if h.ID != team.HuntID && team.HuntID != 0 {
			e.Add(http.StatusBadRequest, "error: teams can not specify a hunt_id that differs from their enclosing hunt")
			break
		}
		// make sure each team has an id specified
		if team.ID == 0 {
			e.Add(http.StatusBadRequest, "id: teams need an id specified to PATCH")
			break
		}
		teamErr := team.TeamDB.PatchValidate(r, team.ID)
		if teamErr != nil {
			e.AddError(teamErr)
		}
	}

	for _, item := range h.Items {
		// make sure each item has the same hunt_id as the hunt being validated or
		// doesn't specify a hunt_id, i.e zero valued
		if h.ID != item.HuntID && item.HuntID != 0 {
			e.Add(http.StatusBadRequest, "error: items can not specify a hunt_id that differs from their enclosing hunt")
			break
		}
		// make sure each item has an id specified
		if item.ID == 0 {
			e.Add(http.StatusBadRequest, "id: items need an id specified to PATCH")
			break
		}
		itemErr := item.ItemDB.PatchValidate(r, item.ID)
		if itemErr != nil {
			e.AddError(itemErr)
		}
	}

	return e.GetError()
}

// Update updates all the non-zero value fields in all of its component structures
func (h *Hunt) Update(ex pgsql.Executioner) *response.Error {
	e := h.HuntDB.Update(ex)
	if e != nil {
		return e
	}

	for _, team := range h.Teams {
		e = team.Update(ex)
		if e != nil {
			return e
		}
	}

	for _, item := range h.Items {
		e = item.Update(ex)
		if e != nil {
			return e
		}
	}

	return nil
}
