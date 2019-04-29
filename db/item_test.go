// +build unit

package db_test

import (
	"net/http"
	"testing"

	"github.com/cljohnson4343/scavenge/db"
)

var r = &http.Request{}

func TestValidate(t *testing.T) {
	var caseStr = `valid case`
	i := &db.ItemDB{Name: "item name", Points: 43, ID: 43, HuntID: 43}

	got := i.Validate(r)
	if got != nil {
		t.Errorf("%s: expecting to recieve nil but got %s", caseStr, got.JSON())
	}

	caseStr = "Invalid zero value"
	z := &db.ItemDB{}

	err := z.Validate(r)
	if err == nil {
		t.Errorf("%s: expected errors but got nil", caseStr)
	}

	errMap := err.ErrorsByKey()

	if _, ok := errMap["name"]; !ok {
		t.Errorf("%s: expecting a name error but go nil", caseStr)
	}

	caseStr = "Negative points"
	i.Points = -43

	err = i.Validate(r)
	errMap = err.ErrorsByKey()
	if _, ok := errMap["points"]; !ok {
		t.Errorf("%s: expection a points error but got nil", caseStr)
	}
}

func TestGetTableColumnMap(t *testing.T) {
	caseStr := "zero value"
	i := db.ItemDB{}

	tblColMap := i.GetTableColumnMap()

	if len(tblColMap["items"]) != 0 {
		t.Errorf("%s: expected zero length map", caseStr)
	}

	tables := []struct {
		item     db.ItemDB
		fieldStr string
		length   int
		expected interface{}
	}{
		{db.ItemDB{HuntID: 43}, "hunt_id", 1, 43},
		{db.ItemDB{ID: 43}, "id", 1, 43},
		{db.ItemDB{Name: "Chris Johnson"}, "name", 1, "Chris Johnson"},
		{db.ItemDB{Points: 43}, "points", 1, 43},
	}

	for _, v := range tables {
		i = v.item
		tblColMap = i.GetTableColumnMap()

		got, ok := tblColMap["items"][v.fieldStr]
		if !ok {
			t.Errorf("%s: expected %v", v.fieldStr, v.expected)
		} else {
			if got != v.expected {
				t.Errorf("%s: expected %v got %v", v.fieldStr, v.expected, got)
			}
		}

	}
}
