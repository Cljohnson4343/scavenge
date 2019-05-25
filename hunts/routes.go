// Package hunts Scavenge API
//
// This package provides all endpoints used to access/manipulate Hunts
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
package hunts

import (
	"github.com/cljohnson4343/scavenge/config"
	"github.com/cljohnson4343/scavenge/users"
	"github.com/go-chi/chi"
)

// Routes returns a router that serves the hunts routes
func Routes(env *config.Env) *chi.Mux {
	router := chi.NewRouter()

	router.Use(users.WithUser)
	router.Use(users.RequireAuth)

	// /hunts routes
	router.Get("/", getHuntsHandler())
	router.Get("/{huntID}", getHuntHandler(env))
	router.Post("/", createHuntHandler(env))           // tested
	router.Delete("/{huntID}", deleteHuntHandler(env)) // tested
	router.Patch("/{huntID}", patchHuntHandler(env))
	router.Post("/populate/", populateDBHandler(env))

	// /hunts/{huntID}/items routes
	router.Get("/{huntID}/items/", getItemsHandler(env))
	router.Delete("/{huntID}/items/{itemID}", deleteItemHandler(env))
	router.Post("/{huntID}/items/", createItemHandler(env))
	router.Patch("/{huntID}/items/{itemID}", patchItemHandler(env))

	router.Get("/{huntID}/players/", getHuntPlayersHandler())
	router.Post("/{huntID}/players/", addHuntPlayersHandler())

	router.Post("/{huntID}/invitations/", createHuntInvitationHandler())

	return router
}
