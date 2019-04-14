package hunts

import (
	"time"
)

// Coord is a representation of a Coord
type Coord struct {
	Latitude  float32 `json:"latitude"`
	Longitude float32 `json:"longitude"`
}

// Location is a representation of a Location
type Location struct {
	Name   string `json:"name"`
	Coords Coord  `json:"coords"`
}

// User is a representation of a User
type User struct {
	First string `json:"first"`
	Last  string `json:"last"`
}

// Team is a representation of a Team
type Team struct {
	Name string `json:"name"`
}

// Item is a representation of a Item
type Item struct {
	Name   string `json:"name"`
	Points uint   `json:"points"`
	IsDone bool   `json:"is_done"`
}

// Hunt is a representation of a Hunt
type Hunt struct {
	Title    string    `json:"title"`
	MaxTeams int       `json:"max_teams"`
	ID       int       `json:"id"`
	Start    time.Time `json:"start"`
	End      time.Time `json:"end"`
	Teams    []Team    `json:"teams"`
	Items    []Item    `json:"items"`
	Location Location  `json:"location"`
}

type HuntDataStore interface {
	AllHunts() ([]*Hunt, error)
	GetHunt(hunt *Hunt, huntID int) error
	GetItems(items *[]Item, huntID int) error
	GetTeams(teams *[]Team, huntID int) error
	InsertHunt(hunt *Hunt) (int, error)
	InsertTeam(team *Team, huntID int) (int, error)
	InsertItem(item *Item, huntID int) (int, error)
	DeleteHunt(huntID int) error
}

// AllHunts returns all Hunts from the database
func (env *Env) AllHunts() ([]*Hunt, error) {
	rows, err := env.db.Query("SELECT * FROM hunts;")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	hunts := make([]*Hunt, 0)
	for rows.Next() {
		hunt := new(Hunt)
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
func (env *Env) GetHunt(hunt *Hunt, huntID int) error {
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
func (env *Env) GetItems(items *[]Item, huntID int) error {
	sqlStatement := `
		SELECT name, points FROM items WHERE items.hunt_id = $1;`

	rows, err := env.db.Query(sqlStatement, huntID)
	if err != nil {
		return err
	}
	defer rows.Close()

	item := Item{}
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
func (env *Env) GetTeams(teams *[]Team, huntID int) error {
	sqlStatement := `
		SELECT name FROM teams WHERE teams.hunt_id = $1;`

	rows, err := env.db.Query(sqlStatement, huntID)
	if err != nil {
		return err
	}
	defer rows.Close()

	team := Team{}
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
func (env *Env) InsertHunt(hunt *Hunt) (int, error) {
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
func (env *Env) InsertTeam(team *Team, huntID int) (int, error) {
	sqlStatement := `
		INSERT INTO teams(hunt_id, name)
		VALUES ($1, $2)
		RETURNING id`

	id := 0
	err := env.db.QueryRow(sqlStatement, huntID, team.Name).Scan(&id)

	return id, err
}

// InsertItem inserts an Item into the db
func (env *Env) InsertItem(item *Item, huntID int) (int, error) {
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
