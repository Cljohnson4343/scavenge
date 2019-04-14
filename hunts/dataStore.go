package hunts

import (
	"github.com/cljohnson4343/scavenge/hunts/models"
)

// HuntDataStore is an interface that comprised of all access methods
// that pkg hunts needs to communicate with the database
type HuntDataStore interface {
	AllHunts() ([]*models.Hunt, error)
	GetHunt(hunt *models.Hunt, huntID int) error
	GetItems(items *[]models.Item, huntID int) error
	GetTeams(teams *[]models.Team, huntID int) error
	InsertHunt(hunt *models.Hunt) (int, error)
	InsertTeam(team *models.Team, huntID int) (int, error)
	InsertItem(item *models.Item, huntID int) (int, error)
	DeleteHunt(huntID int) error
}

// AllHunts returns all Hunts from the database
func (env *Env) AllHunts() ([]*models.Hunt, error) {
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

		err = env.GetTeams(&hunt.Teams, hunt.ID)
		if err != nil {
			return nil, err
		}

		err = env.GetItems(&hunt.Items, hunt.ID)
		if err != nil {
			return nil, err
		}

		hunts = append(hunts, hunt)
	}

	err = rows.Err()

	return hunts, err
}

// GetHunt returns a pointer to the hunt with the given ID.
func (env *Env) GetHunt(hunt *models.Hunt, huntID int) error {
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
	err = env.GetTeams(&hunt.Teams, huntID)
	if err != nil {
		return err
	}

	err = env.GetItems(&hunt.Items, huntID)

	return err
}

// GetItems populates the items slice with all the items for the given hunt
func (env *Env) GetItems(items *[]models.Item, huntID int) error {
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

// GetTeams populates the teams slice with all the teams for the given hunt
func (env *Env) GetTeams(teams *[]models.Team, huntID int) error {
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

// InsertHunt inserts the given hunt into the database and returns the id of the inserted hunt
func (env *Env) InsertHunt(hunt *models.Hunt) (int, error) {
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
		_, err = env.InsertTeam(&v, id)
		if err != nil {
			return id, err
		}
	}

	for _, v := range hunt.Items {
		_, err = env.InsertItem(&v, id)
		if err != nil {
			return id, err
		}
	}

	return id, err
}

// InsertTeam inserts a Team into the db
func (env *Env) InsertTeam(team *models.Team, huntID int) (int, error) {
	sqlStatement := `
		INSERT INTO teams(hunt_id, name)
		VALUES ($1, $2)
		RETURNING id`

	id := 0
	err := env.db.QueryRow(sqlStatement, huntID, team.Name).Scan(&id)

	return id, err
}

// InsertItem inserts an Item into the db
func (env *Env) InsertItem(item *models.Item, huntID int) (int, error) {
	sqlStatement := `
		INSERT INTO items(hunt_id, name, points)
		VALUES ($1, $2, $3)
		RETURNING id`

	id := 0
	err := env.db.QueryRow(sqlStatement, huntID, item.Name, item.Points).Scan(&id)

	return id, err
}

// DeleteHunt deletes the hunt with the given ID. All associated data will also be deleted.
func (env *Env) DeleteHunt(huntID int) error {
	sqlStatement := `
		DELETE FROM hunts
		WHERE hunts.id = $1`

	_, err := env.db.Exec(sqlStatement, huntID)
	return err
}
