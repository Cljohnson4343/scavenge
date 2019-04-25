package main

import (
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"

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

	file, err := os.Open("./db/db_info.json")
	if err != nil {
		log.Panicf("Error configuring db: %s\n", err.Error())
	}
	defer file.Close()

	var dbConfig = new(db.Config)
	err = json.NewDecoder(file).Decode(&dbConfig)
	if err != nil {
		log.Panicf("Error decoding config file: %s\n", err.Error())
	}

	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		dbConfig.Host, dbConfig.Port, dbConfig.User, dbConfig.Password, dbConfig.DBName)

	database, err := db.NewDB(psqlInfo)
	if err != nil {
		log.Panicf("Error initializing the db: %s\n", err.Error())
	}
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
