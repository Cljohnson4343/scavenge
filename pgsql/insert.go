package pgsql

import (
	"github.com/cljohnson4343/scavenge/response"
)

// GetInsertSQLCommand returns sql commands that can insert the information from the
// TableColumnMap and return the values specified in the returning columns slice
func GetInsertSQLCommand(tblColMap TableColumnMap, returningCols [][]string) (*Command, *response.Error) {

	return nil, nil
}
