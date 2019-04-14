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
	"os"

	"github.com/go-chi/chi"
)

// Env is a custom type that wraps the database and allows for
// methods to be added. It is needed to implement the HuntDataStore
// interface.
type Env struct {
	db *sql.DB
}

// Routes returns a router that serves the hunts routes
func Routes(sqlDB *sql.DB) *chi.Mux {
	file, err := os.Open("hunts/test_data.json")
	if err != nil {
		panic(err)
	}
	defer file.Close()

	env := &Env{sqlDB}

	router := chi.NewRouter()

	router.Get("/", getHunts(env))
	router.Get("/{huntID}", getHunt(env))
	router.Post("/", createHunt(env))
	router.Delete("/{huntID}", deleteHunt(env))

	return router
}
