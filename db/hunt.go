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
	Name string `json:"name" valid:"stringlength(1|255)"`

	// The maximum number of teams that can participate in the Hunt.
	//
	// minimum: 1
	// required: true
	MaxTeams int `json:"max_teams" valid:"positive"`

	// The id of the Hunt
	//
	// required: true
	ID int `json:"id" valid:"isNil~id: the id can not be specified,optional"`

	// The start time for the Hunt
	//
	// required: true
	// swagger:strfmt date
	StartTime time.Time `json:"start_time" valid:"timeNotPast"`

	// The end time for the Hunt
	//
	// required: true
	// swagger:strfmt date
	EndTime time.Time `json:"end_time" valid:"timeNotPast,startTimeBeforeEndTime"`

	// The creation time for the Hunt
	//
	// required: true
	// swagger:strfmt date
	CreatedAt time.Time `json:"created_at" valid:"-"`

	// The name of the location of the Hunt
	//
	// required: true
	// maximum length: 80
	LocationName string `json:"location_name" valid:"stringlength(1|80)"`

	// The latitude for the Hunt
	//
	// required: true
	Latitude float32 `json:"latitude" valid:"latitude"`

	// The longitude for the Hunt
	//
	// required: true
	Longitude float32 `json:"longitude" valid:"longitude"`
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

	if z.CreatedAt.Equal(h.CreatedAt) {
		tblColMap[HuntTbl]["created_at"] = h.CreatedAt
	}

	return tblColMap
}

// Validate validates a HuntDB
func (h *HuntDB) Validate(r *http.Request) *response.Error {
	_, err := govalidator.ValidateStruct(h)
	if err != nil {
		return response.NewError(err.Error(), http.StatusBadRequest)
	}

	return nil
}

// PartialValidate only validates the non-zero value fields
func (h *HuntDB) PartialValidate(r *http.Request) *response.Error {
	tblColMap := h.GetTableColumnMap()

	return request.PartialValidate(tblColMap[HuntTbl], h)
}
