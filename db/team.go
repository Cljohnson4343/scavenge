package db

import (
	"fmt"
	"net/http"

	"github.com/cljohnson4343/scavenge/pgsql"

	"github.com/cljohnson4343/scavenge/request"

	"github.com/asaskevich/govalidator"
	"github.com/cljohnson4343/scavenge/response"
)

// TeamTbl is the name of the db table for the TeamDB struct
const TeamTbl string = "teams"

// A TeamDB is a representation of a row in the teams table
//
// swagger:model TeamDB
type TeamDB struct {

	// The id of the Hunt
	//
	// required: true
	HuntID int `json:"hunt_id" valid:"int"`

	// The id of the team
	//
	// required: true
	ID int `json:"id" valid:"int,optional"`

	// the name of the team
	//
	// maximum length: 255
	// required: true
	Name string `json:"name" valid:"stringlength(1|255)"`
}

// Validate validates a TeamDB struct
func (t *TeamDB) Validate(r *http.Request) *response.Error {
	_, err := govalidator.ValidateStruct(t)
	if err != nil {
		return response.NewError(err.Error(), http.StatusBadRequest)
	}

	return nil
}

// GetTableColumnMap returns a mapping between the table, column name,
// and value for each non=zero field in the TeamDB
func (t *TeamDB) GetTableColumnMap() pgsql.TableColumnMap {
	tblColMap := make(pgsql.TableColumnMap)
	tblColMap[TeamTbl] = make(pgsql.ColumnMap)

	// zero value Team for comparison sake
	z := TeamDB{}

	if z.ID != t.ID {
		tblColMap[TeamTbl]["id"] = t.ID
	}

	if z.HuntID != t.HuntID {
		tblColMap[TeamTbl]["hunt_id"] = t.HuntID
	}

	if z.Name != t.Name {
		tblColMap[TeamTbl]["name"] = t.Name
	}

	return tblColMap
}

// PatchValidate validates only the non-zero values fields of a TeamDB
func (t *TeamDB) PatchValidate(r *http.Request, teamID int) *response.Error {
	tblColMap := t.GetTableColumnMap()
	e := response.NewNilError()

	// patching a team requires an id that matches the given teamID,
	// if no id is provided then we can just add one
	id, ok := tblColMap[TeamTbl]["id"]
	if !ok {
		t.ID = teamID
		tblColMap[TeamTbl]["id"] = teamID
		id = teamID
	}

	// if an id is provided that doesn't match or if no teamID is available
	// then we alert the user of a bad request
	if id != teamID {
		e.Add("id: the correct team id must be provided", http.StatusBadRequest)
		// delete the id col name so no new errors will accumulate for this column name
		delete(tblColMap[HuntTbl], "id")
	}

	// patching a team does not support changing the hunt for that team
	if _, ok = tblColMap[TeamTbl]["hunt_id"]; ok {
		e.Add("hunt_id: the hunt can not be changed when patching a team", http.StatusBadRequest)
		// delete the hunt_id col so no new errors will accumulate for this column name
		delete(tblColMap[HuntTbl], "hunt_id")
	}

	patchErr := request.PatchValidate(tblColMap[TeamTbl], t)
	if patchErr != nil {
		e.AddError(patchErr)
	}

	return e.GetError()
}

// Update updates the non-zero value fields of the TeamDB struct
func (t *TeamDB) Update(ex pgsql.Executioner) *response.Error {
	return update(t, ex, t.ID)
}

var teamInsertScript = `
	INSERT INTO teams(hunt_id, name)
	VALUES ($1, $2)
	RETURNING hunt_id, id;`

// Insert inserts the team into the db
func (t *TeamDB) Insert() *response.Error {
	err := teamInsertStmnt.QueryRow(t.HuntID, t.Name).Scan(&t.HuntID, &t.ID)
	if err != nil {
		return response.NewError(fmt.Sprintf("error inserting team %s: %s", t.Name, err.Error()), http.StatusInternalServerError)
	}

	return nil
}

var teamSelectScript = `
	SELECT hunt_id, name, id
	FROM teams
	WHERE id = $1;`

// GetTeam returns a pointer to the team with the given id
func GetTeam(id int) (*TeamDB, *response.Error) {
	team := TeamDB{}
	err := teamSelectStmnt.QueryRow(id).Scan(&team.HuntID, &team.Name, &team.ID)
	if err != nil {
		return nil, response.NewError(fmt.Sprintf("error getting team with id %d: %s", id, err.Error()), http.StatusInternalServerError)
	}

	return &team, nil
}

var teamsSelectScript = `
	SELECT hunt_id, name, id
	FROM teams;`

// GetTeams returns all the teams in the db. NOTE if an error is returned
// the slice still needs to be checked; it is possible to error retrieving
// a single team but still return a collection of other teams
func GetTeams() ([]*TeamDB, *response.Error) {
	rows, err := teamsSelectStmnt.Query()
	if err != nil {
		return nil, response.NewError(fmt.Sprintf("error getting teams: %s", err.Error()), http.StatusInternalServerError)
	}
	defer rows.Close()

	teams := make([]*TeamDB, 0)
	e := response.NewNilError()

	for rows.Next() {
		team := TeamDB{}
		err = rows.Scan(&team.HuntID, &team.Name, &team.ID)
		if err != nil {
			e.Add(fmt.Sprintf("error getting team: %s", err.Error()), http.StatusInternalServerError)
			break
		}
		teams = append(teams, &team)
	}

	err = rows.Err()
	if err != nil {
		e.Add(fmt.Sprintf("error getting team: %s", err.Error()), http.StatusInternalServerError)
	}

	return teams, e.GetError()
}

var teamDeleteScript = `
	DELETE FROM teams
	WHERE id = $1;`

// DeleteTeam deletes the team with the given id. Providing an id that doesn't have a corresponding
// team will result in an error
func DeleteTeam(id int) *response.Error {
	res, err := teamDeleteStmnt.Exec(id)
	if err != nil {
		return response.NewError(fmt.Sprintf("error deleting team with id %d: %s", id, err.Error()), http.StatusInternalServerError)
	}

	numRows, err := res.RowsAffected()
	if err != nil {
		return response.NewError(fmt.Sprintf("error deleting team with id %d: %s", id, err.Error()), http.StatusInternalServerError)
	}

	if numRows < 1 {
		return response.NewError(fmt.Sprintf("there is no team with id %d", id), http.StatusBadRequest)
	}

	return nil
}

var teamsWithHuntIDSelectScript = `
	SELECT hunt_id, name, id
	FROM teams
	WHERE hunt_id = $1;`

// GetTeamsWithHuntID returns a slice of pointers to all the teams with the given hunt id. NOTE
// a returned error does not mean that there weren't any teams returned. It is possible to
// error scanning one of the rows returned from the query, if so, an attempt will be made to
// retrieve the remaining query results.
func GetTeamsWithHuntID(id int) ([]*TeamDB, *response.Error) {
	teams := make([]*TeamDB, 0)

	rows, err := teamsWithHuntIDSelectStmnt.Query(id)
	if err != nil {
		return nil, response.NewError(fmt.Sprintf("error getting teams with hunt id %d: %s", id, err.Error()), http.StatusInternalServerError)
	}
	defer rows.Close()

	e := response.NewNilError()
	for rows.Next() {
		team := new(TeamDB)
		err = rows.Scan(&team.HuntID, &team.Name, &team.ID)
		if err != nil {
			// try to recover and get any other teams that were returned by the query
			e.Add(fmt.Sprintf("error getting teams with hunt id %d: %s", id, err.Error()), http.StatusInternalServerError)
			break
		}

		teams = append(teams, team)
	}

	err = rows.Err()
	if err != nil {
		e.Add(fmt.Sprintf("error getting teams with hunt id %d: %s", id, err.Error()), http.StatusInternalServerError)
	}

	return teams, e.GetError()
}
