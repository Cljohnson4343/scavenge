package hunts

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/cljohnson4343/scavenge/response"

	c "github.com/cljohnson4343/scavenge/config"
	"github.com/cljohnson4343/scavenge/db"
	"github.com/cljohnson4343/scavenge/hunts/models"
)

// GetItems returns the items for the given hunt
func GetItems(env *c.Env, huntID int) (*[]models.Item, *response.Error) {
	sqlStmnt := `
		SELECT name, points, id FROM items WHERE items.hunt_id = $1;`

	rows, err := env.Query(sqlStmnt, huntID)
	if err != nil {
		return nil, response.NewError(err.Error(), http.StatusInternalServerError)
	}
	defer rows.Close()

	e := response.NewNilError()
	items := new([]models.Item)
	item := models.Item{}
	for rows.Next() {
		err = rows.Scan(&item.Name, &item.Points, &item.ID)
		if err != nil {
			e.AddError(err.Error(), http.StatusInternalServerError)
			break
		}

		item.HuntID = huntID
		*items = append(*items, item)
	}

	err = rows.Err()
	if err != nil {
		return items, response.NewError(err.Error(), http.StatusInternalServerError)
	}

	return items, e.GetError()
}

// InsertItem inserts an Item into the db
func InsertItem(env *c.Env, item *models.Item, huntID int) (int, *response.Error) {
	sqlStmnt := `
		INSERT INTO items(hunt_id, name, points)
		VALUES ($1, $2, $3)
		RETURNING id`

	id := 0
	err := env.QueryRow(sqlStmnt, huntID, item.Name, item.Points).Scan(&id)
	if err != nil {
		return id, response.NewError(err.Error(), http.StatusInternalServerError)
	}

	return id, nil
}

func getUpsertItemsSQLStatement(huntID int, newItems []interface{}) (*db.SQLCommand, *response.Error) {
	var sqlValuesSB strings.Builder
	sqlValuesSB.WriteString("(")
	inc := 1

	sqlCmd := new(db.SQLCommand)
	e := response.NewNilError()
	for _, value := range newItems {
		item, ok := value.(map[string]interface{})
		if !ok {
			e.AddError("request json is not valid", http.StatusBadRequest)
			break
		}

		v, ok := item["name"]
		if !ok {
			e.AddError("name field is required", http.StatusBadRequest)
			break
		}

		name, ok := v.(string)
		if !ok {
			e.AddError("name field has to be a string", http.StatusBadRequest)
			break
		}

		ptsV, ok := item["points"]
		if !ok {
			e.AddError("points field is required", http.StatusBadRequest)
			break
		}

		pts, ok := ptsV.(float64)
		if !ok {
			e.AddError("points field has to be a float64 > 0", http.StatusBadRequest)
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
	sqlStmnt := `
		DELETE FROM items
		WHERE id = $1 AND hunt_id = $2;`

	res, err := env.Exec(sqlStmnt, itemID, huntID)
	if err != nil {
		return response.NewError(err.Error(), http.StatusInternalServerError)
	}

	numRows, err := res.RowsAffected()
	if err != nil {
		return response.NewError(err.Error(), http.StatusInternalServerError)
	}

	if numRows < 1 {
		return response.NewError("hunt does not have an item with that id", http.StatusBadRequest)
	}

	return nil
}

// UpdateItem executes a partial update of the item with the given id. NOTE:
// item_id and hunt_id are not eligible to be changed
func UpdateItem(env *c.Env, huntID, itemID int, partialItem *map[string]interface{}) *response.Error {
	sqlCmd, e := getUpdateItemSQLCommand(huntID, itemID, partialItem)
	if e != nil {
		return e
	}

	res, err := sqlCmd.Exec(env)
	if err != nil {
		return response.NewError(err.Error(), http.StatusInternalServerError)
	}

	n, err := res.RowsAffected()
	if err != nil {
		return response.NewError(err.Error(), http.StatusInternalServerError)
	}

	if n < 1 {
		return response.NewError("make sure that huntID and teamID are valid", http.StatusBadRequest)
	}

	return nil
}

// getUpdateItemSQLCommand returns a db.SQLCommand struct for updating an item
// NOTE: the hunt_id and the item_id are not editable
func getUpdateItemSQLCommand(huntID int, itemID int, partialItem *map[string]interface{}) (*db.SQLCommand, *response.Error) {
	var sqlB strings.Builder
	sqlB.WriteString(`
		UPDATE items
		SET `)

	sqlCmd := &db.SQLCommand{}
	e := response.NewNilError()
	inc := 1
	for k, v := range *partialItem {
		switch k {
		case "name":
			newName, ok := v.(string)
			if !ok {
				e.AddError("name field has to be of type string", http.StatusBadRequest)
				break
			}

			sqlB.WriteString(fmt.Sprintf("name=$%d,", inc))
			inc++
			sqlCmd.AppendArgs(newName)

		case "points":
			newPts, ok := v.(float64)
			if !ok {
				e.AddError("points field has to be of type float64", http.StatusBadRequest)
				break
			}

			sqlB.WriteString(fmt.Sprintf("points=$%d,", inc))
			inc++
			sqlCmd.AppendArgs(int(newPts))
		}
	}

	// cut the trailing comma
	sqlStrLen := sqlB.Len()
	sqlCmd.AppendScript(fmt.Sprintf("%s\n\t\tWHERE id = $%d AND hunt_id = $%d;",
		sqlB.String()[:sqlStrLen-1], inc, inc+1))
	sqlCmd.AppendArgs(itemID, huntID)

	return sqlCmd, e.GetError()
}
