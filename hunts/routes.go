package hunts

import (
	"database/sql"
	"os"

	"github.com/go-chi/chi"
)

type Env struct {
	db *sql.DB
}

// Routes returns a router that serves the hunts routes
func Routes(sqlDB *sql.DB) *chi.Mux {
	file, err := os.Open("hunts/test_data.json")
	if err != nil {
		panic(err)
	}
	defer file.Close()

	env := &Env{sqlDB}

	router := chi.NewRouter()

	router.Get("/", getHunts(env))
	router.Get("/{huntID}", getHunt(env))
	router.Post("/", createHunt(env))
	router.Delete("/{huntID}", deleteHunt(env))

	return router
}
