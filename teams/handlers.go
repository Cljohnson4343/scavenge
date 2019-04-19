package teams

import (
	"fmt"
	"net/http"
	"strconv"

	c "github.com/cljohnson4343/scavenge/config"
	"github.com/cljohnson4343/scavenge/response"
	"github.com/go-chi/chi"
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
func getTeamsHandler(env *c.Env) func(http.ResponseWriter, *http.Request) {
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
func getTeamHandler(env *c.Env) func(http.ResponseWriter, *http.Request) {
	return (func(w http.ResponseWriter, r *http.Request) {
		teamID, err := strconv.Atoi(chi.URLParam(r, "teamID"))
		if err != nil {
			e := response.NewError(err.Error(), http.StatusBadRequest)
			e.Handle(w)
			return
		}

		team, e := GetTeam(env, teamID)
		if e != nil {
			e.Handle(w)
			return
		}

		(*team).ID = teamID
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
func deleteTeamHandler(env *c.Env) func(http.ResponseWriter, *http.Request) {
	return (func(w http.ResponseWriter, r *http.Request) {
		teamID, err := strconv.Atoi(chi.URLParam(r, "teamID"))
		if err != nil {
			e := response.NewError(err.Error(), http.StatusBadRequest)
			e.Handle(w)
			return
		}

		e := DeleteTeam(env, teamID)
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
func createTeamHandler(env *c.Env) func(http.ResponseWriter, *http.Request) {
	return (func(w http.ResponseWriter, r *http.Request) {
		team := new(Team)
		err := render.DecodeJSON(r.Body, team)
		if err != nil {
			e := response.NewError(fmt.Sprintf("error decoding request json: %s", err.Error()), http.StatusBadRequest)
			e.Handle(w)
			return
		}

		if team.HuntID < 1 {
			e := response.NewError("'hunt_id' field is required for posting a team", http.StatusBadRequest)
			e.Handle(w)
			return
		}

		teamID, e := InsertTeam(env, team, team.HuntID)
		if e != nil {
			e.Handle(w)
			return
		}

		(*team).ID = teamID
		render.JSON(w, r, team)
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
func patchTeamHandler(env *c.Env) func(http.ResponseWriter, *http.Request) {
	return (func(w http.ResponseWriter, r *http.Request) {
		teamID, err := strconv.Atoi(chi.URLParam(r, "teamID"))
		if err != nil {
			e := response.NewError(err.Error(), http.StatusBadRequest)
			e.Handle(w)
			return
		}

		partialTeam := make(map[string]interface{})
		err = render.DecodeJSON(r.Body, &partialTeam)
		if err != nil {
			e := response.NewError(err.Error(), http.StatusBadRequest)
			e.Handle(w)
			return
		}

		e := UpdateTeam(env, teamID, &partialTeam)
		if e != nil {
			e.Handle(w)
		}

		return
	})
}
