package db

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"os"

	// TODO look into whether this blank import is necessary. GoLint seems to
	// have a problem with it.
	_ "github.com/lib/pq"
)

// Config is a custom type to store info used to configure postgresql db
type Config struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	User     string `json:"user"`
	Password string `json:"password"`
	DBName   string `json:"dbname"`
}

// NewDB returns a newly initialized database that uses the given config.
func NewDB(dataSourceName string) (*sql.DB, error) {
	var err error
	db, err := sql.Open("postgres", dataSourceName)
	if err != nil {
		return nil, err
	}

	if err = db.Ping(); err != nil {
		return nil, err
	}

	err = initStatements(db)
	if err != nil {
		return nil, err
	}

	return db, nil
}

// Shutdown closes all db connections
func Shutdown(db *sql.DB) {

	for _, v := range stmtMap {
		v.Close()
	}

	err := db.Close()
	if err != nil {
		panic(err)
	}

}

// InitDB initializes a db and returns the db. The caller is responible for closing the
// db.
func InitDB(filePath string) *sql.DB {
	file, err := os.Open(filePath)
	if err != nil {
		log.Panicf("Error configuring db: %s\n", err.Error())
	}
	defer file.Close()

	var dbConfig = new(Config)
	err = json.NewDecoder(file).Decode(&dbConfig)
	if err != nil {
		log.Panicf("Error decoding config file: %s\n", err.Error())
	}

	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		dbConfig.Host, dbConfig.Port, dbConfig.User, dbConfig.Password, dbConfig.DBName)

	database, err := NewDB(psqlInfo)
	if err != nil {
		log.Panicf("Error initializing the db: %s\n", err.Error())
	}

	return database
}
