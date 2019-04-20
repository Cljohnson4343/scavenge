package db

import (
	"database/sql"

	// @TODO look into whether this blank import is necessary. GoLint seems to
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

	return db, nil
}
