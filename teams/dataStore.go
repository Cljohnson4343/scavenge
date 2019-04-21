package teams

import (
	"fmt"
	"net/http"
	"strings"

	c "github.com/cljohnson4343/scavenge/config"
	"github.com/cljohnson4343/scavenge/db"
	"github.com/cljohnson4343/scavenge/pgsql"
	"github.com/cljohnson4343/scavenge/response"
)

// GetTeams populates the teams slice with all the teams
func GetTeams(env *c.Env) (*[]Team, *response.Error) {
	sqlStmnt := `
		SELECT name, id, hunt_id FROM teams;`

	rows, err := env.Query(sqlStmnt)
	if err != nil {
		return nil, response.NewError(err.Error(), http.StatusBadRequest)
	}
	defer rows.Close()

	teams := new([]Team)

	e := response.NewNilError()

	team := Team{}
	for rows.Next() {
		err = rows.Scan(&team.Name, &team.ID, &team.HuntID)
		if err != nil {
			e.Add(err.Error(), http.StatusInternalServerError)
		}

		*teams = append(*teams, team)
	}

	err = rows.Err()
	if err != nil {
		e.Add(err.Error(), http.StatusInternalServerError)
	}

	return teams, e.GetError()
}

// GetTeamsForHunt populates the teams slice with all the teams
// of the given hunt
func GetTeamsForHunt(env *c.Env, huntID int) (*[]*Team, *response.Error) {
	sqlStmnt := `
		SELECT name, id, hunt_id FROM teams WHERE hunt_id = $1;`

	rows, err := env.Query(sqlStmnt, huntID)
	if err != nil {
		return nil, response.NewError(err.Error(), http.StatusInternalServerError)
	}
	defer rows.Close()

	teams := make([]*Team, 0)
	e := response.NewNilError()
	for rows.Next() {
		team := Team{}
		err = rows.Scan(&team.Name, &team.ID, &team.HuntID)
		if err != nil {
			e.Add(err.Error(), http.StatusInternalServerError)
		}

		teams = append(teams, &team)
	}

	err = rows.Err()
	if err != nil {
		e.Add(err.Error(), http.StatusInternalServerError)
	}

	return &teams, e.GetError()
}

// GetTeam returns the Team with the given ID
func GetTeam(env *c.Env, teamID int) (*Team, *response.Error) {
	sqlStmnt := `
		SELECT name, hunt_id, id FROM teams WHERE teams.id = $1;`

	team := new(Team)
	err := env.QueryRow(sqlStmnt, teamID).Scan(&team.Name, &team.HuntID, &team.ID)
	if err != nil {
		return nil, response.NewError(err.Error(), http.StatusBadRequest)
	}

	return team, nil
}

// InsertTeam inserts a Team into the db
func InsertTeam(env *c.Env, team *Team, huntID int) (int, *response.Error) {
	sqlStmnt := `
		INSERT INTO teams(hunt_id, name)
		VALUES ($1, $2)
		RETURNING id, hunt_id`

	err := env.QueryRow(sqlStmnt, huntID, team.Name).Scan(&team.ID, &team.HuntID)
	if err != nil {
		return 0, response.NewError(err.Error(), http.StatusBadRequest)
	}

	return team.ID, nil
}

// GetUpsertTeamsSQLCommand returns the pgsql.Command that will update/insert the
// teams described by the slice parameter
func GetUpsertTeamsSQLCommand(huntID int, newTeams []interface{}) (*pgsql.Command, *response.Error) {
	var sqlValuesSB strings.Builder
	sqlValuesSB.WriteString("(")
	inc := 1

	e := response.NewNilError()
	sqlCmd := new(pgsql.Command)
	for k, value := range newTeams {
		team, ok := value.(map[string]interface{})
		if !ok {
			e.Add(fmt.Sprintf("request json is invalid. Check the %d indexed team.", k), http.StatusBadRequest)
			break
		}

		v, ok := team["name"]
		if !ok {
			e.Add("the name field is required", http.StatusBadRequest)
			break
		}

		name, ok := v.(string)
		if !ok {
			e.Add("name field should be of type string", http.StatusBadRequest)
			break
		}

		// make sure all validation is done before writing to sqlValueSB and adding to sqlCmd.args
		sqlValuesSB.WriteString(fmt.Sprintf("$%d, $%d),(", inc, inc+1))
		inc += 2
		sqlCmd.AppendArgs(huntID, name)
	}

	// strip the unnecessary ,( at the end of the string
	valuesStr := (sqlValuesSB.String())[:sqlValuesSB.Len()-2]
	sqlCmd.AppendScript(fmt.Sprintf("\n\tINSERT INTO teams(hunt_id, name)\n\tVALUES\n\t\t%s\n\tON CONFLICT ON CONSTRAINT teams_in_same_hunt_name\n\tDO\n\t\tUPDATE\n\t\tSET name = EXCLUDED.name;", valuesStr))

	return sqlCmd, e.GetError()
}

// DeleteTeam deletes the team with the given teamID
func DeleteTeam(env *c.Env, teamID int) *response.Error {
	sqlStmnt := `
		DELETE FROM teams
		WHERE id = $1;`

	res, err := env.Exec(sqlStmnt, teamID)
	if err != nil {
		return response.NewError(fmt.Sprintf("error deleting team with id %d", teamID), http.StatusInternalServerError)
	}

	numRows, err := res.RowsAffected()
	if err != nil {
		return response.NewError(fmt.Sprintf("error deleting team with id %d", teamID), http.StatusInternalServerError)
	}

	if numRows < 1 {
		return response.NewError(fmt.Sprintf("there is no team with id %d", teamID), http.StatusBadRequest)
	}

	return nil
}

// UpdateTeam executes a partial update of the team with the given id. NOTE:
// team_id and hunt_id are not eligible to be changed
func UpdateTeam(env *c.Env, team *Team) *response.Error {
	tblColMap := team.GetTableColumnMap()
	cmd, e := pgsql.GetUpdateSQLCommand(tblColMap[db.TeamTbl], db.TeamTbl, team.ID)
	if e != nil {
		return e
	}

	res, err := cmd.Exec(env)
	if err != nil {
		return response.NewError(fmt.Sprintf("error updating team %d", team.ID), http.StatusInternalServerError)
	}

	n, err := res.RowsAffected()
	if err != nil {
		return response.NewError(fmt.Sprintf("error updating team %d", team.ID), http.StatusInternalServerError)
	}

	if n < 1 {
		return response.NewError(fmt.Sprintf("team %d was not updated. Check to make sure teamID and huntID are valid",
			team.ID), http.StatusInternalServerError)
	}

	return nil
}
