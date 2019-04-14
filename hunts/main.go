package hunts

import (
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/cljohnson4343/scavenge/models"
	"github.com/go-chi/chi"
	"github.com/go-chi/render"
)

// Routes returns a router that serves the hunts routes
func Routes() *chi.Mux {
	file, err := os.Open("hunts/test_data.json")
	if err != nil {
		panic(err)
	}
	defer file.Close()

	router := chi.NewRouter()

	router.Get("/", getHunts)
	router.Get("/{huntID}", getHunt)
	router.Post("/", createHunt)
	router.Delete("/{huntID}", deleteHunt)

	return router
}

func getHunts(w http.ResponseWriter, r *http.Request) {
	hunts, err := models.AllHunts()
	if err != nil {
		log.Printf("Failed to retrieve hunts: %s\n", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	render.JSON(w, r, hunts)
	return
}

func getHunt(w http.ResponseWriter, r *http.Request) {
	huntID, err := strconv.Atoi(chi.URLParam(r, "huntID"))
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	hunt := new(models.Hunt)
	err = models.GetHunt(hunt, huntID)
	if err == nil {
		render.JSON(w, r, hunt)
		return
	}

	w.WriteHeader(http.StatusNotFound)
	w.Write([]byte("Oops! No such hunt."))
	return
}

func createHunt(w http.ResponseWriter, r *http.Request) {
	hunt := new(models.Hunt)
	err := render.DecodeJSON(r.Body, hunt)
	if err != nil {
		log.Printf("Unable to create hunt: %s\n", err.Error())
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	_, err = models.InsertHunt(hunt)
	if err != nil {
		log.Printf("Unable to create hunt: %s\n", err.Error())
		return
	}

	return
}

func deleteHunt(w http.ResponseWriter, r *http.Request) {

}
