package teams

import (
	"errors"
	"fmt"
	"log"
	"strings"

	c "github.com/cljohnson4343/scavenge/config"
	"github.com/cljohnson4343/scavenge/db"
)

// GetTeams populates the teams slice with all the teams
func GetTeams(env *c.Env) (*[]Team, error) {
	sqlStmnt := `
		SELECT name, id, hunt_id FROM teams;`

	rows, err := env.Query(sqlStmnt)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	teams := new([]Team)

	team := Team{}
	for rows.Next() {
		err = rows.Scan(&team.Name, &team.ID, &team.HuntID)
		if err != nil {
			return nil, err
		}

		*teams = append(*teams, team)
	}

	err = rows.Err()

	return teams, err
}

// GetTeamsForHunt populates the teams slice with all the teams
// of the given hunt
func GetTeamsForHunt(env *c.Env, huntID int) (*[]Team, error) {
	sqlStmnt := `
		SELECT name, id, hunt_id FROM teams WHERE hunt_id = $1;`

	rows, err := env.Query(sqlStmnt, huntID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	teams := new([]Team)

	team := Team{}
	for rows.Next() {
		err = rows.Scan(&team.Name, &team.ID, &team.HuntID)
		if err != nil {
			return nil, err
		}

		*teams = append(*teams, team)
	}

	err = rows.Err()

	return teams, err
}

// GetTeam returns the Team with the given ID
func GetTeam(env *c.Env, teamID int) (*Team, error) {
	sqlStmnt := `
		SELECT name, hunt_id, id FROM teams WHERE teams.id = $1;`

	team := new(Team)
	err := env.QueryRow(sqlStmnt, teamID).Scan(&team.Name, &team.HuntID, &team.ID)
	if err != nil {
		return nil, err
	}

	return team, nil
}

// InsertTeam inserts a Team into the db
func InsertTeam(env *c.Env, team *Team, huntID int) (int, error) {
	sqlStmnt := `
		INSERT INTO teams(hunt_id, name)
		VALUES ($1, $2)
		RETURNING id`

	id := 0
	err := env.QueryRow(sqlStmnt, huntID, team.Name).Scan(&id)

	return id, err
}

// GetUpsertTeamsSQLStatement returns the db.SQLStatement that will update/insert the
// teams described by the slice parameter
func GetUpsertTeamsSQLStatement(huntID int, newTeams []interface{}) (*db.SQLStatement, error) {
	var eb, sqlValuesSB strings.Builder

	eb.WriteString("Error updating teams: \n")
	encounteredError := false

	handleErr := func(errString string) {
		encounteredError = true
		eb.WriteString(errString)
	}

	sqlValuesSB.WriteString("(")
	inc := 1

	sqlStmnt := new(db.SQLStatement)

	for _, value := range newTeams {
		team, ok := value.(map[string]interface{})
		if !ok {
			handleErr(fmt.Sprintf("Expected newTeams to be type map[string]interface{} but got %T\n", value))
			break
		}

		v, ok := team["name"]
		if !ok {
			handleErr("Expected a name value.\n")
			break
		}

		name, ok := v.(string)
		if !ok {
			handleErr(fmt.Sprintf("Expected a name type of string but got %T\n", v))
			break
		}

		// make sure all validation is done before writing to sqlValueSB and adding to sqlStmnt.args
		sqlValuesSB.WriteString(fmt.Sprintf("$%d, $%d),(", inc, inc+1))
		inc += 2
		sqlStmnt.AppendArgs(huntID, name)
	}

	valuesStr := (sqlValuesSB.String())[:sqlValuesSB.Len()-2]

	sqlStmnt.AppendScript(fmt.Sprintf("\n\tINSERT INTO teams(hunt_id, name)\n\tVALUES\n\t\t%s\n\tON CONFLICT ON CONSTRAINT teams_in_same_hunt_name\n\tDO\n\t\tUPDATE\n\t\tSET name = EXCLUDED.name;", valuesStr))

	if encounteredError {
		return sqlStmnt, errors.New(eb.String())
	}

	return sqlStmnt, nil
}

// DeleteTeam deletes the team with the given teamID
func DeleteTeam(env *c.Env, teamID int) error {
	sqlStmnt := `
		DELETE FROM teams
		WHERE id = $1;`

	res, err := env.Exec(sqlStmnt, teamID)
	if err != nil {
		return err
	}

	numRows, err := res.RowsAffected()
	if err != nil {
		return err
	}

	if numRows < 1 {
		return fmt.Errorf("there is no team with id %d", teamID)
	}

	return nil
}

// UpdateTeam executes a partial update of the team with the given id. NOTE:
// team_id and hunt_id are not eligible to be changed
func UpdateTeam(env *c.Env, teamID int, partialTeam *map[string]interface{}) error {
	sqlStmnt, err := getUpdateTeamSQLStatement(teamID, partialTeam)
	if err != nil {
		return err
	}

	res, err := sqlStmnt.Exec(env)
	if err != nil {
		return err
	}

	n, err := res.RowsAffected()
	if err != nil {
		return err
	}

	if n < 1 {
		return errors.New("the team was not updated. Check the URL and request body to make sure teamID and huntID are valid")
	}

	return nil
}

// getUpdateTeamSQLStatement returns a db.SQLStatement struct for updating a team
// NOTE: the hunt_id and the team_id are not editable
func getUpdateTeamSQLStatement(teamID int, partialTeam *map[string]interface{}) (*db.SQLStatement, error) {
	var eb, sqlB strings.Builder

	sqlB.WriteString(`
		UPDATE teams
		SET `)

	eb.WriteString(fmt.Sprintf("error updating team %d:\n", teamID))
	encounteredError := false

	sqlStmnt := &db.SQLStatement{}

	inc := 1
	for k, v := range *partialTeam {
		switch k {
		case "name":
			newName, ok := v.(string)
			if !ok {
				eb.WriteString(fmt.Sprintf("expected name to be of type string but got %T\n", v))
				encounteredError = true
			}

			sqlB.WriteString(fmt.Sprintf("name=$%d,", inc))
			inc++
			sqlStmnt.AppendArgs(newName)
		}
	}

	// cut the trailing comma
	sqlStrLen := sqlB.Len()
	sqlStmnt.AppendScript(fmt.Sprintf("%s\n\t\tWHERE id = $%d;",
		sqlB.String()[:sqlStrLen-1], inc))
	sqlStmnt.AppendArgs(teamID)

	log.Println(sqlStmnt.Script())
	log.Println(sqlStmnt.Args())
	if encounteredError {
		return sqlStmnt, errors.New(eb.String())
	}

	return sqlStmnt, nil
}
