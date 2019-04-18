package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	c "github.com/cljohnson4343/scavenge/config"
	"github.com/cljohnson4343/scavenge/db"
	"github.com/cljohnson4343/scavenge/hunts"
	"github.com/cljohnson4343/scavenge/teams"
	"github.com/go-chi/chi"
)

// Routes inits a router
func Routes(db *sql.DB) *chi.Mux {
	router := chi.NewRouter()

	env := c.CreateEnv(db)
	router.Route("/api/v0", func(r chi.Router) {
		r.Mount("/hunts", hunts.Routes(env))
		r.Mount("/teams", teams.Routes(env))
	})

	return router
}

func main() {
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

	db, err := db.NewDB(psqlInfo)
	if err != nil {
		log.Panicf("Error initializing the db: %s\n", err.Error())
	}

	router := Routes(db)

	walkFunc := func(method string, route string, handler http.Handler, middlewares ...func(http.Handler) http.Handler) error {
		log.Printf("%s %s\n", method, route) // walk and print out all routes
		return nil
	}
	if err := chi.Walk(router, walkFunc); err != nil {
		log.Panicf("Logging err: %s\n", err.Error()) // panic if there is an error
	}

	log.Fatal(http.ListenAndServe(":4343", router))
}
