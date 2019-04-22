package hunts

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/cljohnson4343/scavenge/db"
	"github.com/cljohnson4343/scavenge/response"

	c "github.com/cljohnson4343/scavenge/config"
	"github.com/cljohnson4343/scavenge/hunts/models"
	"github.com/cljohnson4343/scavenge/pgsql"
)

// GetItems returns the items for the given hunt
func GetItems(huntID int) ([]*models.Item, *response.Error) {
	itemDBs, e := db.GetItemsWithHuntID(huntID)
	if itemDBs == nil {
		return nil, e
	}

	items := make([]*models.Item, 0, len(itemDBs))
	for _, itemDB := range itemDBs {
		item := models.Item{*itemDB}
		items = append(items, &item)
	}

	return items, e
}

// GetItemsForHunt returns all the items for the hunt with the given id
func GetItemsForHunt(huntID int) ([]*models.Item, *response.Error) {
	itemDBs, e := db.GetItemsWithHuntID(huntID)
	if itemDBs == nil {
		return nil, e
	}

	items := make([]*models.Item, 0, len(itemDBs))

	for _, v := range itemDBs {
		item := models.Item{*v}
		items = append(items, &item)
	}

	return items, e
}

// InsertItem inserts an Item into the db
func InsertItem(env *c.Env, item *models.Item, huntID int) *response.Error {
	return item.Insert()
}

func getUpsertItemsSQLStatement(huntID int, newItems []interface{}) (*pgsql.Command, *response.Error) {
	var sqlValuesSB strings.Builder
	sqlValuesSB.WriteString("(")
	inc := 1

	sqlCmd := new(pgsql.Command)
	e := response.NewNilError()
	for _, value := range newItems {
		item, ok := value.(map[string]interface{})
		if !ok {
			e.Add("request json is not valid", http.StatusBadRequest)
			break
		}

		v, ok := item["name"]
		if !ok {
			e.Add("name field is required", http.StatusBadRequest)
			break
		}

		name, ok := v.(string)
		if !ok {
			e.Add("name field has to be a string", http.StatusBadRequest)
			break
		}

		ptsV, ok := item["points"]
		if !ok {
			e.Add("points field is required", http.StatusBadRequest)
			break
		}

		pts, ok := ptsV.(float64)
		if !ok {
			e.Add("points field has to be a float64 > 0", http.StatusBadRequest)
			break
		}

		// make sure all validation is done before writing to sqlValueSB and adding to sqlCmd.args
		sqlValuesSB.WriteString(fmt.Sprintf("$%d, $%d, $%d),(", inc, inc+1, inc+2))
		inc += 3
		sqlCmd.AppendArgs(huntID, name, int(pts))
	}

	// drop the extra ',(' from value string
	valuesStr := (sqlValuesSB.String())[:sqlValuesSB.Len()-2]

	sqlCmd.AppendScript(fmt.Sprintf("\n\tINSERT INTO items(hunt_id, name, points)\n\tVALUES\n\t\t%s\n\tON CONFLICT ON CONSTRAINT items_in_same_hunt_name\n\tDO\n\t\tUPDATE\n\t\tSET name = EXCLUDED.name, points = EXCLUDED.points;", valuesStr))

	return sqlCmd, e.GetError()
}

// DeleteItem deletes the item with the given itemID AND huntID
func DeleteItem(env *c.Env, huntID, itemID int) *response.Error {
	return db.DeleteItem(itemID, huntID)
}

// UpdateItem executes a partial update of the item with the given id. NOTE:
// item_id and hunt_id are not eligible to be changed
func UpdateItem(env *c.Env, item *models.Item) *response.Error {
	return item.Update(env)
}
