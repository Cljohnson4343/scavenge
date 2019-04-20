package db

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/asaskevich/govalidator"
	"github.com/cljohnson4343/scavenge/pgsql"
	"github.com/cljohnson4343/scavenge/response"
	"github.com/go-chi/chi"
)

const table string = "items"

// ItemDB is the data representation of a row from items
//
// swagger:model item
type ItemDB struct {

	// The id of the Hunt
	//
	// required: true
	HuntID int `json:"hunt_id" valid:"int,optional"`

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

	_, structErr := govalidator.ValidateStruct(i)
	if structErr != nil {
		e.Add(structErr.Error(), http.StatusBadRequest)
	}

	return e.GetError()
}

// GetTableColumnMap maps an ItemDB's data to its corresponding db table, column, and value
func (i *ItemDB) GetTableColumnMap() pgsql.TableColumnMap {
	t := make(map[string]map[string]interface{})
	t[table] = make(map[string]interface{})

	// map each non-zero valued field to an TableColumnMap value
	zeroed := ItemDB{}

	if i.HuntID != zeroed.HuntID {
		t[table]["hunt_id"] = i.HuntID
	}

	if i.ID != zeroed.ID {
		t[table]["id"] = i.ID
	}

	if i.Name != zeroed.Name {
		t[table]["name"] = i.Name
	}

	if i.Points != zeroed.Points {
		t[table]["points"] = i.Points
	}

	return t
}

// PartialItemDB is a type wrapper for ItemDB that is used to overshadow ItemDB's
// Validate()
type PartialItemDB struct {
	ItemDB `valid:"-"`
}

// Validate PartialItemDB only returns errors for non-zero valued fields
func (pItem *PartialItemDB) Validate(r *http.Request) *response.Error {
	tblColMap := pItem.GetTableColumnMap()

	_, err := govalidator.ValidateStruct(pItem.ItemDB)
	if err == nil {
		return nil
	}

	e := response.NewNilError()
	for col := range tblColMap[table] {
		errStr := govalidator.ErrorByField(err, col)
		if errStr != "" {
			e.Add(fmt.Sprintf("%s: %s", col, errStr), http.StatusBadRequest)
		}
	}

	return e.GetError()
}
