package models

import (
	"database/sql"
	"log"

	_ "github.com/lib/pq"
)

var db *sql.DB

// InitDB initializes the database
func InitDB(dataSourceName string) {
	db, err := sql.Open("postgres", dataSourceName)
	if err != nil {
		log.Panicf("Error connecting to postgresql db: %s\n", err.Error())
	}
	defer db.Close()

	if err = db.Ping(); err != nil {
		log.Panicf("Error pinging postresql db: %s\n", err.Error())
	}
}
