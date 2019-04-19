package db

import (
	"net/http"
	"strconv"
	"time"

	"github.com/asaskevich/govalidator"
	"github.com/cljohnson4343/scavenge/response"
)

func init() {
	govalidator.SetFieldsRequiredByDefault(true)
	govalidator.TagMap["positive"] = govalidator.Validator(func(str string) bool {
		v, err := strconv.Atoi(str)
		if err != nil {
			return false
		}

		return v > 0
	})
	govalidator.TagMap["isNil"] = govalidator.Validator(func(str string) bool {
		return false
	})
	govalidator.CustomTypeTagMap.Set("startTimeBeforeEndTime", govalidator.CustomTypeValidator(func(i interface{}, context interface{}) bool {
		switch v := context.(type) {
		case HuntDB:
			return v.StartTime.Before(v.EndTime)
		}

		return false
	}))
	govalidator.CustomTypeTagMap.Set("timeNotPast", govalidator.CustomTypeValidator(func(i interface{}, context interface{}) bool {
		switch v := i.(type) {
		case time.Time:
			return v.After(time.Now())
		}
		return false
	}))
}

// A TeamDB is a representation of a row in the teams table
//
// swagger:model TeamDB
type TeamDB struct {

	// The id of the Hunt
	//
	// required: true
	HuntID int `json:"hunt_id" valid:"int"`

	// The id of the team
	//
	// required: true
	ID int `json:"id" valid:"isNil~id: field can not be specified,optional"`

	// the name of the team
	//
	// maximum length: 255
	// required: true
	Name string `json:"name" valid:"stringlength(1|255)"`
}

// Validate validates a TeamDB struct
func (t *TeamDB) Validate(r *http.Request) *response.Error {
	_, err := govalidator.ValidateStruct(t)
	if err != nil {
		return response.NewError(err.Error(), http.StatusBadRequest)
	}

	return nil
}
