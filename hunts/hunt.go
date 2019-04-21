package hunts

import (
	"errors"
	"net/http"

	"github.com/cljohnson4343/scavenge/pgsql"

	"github.com/cljohnson4343/scavenge/db"
	"github.com/cljohnson4343/scavenge/response"

	"github.com/cljohnson4343/scavenge/hunts/models"
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

// GetTableColumnMap overshadows the embedded HuntDB GetTableColumnMap and always
// panics. This is a safety measure to prevent calling an unsafe method. GetTableColumnMap
// should not be called on types that have fields outside of an embedded *DB type.
// For example a Hunt can have multiple teams in its Teams field  and so a single
// table, column name, and value mapping can not represent all of the teams.
func (h *Hunt) GetTableColumnMap() pgsql.TableColumnMap {
	panic(errors.New("error: you should use GetTableColumnMaps for this type"))
}

// GetTableColumnMaps returns mappings for each non-zero value
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

// PartialValidate will validate only the non-zero fields of the Hunt
// @TODO think about implementing govalidator's customtagtype validators
// 		for embedded fields and slice fields of Hunt, Item, and Team types
func (h *Hunt) PartialValidate(r *http.Request) *response.Error {
	e := response.NewNilError()
	huntDBErr := h.HuntDB.PartialValidate(r)
	if huntDBErr != nil {
		e.AddError(huntDBErr)
	}

	for _, team := range h.Teams {
		teamErr := team.TeamDB.PartialValidate(r)
		if teamErr != nil {
			e.AddError(teamErr)
		}
	}

	for _, item := range h.Items {
		itemErr := item.ItemDB.Validate(r)
		if itemErr != nil {
			e.AddError(itemErr)
		}
	}

	return e.GetError()
}
