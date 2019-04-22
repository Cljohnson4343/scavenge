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

// GetTeams populates the teams slice with all the teams. If an error
// is returned the team slice still needs to be checked as the error
// might have resulted from getting a single team
func GetTeams(env *c.Env) ([]*Team, *response.Error) {
	teamDBs, e := db.GetTeams()
	if teamDBs == nil {
		return nil, e
	}

	teams := make([]*Team, 0, len(teamDBs))

	for _, teamDB := range teamDBs {
		team := Team{*teamDB}
		teams = append(teams, &team)
	}

	return teams, e
}

// GetTeamsForHunt populates the teams slice with all the teams
// of the given hunt. NOTE if an error is returned then the team slice
// still needs to be checked as the error could have occurred while trying
// to get a single team
func GetTeamsForHunt(huntID int) ([]*Team, *response.Error) {
	teamDBs, e := db.GetTeamsWithHuntID(huntID)
	if teamDBs == nil {
		return nil, e
	}

	teams := make([]*Team, 0, len(teamDBs))
	for _, teamDB := range teamDBs {
		team := Team{*teamDB}
		teams = append(teams, &team)
	}

	return teams, nil
}

// GetTeam returns the Team with the given ID
func GetTeam(env *c.Env, teamID int) (*Team, *response.Error) {
	teamDB, e := db.GetTeam(teamID)
	if e != nil {
		return nil, e
	}

	team := Team{*teamDB}
	return &team, nil
}

// InsertTeam inserts a Team into the db
func InsertTeam(env *c.Env, team *Team) *response.Error {
	return team.Insert()
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
	return db.DeleteTeam(teamID)
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
