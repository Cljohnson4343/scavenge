package routes

import (
	"github.com/cljohnson4343/scavenge/config"
	"github.com/cljohnson4343/scavenge/hunts"
	"github.com/cljohnson4343/scavenge/roles"
	"github.com/cljohnson4343/scavenge/teams"
	"github.com/cljohnson4343/scavenge/users"
	"github.com/go-chi/chi"
)

// Routes inits a router
func Routes(env *config.Env) *chi.Mux {
	router := chi.NewRouter()

	router.Use(users.WithUser)
	router.Use(roles.RequireAuth)

	router.Route(config.BaseAPIURL, func(r chi.Router) {
		r.Mount("/hunts", hunts.Routes(env))
		r.Mount("/teams", teams.Routes(env))
		r.Mount("/users", users.Routes(env))
	})

	return router
}
