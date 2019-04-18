package hunts

import (
	"errors"
	"fmt"
	"strings"
	"time"

	c "github.com/cljohnson4343/scavenge/config"
	"github.com/cljohnson4343/scavenge/teams"
)

// AllHunts returns all Hunts from the database
func AllHunts(env *c.Env) ([]*Hunt, error) {
	rows, err := env.DB().Query("SELECT id, name, max_teams, start_time, end_time, latitude, longitude, location_name, created_at FROM hunts;")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	hunts := make([]*Hunt, 0)
	for rows.Next() {
		hunt := new(Hunt)
		err = rows.Scan(&hunt.ID, &hunt.Name, &hunt.MaxTeams, &hunt.Start,
			&hunt.End, &hunt.Location.Coords.Latitude,
			&hunt.Location.Coords.Longitude, &hunt.Location.Name, &hunt.CreationDate)
		if err != nil {
			return nil, err
		}

		teams, err := teams.GetTeamsForHunt(env, hunt.ID)
		if err != nil {
			return nil, err
		}
		hunt.Teams = *teams

		items, err := GetItems(env, hunt.ID)
		if err != nil {
			return nil, err
		}
		hunt.Items = *items

		hunts = append(hunts, hunt)
	}

	err = rows.Err()

	return hunts, err
}

// GetHunt returns a pointer to the hunt with the given ID.
func GetHunt(env *c.Env, hunt *Hunt, huntID int) error {
	sqlStmnt := `
		SELECT name, max_teams, start_time, end_time, latitude, longitude, location_name FROM hunts
		WHERE hunts.id = $1;`

	err := env.DB().QueryRow(sqlStmnt, huntID).Scan(&hunt.Name, &hunt.MaxTeams, &hunt.Start,
		&hunt.End, &hunt.Location.Coords.Latitude, &hunt.Location.Coords.Longitude, &hunt.Location.Name)
	if err != nil {
		return err
	}

	// @TODO make sure geteams doesnt return an error if no teams are found. we need to still
	// get items
	teams, err := teams.GetTeamsForHunt(env, huntID)
	if err != nil {
		return err
	}
	hunt.Teams = *teams

	items, err := GetItems(env, huntID)
	if err != nil {
		return err
	}
	hunt.Items = *items

	return err
}

// InsertHunt inserts the given hunt into the database and returns the id of the inserted hunt
func InsertHunt(env *c.Env, hunt *Hunt) (int, error) {
	sqlStmnt := `
		INSERT INTO hunts(name, max_teams, start_time, end_time, location_name, latitude, longitude)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id AS hunt_id
		`
	// @TODO look into whether the row from queryrow needs to be closed
	id := 0
	err := env.DB().QueryRow(sqlStmnt, hunt.Name, hunt.MaxTeams, hunt.Start,
		hunt.End, hunt.Location.Name, hunt.Location.Coords.Latitude,
		hunt.Location.Coords.Longitude).Scan(&id)
	if err != nil {
		return id, err
	}

	// @TODO look into better error handling. Right now a failed team or item creation
	// will skip and wipe the error
	for _, v := range hunt.Teams {
		_, err = teams.InsertTeam(env, &v, id)
		if err != nil {
			break
		}
	}

	for _, v := range hunt.Items {
		_, err = InsertItem(env, &v, id)
		if err != nil {
			break
		}
	}

	return id, err
}

// DeleteHunt deletes the hunt with the given ID. All associated data will also be deleted.
func DeleteHunt(env *c.Env, huntID int) error {
	sqlStmnt := `
		DELETE FROM hunts
		WHERE hunts.id = $1`

	_, err := env.DB().Exec(sqlStmnt, huntID)
	return err
}

// UpdateHunt updates the hunt with the given ID using the fields that are not nil in the
// partial hunt. If the hunt was updated then true will be returned. id field can not be
// updated.
func UpdateHunt(env *c.Env, huntID int, partialHunt *map[string]interface{}) (bool, error) {
	sqlStmnts, err := getUpdateHuntSQLStatement(huntID, partialHunt)
	if err != nil {
		return false, err
	}

	tx, err := env.DB().Begin()
	if err != nil {
		return false, err
	}

	// atempt to execute all the statements in this transaction
	for _, v := range *sqlStmnts {
		_, err := v.Exec(tx)
		if err != nil {
			tx.Rollback()
			return false, err
		}
	}

	return true, tx.Commit()
}

func getUpdateHuntSQLStatement(huntID int, partialHunt *map[string]interface{}) (*[]*c.SQLStatement, error) {
	var eb, sqlb strings.Builder

	eb.WriteString("Error updating hunt: \n")
	encounteredError := false

	handleErr := func(errString string) {
		encounteredError = true
		eb.WriteString(errString)
	}

	sqlStmnts := make([]*c.SQLStatement, 0)
	sqlStmnts = append(sqlStmnts, new(c.SQLStatement))

	sqlb.WriteString("\n\t\tUPDATE hunts\n\t\tSET")

	inc := 1
	for k, v := range *partialHunt {
		switch k {
		case "name":
			newName, ok := v.(string)
			if !ok {
				handleErr(fmt.Sprintf("Expected name to be of type string but got %T\n", v))
				break
			}
			sqlb.WriteString(fmt.Sprintf(" name=$%d,", inc))
			inc++
			sqlStmnts[0].AppendArgs(newName)
		case "max_teams":
			newMax, ok := v.(float64)
			if !ok {
				handleErr(fmt.Sprintf("Expected max_teams to be of type float64 but got %T\n", v))
				break
			}
			sqlb.WriteString(fmt.Sprintf(" max_teams=$%d,", inc))
			inc++
			sqlStmnts[0].AppendArgs(int(newMax))

		case "start":
			newStart, ok := v.(string)
			if !ok {
				handleErr(fmt.Sprintf("Expected start to be of type string but got %T\n", v))
				break

			}

			startTime, err := time.Parse(time.RFC3339, newStart)
			if err != nil {
				handleErr(fmt.Sprintf("%s\n", err.Error()))
				break
			}

			sqlb.WriteString(fmt.Sprintf(" start_time=$%d,", inc))
			inc++
			sqlStmnts[0].AppendArgs(startTime)

		case "end":
			newEnd, ok := v.(string)
			if !ok {
				handleErr(fmt.Sprintf("Expected end to be of type string but got %T\n", v))
				break
			}
			endTime, err := time.Parse(time.RFC3339, newEnd)
			if err != nil {
				handleErr(fmt.Sprintf("%s\n", err.Error()))
				break
			}

			sqlb.WriteString(fmt.Sprintf(" end_time=$%d,", inc))
			inc++
			sqlStmnts[0].AppendArgs(endTime)

		case "teams":
			newTeams, ok := v.([]interface{})
			if !ok {
				handleErr(fmt.Sprintf("Expected teams to be of type []interface{} but got %T\n", v))
				break
			}

			// @TODO think about how to handle the case where an error is thrown(should we try partial execution?)
			newTeamsStmnt, err := teams.GetUpsertTeamsSQLStatement(huntID, newTeams)
			if err != nil {
				handleErr(fmt.Sprintf("%s\n", err.Error()))
				break
			}

			sqlStmnts = append(sqlStmnts, newTeamsStmnt)

		case "items":
			newItems, ok := v.([]interface{})
			if !ok {
				handleErr(fmt.Sprintf("Expected items to be of type []interface{} but got %T\n", v))
				break
			}

			// @TODO think about how to handle the case where an error is thrown(should we try partial execution?)
			newItemsStmnt, err := getUpsertItemsSQLStatement(huntID, newItems)
			if err != nil {
				handleErr(fmt.Sprintf("%s\n", err.Error()))
				break
			}

			sqlStmnts = append(sqlStmnts, newItemsStmnt)

		case "location":
			partialLoc, ok := v.(map[string]interface{})
			if !ok {
				handleErr(fmt.Sprintf("Expected location to be of type map[string]interface{} but got %T\n", v))
				break
			}

			locName, ok := partialLoc["name"].(string)
			if ok {
				sqlb.WriteString(fmt.Sprintf(" location_name=$%d,", inc))
				inc++
				sqlStmnts[0].AppendArgs(locName)
			}

			coords, ok := partialLoc["coords"].(map[string]interface{})
			if ok {
				lat, ok := coords["latitude"].(float64)
				if ok {
					sqlb.WriteString(fmt.Sprintf(" latitude=$%d,", inc))
					inc++
					sqlStmnts[0].AppendArgs(lat)
				}
				lon, ok := coords["longitude"].(float64)
				if ok {
					sqlb.WriteString(fmt.Sprintf(" longitude=$%d,", inc))
					inc++
					sqlStmnts[0].AppendArgs(lon)
				}
			}

		default:
		}
	}

	l := sqlb.Len()
	sqlStmnts[0].AppendScript(fmt.Sprintf("%s\n\t\tWHERE id = $%d", sqlb.String()[0:l-1], inc))
	sqlStmnts[0].AppendArgs(huntID)

	if encounteredError {
		return &sqlStmnts, errors.New(eb.String())
	}

	return &sqlStmnts, nil
}
