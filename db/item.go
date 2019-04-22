package db

import (
	"net/http"
	"strconv"

	"github.com/asaskevich/govalidator"
	"github.com/cljohnson4343/scavenge/pgsql"
	"github.com/cljohnson4343/scavenge/request"
	"github.com/cljohnson4343/scavenge/response"
	"github.com/go-chi/chi"
)

// ItemTbl is the name of the items db table
const ItemTbl string = "items"

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
	ID int `json:"id" valid:"int,optional"`

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

var itemSelectScript = `
	SELECT hunt_id, id, name, points
	FROM items
	WHERE id = $1;`

var itemInsertScript = `
	INSERT INTO items(hunt_id, name, points)
	VALUES ($1, $2, $3)
	RETURNING id, hunt_id;
	`
var itemsSelectScript = `
	SELECT hunt_id, id, name, points
	FROM items
	WHERE hunt_id = $1;`

var itemDeleteScript = `
	DELETE FROM items
	WHERE id = $1 AND hunt_id = $2;`

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

// GetTableColumnMap maps all non-zero field in the ItemDB to their corresponding db table, column,
//  and value
func (i *ItemDB) GetTableColumnMap() pgsql.TableColumnMap {
	t := make(pgsql.TableColumnMap)
	t[ItemTbl] = make(pgsql.ColumnMap)

	// map each non-zero valued field to an TableColumnMap value
	zeroed := ItemDB{}

	if i.HuntID != zeroed.HuntID {
		t[ItemTbl]["hunt_id"] = i.HuntID
	}

	if i.ID != zeroed.ID {
		t[ItemTbl]["id"] = i.ID
	}

	if i.Name != zeroed.Name {
		t[ItemTbl]["name"] = i.Name
	}

	if i.Points != zeroed.Points {
		t[ItemTbl]["points"] = i.Points
	}

	return t
}

// PatchValidate only returns errors for non-zero valued fields
func (i *ItemDB) PatchValidate(r *http.Request, itemID int) *response.Error {
	tblColMap := i.GetTableColumnMap()
	e := response.NewNilError()

	// patching an item requires an id that matches the given itemID,
	// if no id is provided then we can just add one
	id, ok := tblColMap[ItemTbl]["id"]
	if !ok {
		i.ID = itemID
		tblColMap[ItemTbl]["id"] = itemID
	}

	// if an id is provided that doesn't match then we alert the user
	// of a bad request
	if id != itemID {
		e.Add("id: the correct item id must be provided", http.StatusBadRequest)
		// delete the id col name so no new errors will accumulate for this column name
		delete(tblColMap[ItemTbl], "id")
	}

	patchErr := request.PatchValidate(tblColMap[ItemTbl], i)
	if patchErr != nil {
		e.AddError(patchErr)
	}

	return e.GetError()
}

// Update updates the non-zero fields of the ItemDB struct
func (i *ItemDB) Update(ex pgsql.Executioner) *response.Error {
	return update(i, ex, i.ID)
}
