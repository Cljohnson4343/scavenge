package hunts

import (
	"errors"
	"fmt"
	"log"
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
	getTeams(teams *[]models.Team, huntID int) error
	insertHunt(hunt *models.Hunt) (int, error)
	insertTeam(team *models.Team, huntID int) (int, error)
	insertItem(item *models.Item, huntID int) (int, error)
	deleteHunt(huntID int) error
	updateHunt(huntID int, partialHunt *map[string]interface{}) (int, error)
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

		err = env.getTeams(&hunt.Teams, hunt.ID)
		if err != nil {
			return nil, err
		}

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
	err = env.getTeams(&hunt.Teams, huntID)
	if err != nil {
		return err
	}

	err = env.getItems(&hunt.Items, huntID)

	return err
}

// getItems populates the items slice with all the items for the given hunt
func (env *Env) getItems(items *[]models.Item, huntID int) error {
	sqlStatement := `
		SELECT name, points FROM items WHERE items.hunt_id = $1;`

	rows, err := env.db.Query(sqlStatement, huntID)
	if err != nil {
		return err
	}
	defer rows.Close()

	item := models.Item{}
	for rows.Next() {
		err = rows.Scan(&item.Name, &item.Points)
		if err != nil {
			return err
		}

		*items = append(*items, item)
	}

	return nil
}

// getTeams populates the teams slice with all the teams for the given hunt
func (env *Env) getTeams(teams *[]models.Team, huntID int) error {
	sqlStatement := `
		SELECT name FROM teams WHERE teams.hunt_id = $1;`

	rows, err := env.db.Query(sqlStatement, huntID)
	if err != nil {
		return err
	}
	defer rows.Close()

	team := models.Team{}
	for rows.Next() {
		err = rows.Scan(&team.Name)
		if err != nil {
			return err
		}

		*teams = append(*teams, team)
	}

	return nil
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

// insertTeam inserts a Team into the db
func (env *Env) insertTeam(team *models.Team, huntID int) (int, error) {
	sqlStatement := `
		INSERT INTO teams(hunt_id, name)
		VALUES ($1, $2)
		RETURNING id`

	id := 0
	err := env.db.QueryRow(sqlStatement, huntID, team.Name).Scan(&id)

	return id, err
}

// insertItem inserts an Item into the db
func (env *Env) insertItem(item *models.Item, huntID int) (int, error) {
	sqlStatement := `
		INSERT INTO items(hunt_id, name, points)
		VALUES ($1, $2, $3)
		RETURNING id`

	id := 0
	err := env.db.QueryRow(sqlStatement, huntID, item.Name, item.Points).Scan(&id)

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

type sqlHuntUpdater struct {
	sqlStatement string
	args         []interface{}
}

// updateHunt updates the hunt with the given ID using the fields that are not nil in the
// partial hunt.
func (env *Env) updateHunt(huntID int, partialHunt *map[string]interface{}) (int, error) {
	sqlUpdater, err := getSQLHuntUpdater(huntID, partialHunt)
	if err != nil {
		log.Print(err.Error())
	}

	res, err := env.db.Exec(sqlUpdater.sqlStatement, sqlUpdater.args...)
	if err != nil {
		return 0, err
	}

	count, err := res.RowsAffected()
	if err != nil {
		return 0, err
	}

	return int(count), nil
}

func getSQLHuntUpdater(huntID int, partialHunt *map[string]interface{}) (*sqlHuntUpdater, error) {
	var eb, sqlb strings.Builder

	eb.WriteString("Error updating hunt: \n")
	encounteredError := false

	handleErr := func(errString string) {
		encounteredError = true
		eb.WriteString(errString)
	}

	update := new(sqlHuntUpdater)

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
			update.args = append(update.args, newTitle)
		case "max_teams":
			newMax, ok := v.(float64)
			if !ok {
				handleErr(fmt.Sprintf("Expected max_teams to be of type float64 but got %T\n", v))
				break
			}
			sqlb.WriteString(fmt.Sprintf(" max_teams=$%d,", inc))
			inc++
			update.args = append(update.args, int(newMax))

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
			update.args = append(update.args, startTime)

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
			update.args = append(update.args, endTime)

		case "teams":
		case "items":
		case "location":
		default:
		}
	}

	l := sqlb.Len()
	update.sqlStatement = fmt.Sprintf("%s\n\t\tWHERE id = $%d", sqlb.String()[0:l-1], inc)
	update.args = append(update.args, huntID)
	log.Printf(update.sqlStatement)
	log.Print(update.args...)

	if encounteredError {
		return update, errors.New(eb.String())
	} else {
		return update, nil
	}
}
