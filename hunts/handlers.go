package hunts

import (
	"log"
	"net/http"
	"strconv"

	"github.com/cljohnson4343/scavenge/hunts/models"
	"github.com/go-chi/chi"
	"github.com/go-chi/render"
)

// swagger:route GET /hunts hunts getHunts
//
// Lists hunts.
//
// This will show all hunts by default.
//
// Consumes:
// 	- application/json
//
// Produces:
//	- application/json
//
// Schemes: http, https
//
// Responses:
// 	200:
//  500:
func getHunts(ds HuntDataStore) func(http.ResponseWriter, *http.Request) {
	return (func(w http.ResponseWriter, r *http.Request) {
		hunts, err := ds.allHunts()
		if err != nil {
			log.Printf("Failed to retrieve hunts: %s\n", err.Error())
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		render.JSON(w, r, hunts)
		return
	})
}

// swagger:route GET /hunts/{id} hunt getHunt
//
// Gets the hunt with given id.
//
// Consumes:
// 	- application/json
//
// Produces:
//	- application/json
//
// Schemes: http, https
//
// Responses:
// 	200:
// 	444:
//  500:
func getHunt(ds HuntDataStore) func(http.ResponseWriter, *http.Request) {
	return (func(w http.ResponseWriter, r *http.Request) {
		huntID, err := strconv.Atoi(chi.URLParam(r, "huntID"))
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		hunt := new(models.Hunt)
		err = ds.getHunt(hunt, huntID)
		if err == nil {
			render.JSON(w, r, hunt)
			return
		}

		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("Oops! No such hunt."))
		return
	})
}

// swagger:route POST /hunts hunts createHunt
//
// Creates the given hunt.
//
// Consumes:
// 	- application/json
//
// Produces:
//	- application/json
//
// Schemes: http, https
//
// Responses:
// 	200:
//  500:
func createHunt(ds HuntDataStore) func(http.ResponseWriter, *http.Request) {
	return (func(w http.ResponseWriter, r *http.Request) {
		hunt := new(models.Hunt)
		err := render.DecodeJSON(r.Body, hunt)
		if err != nil {
			log.Printf("Unable to create hunt: %s\n", err.Error())
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		_, err = ds.insertHunt(hunt)
		if err != nil {
			log.Printf("Unable to create hunt: %s\n", err.Error())
			return
		}

		return
	})
}

// swagger:route DELETE /hunts/{id} hunt deleteHunt
//
// Deletes the given hunt.
//
// Consumes:
// 	- application/json
//
// Produces:
//	- application/json
//
// Schemes: http, https
//
// Responses:
// 	200:
//  400:
func deleteHunt(ds HuntDataStore) func(http.ResponseWriter, *http.Request) {
	return (func(w http.ResponseWriter, r *http.Request) {
		huntID, err := strconv.Atoi(chi.URLParam(r, "huntID"))
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		err = ds.deleteHunt(huntID)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
		}

		return
	})
}
