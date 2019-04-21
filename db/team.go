package db

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/cljohnson4343/scavenge/pgsql"

	"github.com/cljohnson4343/scavenge/request"

	"github.com/asaskevich/govalidator"
	"github.com/cljohnson4343/scavenge/response"
)

// TeamTbl is the name of the db table for the TeamDB struct
const TeamTbl string = "teams"

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

// ValidateWithoutHuntID is a special case func for when the HuntID isn't
// required.
func (t *TeamDB) ValidateWithoutHuntID(r *http.Request) *response.Error {
	_, err := govalidator.ValidateStruct(t)
	errMap := govalidator.ErrorsByField(err)

	delete(errMap, "hunt_id")

	if len(errMap) > 0 {
		e := response.NewNilError()
		for k, v := range errMap {
			e.Add(fmt.Sprintf("%s: %s", k, v), http.StatusBadRequest)
		}
		return e
	}

	return nil
}

// GetTableColumnMap returns a mapping between the table, column name,
// and value for each non=zero field in the TeamDB
func (t *TeamDB) GetTableColumnMap() pgsql.TableColumnMap {
	tblColMap := make(pgsql.TableColumnMap)
	tblColMap[TeamTbl] = make(pgsql.ColumnMap)

	// zero value Team for comparison sake
	z := TeamDB{}

	if z.ID != t.ID {
		tblColMap[TeamTbl]["id"] = t.ID
	}

	if z.HuntID != t.HuntID {
		tblColMap[TeamTbl]["hunt_id"] = t.HuntID
	}

	if z.Name != t.Name {
		tblColMap[TeamTbl]["name"] = t.Name
	}

	return tblColMap
}

// PartialValidate validates only the non-zero values fields of a TeamDB
func (t *TeamDB) PartialValidate(r *http.Request) *response.Error {
	tblColMap := t.GetTableColumnMap()

	return request.PartialValidate(tblColMap[TeamTbl], t)
}
