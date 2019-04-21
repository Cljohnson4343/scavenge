package pgsql

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/cljohnson4343/scavenge/response"
)

// ColumnMap is a type that maps database column names to their values
type ColumnMap map[string]interface{}

// TableColumnMap is a type that maps a database table, the first string key,
// to a column, the second string key, value.
type TableColumnMap map[string]ColumnMap

// TableColumnMapper is an interface that wraps the GetTableColumnMap method.
// This method is used to generate a mapping between data and the associated
// table, column, and value in a database.
type TableColumnMapper interface {
	GetTableColumnMap() TableColumnMap
}

// GetUpdateSQLCommand returns a Command that can be executed to update the given table
// using the ColumnMap
func GetUpdateSQLCommand(colMap ColumnMap, tbl string) (*Command, *response.Error) {
	// get the number of columns to be updated
	numColUpdated := len(colMap)
	var sb strings.Builder

	cmd := GetNewCommand(numColUpdated + 1)
	inc := 1
	for k, v := range colMap {
		_, err := sb.WriteString(fmt.Sprintf("%s = $%d ", k, inc))
		if err != nil {
			return nil, response.NewError(err.Error(), http.StatusInternalServerError)
		}

		inc++
		cmd.AppendArgs(v)
	}

	_, err := cmd.AppendScript(fmt.Sprintf("\n\tUPDATE %s *\n\t\tSET %s\n\t\tWHERE id = $%d;",
		tbl, sb.String(), numColUpdated+1))
	if err != nil {
		return nil, response.NewError(err.Error(), http.StatusInternalServerError)
	}

	return cmd, nil
}
