package hunts

import (
	"log"
	"net/http"
	"strconv"

	"github.com/go-chi/chi"
	"github.com/go-chi/render"
)

func getHunts(ds HuntDataStore) func(http.ResponseWriter, *http.Request) {
	return (func(w http.ResponseWriter, r *http.Request) {
		hunts, err := ds.AllHunts()
		if err != nil {
			log.Printf("Failed to retrieve hunts: %s\n", err.Error())
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		render.JSON(w, r, hunts)
		return
	})
}

func getHunt(ds HuntDataStore) func(http.ResponseWriter, *http.Request) {
	return (func(w http.ResponseWriter, r *http.Request) {
		huntID, err := strconv.Atoi(chi.URLParam(r, "huntID"))
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		hunt := new(Hunt)
		err = ds.GetHunt(hunt, huntID)
		if err == nil {
			render.JSON(w, r, hunt)
			return
		}

		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("Oops! No such hunt."))
		return
	})
}

func createHunt(ds HuntDataStore) func(http.ResponseWriter, *http.Request) {
	return (func(w http.ResponseWriter, r *http.Request) {
		hunt := new(Hunt)
		err := render.DecodeJSON(r.Body, hunt)
		if err != nil {
			log.Printf("Unable to create hunt: %s\n", err.Error())
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		_, err = ds.InsertHunt(hunt)
		if err != nil {
			log.Printf("Unable to create hunt: %s\n", err.Error())
			return
		}

		return
	})
}

func deleteHunt(ds HuntDataStore) func(http.ResponseWriter, *http.Request) {
	return (func(w http.ResponseWriter, r *http.Request) {
		huntID, err := strconv.Atoi(chi.URLParam(r, "huntID"))
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		err = ds.DeleteHunt(huntID)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
		}

		return
	})
}
