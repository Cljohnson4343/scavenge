package db

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/cljohnson4343/scavenge/pgsql"

	"github.com/cljohnson4343/scavenge/response"
)

func update(v pgsql.TableColumnMapper, ex pgsql.Executioner, id int) *response.Error {
	tblColMap := v.GetTableColumnMap()

	if len(tblColMap) != 1 {
		panic(errors.New("incorrect use of db.update. the Updater should return one table"))
	}

	var tblName string
	var colMap pgsql.ColumnMap

	// tblColMap only has one ColumnMap so we can use the for range for assignment
	for tblName, colMap = range tblColMap {
		break
	}

	cmd, e := pgsql.GetUpdateSQLCommand(colMap, tblName, id)
	if e != nil {
		return e
	}

	res, err := ex.Exec(cmd.Script(), cmd.Args()...)
	if err != nil {
		return response.NewError(fmt.Sprintf("%s id %d error: %s", tblName, id, err.Error()), http.StatusInternalServerError)
	}

	numRows, err := res.RowsAffected()
	if err != nil {
		return response.NewError(err.Error(), http.StatusInternalServerError)
	}

	if numRows < 1 {
		return response.NewError(fmt.Sprintf("nothing was updated. Make sure an entity with id %d exists.", id), http.StatusBadRequest)
	}

	return nil
}
