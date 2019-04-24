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
	router.Get("/", getTeamsHandler(env))
	router.Get("/{teamID}", getTeamHandler(env))
	router.Delete("/{teamID}", deleteTeamHandler(env))
	router.Post("/", createTeamHandler(env))
	router.Patch("/{teamID}", patchTeamHandler(env))

	// location routes
	router.Get("/{teamID}/locations/", getLocationsForTeamHandler(env))
	router.Post("/{teamID}/locations/", createLocationHandler(env))
	router.Delete("/{teamID}/locations/{locationID}", deleteLocationHandler(env))

	// media routes
	router.Get("/{teamID}/media/", getMediaForTeamHandler(env))
	router.Post("/{teamID}/media/", createMediaHandler(env))
	router.Delete("/{teamID}/media/{mediaID}", deleteMediaHandler(env))
	router.Post("/populate/", populateMediaDBHandler(env))

	return router
}
