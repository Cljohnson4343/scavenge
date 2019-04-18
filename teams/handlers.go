package teams

import (
	"log"
	"net/http"
	"strconv"

	c "github.com/cljohnson4343/scavenge/config"
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
		teams, err := GetTeams(env)
		if err != nil {
			log.Printf("Failed to retrieve teams: %s\n", err.Error())
			w.WriteHeader(http.StatusInternalServerError)
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
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		team, err := GetTeam(env, teamID)

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
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		err = DeleteTeam(env, teamID)
		if err != nil {
			log.Printf("Error deleting team: %s\n", err.Error())
			w.WriteHeader(http.StatusBadRequest)
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
			log.Printf("Unable to create team: %s\n", err.Error())
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		teamID, err := InsertTeam(env, team, team.HuntID)
		if err != nil {
			log.Printf("Error creating a team: %s\n", err.Error())
			w.WriteHeader(http.StatusInternalServerError)
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
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		partialTeam := make(map[string]interface{})
		err = render.DecodeJSON(r.Body, &partialTeam)
		if err != nil {
			log.Printf("unable to patch team: %s\n", err.Error())
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		err = UpdateTeam(env, teamID, &partialTeam)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			log.Printf("error patching team: %s\n", err.Error())
			return
		}

		return
	})
}
