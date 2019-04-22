package db

import (
	"fmt"
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

// GetItem returns the item with the given id
func GetItem(id int) (*ItemDB, *response.Error) {
	item := ItemDB{}

	err := stmtMap["itemSelect"].QueryRow(id).Scan(&item.HuntID, &item.ID, &item.Name, &item.Points)
	if err != nil {
		return nil, response.NewError(fmt.Sprintf("error getting item with id %d: %s", id, err.Error()), http.StatusInternalServerError)
	}

	return &item, nil
}

var itemInsertScript = `
	INSERT INTO items(hunt_id, name, points)
	VALUES ($1, $2, $3)
	RETURNING id;
	`

// Insert inserts the item into the items table
func (i *ItemDB) Insert() *response.Error {
	err := stmtMap["itemInsert"].QueryRow(i.HuntID, i.Name, i.Points).Scan(&i.ID)
	if err != nil {
		return response.NewError(fmt.Sprintf("error inserting item: %s", err.Error()), http.StatusInternalServerError)
	}

	return nil
}

var itemsSelectScript = `
	SELECT hunt_id, id, name, points
	FROM items
	WHERE hunt_id = $1;`

// GetItemsWithHuntID returns all the items for the given hunt id
func GetItemsWithHuntID(huntID int) ([]*ItemDB, *response.Error) {
	rows, err := stmtMap["itemsSelect"].Query(huntID)
	if err != nil {
		return nil, response.NewError(fmt.Sprintf("error getting items with hunt id %d: %s", huntID, err.Error()), http.StatusInternalServerError)
	}

	items := make([]*ItemDB, 0)
	e := response.NewNilError()

	for rows.Next() {
		item := ItemDB{}
		err := rows.Scan(&item.HuntID, &item.ID, &item.Name, &item.Points)
		if err != nil {
			e.Add(fmt.Sprintf("error getting item with hunt id %d: %s", huntID, err.Error()), http.StatusInternalServerError)
			break
		}
		items = append(items, &item)
	}

	err = rows.Err()
	if err != nil {
		e.Add(fmt.Sprintf("error getting item with hunt id %d: %s", huntID, err.Error()), http.StatusInternalServerError)
	}

	return items, e.GetError()
}

var itemDeleteScript = `
	DELETE FROM items
	WHERE id = $1 AND hunt_id = $2;`

// DeleteItem deletes the item with the given id AND huntID
func DeleteItem(id int, huntID int) *response.Error {
	res, err := stmtMap["itemDelete"].Exec(id, huntID)
	if err != nil {
		return response.NewError(fmt.Sprintf("error deleting item with id %d: %s", id, err.Error()), http.StatusInternalServerError)
	}

	numRows, err := res.RowsAffected()
	if err != nil {
		return response.NewError(fmt.Sprintf("error deleting item with id %d: %s", id, err.Error()), http.StatusInternalServerError)
	}

	if numRows < 1 {
		return response.NewError(fmt.Sprintf("there is no item with id %d and hunt id %d", id, huntID), http.StatusBadRequest)
	}

	return nil
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
		id = itemID
	}

	// if an id is provided that doesn't match then we alert the user
	// of a bad request
	if id != itemID && itemID != 0 {
		e.Add("id: the correct item id must be provided", http.StatusBadRequest)
		// delete the id col name so no new errors will accumulate for this column name
		delete(tblColMap[ItemTbl], "id")
	}

	// changing an item's hunt is not supported
	if _, ok = tblColMap[ItemTbl]["hunt_id"]; ok {
		e.Add("hunt_id: an item's hunt can not be changed with a PATCH", http.StatusBadRequest)
		delete(tblColMap[ItemTbl], "hunt_id")
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
