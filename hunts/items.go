package hunts

import (
	"errors"
	"fmt"
	"strings"

	c "github.com/cljohnson4343/scavenge/config"
	"github.com/cljohnson4343/scavenge/hunts/models"
)

// GetItems returns the items for the given hunt
func GetItems(env *c.Env, huntID int) (*[]models.Item, error) {
	sqlStmnt := `
		SELECT name, points, id FROM items WHERE items.hunt_id = $1;`

	rows, err := env.DB().Query(sqlStmnt, huntID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := new([]models.Item)
	item := models.Item{}
	for rows.Next() {
		err = rows.Scan(&item.Name, &item.Points, &item.ID)
		if err != nil {
			return nil, err
		}

		item.HuntID = huntID
		*items = append(*items, item)
	}

	err = rows.Err()

	return items, err
}

// InsertItem inserts an Item into the db
func InsertItem(env *c.Env, item *models.Item, huntID int) (int, error) {
	sqlStmnt := `
		INSERT INTO items(hunt_id, name, points)
		VALUES ($1, $2, $3)
		RETURNING id`

	id := 0
	err := env.DB().QueryRow(sqlStmnt, huntID, item.Name, item.Points).Scan(&id)

	return id, err
}

func getUpsertItemsSQLStatement(huntID int, newItems []interface{}) (*c.SQLStatement, error) {
	var eb, sqlValuesSB strings.Builder

	eb.WriteString("Error updating items: \n")
	encounteredError := false

	handleErr := func(errString string) {
		encounteredError = true
		eb.WriteString(errString)
	}

	sqlValuesSB.WriteString("(")
	inc := 1

	sqlStmnt := new(c.SQLStatement)

	for _, value := range newItems {
		item, ok := value.(map[string]interface{})
		if !ok {
			handleErr(fmt.Sprintf("Expected newItems to be type map[string]interface{} but got %T\n", value))
			break
		}

		v, ok := item["name"]
		if !ok {
			handleErr("Expected a name value for items.\n")
			break
		}

		name, ok := v.(string)
		if !ok {
			handleErr(fmt.Sprintf("Expected a name type of string but got %T\n", v))
			break
		}

		ptsV, ok := item["points"]
		if !ok {
			handleErr(fmt.Sprintf("Expected a points value for item with name %s\n", name))
			break
		}

		pts, ok := ptsV.(float64)
		if !ok {
			handleErr(fmt.Sprintf("Expected a points type of float64 but got %T for item with name %s\n", ptsV, name))
			pts = 1
		}

		// make sure all validation is done before writing to sqlValueSB and adding to sqlStmnt.args
		sqlValuesSB.WriteString(fmt.Sprintf("$%d, $%d, $%d),(", inc, inc+1, inc+2))
		inc += 3
		sqlStmnt.AppendArgs(huntID, name, int(pts))
	}

	// drop the extra ',(' from value string
	valuesStr := (sqlValuesSB.String())[:sqlValuesSB.Len()-2]

	sqlStmnt.AppendScript(fmt.Sprintf("\n\tINSERT INTO items(hunt_id, name, points)\n\tVALUES\n\t\t%s\n\tON CONFLICT ON CONSTRAINT items_in_same_hunt_name\n\tDO\n\t\tUPDATE\n\t\tSET name = EXCLUDED.name, points = EXCLUDED.points;", valuesStr))

	if encounteredError {
		return sqlStmnt, errors.New(eb.String())
	}

	return sqlStmnt, nil
}

// DeleteItem deletes the item with the given itemID AND huntID
func DeleteItem(env *c.Env, huntID, itemID int) error {
	sqlStmnt := `
		DELETE FROM items
		WHERE id = $1 AND hunt_id = $2;`

	res, err := env.DB().Exec(sqlStmnt, itemID, huntID)
	if err != nil {
		return err
	}

	numRows, err := res.RowsAffected()
	if err != nil {
		return err
	}

	if numRows < 1 {
		return fmt.Errorf("hunt w/ id %d does not have an item w/ id %d", huntID, itemID)
	}

	return nil
}

// UpdateItem executes a partial update of the item with the given id. NOTE:
// item_id and hunt_id are not eligible to be changed
func UpdateItem(env *c.Env, huntID, itemID int, partialItem *map[string]interface{}) error {
	sqlStmnt, err := getUpdateItemSQLStatement(huntID, itemID, partialItem)
	if err != nil {
		return err
	}

	res, err := sqlStmnt.Exec(env.DB())
	if err != nil {
		return err
	}

	n, err := res.RowsAffected()
	if err != nil {
		return err
	}

	if n < 1 {
		return errors.New("the item was not updated. Check the URL and request body to make sure itemID and huntID are valid")
	}

	return nil
}

// getUpdateItemSQLStatement returns a c.SQLStatement struct for updating an item
// NOTE: the hunt_id and the item_id are not editable
func getUpdateItemSQLStatement(huntID int, itemID int, partialItem *map[string]interface{}) (*c.SQLStatement, error) {
	var eb, sqlB strings.Builder

	sqlB.WriteString(`
		UPDATE items
		SET `)

	eb.WriteString(fmt.Sprintf("error updating item %d:\n", itemID))
	encounteredError := false

	sqlStmnt := &c.SQLStatement{}

	inc := 1
	for k, v := range *partialItem {
		switch k {
		case "name":
			newName, ok := v.(string)
			if !ok {
				eb.WriteString(fmt.Sprintf("expected name to be of type string but got %T\n", v))
				encounteredError = true
				break
			}

			sqlB.WriteString(fmt.Sprintf("name=$%d,", inc))
			inc++
			sqlStmnt.AppendArgs(newName)

		case "points":
			newPts, ok := v.(float64)
			if !ok {
				eb.WriteString(fmt.Sprintf("expected points to be of type float64 but got %T\n", v))
				encounteredError = true
				break
			}

			sqlB.WriteString(fmt.Sprintf("points=$%d,", inc))
			inc++
			sqlStmnt.AppendArgs(int(newPts))
		}
	}

	// cut the trailing comma
	sqlStrLen := sqlB.Len()
	sqlStmnt.AppendScript(fmt.Sprintf("%s\n\t\tWHERE id = $%d AND hunt_id = $%d;",
		sqlB.String()[:sqlStrLen-1], inc, inc+1))
	sqlStmnt.AppendArgs(itemID, huntID)

	if encounteredError {
		return sqlStmnt, errors.New(eb.String())
	}

	return sqlStmnt, nil
}
