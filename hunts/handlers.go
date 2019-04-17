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

// swagger:route GET /hunts/{huntID} hunt getHunt
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
// 	404:
//  400:
func getHunt(ds HuntDataStore) func(http.ResponseWriter, *http.Request) {
	return (func(w http.ResponseWriter, r *http.Request) {
		huntID, err := strconv.Atoi(chi.URLParam(r, "huntID"))
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		hunt := new(models.Hunt)
		err = ds.getHunt(hunt, huntID)
		if err != nil {
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte("Oops! No such hunt."))
			return
		}

		(*hunt).ID = huntID
		render.JSON(w, r, hunt)
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
//  400:
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

// swagger:route DELETE /hunts/{huntID} hunt delete deleteHunt
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

// swagger:route PATCH /hunts/{huntID} hunt patchHunt
//
// Partial update on the hunt with the given id.
// The data that will be updated will be retrieved from
// the request body. All valid keys from the request body
// will update the corresponding hunt's value with that
// key's value. To update the name of the hunt send
// body: {"title": "New Hunt Name"}.
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
// 	400:
// 	404:
// 	500:
func patchHunt(ds HuntDataStore) func(http.ResponseWriter, *http.Request) {
	return (func(w http.ResponseWriter, r *http.Request) {
		huntID, err := strconv.Atoi(chi.URLParam(r, "huntID"))
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		partialHunt := make(map[string]interface{})
		err = render.DecodeJSON(r.Body, &partialHunt)
		if err != nil {
			log.Printf("Unable to patch hunt: %s\n", err.Error())
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		rowsAffected, err := ds.updateHunt(huntID, &partialHunt)
		if err != nil {
			log.Printf("Error patching hunt: %s\n", err.Error())
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		if !rowsAffected {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		render.JSON(w, r, partialHunt)
		return
	})
}

// swagger:route GET /hunts/{huntID}/teams teams getTeams
//
// Lists the teams for {huntID}.
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
// 	400:
//  500:
func getTeams(ds HuntDataStore) func(http.ResponseWriter, *http.Request) {
	return (func(w http.ResponseWriter, r *http.Request) {
		huntID, err := strconv.Atoi(chi.URLParam(r, "huntID"))
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		teams, err := ds.getTeams(huntID)
		if err != nil {
			log.Printf("Failed to retrieve teams for hunt %d: %s\n", huntID, err.Error())
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		render.JSON(w, r, teams)
		return
	})
}

// swagger:route GET /hunts/{huntID}/teams/{teamID} team getTeam
//
// Gets the team with given id.
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
// 	400:
// 	404:
func getTeam(ds HuntDataStore) func(http.ResponseWriter, *http.Request) {
	return (func(w http.ResponseWriter, r *http.Request) {
		huntID, err := strconv.Atoi(chi.URLParam(r, "huntID"))
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		teamID, err := strconv.Atoi(chi.URLParam(r, "teamID"))
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		team, err := ds.getTeam(teamID)
		if err != nil || team.HuntID != huntID {
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte("Oops! No such team."))
			return
		}

		(*team).ID = teamID
		render.JSON(w, r, team)
		return
	})
}

// swagger:route DELETE /hunts/{huntID}/teams/{teamID} delete team deleteTeam
//
// Deletes the given team.
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
func deleteTeam(ds HuntDataStore) func(http.ResponseWriter, *http.Request) {
	return (func(w http.ResponseWriter, r *http.Request) {
		huntID, err := strconv.Atoi(chi.URLParam(r, "huntID"))
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		teamID, err := strconv.Atoi(chi.URLParam(r, "teamID"))
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		err = ds.deleteTeam(huntID, teamID)
		if err != nil {
			log.Printf("Error deleting team: %s\n", err.Error())
			w.WriteHeader(http.StatusBadRequest)
		}

		return
	})
}

// swagger:route POST /hunts/{huntID}/teams team create createTeam
//
// Creates the given team for the given hunt.
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
//  500:
func createTeam(ds HuntDataStore) func(http.ResponseWriter, *http.Request) {
	return (func(w http.ResponseWriter, r *http.Request) {
		huntID, err := strconv.Atoi(chi.URLParam(r, "huntID"))
		if err != nil {
			log.Printf("Error creating a team: %s\n", err.Error())
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		team := new(models.Team)
		err = render.DecodeJSON(r.Body, team)
		if err != nil {
			log.Printf("Unable to create team: %s\n", err.Error())
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		teamID, err := ds.insertTeam(team, huntID)
		if err != nil {
			log.Printf("Error creating a team: %s\n", err.Error())
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		(*team).ID = teamID
		render.JSON(w, r, team)
		return
	})
}

// swagger:route PATCH /hunts/{huntID}/teams/{teamID} team patchTeam
//
// Partial update on the team with the given id.
// The data that will be updated will be retrieved from
// the request body. All valid keys from the request body
// will update the corresponding team's value with that
// key's value. To update the name of the team send
// body: {"name": "New Team Name"}. NOTE that the id and
// the hunt_id are not eligible to be changed.
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
// 	400:
func patchTeam(ds HuntDataStore) func(http.ResponseWriter, *http.Request) {
	return (func(w http.ResponseWriter, r *http.Request) {
		huntID, err := strconv.Atoi(chi.URLParam(r, "huntID"))
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		teamID, err := strconv.Atoi(chi.URLParam(r, "teamID"))
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		partialTeam := make(map[string]interface{})
		err = render.DecodeJSON(r.Body, &partialTeam)
		if err != nil {
			log.Printf("unable to patch team: %s\n", err.Error())
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		err = ds.updateTeam(huntID, teamID, &partialTeam)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			log.Printf("error patching team: %s\n", err.Error())
			return
		}

		return
	})
}

// swagger:route GET /hunts/{huntID}/items items getItems
//
// Lists the items for hunt with {huntID}.
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
// 	400:
//  500:
func getItems(ds HuntDataStore) func(http.ResponseWriter, *http.Request) {
	return (func(w http.ResponseWriter, r *http.Request) {
		huntID, err := strconv.Atoi(chi.URLParam(r, "huntID"))
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		items, err := ds.getItems(huntID)
		if err != nil {
			log.Printf("Failed to retrieve items for hunt %d: %s\n", huntID, err.Error())
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		render.JSON(w, r, items)
		return
	})
}

// swagger:route DELETE /hunts/{huntID}/items/{itemID} delete item deleteItem
//
// Deletes the given item.
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
func deleteItem(ds HuntDataStore) func(http.ResponseWriter, *http.Request) {
	return (func(w http.ResponseWriter, r *http.Request) {
		huntID, err := strconv.Atoi(chi.URLParam(r, "huntID"))
		if err != nil {
			log.Printf("error deleting item: %s\n", err.Error())
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		itemID, err := strconv.Atoi(chi.URLParam(r, "itemID"))
		if err != nil {
			log.Printf("error deleting item: %s\n", err.Error())
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		err = ds.deleteItem(huntID, itemID)
		if err != nil {
			log.Printf("error deleting item: %s\n", err.Error())
			w.WriteHeader(http.StatusBadRequest)
		}

		return
	})
}

// swagger:route POST /hunts/{huntID}/items item create createItem
//
// Creates the item described in the request body for the given hunt.
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
//  500:
func createItem(ds HuntDataStore) func(http.ResponseWriter, *http.Request) {
	return (func(w http.ResponseWriter, r *http.Request) {
		huntID, err := strconv.Atoi(chi.URLParam(r, "huntID"))
		if err != nil {
			log.Printf("Error creating item: %s\n", err.Error())
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		item := new(models.Item)
		err = render.DecodeJSON(r.Body, item)
		if err != nil {
			log.Printf("Unable to create item: %s\n", err.Error())
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		itemID, err := ds.insertItem(item, huntID)
		if err != nil {
			log.Printf("Error creating a item: %s\n", err.Error())
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		(*item).ID = itemID
		render.JSON(w, r, item)
		return
	})
}

// swagger:route PATCH /hunts/{huntID}/items/{itemID} item update partial patchItem
//
// Partial update on the item with the given id.
// The data that will be updated will be retrieved from
// the request body. All valid keys from the request body
// will update the corresponding item's value with that
// key's value. To update the name of the item send
// body: {"name": "New Item Name"}. NOTE that the id and hunt_id
// are not eligible to be changed
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
// 	400:
func patchItem(ds HuntDataStore) func(http.ResponseWriter, *http.Request) {
	return (func(w http.ResponseWriter, r *http.Request) {
		huntID, err := strconv.Atoi(chi.URLParam(r, "huntID"))
		if err != nil {
			log.Printf("error updating item: %s\n", err.Error())
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		itemID, err := strconv.Atoi(chi.URLParam(r, "itemID"))
		if err != nil {
			log.Printf("error updating item: %s\n", err.Error())
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		partialItem := make(map[string]interface{})
		err = render.DecodeJSON(r.Body, &partialItem)
		if err != nil {
			log.Printf("error updating item: %s\n", err.Error())
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		err = ds.updateItem(huntID, itemID, &partialItem)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			log.Printf("error updating item: %s\n", err.Error())
			return
		}

		return
	})
}
