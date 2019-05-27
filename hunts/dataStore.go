package hunts

import (
	"context"
	"net/http"

	"github.com/cljohnson4343/scavenge/config"
	"github.com/cljohnson4343/scavenge/db"
	"github.com/cljohnson4343/scavenge/pgsql"
	"github.com/cljohnson4343/scavenge/response"
	"github.com/cljohnson4343/scavenge/roles"
	"github.com/cljohnson4343/scavenge/teams"
	"github.com/cljohnson4343/scavenge/users"
)

// AllHunts returns all Hunts from the database
func AllHunts() ([]*Hunt, *response.Error) {
	huntDBs, e := db.GetHunts()
	if huntDBs == nil {
		return nil, e
	}

	if e == nil {
		e = response.NewNilError()
	}

	hunts := make([]*Hunt, 0, len(huntDBs))

	for _, h := range huntDBs {
		ts, teamErr := teams.GetTeamsForHunt(h.ID)
		if teamErr != nil {
			e.AddError(teamErr)
		}

		items, itemErr := GetItemsForHunt(h.ID)
		if itemErr != nil {
			e.AddError(itemErr)
		}

		hunt := Hunt{HuntDB: *h, Teams: ts, Items: items}

		hunts = append(hunts, &hunt)
	}

	return hunts, e.GetError()
}

// GetHunt returns a pointer to the hunt with the given ID.
func GetHunt(huntID int) (*Hunt, *response.Error) {
	huntDB, e := db.GetHunt(huntID)
	if e != nil {
		return nil, e
	}

	e = response.NewNilError()

	teams, teamErr := teams.GetTeamsForHunt(huntID)
	if teamErr != nil {
		e.AddError(teamErr)
	}

	items, itemErr := GetItems(huntID)
	if itemErr != nil {
		e.AddError(itemErr)
	}

	return &Hunt{HuntDB: *huntDB, Teams: teams, Items: items}, e.GetError()
}

// GetHuntByCreatorAndName returns a pointer to the hunt with the given
// creator username and hunt name.
func GetHuntByCreatorAndName(creator, name string) (*Hunt, *response.Error) {
	huntDB, e := db.GetHuntByCreatorAndName(creator, name)
	if e != nil {
		return nil, e
	}

	e = response.NewNilError()

	teams, teamErr := teams.GetTeamsForHunt(huntDB.ID)
	if teamErr != nil {
		e.AddError(teamErr)
	}

	items, itemErr := GetItems(huntDB.ID)
	if itemErr != nil {
		e.AddError(itemErr)
	}

	return &Hunt{HuntDB: *huntDB, Teams: teams, Items: items}, e.GetError()
}

// GetHuntsByUserID returns all Hunts for the given user
func GetHuntsByUserID(userID int) ([]*Hunt, *response.Error) {
	huntDBs, e := db.GetHuntsByUserID(userID)
	if huntDBs == nil {
		return nil, e
	}

	if e == nil {
		e = response.NewNilError()
	}

	hunts := make([]*Hunt, 0, len(huntDBs))

	for _, h := range huntDBs {
		ts, teamErr := teams.GetTeamsForHunt(h.ID)
		if teamErr != nil {
			e.AddError(teamErr)
		}

		items, itemErr := GetItemsForHunt(h.ID)
		if itemErr != nil {
			e.AddError(itemErr)
		}

		hunt := Hunt{HuntDB: *h, Teams: ts, Items: items}

		hunts = append(hunts, &hunt)
	}

	return hunts, e.GetError()
}

// InsertHunt inserts the given hunt into the database and updates the hunt
// with the new id and created_at timestamp
func InsertHunt(ctx context.Context, hunt *Hunt) *response.Error {
	// set the creator field
	userID, e := users.GetUserID(ctx)
	if e != nil {
		return e
	}
	hunt.CreatorID = userID

	// TODO I don't think I like how hunts are inserted. Go over and see about refactoring
	e = hunt.HuntDB.Insert()
	if e != nil {
		return e
	}

	huntOwner := roles.New("hunt_owner", hunt.ID)
	e = huntOwner.AddTo(userID)
	if e != nil {
		return e
	}

	e = response.NewNilError()

	for _, team := range hunt.Teams {
		team.HuntID = hunt.ID
		teamErr := teams.InsertTeam(ctx, team)
		if teamErr != nil {
			e.AddError(teamErr)
			break
		}
	}

	for _, item := range hunt.Items {
		item.HuntID = hunt.ID
		itemErr := InsertItem(item)
		if itemErr != nil {
			e.AddError(itemErr)
			break
		}
	}

	// TODO add the hunt creator automatically
	for _, player := range hunt.Players {
		player.HuntID = hunt.ID
		playerErr := player.Invite(userID)
		if playerErr != nil {
			e.AddError(playerErr)
			break
		}
	}

	return e.GetError()
}

// DeleteHunt deletes the hunt with the given ID. All associated data will also be deleted.
func DeleteHunt(huntID int) *response.Error {
	teams, e := db.TeamsForHunt(huntID)
	if e != nil {
		return e
	}

	e = db.DeleteHunt(huntID)
	if e != nil {
		return e
	}

	e = roles.DeleteRolesForHunt(huntID, teams)
	if e != nil {
		return e
	}

	return nil
}

// UpdateHunt updates the hunt with the given ID using the fields that are not nil in the
// partial hunt. If the hunt was updated then true will be returned. id field can not be
// updated.
func UpdateHunt(env *config.Env, hunt *Hunt) (bool, *response.Error) {
	tx, err := env.Begin()
	if err != nil {
		return false, response.NewError(http.StatusInternalServerError, err.Error())
	}

	e := hunt.Update(tx)
	if e != nil {
		err = tx.Rollback()
		if err != nil {
			e.Add(http.StatusInternalServerError, err.Error())
			return false, e.GetError()
		}

		return false, e
	}

	err = tx.Commit()
	if err != nil {
		return false, response.NewError(http.StatusInternalServerError, err.Error())
	}

	return true, nil
}

// getUpdateHuntSQLCommand returns the commands to update the db based on the provided hunt
// the partial hunt. The given hunt should only provide data that needs to be updated AND
// the hunt.ID field MUST be set
func getUpdateHuntSQLCommand(hunt *Hunt) (*[]*pgsql.Command, *response.Error) {
	// get all the mappings for all hunt's entities to their table, column name, and value
	tblColMaps := hunt.GetTableColumnMaps()

	// use the number of cmds that will be needed to avoid unnecessary memory allocations
	sqlCmds := make([]*pgsql.Command, 0, len(tblColMaps))

	e := response.NewNilError()

	for _, tblColMap := range tblColMaps {
		for tbl, colMap := range tblColMap {
			cmd, cmdErr := pgsql.GetUpdateSQLCommand(colMap, tbl, hunt.ID)
			if cmdErr != nil {
				e.AddError(cmdErr)
				break
			}
			sqlCmds = append(sqlCmds, cmd)
		}
	}

	return &sqlCmds, nil
}

// AddPlayer adds a player to the given hunt and assigns the necessary roles
func AddPlayer(huntID int, player *db.PlayerDB) *response.Error {
	player.HuntID = huntID
	e := player.AddToHunt()
	if e != nil {
		return e
	}

	huntMember := roles.New("hunt_member", huntID)
	e = huntMember.AddTo(player.ID)
	if e != nil {
		return e
	}

	return nil
}
