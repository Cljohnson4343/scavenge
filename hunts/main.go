package hunts

import (
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/go-chi/chi"
	"github.com/go-chi/render"
)

type Coord struct {
	Latitude  float32 `json:"latitude"`
	Longitude float32 `json:"longitude"`
}

type Location struct {
	Name   string `json:"name"`
	Coords Coord  `json:"coords"`
}

type User struct {
	First string `json:"first"`
	Last  string `json:"last"`
}
type Team struct {
	Name string `json:"name"`
}

type Item struct {
	Name   string `json:"name"`
	Points uint   `json:"points"`
	IsDone bool   `json:"is_done"`
}

type Hunt struct {
	Title    string    `json:"title"`
	MaxTeams int       `json:"max_teams"`
	ID       int       `json:"id"`
	Start    time.Time `json:"start"`
	End      time.Time `json:"end"`
	Teams    []Team    `json:"teams"`
	Items    []Item    `json:"items"`
	Location Location  `json:"location"`
}

var hunts []Hunt

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
