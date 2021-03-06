package db

import (
	"net/http"

	"github.com/asaskevich/govalidator"
	"github.com/cljohnson4343/scavenge/response"
	"github.com/lib/pq"
)

// MediaMetaDB is info  associated with a media file
type MediaMetaDB struct {

	// The id of the media row
	//
	// required: false
	ID int `json:"mediaID" valid:"int,optional"`

	// The id of the item, if available, for the media file described by this object
	ItemID int `json:"itemID" valid:"int,optional"`

	// The id of the team associated with the media file described by this object
	//
	// required: true
	TeamID int `json:"teamID" valid:"int"`

	// The location associated with the media file described by this object
	//
	// required: true
	Location LocationDB `json:"location" valid:"-"`

	// The url where the media file can be retrieved
	//
	// required: true
	// maximum length: 2083
	// minimum length: 3
	URL string `json:"url" valid:"url"`
}

// Validate validates the struct
func (m *MediaMetaDB) Validate(r *http.Request) *response.Error {
	_, err := govalidator.ValidateStruct(m)
	if err != nil {
		return response.NewErrorf(http.StatusBadRequest, "error validating media meta info: %v", err)
	}

	_, err = govalidator.ValidateStruct(m.Location)
	if err != nil {
		return response.NewErrorf(http.StatusBadRequest, "error validating media location info: %v", err)
	}

	return nil
}

var mediaMetasForTeamScript = `
	WITH loc_for_team AS (
		SELECT l.latitude, l.longitude, l.time_stamp, l.team_id AS loc_team_id, l.id
		FROM locations l
		WHERE l.team_id = $1
	), media_for_team AS (
		SELECT m.id, m.team_id, COALESCE(m.item_id, 0), m.url, m.location_id
		FROM media m
		WHERE m.team_id = $1
	)
	SELECT * 
	FROM media_for_team m INNER JOIN loc_for_team l
	ON m.location_id = l.id;
	`

// GetMediaMetasForTeam returns all the meta information for all media files associated w/
// this team. A result with both media meta objects and an error is possible
func GetMediaMetasForTeam(teamID int) ([]*MediaMetaDB, *response.Error) {
	rows, err := stmtMap["mediaMetasForTeam"].Query(teamID)
	if err != nil {
		return nil, response.NewErrorf(
			http.StatusInternalServerError,
			"error getting all media meta info for team %d: %v",
			teamID,
			err,
		)
	}
	defer rows.Close()

	e := response.NewNilError()
	metas := make([]*MediaMetaDB, 0)

	for rows.Next() {
		m := MediaMetaDB{}

		err = rows.Scan(
			&m.ID,
			&m.TeamID,
			&m.ItemID,
			&m.URL,
			&m.Location.ID,
			&m.Location.Latitude,
			&m.Location.Longitude,
			&m.Location.TimeStamp,
			&m.Location.TeamID,
			&m.Location.ID,
		)
		if err != nil {
			e.Addf(
				http.StatusInternalServerError,
				"error getting media meta info for team %d: %v",
				teamID,
				err,
			)
			break
		}
		metas = append(metas, &m)
	}

	if err = rows.Err(); err != nil {
		e.Addf(
			http.StatusInternalServerError,
			"error getting media meta info for team %d: %v",
			teamID,
			err,
		)
	}

	return metas, e.GetError()
}

var mediaMetaInsertScript = `
	WITH loc_ins AS (
		INSERT INTO locations(team_id, latitude, longitude, time_stamp)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT ON CONSTRAINT team_same_loc_and_time DO NOTHING
		RETURNING id locations_id
	), loc_get AS (
		SELECT id locations_id
		FROM locations
		WHERE team_id = $1 AND time_stamp = $4
	), loc AS (
		SELECT locations_id FROM loc_get
		UNION ALL 
		SELECT locations_id FROM loc_ins
	)
	INSERT INTO media(team_id, item_id, location_id, url)
	VALUES ($5, NULLIF($6, 0), (SELECT locations_id FROM loc), $7)
	RETURNING location_id, id media_id;
	`

// Insert inserts the given data into the db. The id of the locations row
// and the id of the media row are written back to the MediaMetaDB struct
func (m *MediaMetaDB) Insert(teamID int) *response.Error {
	// make sure the given teamID matches the teamID's for the structs
	if teamID != m.TeamID || teamID != m.Location.TeamID {
		return response.NewError(
			http.StatusBadRequest,
			"invalid insert request: the teamID's provided don't match",
		)
	}

	err := stmtMap["mediaMetaInsert"].QueryRow(m.TeamID, m.Location.Latitude,
		m.Location.Longitude, m.Location.TimeStamp, m.TeamID, m.ItemID,
		m.URL).Scan(&m.Location.ID, &m.ID)
	if err != nil {
		return m.ParseError(err, "insert")
	}

	return nil
}

var mediaMetaDeleteScript = `
	DELETE FROM media
	WHERE id = $1 AND team_id = $2
	RETURNING url; 
	`

// DeleteMedia deletes the row from the media table but leaves the location data in
// the location table. If you want to delete both then delete the associated Location row.
func DeleteMedia(mediaID, teamID int) (string, *response.Error) {
	var url string
	err := stmtMap["mediaMetaDelete"].QueryRow(mediaID, teamID).Scan(&url)
	if err != nil {
		return url, response.NewErrorf(
			http.StatusInternalServerError,
			"error deleting media with id %d: %v",
			mediaID,
			err,
		)
	}

	return url, nil
}

var teamPointsScript = `
	WITH media_for_team AS (
		SELECT item_id
		FROM media
		WHERE media.team_id = $1 AND media.item_id IS NOT NULL
	)
		SELECT COALESCE(SUM(i.points), 0)
		FROM media_for_team m
		INNER JOIN items i ON m.item_id = i.id; 
	`

// GetTeamPoints returns the integer number of points the team with the given
// id has accumulated thus far
func GetTeamPoints(teamID int) (int, *response.Error) {
	var pts int
	err := stmtMap["teamPoints"].QueryRow(teamID).Scan(&pts)
	if err != nil {
		return 0, response.NewErrorf(http.StatusInternalServerError, "error getting pts for team %d: %v", teamID, err)
	}

	return pts, nil
}

// ParseError maps a pq driver error to a response.Error
func (m *MediaMetaDB) ParseError(err error, op string) *response.Error {
	pqErr, ok := err.(*pq.Error)
	if ok {
		if pqErr.Constraint != "" {
			switch pqErr.Constraint {
			case "media_item_id_fkey":
				return response.NewErrorf(
					http.StatusBadRequest,
					"item_id: item %d does not exist",
					m.ItemID,
				)
			case "media_location_id_fkey":
				return response.NewErrorf(
					http.StatusBadRequest,
					"location.id: location %d does not exist",
					m.Location.ID,
				)
			case "media_team_id_fkey":
				return response.NewErrorf(
					http.StatusBadRequest,
					"team_id: team %d does not exist",
					m.TeamID,
				)
			}
		}
	}

	return m.Location.ParseError(err, op)
}
