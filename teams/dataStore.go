package teams

import (
	"net/http"

	c "github.com/cljohnson4343/scavenge/config"
	"github.com/cljohnson4343/scavenge/db"
	"github.com/cljohnson4343/scavenge/pgsql"
	"github.com/cljohnson4343/scavenge/response"
)

// GetTeams populates the teams slice with all the teams. If an error
// is returned the team slice still needs to be checked as the error
// might have resulted from getting a single team
func GetTeams() ([]*Team, *response.Error) {
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
		team := Team{TeamDB: *teamDB}
		teams = append(teams, &team)
	}

	return teams, nil
}

// GetTeam returns the Team with the given ID
func GetTeam(teamID int) (*Team, *response.Error) {
	teamDB, e := db.GetTeam(teamID)
	if e != nil {
		return nil, e
	}

	team := Team{*teamDB}
	return &team, nil
}

// InsertTeam inserts a Team into the db
func InsertTeam(team *Team) *response.Error {
	// inserting a team that has a non-zero id is not valid
	if team.ID != 0 {
		return response.NewError(
			http.StatusBadRequest,
			"id: can not provide id when creating team",
		)
	}
	return team.Insert()
}

// DeleteTeam deletes the team with the given teamID
func DeleteTeam(teamID int) *response.Error {
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
		return response.NewErrorf(
			http.StatusInternalServerError,
			"error updating team %d: %v",
			team.ID,
			err,
		)
	}

	n, err := res.RowsAffected()
	if err != nil {
		return response.NewErrorf(
			http.StatusInternalServerError,
			"error updating team %d: %v",
			team.ID,
			err,
		)
	}

	if n < 1 {
		return response.NewErrorf(
			http.StatusInternalServerError,
			"team %d was not updated. Check to make sure teamID and huntID are valid",
			team.ID,
		)
	}

	return nil
}
