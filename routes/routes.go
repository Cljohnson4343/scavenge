package routes

import (
	"github.com/cljohnson4343/scavenge/config"
	"github.com/cljohnson4343/scavenge/hunts"
	"github.com/cljohnson4343/scavenge/response"
	"github.com/cljohnson4343/scavenge/teams"
	"github.com/cljohnson4343/scavenge/users"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
)

// Routes inits a router
func Routes(env *config.Env) *chi.Mux {
	router := chi.NewRouter()

	router.Route(config.BaseAPIURL, func(r chi.Router) {
		r.Use(middleware.RequestID)
		r.Use(middleware.Logger)
		r.Use(response.AllowCORS)

		r.Mount("/hunts", hunts.Routes(env))
		r.Mount("/teams", teams.Routes(env))
		r.Mount("/users", users.Routes(env))
	})

	return router
}
