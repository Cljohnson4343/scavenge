package hunts

import (
	"errors"
	"fmt"
	"strings"

	"github.com/cljohnson4343/scavenge/hunts/models"
)

// getTeams populates the teams slice with all the teams for the given hunt
func (env *Env) getTeams(huntID int) (*[]models.Team, error) {
	sqlStatement := `
		SELECT name, id, hunt_id FROM teams WHERE teams.hunt_id = $1;`

	rows, err := env.db.Query(sqlStatement, huntID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	teams := new([]models.Team)

	team := models.Team{}
	for rows.Next() {
		err = rows.Scan(&team.Name, &team.ID, &team.HuntID)
		if err != nil {
			return nil, err
		}

		*teams = append(*teams, team)
	}

	return teams, nil
}

// getTeam returns the Team with the given ID
func (env *Env) getTeam(teamID int) (*models.Team, error) {
	sqlStatement := `
		SELECT name, hunt_id, id FROM teams WHERE teams.id = $1;`

	team := new(models.Team)
	err := env.db.QueryRow(sqlStatement, teamID).Scan(&team.Name, &team.HuntID, &team.ID)
	if err != nil {
		return nil, err
	}

	return team, nil
}

// insertTeam inserts a Team into the db
func (env *Env) insertTeam(team *models.Team, huntID int) (int, error) {
	sqlStatement := `
		INSERT INTO teams(hunt_id, name)
		VALUES ($1, $2)
		RETURNING id`

	id := 0
	err := env.db.QueryRow(sqlStatement, huntID, team.Name).Scan(&id)

	return id, err
}

func getUpsertTeamsSQLStatement(huntID int, newTeams []interface{}) (*sqlStatement, error) {
	var eb, sqlValuesSB strings.Builder

	eb.WriteString("Error updating teams: \n")
	encounteredError := false

	handleErr := func(errString string) {
		encounteredError = true
		eb.WriteString(errString)
	}

	sqlValuesSB.WriteString("(")
	inc := 1

	sqlStmnt := new(sqlStatement)

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
		sqlStmnt.args = append(sqlStmnt.args, huntID, name)
	}

	valuesStr := (sqlValuesSB.String())[:sqlValuesSB.Len()-2]

	sqlStmnt.sql = fmt.Sprintf("\n\tINSERT INTO teams(hunt_id, name)\n\tVALUES\n\t\t%s\n\tON CONFLICT ON CONSTRAINT teams_in_same_hunt_name\n\tDO\n\t\tUPDATE\n\t\tSET name = EXCLUDED.name;", valuesStr)

	if encounteredError {
		return sqlStmnt, errors.New(eb.String())
	}

	return sqlStmnt, nil
}

// deleteTeam deletes the team with the given teamID AND huntID
func (env *Env) deleteTeam(huntID, teamID int) error {
	sqlStatement := `
		DELETE FROM teams
		WHERE id = $1 AND hunt_id = $2;`

	res, err := env.db.Exec(sqlStatement, teamID, huntID)
	if err != nil {
		return err
	}

	numRows, err := res.RowsAffected()
	if err != nil {
		return err
	}

	if numRows < 1 {
		return fmt.Errorf("hunt %d does not have a team %d", huntID, teamID)
	}

	return nil
}

// updateTeam executes a partial update of the team with the given id. NOTE:
// team_id and hunt_id are not eligible to be changed
func (env *Env) updateTeam(huntID, teamID int, partialTeam *map[string]interface{}) error {
	sqlStmnt, err := getUpdateTeamSQLStatement(huntID, teamID, partialTeam)
	if err != nil {
		return err
	}

	res, err := sqlStmnt.exec(env.db)
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

// getUpdateTeamSQLStatement returns a sqlStatement struct for updating a team
// NOTE: the hunt_id and the team_id are not editable
func getUpdateTeamSQLStatement(huntID int, teamID int, partialTeam *map[string]interface{}) (*sqlStatement, error) {
	var eb, sqlB strings.Builder

	sqlB.WriteString(`
		UPDATE teams
		SET `)

	eb.WriteString(fmt.Sprintf("error updating team %d:\n", teamID))
	encounteredError := false

	sqlStmnt := &sqlStatement{}

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
			sqlStmnt.args = append(sqlStmnt.args, newName)
		}
	}

	// cut the trailing comma
	sqlStrLen := sqlB.Len()
	sqlStmnt.sql = fmt.Sprintf("%s\n\t\tWHERE id = $%d AND hunt_id = $%d;",
		sqlB.String()[:sqlStrLen-1], inc, inc+1)
	sqlStmnt.args = append(sqlStmnt.args, teamID, huntID)

	if encounteredError {
		return sqlStmnt, errors.New(eb.String())
	}

	return sqlStmnt, nil
}
