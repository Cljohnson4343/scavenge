package hunts

import (
	"net/http"
	"os"
	"strconv"

	"github.com/cljohnson4343/scavenge/models"
	"github.com/go-chi/chi"
	"github.com/go-chi/render"
)

var hunts []models.Hunt

// Routes returns a router that serves the hunts routes
func Routes() *chi.Mux {
	file, err := os.Open("hunts/test_data.json")
	if err != nil {
		panic(err)
	}
	defer file.Close()

	render.DecodeJSON(file, &hunts)

	router := chi.NewRouter()

	router.Get("/", getHunts)
	router.Get("/{huntID}", getHunt)
	router.Post("/", createHunt)
	router.Delete("/{huntID}", deleteHunt)

	return router
}

func getHunts(w http.ResponseWriter, r *http.Request) {
	render.JSON(w, r, hunts)
	return
}

func getHunt(w http.ResponseWriter, r *http.Request) {
	huntID, err := strconv.Atoi(chi.URLParam(r, "huntID"))
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	for _, v := range hunts {
		if v.ID == huntID {
			render.JSON(w, r, v)
			return
		}
	}

	w.WriteHeader(http.StatusNotFound)
	w.Write([]byte("Oops! No such hunt."))
	return
}

func createHunt(w http.ResponseWriter, r *http.Request) {

}

func deleteHunt(w http.ResponseWriter, r *http.Request) {

}
