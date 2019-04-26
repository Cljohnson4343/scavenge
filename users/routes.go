// Package users Scavenge API
//
// This package provides all endpoints used to access/manipulate Users
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
package users

import (
	c "github.com/cljohnson4343/scavenge/config"
	"github.com/go-chi/chi"
)

// Routes returns a router that serves the users routes
func Routes(env *c.Env) *chi.Mux {
	router := chi.NewRouter()

	// /users routes
	router.Get("/{userID}", getSelectUserHandler(env))
	router.Post("/login/", getLoginHandler(env))
	router.Delete("/{userID}", getDeleteUserHandler(env))
	router.Patch("/{userID}", getUpdateUserHandler(env))

	return router
}
