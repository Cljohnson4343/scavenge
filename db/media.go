package db

import (
	"fmt"
	"net/http"

	"github.com/asaskevich/govalidator"
	"github.com/cljohnson4343/scavenge/response"
)

// MediaMetaDB is a representation of a row in the media table
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

	// The id of the location associated with the media file described by this object
	//
	// required: true
	LocationID int `json:"location_id" valid:"int"`

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

	return nil
}

var mediaMetasForTeamScript = `
	SELECT id, team_id, item_id, location_id, url
	FROM media
	WHERE team_id = $1;`

// GetMediaMetasForTeam returns all the meta information for all media files associated w/
// this team. A result with both media meta objects and an error is possible
func GetMediaMetasForTeam(teamID int) ([]*MediaMetaDB, *response.Error) {
	rows, err := stmtMap["mediaMetasForTeam"].Query(teamID)
	if err != nil {
		return nil, response.NewError(fmt.Sprintf("error getting all media meta info for team %d: %v",
			teamID, err), http.StatusInternalServerError)
	}
	rows.Close()

	e := response.NewNilError()
	metas := make([]*MediaMetaDB, 0)

	for rows.Next() {
		meta := MediaMetaDB{}

		err = rows.Scan(&meta.ID, &meta.TeamID, &meta.ItemID, &meta.LocationID, &meta.URL)
		if err != nil {
			e.Add(fmt.Sprintf("error getting media meta info for team %d: %v", teamID,
				err), http.StatusInternalServerError)
			break
		}
		metas = append(metas, &meta)
	}

	if err = rows.Err(); err != nil {
		e.Add(fmt.Sprintf("error getting media meta info for team %d: %v", teamID,
			err), http.StatusInternalServerError)
	}

	return metas, e.GetError()
}

var mediaMetaInsertScript = ``
var mediaMetaDeleteScript = ``
