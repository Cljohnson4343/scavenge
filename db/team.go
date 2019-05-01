package db

import (
	"database/sql"
	"net/http"

	"github.com/asaskevich/govalidator"
	"github.com/lib/pq"

	"github.com/cljohnson4343/scavenge/pgsql"
	"github.com/cljohnson4343/scavenge/request"
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
	HuntID int `json:"hunt_id" valid:"int,optional"`

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
		return response.NewError(http.StatusBadRequest, err.Error())
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
		e.Add(http.StatusBadRequest, "id: the correct team id must be provided")
		// delete the id col name so no new errors will accumulate for this column name
		delete(tblColMap[HuntTbl], "id")
	}

	// patching a team does not support changing the hunt for that team
	if _, ok = tblColMap[TeamTbl]["hunt_id"]; ok {
		e.Add(http.StatusBadRequest, "hunt_id: the hunt can not be changed when patching a team")
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
	RETURNING id;`

// Insert inserts the team into the db
func (t *TeamDB) Insert() *response.Error {
	err := stmtMap["teamInsert"].QueryRow(t.HuntID, t.Name).Scan(&t.ID)
	if err != nil {
		return t.ParseError(err, "insert")
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
	err := stmtMap["teamSelect"].QueryRow(id).Scan(&team.HuntID, &team.Name, &team.ID)
	if err == nil {
		return &team, nil
	}

	if err == sql.ErrNoRows {
		return nil, response.NewErrorf(
			http.StatusBadRequest,
			"team_id: no team with id %d",
			id,
		)
	}

	return nil, response.NewErrorf(
		http.StatusInternalServerError,
		"error getting team with id %d: %s",
		id,
		err.Error(),
	)
}

var teamsSelectScript = `
	SELECT hunt_id, name, id
	FROM teams;`

// GetTeams returns all the teams in the db. NOTE if an error is returned
// the slice still needs to be checked; it is possible to error retrieving
// a single team but still return a collection of other teams
func GetTeams() ([]*TeamDB, *response.Error) {
	rows, err := stmtMap["teamsSelect"].Query()
	if err != nil {
		return nil, response.NewErrorf(http.StatusInternalServerError, "error getting teams: %s", err.Error())
	}
	defer rows.Close()

	teams := make([]*TeamDB, 0)
	e := response.NewNilError()

	for rows.Next() {
		team := TeamDB{}
		err = rows.Scan(&team.HuntID, &team.Name, &team.ID)
		if err != nil {
			e.Addf(http.StatusInternalServerError, "error getting team: %s", err.Error())
			break
		}
		teams = append(teams, &team)
	}

	err = rows.Err()
	if err != nil {
		e.Addf(http.StatusInternalServerError, "error getting team: %s", err.Error())
	}

	return teams, e.GetError()
}

var teamDeleteScript = `
	DELETE FROM teams
	WHERE id = $1;`

// DeleteTeam deletes the team with the given id. Providing an id that doesn't have a corresponding
// team will result in an error
func DeleteTeam(id int) *response.Error {
	res, err := stmtMap["teamDelete"].Exec(id)
	if err != nil {
		return response.NewErrorf(http.StatusInternalServerError, "error deleting team with id %d: %s", id, err.Error())
	}

	numRows, err := res.RowsAffected()
	if err != nil {
		return response.NewErrorf(http.StatusInternalServerError, "error deleting team with id %d: %s", id, err.Error())
	}

	if numRows < 1 {
		return response.NewErrorf(http.StatusBadRequest, "there is no team with id %d", id)
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

	rows, err := stmtMap["teamsWithHuntIDSelect"].Query(id)
	if err != nil {
		return nil, response.NewErrorf(http.StatusInternalServerError, "error getting teams with hunt id %d: %s", id, err.Error())
	}
	defer rows.Close()

	e := response.NewNilError()
	for rows.Next() {
		team := new(TeamDB)
		err = rows.Scan(&team.HuntID, &team.Name, &team.ID)
		if err != nil {
			// try to recover and get any other teams that were returned by the query
			e.Addf(http.StatusInternalServerError, "error getting teams with hunt id %d: %s", id, err.Error())
			break
		}

		teams = append(teams, team)
	}

	err = rows.Err()
	if err != nil {
		e.Addf(http.StatusInternalServerError, "error getting teams with hunt id %d: %s", id, err.Error())
	}

	return teams, e.GetError()
}

var teamGetPlayersScript = `
	SELECT 
		u.id, 
		u.first_name, 
		u.last_name, 
		u.username, 
		u.joined_at, 
		u.last_visit, 
		u.image_url, 
		u.email
	FROM users_teams u_t 
	INNER JOIN users u 
		ON u_t.team_id = $1 AND u_t.user_id = u.id;`

// GetUsersForTeam returns all the players, users, for a given team
func GetUsersForTeam(teamID int) ([]*UserDB, *response.Error) {
	rows, err := stmtMap["teamGetPlayers"].Query(teamID)
	if err != nil {
		return nil, response.NewErrorf(http.StatusInternalServerError,
			"GetUsersForTeam: error getting users for team %d: %v", teamID, err)
	}
	defer rows.Close()

	e := response.NewNilError()
	users := make([]*UserDB, 0)
	for rows.Next() {
		u := UserDB{}
		err = rows.Scan(&u.ID, &u.FirstName, &u.LastName, &u.Username, &u.JoinedAt,
			&u.ImageURL, &u.Email)
		if err != nil {
			e.Addf(http.StatusInternalServerError,
				"GetUsersForTeam: error getting user for team %d: %v", teamID, err)
			break
		}

		users = append(users, &u)
	}

	if err = rows.Err(); err != nil {
		e.Addf(http.StatusInternalServerError,
			"GetUsersForTeam: error getting user for team %d: %v", teamID, err)
	}

	return users, e.GetError()
}

var teamAddPlayerScript = `
	INSERT INTO users_teams(team_id, user_id)
	VALUES ($1, $2);`

// TeamAddPlayer adds the user with the given id to the team with the given id
func TeamAddPlayer(teamID, userID int) *response.Error {
	res, err := stmtMap["teamAddPlayer"].Exec(teamID, userID)
	if err != nil {
		team := TeamDB{ID: teamID}
		return team.ParseError(err, "insert")
	}

	numRows, err := res.RowsAffected()
	if err != nil {
		return response.NewErrorf(http.StatusInternalServerError,
			"TeamAddPlayer: error adding player %d to team %d: %v", userID, teamID, err)
	}

	if numRows < 1 {
		return response.NewErrorf(http.StatusBadRequest,
			"TeamAddPlayer: error adding player %d to team %d", userID, teamID)
	}

	return nil
}

var teamRemovePlayersScript = `
	DELETE FROM users_teams
	WHERE team_id = $1 AND user_id = $2;`

// TeamRemovePlayers removes the users from the given team
func TeamRemovePlayers(teamID int, players []int) *response.Error {
	e := response.NewNilError()
	for _, id := range players {
		res, err := stmtMap["teamRemovePlayers"].Exec(teamID, id)
		if err != nil {
			e.Addf(http.StatusInternalServerError,
				"TeamAddPlayer: error removing player %d from team %d: %v", id, teamID, err)
		}

		numRows, err := res.RowsAffected()
		if err != nil {
			e.Addf(http.StatusInternalServerError,
				"TeamAddPlayer: error removing player %d from team %d: %v", id, teamID, err)
		}

		if numRows < 1 {
			e.Addf(http.StatusInternalServerError,
				"TeamAddPlayer: error removing player %d from team %d", id, teamID)
		}
	}

	return e.GetError()
}

// ParseError maps a pq driver error to a response.Error that contains the information a
// client needs to know.
func (t *TeamDB) ParseError(err error, op string) *response.Error {
	pqErr, ok := err.(*pq.Error)
	if ok {
		if pqErr.Constraint != "" {
			switch pqErr.Constraint {
			case "teams_in_same_hunt_name":
				return response.NewErrorf(
					http.StatusBadRequest,
					"name: %s is already in use for this hunt",
					t.Name,
				)
			case "teams_hunt_id_fkey":
				return response.NewErrorf(
					http.StatusBadRequest,
					"hunt_id: hunt %d does not exist",
					t.HuntID,
				)
			case "users_teams_team_id_fkey":
				return response.NewErrorf(
					http.StatusBadRequest,
					"team_id: team %d does not exist",
					t.ID,
				)
			case "users_teams_user_id_fkey":
				return response.NewError(
					http.StatusBadRequest,
					"user_id: user being added does not exist",
				)
			}
		}
	}

	return response.NewErrorf(
		http.StatusInternalServerError,
		"error performing operation %s: %v",
		op,
		err,
	)
}
