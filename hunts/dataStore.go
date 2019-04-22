package hunts

import (
	"net/http"

	"github.com/cljohnson4343/scavenge/response"

	c "github.com/cljohnson4343/scavenge/config"
	"github.com/cljohnson4343/scavenge/pgsql"
	"github.com/cljohnson4343/scavenge/teams"
)

// AllHunts returns all Hunts from the database
func AllHunts(env *c.Env) ([]*Hunt, *response.Error) {
	rows, err := env.Query("SELECT id, name, max_teams, start_time, end_time, latitude, longitude, location_name, created_at FROM hunts;")
	if err != nil {
		return nil, response.NewError(err.Error(), http.StatusInternalServerError)
	}
	defer rows.Close()

	e := response.NewNilError()
	hunts := make([]*Hunt, 0)
	for rows.Next() {
		hunt := new(Hunt)
		err = rows.Scan(&hunt.ID, &hunt.Name, &hunt.MaxTeams, &hunt.StartTime, &hunt.EndTime,
			&hunt.Latitude, &hunt.Longitude, &hunt.LocationName, &hunt.CreatedAt)
		if err != nil {
			e.Add(err.Error(), http.StatusInternalServerError)
			break
		}

		teams, teamErr := teams.GetTeamsForHunt(env, hunt.ID)
		if teamErr != nil {
			e.Add(teamErr.Error(), teamErr.Code())
		}
		hunt.Teams = *teams

		items, itemErr := GetItems(env, hunt.ID)
		if itemErr != nil {
			e.Add(itemErr.Error(), itemErr.Code())
		}
		hunt.Items = *items

		hunts = append(hunts, hunt)
	}

	err = rows.Err()
	if err != nil {
		e.Add(err.Error(), http.StatusInternalServerError)
	}

	return hunts, e.GetError()
}

// GetHunt returns a pointer to the hunt with the given ID.
func GetHunt(env *c.Env, hunt *Hunt, huntID int) *response.Error {
	sqlStmnt := `
		SELECT name, max_teams, start_time, end_time, latitude, longitude, location_name, created_at FROM hunts
		WHERE hunts.id = $1;`

	err := env.QueryRow(sqlStmnt, huntID).Scan(&hunt.Name, &hunt.MaxTeams, &hunt.StartTime,
		&hunt.EndTime, &hunt.Latitude, &hunt.Longitude, &hunt.LocationName, &hunt.CreatedAt)
	if err != nil {
		return response.NewError(err.Error(), http.StatusBadRequest)
	}

	e := response.NewNilError()
	// @TODO make sure geteams doesnt return an error if no teams are found. we need to still
	// get items
	teams, teamErr := teams.GetTeamsForHunt(env, huntID)
	if teamErr != nil {
		e.Add(teamErr.Error(), teamErr.Code())
	} else {
		hunt.Teams = *teams
	}

	items, itemErr := GetItems(env, huntID)
	if itemErr != nil {
		e.Add(itemErr.Error(), itemErr.Code())
	} else {
		hunt.Items = *items
	}

	return e.GetError()
}

// InsertHunt inserts the given hunt into the database and returns the id of the inserted hunt
func InsertHunt(env *c.Env, hunt *Hunt) (int, *response.Error) {
	sqlStmnt := `
		INSERT INTO hunts(name, max_teams, start_time, end_time, location_name, latitude, longitude)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, created_at;
		`

	//jsonColMap := hunt.GetJSONColumnMap()
	//tblColMap := hunt.GetTableColumnMap()

	// @TODO look into whether the row from queryrow needs to be closed
	err := env.QueryRow(sqlStmnt, hunt.Name, hunt.MaxTeams, hunt.StartTime,
		hunt.EndTime, hunt.LocationName, hunt.Latitude,
		hunt.Longitude).Scan(&hunt.ID, &hunt.CreatedAt)
	if err != nil {
		return hunt.ID, response.NewError(err.Error(), http.StatusInternalServerError)
	}

	e := response.NewNilError()
	// @TODO look into better error handling. Right now a failed team or item creation
	// will skip and wipe the error
	for _, v := range hunt.Teams {
		_, teamErr := teams.InsertTeam(env, v, hunt.ID)
		if teamErr != nil {
			e.Add(teamErr.Error(), teamErr.Code())
			break
		}
	}

	for _, v := range hunt.Items {
		_, itemErr := InsertItem(env, v, hunt.ID)
		if itemErr != nil {
			e.Add(itemErr.Error(), itemErr.Code())
			break
		}
	}

	return hunt.ID, e.GetError()
}

// DeleteHunt deletes the hunt with the given ID. All associated data will also be deleted.
func DeleteHunt(env *c.Env, huntID int) *response.Error {
	sqlStmnt := `
		DELETE FROM hunts
		WHERE hunts.id = $1`

	_, err := env.Exec(sqlStmnt, huntID)
	if err != nil {
		return response.NewError(err.Error(), http.StatusInternalServerError)
	}

	return nil
}

// UpdateHunt updates the hunt with the given ID using the fields that are not nil in the
// partial hunt. If the hunt was updated then true will be returned. id field can not be
// updated.
func UpdateHunt(env *c.Env, hunt *Hunt) (bool, *response.Error) {
	tx, err := env.Begin()
	if err != nil {
		return false, response.NewError(err.Error(), http.StatusInternalServerError)
	}

	e := hunt.Update(tx)
	if e != nil {
		err = tx.Rollback()
		if err != nil {
			e.Add(err.Error(), http.StatusInternalServerError)
			return false, e.GetError()
		}

		return false, e
	}

	err = tx.Commit()
	if err != nil {
		return false, response.NewError(err.Error(), http.StatusInternalServerError)
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
