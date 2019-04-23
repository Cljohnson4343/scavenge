package teams

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/cljohnson4343/scavenge/db"

	"github.com/cljohnson4343/scavenge/response"

	"github.com/go-chi/chi"

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

		e = InsertTeam(env, &team)
		if e != nil {
			e.Handle(w)
			return
		}

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

		team := Team{}

		e = request.DecodeAndPatchValidate(r, &team, teamID)
		if e != nil {
			e.Handle(w)
			return
		}

		e = UpdateTeam(env, &team)
		if e != nil {
			e.Handle(w)
		}

		return
	})
}

// swagger:route GET /teams/{teamID}/locations/ locations getLocationsForTeamHandler
//
// Lists all the locations for a team.
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
func getLocationsForTeamHandler(env *c.Env) http.HandlerFunc {
	return (func(w http.ResponseWriter, r *http.Request) {
		e := response.NewNilError()
		teamID, err := strconv.Atoi(chi.URLParam(r, "teamID"))
		if err != nil {
			e.Add(fmt.Sprintf("error getting teamID URL param: %s",
				err.Error()), http.StatusBadRequest)

			e.GetError().Handle(w)
			return
		}

		locationDBs, e := db.GetLocationsForTeam(teamID)
		if e != nil {
			e.Handle(w)
		}

		render.JSON(w, r, locationDBs)
		return
	})
}

// swagger:route POST /teams/{teamID}/locations/ location create createLocationHandler
//
// Creates the given location.
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
func createLocationHandler(env *c.Env) http.HandlerFunc {
	return (func(w http.ResponseWriter, r *http.Request) {
		teamID, err := strconv.Atoi(chi.URLParam(r, "teamID"))
		if err != nil {
			e := response.NewError(fmt.Sprintf("error getting teamID from URL: %s",
				err.Error()), http.StatusBadRequest)
			e.Handle(w)
			return
		}

		location := db.LocationDB{}

		e := request.DecodeAndValidate(r, &location)
		if e != nil {
			e.Handle(w)
			return
		}

		e = location.Insert(teamID)
		if e != nil {
			e.Handle(w)
			return
		}

		render.JSON(w, r, &location)
		return
	})
}
