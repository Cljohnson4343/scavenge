package hunts

import (
	"log"
	"net/http"
	"strconv"

	c "github.com/cljohnson4343/scavenge/config"
	"github.com/cljohnson4343/scavenge/hunts/models"
	"github.com/go-chi/chi"
	"github.com/go-chi/render"
)

// swagger:route GET /hunts hunts getHuntsHandler
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
func getHuntsHandler(env *c.Env) func(http.ResponseWriter, *http.Request) {
	return (func(w http.ResponseWriter, r *http.Request) {
		hunts, err := AllHunts(env)
		if err != nil {
			log.Printf("Failed to retrieve hunts: %s\n", err.Error())
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		render.JSON(w, r, hunts)
		return
	})
}

// swagger:route GET /hunts/{huntID} hunt getHuntHandler
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
func getHuntHandler(env *c.Env) func(http.ResponseWriter, *http.Request) {
	return (func(w http.ResponseWriter, r *http.Request) {
		huntID, err := strconv.Atoi(chi.URLParam(r, "huntID"))
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		hunt := new(Hunt)
		err = GetHunt(env, hunt, huntID)
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

// swagger:route POST /hunts hunts createHuntHandler
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
func createHuntHandler(env *c.Env) func(http.ResponseWriter, *http.Request) {
	return (func(w http.ResponseWriter, r *http.Request) {
		hunt := new(Hunt)
		err := render.DecodeJSON(r.Body, hunt)
		if err != nil {
			log.Printf("error decoding request: %s\n", err.Error())
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		_, err = InsertHunt(env, hunt)
		if err != nil {
			log.Printf("Unable to create hunt: %s\n", err.Error())
			return
		}

		return
	})
}

// swagger:route DELETE /hunts/{huntID} hunt delete deleteHuntHandler
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
func deleteHuntHandler(env *c.Env) func(http.ResponseWriter, *http.Request) {
	return (func(w http.ResponseWriter, r *http.Request) {
		huntID, err := strconv.Atoi(chi.URLParam(r, "huntID"))
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		err = DeleteHunt(env, huntID)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
		}

		return
	})
}

// swagger:route PATCH /hunts/{huntID} hunt patchHuntHandler
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
func patchHuntHandler(env *c.Env) func(http.ResponseWriter, *http.Request) {
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

		rowsAffected, err := UpdateHunt(env, huntID, &partialHunt)
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

// swagger:route GET /hunts/{huntID}/items items getItemsHandler
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
func getItemsHandler(env *c.Env) func(http.ResponseWriter, *http.Request) {
	return (func(w http.ResponseWriter, r *http.Request) {
		huntID, err := strconv.Atoi(chi.URLParam(r, "huntID"))
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		items, err := GetItems(env, huntID)
		if err != nil {
			log.Printf("Failed to retrieve items for hunt %d: %s\n", huntID, err.Error())
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		render.JSON(w, r, items)
		return
	})
}

// swagger:route DELETE /hunts/{huntID}/items/{itemID} delete item deleteItemHandler
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
func deleteItemHandler(env *c.Env) func(http.ResponseWriter, *http.Request) {
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

		err = DeleteItem(env, huntID, itemID)
		if err != nil {
			log.Printf("error deleting item: %s\n", err.Error())
			w.WriteHeader(http.StatusBadRequest)
		}

		return
	})
}

// swagger:route POST /hunts/{huntID}/items item create createItemHandler
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
func createItemHandler(env *c.Env) func(http.ResponseWriter, *http.Request) {
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

		itemID, err := InsertItem(env, item, huntID)
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

// swagger:route PATCH /hunts/{huntID}/items/{itemID} item update partial patchItemHandler
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
func patchItemHandler(env *c.Env) func(http.ResponseWriter, *http.Request) {
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

		err = UpdateItem(env, huntID, itemID, &partialItem)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			log.Printf("error updating item: %s\n", err.Error())
			return
		}

		return
	})
}
