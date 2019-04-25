package db

import (
	"net/http"
	"time"

	"github.com/asaskevich/govalidator"
	"github.com/cljohnson4343/scavenge/response"
)

// A LocationDB is a representation of a row in the locations table
//
// swagger:model LocationDB
type LocationDB struct {

	// The id of the team
	//
	// required: true
	TeamID int `json:"team_id" valid:"int"`

	// The id of the location
	//
	// required: true
	ID int `json:"id" valid:"int,optional"`

	// the latitude
	//
	// required: true
	Latitude float32 `json:"latitude" valid:"latitude"`

	// the longitude
	//
	// required: true
	Longitude float32 `json:"longitude" valid:"longitude"`

	// the time stamp for this location
	//
	// required: true
	// swagger:strfmt date
	TimeStamp time.Time `json:"time_stamp" valid:"timePast"`
}

// Validate validates a locationDB struct
func (l *LocationDB) Validate(r *http.Request) *response.Error {
	_, err := govalidator.ValidateStruct(l)
	if err != nil {
		return response.NewErrorf(http.StatusBadRequest, "error validating location: %s", err.Error())

	}

	return nil
}

var locationsForTeamScript = `
	SELECT team_id, id, latitude, longitude, time_stamp
	FROM locations
	WHERE team_id = $1;`

// GetLocationsForTeam returns all the locationDBs for the team with the given id
// It is possible to return both results and an error
func GetLocationsForTeam(teamID int) ([]*LocationDB, *response.Error) {
	rows, err := stmtMap["locationsForTeam"].Query(teamID)
	if err != nil {
		return nil, response.NewErrorf(http.StatusInternalServerError, "error getting locations for team %d: %s", teamID, err.Error())

	}
	defer rows.Close()

	e := response.NewNilError()
	locs := make([]*LocationDB, 0)
	for rows.Next() {
		l := LocationDB{}
		err := rows.Scan(&l.TeamID, &l.ID, &l.Latitude, &l.Longitude, &l.TimeStamp)
		if err != nil {
			e.Addf(http.StatusInternalServerError, "error getting locations for team %d: %s", teamID, err.Error())
			break
		}

		locs = append(locs, &l)
	}

	if err = rows.Err(); err != nil {
		e.Addf(http.StatusInternalServerError, "error getting locations for team %d: %s", teamID, err.Error())
	}

	return locs, e.GetError()
}

var locationInsertScript = `
	INSERT INTO locations(team_id, latitude, longitude, time_stamp)
	VALUES ($1, $2, $3, $4)
	RETURNING id;`

// Insert inserts the locationDB into the locations table and writes the
// id into the locationDB struct
func (l *LocationDB) Insert(teamID int) *response.Error {
	// make sure that the location's teamID and teamID match
	if l.TeamID != teamID {
		return response.NewErrorf(http.StatusBadRequest, "team_id: team_id must match the URL team id %d", teamID)

	}

	err := stmtMap["locationInsert"].QueryRow(l.TeamID, l.Latitude, l.Longitude, l.TimeStamp).Scan(&l.ID)
	if err != nil {
		return response.NewErrorf(http.StatusInternalServerError, "error inserting location: %s", err.Error())

	}

	return nil
}

var locationDeleteScript = `
	DELETE FROM locations
	WHERE id = $1 AND team_id = $2;`

// DeleteLocation deletes the locationDB with the given id AND teamID
func DeleteLocation(id int, teamID int) *response.Error {
	res, err := stmtMap["locationDelete"].Exec(id, teamID)
	if err != nil {
		return response.NewErrorf(http.StatusBadRequest, "error deleting location with teamID %d and id %d: %s", teamID, id, err.Error())

	}

	numRows, err := res.RowsAffected()
	if err != nil {
		return response.NewErrorf(http.StatusBadRequest, "error deleting location with teamID %d and id %d: %s", teamID, id, err.Error())

	}

	if numRows < 1 {
		return response.NewErrorf(http.StatusBadRequest, "error deleting location with teamID %d and id %d", teamID, id)

	}

	return nil
}
