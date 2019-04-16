package hunts

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/cljohnson4343/scavenge/hunts/models"
)

// HuntDataStore is an interface that comprised of all access methods
// that pkg hunts needs to communicate with the database
type HuntDataStore interface {
	allHunts() ([]*models.Hunt, error)
	getHunt(hunt *models.Hunt, huntID int) error
	getItems(items *[]models.Item, huntID int) error
	getTeams(huntID int) (*[]models.Team, error)
	insertHunt(hunt *models.Hunt) (int, error)
	insertTeam(team *models.Team, huntID int) (int, error)
	insertItem(item *models.Item, huntID int) (int, error)
	deleteHunt(huntID int) error
	updateHunt(huntID int, partialHunt *map[string]interface{}) (bool, error)
	getTeam(teamID int) (*models.Team, error)
}

// AllHunts returns all Hunts from the database
func (env *Env) allHunts() ([]*models.Hunt, error) {
	rows, err := env.db.Query("SELECT * FROM hunts;")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	hunts := make([]*models.Hunt, 0)
	for rows.Next() {
		hunt := new(models.Hunt)
		err = rows.Scan(&hunt.ID, &hunt.Title, &hunt.MaxTeams, &hunt.Start,
			&hunt.End, &hunt.Location.Coords.Latitude,
			&hunt.Location.Coords.Longitude, &hunt.Location.Name)
		if err != nil {
			return nil, err
		}

		teams, err := env.getTeams(hunt.ID)
		if err != nil {
			return nil, err
		}
		hunt.Teams = *teams

		err = env.getItems(&hunt.Items, hunt.ID)
		if err != nil {
			return nil, err
		}

		hunts = append(hunts, hunt)
	}

	err = rows.Err()

	return hunts, err
}

// getHunt returns a pointer to the hunt with the given ID.
func (env *Env) getHunt(hunt *models.Hunt, huntID int) error {
	sqlStatement := `
		SELECT title, max_teams, start_time, end_time, latitude, longitude, location_name FROM hunts
		WHERE hunts.id = $1;`

	err := env.db.QueryRow(sqlStatement, huntID).Scan(&hunt.Title, &hunt.MaxTeams, &hunt.Start,
		&hunt.End, &hunt.Location.Coords.Latitude, &hunt.Location.Coords.Longitude, &hunt.Location.Name)
	if err != nil {
		return err
	}

	// @TODO make sure getteams doesnt return an error if no teams are found. we need to still
	// get items
	teams, err := env.getTeams(huntID)
	if err != nil {
		return err
	}
	hunt.Teams = *teams

	err = env.getItems(&hunt.Items, huntID)

	return err
}

// insertHunt inserts the given hunt into the database and returns the id of the inserted hunt
func (env *Env) insertHunt(hunt *models.Hunt) (int, error) {
	sqlStatement := `
		INSERT INTO hunts(title, max_teams, start_time, end_time, location_name, latitude, longitude)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id AS hunt_id
		`
	// @TODO look into whether the row from queryrow needs to be closed
	id := 0
	err := env.db.QueryRow(sqlStatement, hunt.Title, hunt.MaxTeams, hunt.Start,
		hunt.End, hunt.Location.Name, hunt.Location.Coords.Latitude,
		hunt.Location.Coords.Longitude).Scan(&id)
	if err != nil {
		return id, err
	}

	for _, v := range hunt.Teams {
		_, err = env.insertTeam(&v, id)
		if err != nil {
			return id, err
		}
	}

	for _, v := range hunt.Items {
		_, err = env.insertItem(&v, id)
		if err != nil {
			return id, err
		}
	}

	return id, err
}

// deleteHunt deletes the hunt with the given ID. All associated data will also be deleted.
func (env *Env) deleteHunt(huntID int) error {
	sqlStatement := `
		DELETE FROM hunts
		WHERE hunts.id = $1`

	_, err := env.db.Exec(sqlStatement, huntID)
	return err
}

type sqlStatement struct {
	sql  string
	args []interface{}
}

// Executioner is an interface that is needed for database/sql polymorphism
type Executioner interface {
	Exec(query string, args ...interface{}) (sql.Result, error)
}

// updateHunt updates the hunt with the given ID using the fields that are not nil in the
// partial hunt. If the hunt was updated then true will be returned.
func (env *Env) updateHunt(huntID int, partialHunt *map[string]interface{}) (bool, error) {
	sqlStmnts, err := getSQLHuntUpdater(huntID, partialHunt)
	if err != nil {
		return false, err
	}

	tx, err := env.db.Begin()
	if err != nil {
		return false, err
	}

	// attempt to execute all the statements in this transaction
	for _, v := range *sqlStmnts {
		_, err := v.exec(tx)
		if err != nil {
			tx.Rollback()
			return false, err
		}
	}

	return true, tx.Commit()
}

func (sqlStmnt *sqlStatement) exec(ex Executioner) (sql.Result, error) {
	return ex.Exec(sqlStmnt.sql, sqlStmnt.args...)
}

func getSQLHuntUpdater(huntID int, partialHunt *map[string]interface{}) (*[]*sqlStatement, error) {
	var eb, sqlb strings.Builder

	eb.WriteString("Error updating hunt: \n")
	encounteredError := false

	handleErr := func(errString string) {
		encounteredError = true
		eb.WriteString(errString)
	}

	sqlStmnts := make([]*sqlStatement, 0)
	sqlStmnts = append(sqlStmnts, new(sqlStatement))

	sqlb.WriteString("\n\t\tUPDATE hunts\n\t\tSET")

	inc := 1
	for k, v := range *partialHunt {
		switch k {
		case "title":
			newTitle, ok := v.(string)
			if !ok {
				handleErr(fmt.Sprintf("Expected title to be of type string but got %T\n", v))
				break
			}
			sqlb.WriteString(fmt.Sprintf(" title=$%d,", inc))
			inc++
			sqlStmnts[0].args = append(sqlStmnts[0].args, newTitle)
		case "max_teams":
			newMax, ok := v.(float64)
			if !ok {
				handleErr(fmt.Sprintf("Expected max_teams to be of type float64 but got %T\n", v))
				break
			}
			sqlb.WriteString(fmt.Sprintf(" max_teams=$%d,", inc))
			inc++
			sqlStmnts[0].args = append(sqlStmnts[0].args, int(newMax))

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
			sqlStmnts[0].args = append(sqlStmnts[0].args, startTime)

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
			sqlStmnts[0].args = append(sqlStmnts[0].args, endTime)

		case "teams":
			newTeams, ok := v.([]interface{})
			if !ok {
				handleErr(fmt.Sprintf("Expected teams to be of type []interface{} but got %T\n", v))
				break
			}

			// @TODO think about how to handle the case where an error is thrown(should we try partial execution?)
			newTeamsStmnt, err := getUpsertTeamsSQLStatement(huntID, newTeams)
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
			_, ok := v.(map[string]interface{})
			if !ok {
				handleErr(fmt.Sprintf("Expected location to be of type map[string]interface{} but got %T\n", v))
				break
			}
		default:
		}
	}

	l := sqlb.Len()
	sqlStmnts[0].sql = fmt.Sprintf("%s\n\t\tWHERE id = $%d", sqlb.String()[0:l-1], inc)
	sqlStmnts[0].args = append(sqlStmnts[0].args, huntID)

	if encounteredError {
		return &sqlStmnts, errors.New(eb.String())
	}

	return &sqlStmnts, nil
}
