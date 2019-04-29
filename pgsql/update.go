package pgsql

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/cljohnson4343/scavenge/response"
)

// GetUpdateSQLCommand returns a Command that can be executed to update the given table
// using the ColumnMap
func GetUpdateSQLCommand(colMap ColumnMap, tbl string, id int) (*Command, *response.Error) {
	// get the number of columns to be updated
	numColUpdated := len(colMap)
	names := make([]string, 0, numColUpdated)

	cmd := GetNewCommand(numColUpdated + 1)
	inc := 1
	for k, v := range colMap {
		names = append(names, fmt.Sprintf("%s=$%d", k, inc))
		inc++
		cmd.AppendArgs(v)
	}

	nameExpStr := strings.Join(names, ", ")

	_, err := cmd.AppendScript(
		fmt.Sprintf("\n\t\tUPDATE %s\n\t\tSET %s\n\t\tWHERE id=$%d;",
			tbl,
			nameExpStr,
			numColUpdated+1,
		),
	)
	if err != nil {
		return nil, response.NewError(http.StatusInternalServerError, err.Error())
	}

	// add the final arg, the id of the WHERE constraint
	cmd.AppendArgs(id)

	return cmd, nil
}
