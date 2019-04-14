package models

import (
	"database/sql"
	"log"

	_ "github.com/lib/pq"
)

var db *sql.DB

// InitDB initializes the database
func InitDB(dataSourceName string) {
	var err error
	db, err = sql.Open("postgres", dataSourceName)
	if err != nil {
		log.Panicf("Error connecting to postgresql db: %s\n", err.Error())
	}

	log.Print("Pinging...")
	if err = db.Ping(); err != nil {
		log.Panicf("Error pinging postresql db: %s\n", err.Error())
	}
}

func CloseDB() error {
	return db.Close()
}
