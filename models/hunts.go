package models

import (
	"log"
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
	rows, err := db.Query("")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	cols, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	log.Print(cols)

	hunts := make([]*Hunt, 0)
	/*
		for rows.Next() {

		}
	*/

	return hunts, nil

}
