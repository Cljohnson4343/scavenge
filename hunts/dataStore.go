package hunts

import (
	"fmt"
	"net/http"
	"strings"
	"time"

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
func UpdateHunt(env *c.Env, huntID int, partialHunt *map[string]interface{}) (bool, *response.Error) {
	sqlCmds, e := getUpdateHuntSQLCommand(huntID, partialHunt)
	if e != nil {
		return false, e
	}

	tx, err := env.Begin()
	if err != nil {
		return false, response.NewError(err.Error(), http.StatusInternalServerError)
	}

	// attempt to execute all the statements in this transaction
	for _, v := range *sqlCmds {
		_, err := v.Exec(tx)
		if err != nil {
			tx.Rollback()
			return false, response.NewError(err.Error(), http.StatusInternalServerError)
		}
	}

	err = tx.Commit()
	if err != nil {
		return false, response.NewError(err.Error(), http.StatusInternalServerError)
	}

	return true, nil
}

func getUpdateHuntSQLCommand(huntID int, partialHunt *map[string]interface{}) (*[]*pgsql.Command, *response.Error) {
	var sqlb strings.Builder
	sqlb.WriteString("\n\t\tUPDATE hunts\n\t\tSET")

	sqlCmds := make([]*pgsql.Command, 0)
	sqlCmds = append(sqlCmds, new(pgsql.Command))

	e := response.NewNilError()
	inc := 1
	for k, v := range *partialHunt {
		switch k {
		case "name":
			newName, ok := v.(string)
			if !ok {
				e.Add("name field has to be type string", http.StatusBadRequest)
				break
			}
			sqlb.WriteString(fmt.Sprintf(" name=$%d,", inc))
			inc++
			sqlCmds[0].AppendArgs(newName)
		case "max_teams":
			newMax, ok := v.(float64)
			if !ok {
				e.Add("max_teams field has to be type float64", http.StatusBadRequest)
				break
			}
			sqlb.WriteString(fmt.Sprintf(" max_teams=$%d,", inc))
			inc++
			sqlCmds[0].AppendArgs(int(newMax))

		case "start_time":
			newStart, ok := v.(string)
			if !ok {
				e.Add("start_time field has to be type float64", http.StatusBadRequest)
				break

			}

			startTime, err := time.Parse(time.RFC3339, newStart)
			if err != nil {
				e.Add(err.Error(), http.StatusBadRequest)
				break
			}

			sqlb.WriteString(fmt.Sprintf(" start_time=$%d,", inc))
			inc++
			sqlCmds[0].AppendArgs(startTime)

		case "end_time":
			newEnd, ok := v.(string)
			if !ok {
				e.Add("end_time field has to be of type string", http.StatusBadRequest)
				break
			}
			endTime, err := time.Parse(time.RFC3339, newEnd)
			if err != nil {
				e.Add(err.Error(), http.StatusBadRequest)
				break
			}

			sqlb.WriteString(fmt.Sprintf(" end_time=$%d,", inc))
			inc++
			sqlCmds[0].AppendArgs(endTime)

		case "teams":
			newTeams, ok := v.([]interface{})
			if !ok {
				e.Add("teams field is invalid", http.StatusBadRequest)
				break
			}

			newTeamsCmd, teamErr := teams.GetUpsertTeamsSQLCommand(huntID, newTeams)
			if teamErr != nil {
				e.Add(teamErr.Error(), teamErr.Code())
				break
			}

			sqlCmds = append(sqlCmds, newTeamsCmd)

		case "items":
			newItems, ok := v.([]interface{})
			if !ok {
				e.Add("items field is invalid", http.StatusBadRequest)
				break
			}

			// @TODO think about how to handle the case where an error is thrown(should we try partial execution?)
			newItemsCmd, itemErr := getUpsertItemsSQLStatement(huntID, newItems)
			if itemErr != nil {
				e.Add(itemErr.Error(), itemErr.Code())
				break
			}

			sqlCmds = append(sqlCmds, newItemsCmd)

		case "location_name":
			locName, ok := v.(string)
			if !ok {
				e.Add("location_name field has to be of type string", http.StatusBadRequest)
				break
			}

			sqlb.WriteString(fmt.Sprintf(" location_name=$%d,", inc))
			inc++
			sqlCmds[0].AppendArgs(locName)

		case "latitude":
			lat, ok := v.(float64)
			if !ok {
				e.Add("latitude field has to be of type float64", http.StatusBadRequest)
				break
			}

			sqlb.WriteString(fmt.Sprintf(" latitude=$%d,", inc))
			inc++
			sqlCmds[0].AppendArgs(lat)

		case "longitude":
			lon, ok := v.(float64)
			if !ok {
				e.Add("longitude field has to be of type float64", http.StatusBadRequest)
				break
			}

			sqlb.WriteString(fmt.Sprintf(" longitude=$%d,", inc))
			inc++
			sqlCmds[0].AppendArgs(lon)
		}
	}

	l := sqlb.Len()
	sqlCmds[0].AppendScript(fmt.Sprintf("%s\n\t\tWHERE id = $%d", sqlb.String()[0:l-1], inc))
	sqlCmds[0].AppendArgs(huntID)

	return &sqlCmds, e.GetError()
}
