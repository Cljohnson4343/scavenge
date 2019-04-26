package db_test

import (
	"net/http"
	"testing"

	"github.com/cljohnson4343/scavenge/db"
)

var req = &http.Request{}

func TestValidateTeamDB(t *testing.T) {
	caseStr := "valid team"
	team := &db.TeamDB{ID: 43, HuntID: 43, Name: "Chris Johnson"}

	err := team.Validate(req)
	if err != nil {
		t.Errorf("%s: expected nil", caseStr)
	}

	caseStr = "name"
	team = &db.TeamDB{ID: 43, HuntID: 43}
	err = team.Validate(req)
	errMap := err.ErrorsByKey()
	if len(errMap) != 1 {
		t.Errorf("%s: expected length 1 got %d", caseStr, len(errMap))
	}

	_, ok := errMap[caseStr]
	if !ok {
		t.Errorf("%s: expected a %s key in error map", caseStr, caseStr)
	}
}

func TestGetTableColumnMapTeamDB(t *testing.T) {
	caseStr := "zero case"
	team := db.TeamDB{}

	tblColMap := team.GetTableColumnMap()
	if len(tblColMap["teams"]) != 0 {
		t.Errorf("%s: expected no errors", caseStr)
	}

	tables := []struct {
		team     db.TeamDB
		fieldStr string
		length   int
		expected interface{}
	}{
		{db.TeamDB{HuntID: 43}, "hunt_id", 1, 43},
		{db.TeamDB{ID: 43}, "id", 1, 43},
		{db.TeamDB{Name: "Chris Johnson"}, "name", 1, "Chris Johnson"},
	}

	for _, c := range tables {
		tblColMap = c.team.GetTableColumnMap()

		if len(tblColMap["teams"]) != c.length {
			t.Errorf("%s: expected length %d but got %d", c.fieldStr, c.length, len(tblColMap["teams"]))
			break
		}

		v, ok := tblColMap["teams"][c.fieldStr]
		if !ok {
			t.Errorf("%s: expected a %s value", c.fieldStr, c.fieldStr)
			break
		}
		if v != c.expected {
			t.Errorf("%s: expected %v but got %v", c.fieldStr, c.expected, v)
		}
	}
}
