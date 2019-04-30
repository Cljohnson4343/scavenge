// Package teams Scavenge API
//
// This package provides all endpoints used to access/manipulate Teams
//
// Terms Of Service:
//
// there are no TOS at this moment, use at your own risk we take no responsibility
//
//     Schemes: http, https
//     Host: localhost
//     BasePath: /api/v0
//     Version: 0.0.1
//     Contact: Chris Johnson<cljohnson4343@gmail.com>
//
//     Consumes:
//     - application/json
//
//     Produces:
//     - application/json
//
//     Security:
//			Not yet implemented
//
// swagger:meta
package teams

import (
	c "github.com/cljohnson4343/scavenge/config"
	"github.com/go-chi/chi"
)

// Routes returns a router that serves the teams routes
func Routes(env *c.Env) *chi.Mux {
	router := chi.NewRouter()

	// /teams routes
	router.Get("/", getTeamsHandler(env))        // tested
	router.Get("/{teamID}", getTeamHandler(env)) // tested
	router.Get("/{teamID}/points/", getTeamPointsHandler(env))
	router.Get("/{teamID}/players/", getTeamPlayersHandler(env))
	router.Post("/{teamID}/join/", getJoinTeamHandler(env))
	router.Post("/{teamID}/remove/", getRemovePlayersHandler(env))
	router.Delete("/{teamID}", deleteTeamHandler(env)) // tested
	router.Post("/", createTeamHandler(env))           // tested
	router.Patch("/{teamID}", patchTeamHandler(env))

	// location routes
	router.Get("/{teamID}/locations/", getLocationsForTeamHandler(env))
	router.Post("/{teamID}/locations/", createLocationHandler(env))               // tested
	router.Delete("/{teamID}/locations/{locationID}", deleteLocationHandler(env)) // tested

	// media routes
	router.Get("/{teamID}/media/", getMediaForTeamHandler(env))         // tested
	router.Post("/{teamID}/media/", createMediaHandler(env))            // tested
	router.Delete("/{teamID}/media/{mediaID}", deleteMediaHandler(env)) // tested
	router.Post("/populate/", populateMediaDBHandler(env))

	return router
}
