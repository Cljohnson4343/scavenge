package main

import (
	"log"
	"net/http"

	"github.com/cljohnson4343/scavenge/hunts"
	"github.com/go-chi/chi"
)

// Routes inits a router
func Routes() *chi.Mux {
	router := chi.NewRouter()

	router.Route("/v0", func(r chi.Router) {
		r.Mount("/api/hunts", hunts.Routes())
	})

	return router
}

func main() {
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
