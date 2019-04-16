package hunts

import (
	"errors"
	"fmt"
	"strings"

	"github.com/cljohnson4343/scavenge/hunts/models"
)

// getItems populates the items slice with all the items for the given hunt
func (env *Env) getItems(items *[]models.Item, huntID int) error {
	sqlStatement := `
		SELECT name, points FROM items WHERE items.hunt_id = $1;`

	rows, err := env.db.Query(sqlStatement, huntID)
	if err != nil {
		return err
	}
	defer rows.Close()

	item := models.Item{}
	for rows.Next() {
		err = rows.Scan(&item.Name, &item.Points)
		if err != nil {
			return err
		}

		*items = append(*items, item)
	}

	return nil
}

// insertItem inserts an Item into the db
func (env *Env) insertItem(item *models.Item, huntID int) (int, error) {
	sqlStatement := `
		INSERT INTO items(hunt_id, name, points)
		VALUES ($1, $2, $3)
		RETURNING id`

	id := 0
	err := env.db.QueryRow(sqlStatement, huntID, item.Name, item.Points).Scan(&id)

	return id, err
}

func getUpsertItemsSQLStatement(huntID int, newItems []interface{}) (*sqlStatement, error) {
	var eb, sqlValuesSB strings.Builder

	eb.WriteString("Error updating items: \n")
	encounteredError := false

	handleErr := func(errString string) {
		encounteredError = true
		eb.WriteString(errString)
	}

	sqlValuesSB.WriteString("(")
	inc := 1

	sqlStmnt := new(sqlStatement)

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
		sqlStmnt.args = append(sqlStmnt.args, huntID, name, int(pts))
	}

	// drop the extra ',(' from value string
	valuesStr := (sqlValuesSB.String())[:sqlValuesSB.Len()-2]

	sqlStmnt.sql = fmt.Sprintf("\n\tINSERT INTO items(hunt_id, name, points)\n\tVALUES\n\t\t%s\n\tON CONFLICT ON CONSTRAINT items_in_same_hunt_name\n\tDO\n\t\tUPDATE\n\t\tSET name = EXCLUDED.name, points = EXCLUDED.points;", valuesStr)

	if encounteredError {
		return sqlStmnt, errors.New(eb.String())
	}

	return sqlStmnt, nil
}
