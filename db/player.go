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
		uh.hunt_id
	FROM users u
	INNER JOIN users_hunts uh 
		ON uh.hunt_id = $1 AND u.id = uh.user_id
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

var playerAddToHuntScript = `
	SELECT COALESCE(ins_hunt_player($1, $2, COALESCE($3, 0)), 0);
`

// AddToHunt adds the given player to a hunt. If the player's
// teamID field is not nil, then the player will also be
// added to the given hunt, if the team is part of the hunt.
// If the teamID field is valid and player is on another team
// in the same hunt, then the player will be removed from
// old team and added to the new team.
func (p *PlayerDB) AddToHunt() *response.Error {
	err := stmtMap["playerAddToHunt"].QueryRow(
		p.HuntID,
		p.ID,
		p.TeamID,
	).Scan(&p.TeamID)
	if err != nil {
		return response.NewErrorf(
			http.StatusInternalServerError,
			"error adding player to hunt in AddToHunt: %v",
			err,
		)
	}

	return nil
}

var playerRemoveFromHuntScript = `
	DELETE FROM users_hunts
	WHERE user_id = $1 AND hunt_id = $2;
`

// RemoveFromHunt removes the player from a hunt. This method
// depends on the player's ID and HuntID fields for removal.
func (p *PlayerDB) RemoveFromHunt() *response.Error {
	res, err := stmtMap["playerRemoveFromHunt"].Exec(p.ID, p.HuntID)
	if err != nil {
		return response.NewErrorf(
			http.StatusInternalServerError,
			"error removing player from hunt: %v",
			err,
		)
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return response.NewErrorf(
			http.StatusInternalServerError,
			"error getting rows affected for player removal from hunt: %v",
			err,
		)
	}

	if rowsAffected == 0 {
		return response.NewError(
			http.StatusBadRequest,
			"Can't remove a player from a hunt that they haven't joined.",
		)
	}

	return nil
}

// Invite invites a player to a hunt. The player struct should
// have the HuntID and Email fields set. The player does not
// have to be a registered user to be invited to a hunt.
func (p *PlayerDB) Invite(inviterID int) *response.Error {
	invite := HuntInvitationDB{
		HuntID:    p.HuntID,
		Email:     p.Email,
		InviterID: inviterID,
	}

	return invite.Insert()
}
