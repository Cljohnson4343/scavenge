package hunts

import (
	"net/http"

	"github.com/cljohnson4343/scavenge/db"
	"github.com/cljohnson4343/scavenge/roles"

	"github.com/cljohnson4343/scavenge/response"

	c "github.com/cljohnson4343/scavenge/config"
	"github.com/cljohnson4343/scavenge/pgsql"
	"github.com/cljohnson4343/scavenge/teams"
)

// AllHunts returns all Hunts from the database
func AllHunts() ([]*Hunt, *response.Error) {
	huntDBs, e := db.GetHunts()
	if huntDBs == nil {
		return nil, e
	}

	if e == nil {
		e = response.NewNilError()
	}

	hunts := make([]*Hunt, 0, len(huntDBs))

	for _, h := range huntDBs {
		ts, teamErr := teams.GetTeamsForHunt(h.ID)
		if teamErr != nil {
			e.AddError(teamErr)
		}

		items, itemErr := GetItemsForHunt(h.ID)
		if itemErr != nil {
			e.AddError(itemErr)
		}

		hunt := Hunt{HuntDB: *h, Teams: ts, Items: items}

		hunts = append(hunts, &hunt)
	}

	return hunts, e.GetError()
}

// GetHunt returns a pointer to the hunt with the given ID.
func GetHunt(huntID int) (*Hunt, *response.Error) {
	huntDB, e := db.GetHunt(huntID)
	if e != nil {
		return nil, e
	}

	e = response.NewNilError()

	teams, teamErr := teams.GetTeamsForHunt(huntID)
	if teamErr != nil {
		e.AddError(teamErr)
	}

	items, itemErr := GetItems(huntID)
	if itemErr != nil {
		e.AddError(itemErr)
	}

	return &Hunt{HuntDB: *huntDB, Teams: teams, Items: items}, e.GetError()
}

// InsertHunt inserts the given hunt into the database and updates the hunt
// with the new id and created_at timestamp
func InsertHunt(hunt *Hunt) *response.Error {
	e := hunt.HuntDB.Insert()
	if e != nil {
		return e
	}

	e = response.NewNilError()

	for _, team := range hunt.Teams {
		team.HuntID = hunt.ID
		teamErr := team.Insert()
		if teamErr != nil {
			e.AddError(teamErr)
			break
		}
	}

	for _, item := range hunt.Items {
		item.HuntID = hunt.ID
		itemErr := item.Insert()
		if itemErr != nil {
			e.AddError(itemErr)
			break
		}
	}

	return e.GetError()
}

// DeleteHunt deletes the hunt with the given ID. All associated data will also be deleted.
func DeleteHunt(huntID int) *response.Error {
	e := db.DeleteHunt(huntID)
	if e != nil {
		return e
	}

	return roles.DeleteRolesForHunt(huntID)
}

// UpdateHunt updates the hunt with the given ID using the fields that are not nil in the
// partial hunt. If the hunt was updated then true will be returned. id field can not be
// updated.
func UpdateHunt(env *c.Env, hunt *Hunt) (bool, *response.Error) {
	tx, err := env.Begin()
	if err != nil {
		return false, response.NewError(http.StatusInternalServerError, err.Error())
	}

	e := hunt.Update(tx)
	if e != nil {
		err = tx.Rollback()
		if err != nil {
			e.Add(http.StatusInternalServerError, err.Error())
			return false, e.GetError()
		}

		return false, e
	}

	err = tx.Commit()
	if err != nil {
		return false, response.NewError(http.StatusInternalServerError, err.Error())
	}

	return true, nil
}

// getUpdateHuntSQLCommand returns the commands to update the db based on the provided hunt
// the partial hunt. The given hunt should only provide data that needs to be updated AND
// the hunt.ID field MUST be set
func getUpdateHuntSQLCommand(hunt *Hunt) (*[]*pgsql.Command, *response.Error) {
	// get all the mappings for all hunt's entities to their table, column name, and value
	tblColMaps := hunt.GetTableColumnMaps()

	// use the number of cmds that will be needed to avoid unnecessary memory allocations
	sqlCmds := make([]*pgsql.Command, 0, len(tblColMaps))

	e := response.NewNilError()

	for _, tblColMap := range tblColMaps {
		for tbl, colMap := range tblColMap {
			cmd, cmdErr := pgsql.GetUpdateSQLCommand(colMap, tbl, hunt.ID)
			if cmdErr != nil {
				e.AddError(cmdErr)
				break
			}
			sqlCmds = append(sqlCmds, cmd)
		}
	}

	return &sqlCmds, nil
}
