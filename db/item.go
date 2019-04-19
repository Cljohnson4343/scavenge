package db

import (
	"net/http"
	"strconv"

	"github.com/asaskevich/govalidator"
	"github.com/cljohnson4343/scavenge/response"
	"github.com/go-chi/chi"
)

// ItemDB is the data representation of a row from items
//
// swagger:model item
type ItemDB struct {

	// The id of the Hunt
	//
	// required: true
	HuntID int `json:"hunt_id" valid:"-"`

	// The id of the item
	//
	// required: true
	ID int `json:"id" valid:"isNil~id: can not be specified,optional"`

	// the name of the item
	//
	// maximum length: 255
	// required: true
	Name string `json:"name" valid:"stringlength(1|255)"`

	// the amount of points this item is worth
	//
	// minimum: 1
	// default: 1
	Points int `json:"points,omitempty" valid:"positive,optional"`
}

// Validate validates a ItemDB struct
func (i *ItemDB) Validate(r *http.Request) *response.Error {
	// it is possible to get here without the huntID parameter being specified so don't catch
	// the Atoi error
	huntID, _ := strconv.Atoi(chi.URLParam(r, "huntID"))

	e := response.NewNilError()

	// make sure HuntID is either 0, client prob didn't specify it, or == to URL hunt_id
	if huntID != i.HuntID && i.HuntID != 0 {
		e.Add("hunt_id: field must either be the same as the URL huntID or not specified", http.StatusBadRequest)
	}

	// set points to default if they weren't specified
	if i.Points == 0 {
		i.Points = 1
	}

	_, structErr := govalidator.ValidateStruct(i)
	if structErr != nil {
		e.Add(structErr.Error(), http.StatusBadRequest)
	}

	return e.GetError()
}
