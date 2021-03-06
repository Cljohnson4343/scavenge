package config

import (
	"database/sql"

	// necessary for database/sql package
	_ "github.com/lib/pq"
)

// Env is a custom type that wraps the database and allows for
// methods to be added. It is needed to implement the DataStore
// interfaces of the other packages.
type Env struct {
	*sql.DB
}

// CreateEnv instantiates a Env type
func CreateEnv(db *sql.DB) *Env {
	return &Env{db}
}

// BaseAPIURL is the base of the api's url.
const BaseAPIURL = `/api/v0/`
