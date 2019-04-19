package hunts

import (
	"net/http"
	"strconv"

	"github.com/cljohnson4343/scavenge/request"

	"github.com/cljohnson4343/scavenge/hunts/models"
	"github.com/cljohnson4343/scavenge/response"

	c "github.com/cljohnson4343/scavenge/config"
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
func getHuntsHandler(env *c.Env) http.HandlerFunc {
	return (func(w http.ResponseWriter, r *http.Request) {
		hunts, e := AllHunts(env)
		if e != nil {
			e.Handle(w)
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
func getHuntHandler(env *c.Env) http.HandlerFunc {
	return (func(w http.ResponseWriter, r *http.Request) {
		huntID, err := strconv.Atoi(chi.URLParam(r, "huntID"))
		if err != nil {
			e := response.NewError(err.Error(), http.StatusBadRequest)
			e.Handle(w)
			return
		}

		hunt := Hunt{}
		e := GetHunt(env, &hunt, huntID)
		if e != nil {
			e.Handle(w)
			return
		}

		hunt.ID = huntID
		render.JSON(w, r, &hunt)
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
func createHuntHandler(env *c.Env) http.HandlerFunc {
	return (func(w http.ResponseWriter, r *http.Request) {
		hunt := Hunt{}
		e := request.DecodeAndValidate(r, &hunt)
		if e != nil {
			e.Handle(w)
			return
		}

		_, e = InsertHunt(env, &hunt)
		if e != nil {
			e.Handle(w)
		}

		render.JSON(w, r, &hunt)
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
func deleteHuntHandler(env *c.Env) http.HandlerFunc {
	return (func(w http.ResponseWriter, r *http.Request) {
		huntID, err := strconv.Atoi(chi.URLParam(r, "huntID"))
		if err != nil {
			e := response.NewError(err.Error(), http.StatusBadRequest)
			e.Handle(w)
			return
		}

		e := DeleteHunt(env, huntID)
		if e != nil {
			e.Handle(w)
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
func patchHuntHandler(env *c.Env) http.HandlerFunc {
	return (func(w http.ResponseWriter, r *http.Request) {
		huntID, err := strconv.Atoi(chi.URLParam(r, "huntID"))
		if err != nil {
			e := response.NewError(err.Error(), http.StatusBadRequest)
			e.Handle(w)
			return
		}

		partialHunt := make(map[string]interface{})
		err = render.DecodeJSON(r.Body, partialHunt)
		if err != nil {
			e := response.NewError(err.Error(), http.StatusBadRequest)
			e.Handle(w)
			return
		}

		rowsAffected, e := UpdateHunt(env, huntID, &partialHunt)
		if e != nil {
			e.Handle(w)
			return
		}

		if !rowsAffected {
			e := response.NewError("cannot patch a hunt that doesn't exist", http.StatusBadRequest)
			e.Handle(w)
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
func getItemsHandler(env *c.Env) http.HandlerFunc {
	return (func(w http.ResponseWriter, r *http.Request) {
		huntID, err := strconv.Atoi(chi.URLParam(r, "huntID"))
		if err != nil {
			e := response.NewError(err.Error(), http.StatusBadRequest)
			e.Handle(w)
			return
		}

		items, e := GetItems(env, huntID)
		if e != nil {
			e.Handle(w)
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
func deleteItemHandler(env *c.Env) http.HandlerFunc {
	return (func(w http.ResponseWriter, r *http.Request) {
		huntID, err := strconv.Atoi(chi.URLParam(r, "huntID"))
		if err != nil {
			e := response.NewError(err.Error(), http.StatusBadRequest)
			e.Handle(w)
			return
		}

		itemID, err := strconv.Atoi(chi.URLParam(r, "itemID"))
		if err != nil {
			e := response.NewError(err.Error(), http.StatusBadRequest)
			e.Handle(w)
			return
		}

		e := DeleteItem(env, huntID, itemID)
		if e != nil {
			e.Handle(w)
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
func createItemHandler(env *c.Env) http.HandlerFunc {
	return (func(w http.ResponseWriter, r *http.Request) {
		huntID, err := strconv.Atoi(chi.URLParam(r, "huntID"))
		if err != nil {
			e := response.NewError(err.Error(), http.StatusBadRequest)
			e.Handle(w)
			return
		}

		item := models.Item{}
		e := request.DecodeAndValidate(r, &item)
		if e != nil {
			e.Handle(w)
			return
		}

		_, e = InsertItem(env, &item, huntID)
		if e != nil {
			e.Handle(w)
			return
		}

		render.JSON(w, r, &item)
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
func patchItemHandler(env *c.Env) http.HandlerFunc {
	return (func(w http.ResponseWriter, r *http.Request) {
		huntID, err := strconv.Atoi(chi.URLParam(r, "huntID"))
		if err != nil {
			e := response.NewError(err.Error(), http.StatusBadRequest)
			e.Handle(w)
			return
		}

		itemID, err := strconv.Atoi(chi.URLParam(r, "itemID"))
		if err != nil {
			e := response.NewError(err.Error(), http.StatusBadRequest)
			e.Handle(w)
			return
		}

		partialItem := make(map[string]interface{})
		err = render.DecodeJSON(r.Body, &partialItem)
		if err != nil {
			e := response.NewError(err.Error(), http.StatusBadRequest)
			e.Handle(w)
			return
		}

		e := UpdateItem(env, huntID, itemID, &partialItem)
		if e != nil {
			e.Handle(w)
		}

		return
	})
}
