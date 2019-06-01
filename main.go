package main

import (
	"flag"
	"log"
	"net/http"

	"github.com/cljohnson4343/scavenge/config"
	"github.com/cljohnson4343/scavenge/db"
	"github.com/cljohnson4343/scavenge/populate"
	"github.com/cljohnson4343/scavenge/response"
	"github.com/cljohnson4343/scavenge/routes"
	"github.com/go-chi/chi"
)

// Flags for scavenge
var devModeFlag bool
var populateFlag bool

func init() {
	flag.BoolVar(&devModeFlag, "dev-mode", false, "set the server to dev mode")
	flag.BoolVar(&populateFlag, "populate", false, "populate the database with dummy data")

}

func main() {
	flag.Parse()
	if devModeFlag {
		response.SetDevMode(true)
	}

	database := db.InitDB("./db/db_info.json")
	defer db.Shutdown(database)

	env := config.CreateEnv(database)
	router := routes.Routes(env)

	populate.Populate(populateFlag)

	walkFunc := func(method string, route string, handler http.Handler, middlewares ...func(http.Handler) http.Handler) error {
		log.Printf("%s %s\n", method, route) // walk and print out all routes
		return nil
	}
	if err := chi.Walk(router, walkFunc); err != nil {
		log.Panicf("Logging err: %s\n", err.Error()) // panic if there is an error
	}

	log.Fatal(http.ListenAndServe(":4343", router))
}
