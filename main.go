package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/cljohnson4343/scavenge/models"

	"github.com/cljohnson4343/scavenge/hunts"
	"github.com/go-chi/chi"
)

// DBConfig is a custom type to store info used to configure postgresql db
type DBConfig struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	User     string `json:"user"`
	Password string `json:"password"`
	DBName   string `json:"dbname"`
}

// Routes inits a router
func Routes() *chi.Mux {
	router := chi.NewRouter()

	router.Route("/api/v0", func(r chi.Router) {
		r.Mount("/hunts", hunts.Routes())
	})

	return router
}

func main() {
	file, err := os.Open("./models/db_info.json")
	if err != nil {
		log.Panicf("Error configuring db: %s\n", err.Error())
	}
	defer file.Close()

	var dbConfig = new(DBConfig)
	err = json.NewDecoder(file).Decode(&dbConfig)
	if err != nil {
		log.Panicf("Error decoding config file: %s\n", err.Error())
	}

	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		dbConfig.Host, dbConfig.Port, dbConfig.User, dbConfig.Password, dbConfig.DBName)

	models.InitDB(psqlInfo)

	router := Routes()

	walkFunc := func(method string, route string, handler http.Handler, middlewares ...func(http.Handler) http.Handler) error {
		log.Printf("%s %s\n", method, route) // walk and print out all routes
		return nil
	}
	if err := chi.Walk(router, walkFunc); err != nil {
		log.Panicf("Logging err: %s\n", err.Error()) // panic if there is an error
	}

	log.Fatal(http.ListenAndServe(":4343", router))
}
