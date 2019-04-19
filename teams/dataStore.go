package teams

import (
	"fmt"
	"net/http"
	"strings"

	c "github.com/cljohnson4343/scavenge/config"
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
			e.AddError(err.Error(), http.StatusInternalServerError)
		}

		*teams = append(*teams, team)
	}

	err = rows.Err()
	if err != nil {
		e.AddError(err.Error(), http.StatusInternalServerError)
	}

	return teams, e.GetError()
}

// GetTeamsForHunt populates the teams slice with all the teams
// of the given hunt
func GetTeamsForHunt(env *c.Env, huntID int) (*[]Team, *response.Error) {
	sqlStmnt := `
		SELECT name, id, hunt_id FROM teams WHERE hunt_id = $1;`

	rows, err := env.Query(sqlStmnt, huntID)
	if err != nil {
		return nil, response.NewError(err.Error(), http.StatusInternalServerError)
	}
	defer rows.Close()

	teams := new([]Team)

	e := response.NewNilError()

	team := Team{}
	for rows.Next() {
		err = rows.Scan(&team.Name, &team.ID, &team.HuntID)
		if err != nil {
			e.AddError(err.Error(), http.StatusInternalServerError)
		}

		*teams = append(*teams, team)
	}

	err = rows.Err()
	if err != nil {
		e.AddError(err.Error(), http.StatusInternalServerError)
	}

	return teams, e.GetError()
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
		RETURNING id`

	id := 0
	err := env.QueryRow(sqlStmnt, huntID, team.Name).Scan(&id)
	if err != nil {
		return 0, response.NewError(err.Error(), http.StatusBadRequest)
	}

	return id, nil
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
			e.AddError(fmt.Sprintf("request json is invalid. Check the %d indexed team.", k), http.StatusBadRequest)
			break
		}

		v, ok := team["name"]
		if !ok {
			e.AddError("the name field is required", http.StatusBadRequest)
			break
		}

		name, ok := v.(string)
		if !ok {
			e.AddError("name field should be of type string", http.StatusBadRequest)
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
func UpdateTeam(env *c.Env, teamID int, partialTeam *map[string]interface{}) *response.Error {
	sqlCmd, e := getUpdateTeamSQLCommand(teamID, partialTeam)
	if e != nil {
		return e
	}

	res, err := sqlCmd.Exec(env)
	if err != nil {
		return response.NewError(fmt.Sprintf("error updating team %d", teamID), http.StatusInternalServerError)
	}

	n, err := res.RowsAffected()
	if err != nil {
		return response.NewError(fmt.Sprintf("error updating team %d", teamID), http.StatusInternalServerError)
	}

	if n < 1 {
		return response.NewError(fmt.Sprintf("team %d was not updated. Check to make sure teamID and huntID are valid",
			teamID), http.StatusInternalServerError)
	}

	return nil
}

// getUpdateTeamSQLCommand returns a pgsql.Command struct for updating a team
// NOTE: the hunt_id and the team_id are not editable
func getUpdateTeamSQLCommand(teamID int, partialTeam *map[string]interface{}) (*pgsql.Command, *response.Error) {
	var sqlB strings.Builder
	sqlB.WriteString(`
		UPDATE teams
		SET `)

	e := response.NewNilError()
	sqlCmd := &pgsql.Command{}
	inc := 1
	for k, v := range *partialTeam {
		switch k {
		case "name":
			newName, ok := v.(string)
			if !ok {
				e.AddError("name field has to be of type string", http.StatusBadRequest)
			}

			sqlB.WriteString(fmt.Sprintf("name=$%d,", inc))
			inc++
			sqlCmd.AppendArgs(newName)
		}
	}

	// cut the trailing comma
	sqlStrLen := sqlB.Len()
	sqlCmd.AppendScript(fmt.Sprintf("%s\n\t\tWHERE id = $%d;", sqlB.String()[:sqlStrLen-1], inc))
	sqlCmd.AppendArgs(teamID)

	return sqlCmd, e.GetError()
}
