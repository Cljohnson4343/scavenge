package db

import (
	"net/http"

	"github.com/cljohnson4343/scavenge/pgsql"

	"github.com/cljohnson4343/scavenge/request"

	"github.com/asaskevich/govalidator"
	"github.com/cljohnson4343/scavenge/response"
)

// TeamTbl is the name of the db table for the TeamDB struct
const TeamTbl string = "teams"

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
	ID int `json:"id" valid:"int,optional"`

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

// Update updates the non-zero value fields of the TeamDB struct
func (t *TeamDB) Update(ex pgsql.Executioner) *response.Error {
	return update(t, ex, t.ID)
}
