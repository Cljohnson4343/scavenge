package models

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

// AllHunts returns all Hunts from the database
func AllHunts() ([]*Hunt, error) {
	rows, err := db.Query("SELECT * FROM hunts;")
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

		hunts = append(hunts, hunt)
	}

	err = rows.Err()

	return hunts, err
}

// InsertHunt inserts the given hunt into the database and returns the id of the inserted hunt
func InsertHunt(hunt *Hunt) (int, error) {
	sqlStatement := `
		INSERT INTO hunts(title, max_teams, start_time, end_time, location_name, latitude, longitude)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id AS hunt_id
		`
	/*
		`
		insert_teams AS
			INSERT INTO teams(hunt_id, name)
			SELECT hunt_id, $7 FROM insert_hunt),
		INSERT INTO items(hunt_id, name, points)
		SELECT hunt_id, $8, $9 FROM insert_hunt;`
	*/

	id := 0
	err := db.QueryRow(sqlStatement, hunt.Title, hunt.MaxTeams, hunt.Start,
		hunt.End, hunt.Location.Name, hunt.Location.Coords.Latitude,
		hunt.Location.Coords.Longitude).Scan(&id)

	return id, err
}
