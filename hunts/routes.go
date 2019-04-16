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
	"database/sql"

	"github.com/go-chi/chi"
)

// Env is a custom type that wraps the database and allows for
// methods to be added. It is needed to implement the HuntDataStore
// interface.
type Env struct {
	db *sql.DB
}

// CreateEnv instantiates a Env type
func CreateEnv(db *sql.DB) *Env {
	return &Env{db}
}

// Routes returns a router that serves the hunts routes
func Routes(env HuntDataStore) *chi.Mux {
	router := chi.NewRouter()

	// /hunts routes
	router.Get("/", getHunts(env))
	router.Get("/{huntID}", getHunt(env))
	router.Post("/", createHunt(env))
	router.Delete("/{huntID}", deleteHunt(env))
	router.Patch("/{huntID}", patchHunt(env))

	// /hunts/{huntID}/teams routes
	router.Get("/{huntID}/teams/", getTeams(env))
	router.Get("/{huntID}/teams/{teamID}", getTeam(env))
	router.Delete("/{huntID}/teams/{teamID}", deleteTeam(env))
	router.Post("/{huntID}/teams/", createTeam(env))

	return router
}
