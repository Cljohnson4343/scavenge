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

	return nil
}

var mediaMetasForTeamScript = `
	SELECT m.id, m.team_id, m.item_id, m.url, l.latitude, l.longitude, l.time_stamp
	FROM media as m 
		INNER JOIN locations as l
		ON m.team_id = l.team_id
	WHERE m.team_id = $1;`

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
		m := MediaMetaDB{}

		err = rows.Scan(&m.ID, &m.TeamID, &m.ItemID, &m.URL, &m.Location.Latitude,
			&m.Location.Longitude, &m.Location.TimeStamp)
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

var mediaMetaInsertScript = ``
var mediaMetaDeleteScript = ``
