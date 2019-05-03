package main

import (
	"database/sql"
	"flag"
	"log"
	"net/http"

	"github.com/cljohnson4343/scavenge/response"

	c "github.com/cljohnson4343/scavenge/config"
	"github.com/cljohnson4343/scavenge/db"
	"github.com/cljohnson4343/scavenge/hunts"
	"github.com/cljohnson4343/scavenge/teams"
	"github.com/cljohnson4343/scavenge/users"
	"github.com/go-chi/chi"
)

// Flags for scavenge
var devModeFlag bool

func init() {
	flag.BoolVar(&devModeFlag, "dev-mode", false, "set the server to dev mode")
}

// Routes inits a router
func Routes(db *sql.DB) *chi.Mux {
	router := chi.NewRouter()

	env := c.CreateEnv(db)
	router.Route("/api/v0", func(r chi.Router) {
		r.Mount("/hunts", hunts.Routes(env))
		r.Mount("/teams", teams.Routes(env))
		r.Mount("/users", users.Routes(env))
	})

	return router
}

func main() {
	flag.Parse()
	if devModeFlag {
		response.SetDevMode(true)
	}

	database := db.InitDB("./db/db_info.json")
	defer db.Shutdown(database)

	router := Routes(database)

	walkFunc := func(method string, route string, handler http.Handler, middlewares ...func(http.Handler) http.Handler) error {
		log.Printf("%s %s\n", method, route) // walk and print out all routes
		return nil
	}
	if err := chi.Walk(router, walkFunc); err != nil {
		log.Panicf("Logging err: %s\n", err.Error()) // panic if there is an error
	}

	log.Fatal(http.ListenAndServe(":4343", router))
}
