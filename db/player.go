package db

import (
	"net/http"

	"github.com/cljohnson4343/scavenge/response"
)

// A PlayerDB is a user that has joined a particular hunt.
//
// swagger:model PlayerDB
type PlayerDB struct {
	UserDB `valid:"-"`

	// the team, if any, this player is on
	TeamID int `json:"teamID" valid:"-"`

	// the hunt being referred to
	HuntID int `json:"huntID" valid:"-"`
}

// Validate is a dummy function because PlayerDB model is server generated and
// not meant to be posted by clients
func (p *PlayerDB) Validate(r *http.Request) *response.Error {
	return nil
}

var playersGetForHuntScript = `
WITH teams_for_hunt AS(
	SELECT t.id AS hunt_team
	FROM teams t 
	WHERE t.hunt_id = $1
), players_for_teams AS (
	SELECT ut.user_id AS player, ut.team_id AS players_team
	FROM users_teams ut
	INNER JOIN teams_for_hunt tfh
		ON tfh.hunt_team = ut.team_id
)
	SELECT
		u.id,
		u.first_name,
		u.last_name,
		u.username,
		u.joined_at,
		u.last_visit,
		COALESCE(u.image_url, ''),
		u.email,
		COALESCE(pft.players_team, 0),
		hu.hunt_id
	FROM users u
	INNER JOIN hunts_users hu 
		ON hu.hunt_id = $1 AND u.id = hu.user_id
	LEFT OUTER JOIN players_for_teams pft 
		ON pft.player = u.id;
`

// GetPlayersForHunt returns all the players for a hunt
func GetPlayersForHunt(huntID int) ([]*PlayerDB, *response.Error) {
	rows, err := stmtMap["playersGetForHunt"].Query(huntID)
	if err != nil {

	}
	defer rows.Close()

	players := make([]*PlayerDB, 0)
	e := response.NewNilError()

	for rows.Next() {
		p := PlayerDB{}
		err = rows.Scan(
			&p.ID,
			&p.FirstName,
			&p.LastName,
			&p.Username,
			&p.JoinedAt,
			&p.LastVisit,
			&p.ImageURL,
			&p.Email,
			&p.TeamID,
			&p.HuntID,
		)
		if err != nil {
			e.Addf(
				http.StatusInternalServerError,
				"error scanning row in GetPlayersForHunt: %s",
				err.Error(),
			)
			break
		}
		players = append(players, &p)
	}

	if err = rows.Err(); err != nil {
		e.Addf(
			http.StatusInternalServerError,
			"an error occurred while retrieving players: %s",
			err.Error(),
		)
	}

	return players, e.GetError()
}
