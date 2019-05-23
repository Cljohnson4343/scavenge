package db

import (
	"net/http"
	"time"

	"github.com/cljohnson4343/scavenge/request"

	"github.com/cljohnson4343/scavenge/pgsql"

	"github.com/asaskevich/govalidator"
	"github.com/cljohnson4343/scavenge/response"
)

// HuntTbl is the name of the hunts db table
const HuntTbl string = "hunts"

// A HuntDB is the representation of a row from the hunts table
//
// swagger:model Hunt
type HuntDB struct {

	// The name of the Hunt
	//
	// required: true
	// maximum length: 255
	Name string `json:"huntName" valid:"stringlength(1|255)"`

	// The maximum number of teams that can participate in the Hunt.
	//
	// minimum: 1
	// required: true
	MaxTeams int `json:"maxTeams" valid:"positive"`

	// The id of the Hunt
	//
	// required: true
	ID int `json:"huntID" valid:"int,optional"`

	// The start time for the Hunt
	//
	// required: true
	// swagger:strfmt date
	StartTime time.Time `json:"startTime" valid:"timeNotPast"`

	// The end time for the Hunt
	//
	// required: true
	// swagger:strfmt date
	EndTime time.Time `json:"endTime" valid:"timeNotPast,startTimeBeforeEndTime"`

	// The creation time for the Hunt
	//
	// required: true
	// swagger:strfmt date
	CreatedAt time.Time `json:"createdAt" valid:"-"`

	// The creator of the hunt
	//
	// required: true
	CreatorID int `json:"creatorID" valid:"-"`

	// The name of the location of the Hunt
	//
	// required: true
	// maximum length: 80
	LocationName string `json:"locationName" valid:"stringlength(1|80),optional"`

	// The latitude for the Hunt
	//
	// required: true
	Latitude float32 `json:"latitude" valid:"latitude,optional"`

	// The longitude for the Hunt
	//
	// required: true
	Longitude float32 `json:"longitude" valid:"longitude,optional"`
}

// Update updates the non-zero value fields in the HuntDB struct
func (h *HuntDB) Update(ex pgsql.Executioner) *response.Error {
	return update(h, ex, h.ID)
}

// GetTableColumnMap maps all non-zero value fields of a HuntDB to the
// associated table, column name, and value
func (h *HuntDB) GetTableColumnMap() pgsql.TableColumnMap {
	tblColMap := make(pgsql.TableColumnMap)
	tblColMap[HuntTbl] = make(pgsql.ColumnMap)

	// get zero value HuntDB
	z := HuntDB{}

	if z.ID != h.ID {
		tblColMap[HuntTbl]["id"] = h.ID
	}

	if z.Name != h.Name {
		tblColMap[HuntTbl]["name"] = h.Name
	}

	if z.MaxTeams != h.MaxTeams {
		tblColMap[HuntTbl]["max_teams"] = h.MaxTeams
	}

	if !h.StartTime.IsZero() {
		tblColMap[HuntTbl]["start_time"] = h.StartTime
	}

	if !h.EndTime.IsZero() {
		tblColMap[HuntTbl]["end_time"] = h.EndTime
	}

	// because we are comparing whether or not h is the zero
	// value we can use regular comparison for float value
	if z.Latitude != h.Latitude {
		tblColMap[HuntTbl]["latitude"] = h.Latitude
	}

	// because we are comparing whether or not h is the zero
	// value we can use regular comparison for float value
	if z.Longitude != h.Longitude {
		tblColMap[HuntTbl]["longitude"] = h.Longitude
	}

	if z.LocationName != h.LocationName {
		tblColMap[HuntTbl]["location_name"] = h.LocationName
	}

	if !h.CreatedAt.IsZero() {
		tblColMap[HuntTbl]["created_at"] = h.CreatedAt
	}

	if z.CreatorID != h.CreatorID {
		tblColMap[HuntTbl]["creator_id"] = h.CreatorID
	}

	return tblColMap
}

// Validate validates a HuntDB
func (h *HuntDB) Validate(r *http.Request) *response.Error {
	_, err := govalidator.ValidateStruct(h)
	if err != nil {
		return response.NewError(http.StatusBadRequest, err.Error())
	}

	return nil
}

// PatchValidate only validates the non-zero value fields
func (h *HuntDB) PatchValidate(r *http.Request, huntID int) *response.Error {
	tblColMap := h.GetTableColumnMap()
	e := response.NewNilError()

	// patching a hunt requires an id that matches the given huntID,
	// if no id is provided then we can just add one
	id, ok := tblColMap[HuntTbl]["id"]
	if !ok {
		h.ID = huntID
		tblColMap[HuntTbl]["id"] = huntID
	}

	// if an id is provided that doesn't match then we alert the user
	// of a bad request
	if id != huntID {
		e.Add(http.StatusBadRequest, "id: the correct hunt id must be provided")
		// delete the id col name so no new errors will accumulate for this column name
		delete(tblColMap[HuntTbl], "id")
	}

	// changing a hunt's created_at field is not supported
	if _, ok = tblColMap[HuntTbl]["created_at"]; ok {
		e.Add(http.StatusBadRequest, "created_at: changing a hunt's created_at field is not supported with PATCH")
		delete(tblColMap[HuntTbl], "created_at")
	}

	// changing a hunt's creator_id field is not supported
	if _, ok = tblColMap[HuntTbl]["creator_id"]; ok {
		e.Add(http.StatusBadRequest, "creator_id: changing a hunt's creator_id field is not supported with PATCH")
		delete(tblColMap[HuntTbl], "creator_id")
	}

	patchErr := request.PatchValidate(tblColMap[HuntTbl], h)
	if patchErr != nil {
		e.AddError(patchErr)
	}

	return e.GetError()
}

var huntsSelectScript = `
	SELECT 
		name, 
		id, 
		start_time, 
		end_time, 
		location_name, 
		latitude, 
		longitude, 
		max_teams, 
		created_at,
		creator_id
	FROM hunts;`

// GetHunts returns all the huntDBs in the db. NOTE that it is possible to have returned hunts and
// an error, check both
func GetHunts() ([]*HuntDB, *response.Error) {
	rows, err := stmtMap["huntsSelect"].Query()
	if err != nil {
		return nil, response.NewErrorf(http.StatusInternalServerError, "error getting hunts: %s", err.Error())
	}
	defer rows.Close()

	hunts := make([]*HuntDB, 0)
	e := response.NewNilError()
	for rows.Next() {
		hunt := HuntDB{}
		huntErr := rows.Scan(&hunt.Name, &hunt.ID, &hunt.StartTime, &hunt.EndTime, &hunt.LocationName,
			&hunt.Latitude, &hunt.Longitude, &hunt.MaxTeams, &hunt.CreatedAt, &hunt.CreatorID)
		if huntErr != nil {
			e.Addf(http.StatusInternalServerError, "error getting hunt: %s", huntErr.Error())
			break
		}
		hunts = append(hunts, &hunt)
	}

	err = rows.Err()
	if err != nil {
		e.Addf(http.StatusInternalServerError, "error getting hunt: %s", err.Error())
	}

	return hunts, e.GetError()
}

var huntsByUserIDSelectScript = `
	WITH teams_for_user AS (
		SELECT t.hunt_id 
		FROM teams t 
		INNER JOIN users_teams ut 
		ON ut.user_id = $1 AND t.id = ut.team_id
	)
	SELECT 
		name, 
		id, 
		start_time, 
		end_time, 
		location_name, 
		latitude, 
		longitude, 
		max_teams, 
		created_at,
		creator_id
	FROM teams_for_user tfs 
	INNER JOIN hunts h 
	ON tfs.hunt_id = h.id;
	`

// GetHuntsByUserID returns all the huntDBs in the db. NOTE that it is possible to have returned hunts and
// an error, check both
func GetHuntsByUserID(userID int) ([]*HuntDB, *response.Error) {
	rows, err := stmtMap["huntsByUserIDSelect"].Query(userID)
	if err != nil {
		return nil, response.NewErrorf(http.StatusInternalServerError, "error getting hunts for user: %s", err.Error())
	}
	defer rows.Close()

	hunts := make([]*HuntDB, 0)
	e := response.NewNilError()
	for rows.Next() {
		hunt := HuntDB{}
		huntErr := rows.Scan(&hunt.Name, &hunt.ID, &hunt.StartTime, &hunt.EndTime, &hunt.LocationName,
			&hunt.Latitude, &hunt.Longitude, &hunt.MaxTeams, &hunt.CreatedAt, &hunt.CreatorID)
		if huntErr != nil {
			e.Addf(http.StatusInternalServerError, "error getting hunts for user: %s", huntErr.Error())
			break
		}
		hunts = append(hunts, &hunt)
	}

	err = rows.Err()
	if err != nil {
		e.Addf(http.StatusInternalServerError, "error getting hunts for user: %s", err.Error())
	}

	return hunts, e.GetError()
}

var huntSelectScript = `
	SELECT 
		name, 
		id, 
		start_time, 
		end_time, 
		location_name, 
		latitude, 
		longitude, 
		max_teams, 
		created_at,
		creator_id
	FROM hunts
	WHERE id = $1;`

// GetHunt returns the huntDB with the given id
func GetHunt(huntID int) (*HuntDB, *response.Error) {
	h := HuntDB{}

	err := stmtMap["huntSelect"].QueryRow(huntID).Scan(
		&h.Name,
		&h.ID,
		&h.StartTime,
		&h.EndTime,
		&h.LocationName,
		&h.Latitude,
		&h.Longitude,
		&h.MaxTeams,
		&h.CreatedAt,
		&h.CreatorID,
	)
	if err != nil {
		return nil, response.NewErrorf(
			http.StatusInternalServerError,
			"error getting the hunt with id %d: %s",
			huntID,
			err.Error(),
		)
	}

	return &h, nil
}

var huntDeleteScript = `
	DELETE FROM hunts
	WHERE id = $1;`

// DeleteHunt deletes the huntdb with the given id
func DeleteHunt(id int) *response.Error {
	res, err := stmtMap["huntDelete"].Exec(id)
	if err != nil {
		return response.NewErrorf(
			http.StatusInternalServerError,
			"error deleting hunt with id %d: %s",
			id,
			err.Error(),
		)
	}

	numRows, err := res.RowsAffected()
	if err != nil {
		return response.NewErrorf(
			http.StatusInternalServerError,
			"error deleting hunt with id %d: %s",
			id,
			err.Error(),
		)
	}

	if numRows < 1 {
		return response.NewErrorf(
			http.StatusInternalServerError,
			"error deleting hunt with id %d: %s",
			id,
			err.Error(),
		)
	}

	return nil
}

var huntInsertScript = `
	INSERT INTO hunts(
		name, 
		max_teams, 
		start_time, 
		end_time, 
		location_name, 
		latitude, 
		longitude,
		creator_id
	)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	RETURNING id, created_at;
	`

// Insert inserts the given huntDB into the db and returns, by writing to the huntDB the id and
// create_at timestamp
func (h *HuntDB) Insert() *response.Error {
	err := stmtMap["huntInsert"].QueryRow(h.Name, h.MaxTeams, h.StartTime, h.EndTime,
		h.LocationName, h.Latitude, h.Longitude, h.CreatorID).Scan(&h.ID, &h.CreatedAt)
	if err != nil {
		return response.NewErrorf(http.StatusInternalServerError, "error inserting hunt with name %s: %s", h.Name, err.Error())

	}

	return nil
}
