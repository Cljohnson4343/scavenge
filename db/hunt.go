package db

import (
	"net/http"
	"time"

	"github.com/asaskevich/govalidator"
	"github.com/cljohnson4343/scavenge/response"
)

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

// Validate validates a HuntDB
func (h *HuntDB) Validate(r *http.Request) *response.Error {
	_, err := govalidator.ValidateStruct(h)
	if err != nil {
		return response.NewError(err.Error(), http.StatusBadRequest)
	}

	return nil
}
