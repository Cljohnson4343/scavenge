package db

import (
	"fmt"
	"net/http"

	"github.com/asaskevich/govalidator"
	"github.com/cljohnson4343/scavenge/response"
)

// MediaMetaDB is info  associated with a media file
type MediaMetaDB struct {

	// The id of the media row
	//
	// required: true
	ID int `json:"id" valid:"int,optional"`

	// The id of the item, if available, for the media file described by this object
	//
	// required: true
	ItemID int `json:"item_id" valid:"int,optional"`

	// The id of the team associated with the media file described by this object
	//
	// required: true
	TeamID int `json:"team_id" valid:"int"`

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
		return response.NewError(fmt.Sprintf("error validating media meta info: %v", err),
			http.StatusBadRequest)
	}

	_, err = govalidator.ValidateStruct(m.Location)
	if err != nil {
		return response.NewError(fmt.Sprintf("error validating media location info: %v", err),
			http.StatusBadRequest)
	}

	return nil
}

var mediaMetasForTeamScript = `
	WITH loc_and_media AS (
		SELECT m.id, m.team_id, COALESCE(m.item_id, 0), m.url, m.location_id, l.latitude, l.longitude, l.time_stamp, l.team_id AS loc_team
		FROM media m 
		INNER JOIN locations l
		ON m.location_id = l.id
	)
	SELECT * 
	FROM loc_and_media
	WHERE team_id = $1;`

// GetMediaMetasForTeam returns all the meta information for all media files associated w/
// this team. A result with both media meta objects and an error is possible
func GetMediaMetasForTeam(teamID int) ([]*MediaMetaDB, *response.Error) {
	rows, err := stmtMap["mediaMetasForTeam"].Query(teamID)
	if err != nil {
		return nil, response.NewError(fmt.Sprintf("error getting all media meta info for team %d: %v",
			teamID, err), http.StatusInternalServerError)
	}
	defer rows.Close()

	e := response.NewNilError()
	metas := make([]*MediaMetaDB, 0)

	for rows.Next() {
		m := MediaMetaDB{}

		err = rows.Scan(&m.ID, &m.TeamID, &m.ItemID, &m.URL, &m.Location.ID, &m.Location.Latitude,
			&m.Location.Longitude, &m.Location.TimeStamp, &m.Location.TeamID)
		if err != nil {
			e.Add(fmt.Sprintf("error getting media meta info for team %d: %v", teamID,
				err), http.StatusInternalServerError)
			break
		}
		metas = append(metas, &m)
	}

	if err = rows.Err(); err != nil {
		e.Add(fmt.Sprintf("error getting media meta info for team %d: %v", teamID,
			err), http.StatusInternalServerError)
	}

	return metas, e.GetError()
}

var mediaMetaInsertScript = `
	WITH loc_ins AS (
		INSERT INTO locations(team_id, latitude, longitude, time_stamp)
		VALUES ($1, $2, $3, $4)
		RETURNING id locations_id
	)
	INSERT INTO media(team_id, item_id, location_id, url)
	VALUES ($5, NULLIF($6, 0), (SELECT locations_id FROM loc_ins), $7)
	RETURNING location_id, id media_id;
	`

// Insert inserts the given data into the db. The id of the locations row
// and the id of the media row are written back to the MediaMetaDB struct
func (m *MediaMetaDB) Insert(teamID int) *response.Error {
	// make sure the given teamID matches the teamID's for the structs
	if teamID != m.TeamID || teamID != m.Location.TeamID {
		return response.NewError("invalid insert request: the teamID's provided don't match",
			http.StatusBadRequest)
	}

	err := stmtMap["mediaMetaInsert"].QueryRow(m.TeamID, m.Location.Latitude,
		m.Location.Longitude, m.Location.TimeStamp, m.TeamID, m.ItemID,
		m.URL).Scan(&m.Location.ID, &m.ID)
	if err != nil {
		return response.NewError(
			fmt.Sprintf("error inserting media meta info with team %d: %v",
				teamID, err), http.StatusInternalServerError)

	}

	return nil
}

var mediaMetaDeleteScript = `
	DELETE FROM media
	WHERE id = $1 AND team_id = $2;`

// DeleteMedia deletes the row from the media table but leaves the location data in
// the location table. If you want to delete both then delete the associated Location row.
func DeleteMedia(mediaID, teamID int) *response.Error {
	res, err := stmtMap["mediaMetaDelete"].Exec(mediaID, teamID)
	if err != nil {
		return response.NewError(fmt.Sprintf("error deleting media with id %d: %v",
			mediaID, err), http.StatusInternalServerError)
	}

	numRows, err := res.RowsAffected()
	if err != nil {
		return response.NewError(fmt.Sprintf("error deleting media with id %d: %v",
			mediaID, err), http.StatusInternalServerError)
	}

	if numRows < 1 {
		return response.NewError(fmt.Sprintf("error deleting media with id %d and team id %d",
			mediaID, teamID), http.StatusBadRequest)
	}

	return nil
}

var teamPointsScript = `
	WITH media_for_team AS (
		SELECT media.item_id
		FROM media
		WHERE media.team_id = $1
	)
		SELECT SUM(i.points) AS total_points
		FROM media_for_team m
		INNER JOIN items i ON m.item_id = i.id; 
	`

// GetTeamPoints returns the integer number of points the team with the given
// id has accumulated thus far
func GetTeamPoints(teamID int) (int, *response.Error) {
	var pts int
	err := stmtMap["teamPoints"].QueryRow(teamID).Scan(&pts)
	if err != nil {
		return 0, response.NewError(fmt.Sprintf("error getting pts for team %d: %v",
			teamID, err), http.StatusInternalServerError)
	}

	return pts, nil
}
