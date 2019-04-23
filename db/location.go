package db

import (
	"fmt"
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
		return response.NewError(fmt.Sprintf("error validating location: %s", err.Error()),
			http.StatusBadRequest)
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
		return nil, response.NewError(fmt.Sprintf("error getting locations for team %d: %s", teamID,
			err.Error()), http.StatusInternalServerError)
	}
	defer rows.Close()

	e := response.NewNilError()
	locs := make([]*LocationDB, 0)
	for rows.Next() {
		l := LocationDB{}
		err := rows.Scan(&l.TeamID, &l.ID, &l.Latitude, &l.Longitude, &l.TimeStamp)
		if err != nil {
			e.Add(fmt.Sprintf("error getting locations for team %d: %s", teamID,
				err.Error()), http.StatusInternalServerError)
			break
		}

		locs = append(locs, &l)
	}

	if err = rows.Err(); err != nil {
		e.Add(fmt.Sprintf("error getting locations for team %d: %s", teamID,
			err.Error()), http.StatusInternalServerError)
	}

	return locs, e.GetError()
}
