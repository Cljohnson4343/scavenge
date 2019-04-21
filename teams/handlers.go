package teams

import (
	"net/http"

	c "github.com/cljohnson4343/scavenge/config"
	"github.com/cljohnson4343/scavenge/request"
	"github.com/go-chi/render"
)

// swagger:route GET /teams/ teams getTeamsHandler
//
// Lists all the teams.
//
// Consumes:
// 	- application/json
//
// Produces:
//	- application/json
//
// Schemes: http, https
//
// Responses:
// 	200:
// 	400:
//  500:
func getTeamsHandler(env *c.Env) http.HandlerFunc {
	return (func(w http.ResponseWriter, r *http.Request) {
		teams, e := GetTeams(env)
		if e != nil {
			e.Handle(w)
			return
		}

		render.JSON(w, r, teams)
		return
	})
}

// swagger:route GET /teams/{teamID} team getTeamHandler
//
// Gets the team with given id.
//
// Consumes:
// 	- application/json
//
// Produces:
//	- application/json
//
// Schemes: http, https
//
// Responses:
// 	200:
// 	400:
// 	404:
func getTeamHandler(env *c.Env) http.HandlerFunc {
	return (func(w http.ResponseWriter, r *http.Request) {
		teamID, e := request.GetIntURLParam(r, "teamID")
		if e != nil {
			e.Handle(w)
			return
		}

		team, e := GetTeam(env, teamID)
		if e != nil {
			e.Handle(w)
			return
		}

		render.JSON(w, r, team)
		return
	})
}

// swagger:route DELETE /teams/{teamID} delete team deleteTeamHandler
//
// Deletes the given team.
//
// Consumes:
// 	- application/json
//
// Produces:
//	- application/json
//
// Schemes: http, https
//
// Responses:
// 	200:
//  400:
func deleteTeamHandler(env *c.Env) http.HandlerFunc {
	return (func(w http.ResponseWriter, r *http.Request) {
		teamID, e := request.GetIntURLParam(r, "teamID")
		if e != nil {
			e.Handle(w)
			return
		}

		e = DeleteTeam(env, teamID)
		if e != nil {
			e.Handle(w)
		}

		return
	})
}

// swagger:route POST /teams/ team create createTeamHandler
//
// Creates the given team.
//
// Consumes:
// 	- application/json
//
// Produces:
//	- application/json
//
// Schemes: http, https
//
// Responses:
// 	200:
//  400:
//  500:
func createTeamHandler(env *c.Env) http.HandlerFunc {
	return (func(w http.ResponseWriter, r *http.Request) {
		team := Team{}
		e := request.DecodeAndValidate(r, &team)
		if e != nil {
			e.Handle(w)
			return
		}

		teamID, e := InsertTeam(env, &team, team.HuntID)
		if e != nil {
			e.Handle(w)
			return
		}

		team.ID = teamID
		render.JSON(w, r, &team)
		return
	})
}

// swagger:route PATCH /teams/{teamID} team patchTeamHandler
//
// Partial update on the team with the given id.
// The data that will be updated will be retrieved from
// the request body. All valid keys from the request body
// will update the corresponding team's value with that
// key's value. To update the name of the team send
// body: {"name": "New Team Name"}. NOTE that the id and
// the hunt_id are not eligible to be changed.
//
// Consumes:
// 	- application/json
//
// Produces:
//	- application/json
//
// Schemes: http, https
//
// Responses:
// 	200:
// 	400:
func patchTeamHandler(env *c.Env) http.HandlerFunc {
	return (func(w http.ResponseWriter, r *http.Request) {
		teamID, e := request.GetIntURLParam(r, "teamID")
		if e != nil {
			e.Handle(w)
			return
		}

		partialTeam := PartialTeam{}
		partialTeam.ID = teamID

		e = request.DecodeAndValidate(r, &partialTeam)
		if e != nil {
			e.Handle(w)
			return
		}

		e = UpdateTeam(env, teamID, &partialTeam)
		if e != nil {
			e.Handle(w)
		}

		return
	})
}
