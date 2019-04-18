package config

import (
	"database/sql"

	_ "github.com/lib/pq"
)

// Env is a custom type that wraps the database and allows for
// methods to be added. It is needed to implement the DataStore
// interfaces of the other packages.
type Env struct {
	db *sql.DB
}

func (env *Env) DB() *sql.DB {
	return env.db
}

// CreateEnv instantiates a Env type
func CreateEnv(db *sql.DB) *Env {
	return &Env{db}
}
