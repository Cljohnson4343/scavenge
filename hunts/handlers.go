package hunts

import (
	"net/http"

	"github.com/cljohnson4343/scavenge/request"

	"github.com/cljohnson4343/scavenge/hunts/models"
	"github.com/cljohnson4343/scavenge/response"

	c "github.com/cljohnson4343/scavenge/config"
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
		hunts, e := AllHunts()
		if e != nil {
			e.Handle(w)
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
		huntID, e := request.GetIntURLParam(r, "huntID")
		if e != nil {
			e.Handle(w)
			return
		}

		hunt, e := GetHunt(huntID)
		if e != nil {
			e.Handle(w)
		}

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

		e = InsertHunt(&hunt)
		if e != nil {
			e.Handle(w)
			return
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
		huntID, e := request.GetIntURLParam(r, "huntID")
		if e != nil {
			e.Handle(w)
			return
		}

		e = DeleteHunt(huntID)
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
		huntID, e := request.GetIntURLParam(r, "huntID")
		if e != nil {
			e.Handle(w)
			return
		}

		hunt := Hunt{}

		e = request.DecodeAndPatchValidate(r, &hunt, huntID)
		if e != nil {
			e.Handle(w)
			return
		}

		rowsAffected, e := UpdateHunt(env, &hunt)
		if e != nil {
			e.Handle(w)
			return
		}

		if !rowsAffected {
			e := response.NewError("cannot patch a hunt that doesn't exist", http.StatusBadRequest)
			e.Handle(w)
			return
		}

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
		huntID, e := request.GetIntURLParam(r, "huntID")
		if e != nil {
			e.Handle(w)
			return
		}

		items, e := GetItems(huntID)
		if e != nil {
			e.Handle(w)
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
		huntID, e := request.GetIntURLParam(r, "huntID")
		if e != nil {
			e.Handle(w)
			return
		}

		itemID, e := request.GetIntURLParam(r, "itemID")
		if e != nil {
			e.Handle(w)
			return
		}

		e = DeleteItem(env, huntID, itemID)
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
		huntID, e := request.GetIntURLParam(r, "huntID")
		if e != nil {
			e.Handle(w)
			return
		}

		item := models.Item{}
		e = request.DecodeAndValidate(r, &item)
		if e != nil {
			e.Handle(w)
			return
		}

		e = InsertItem(env, &item, huntID)
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
		huntID, e := request.GetIntURLParam(r, "huntID")
		if e != nil {
			e.Handle(w)
			return
		}

		itemID, e := request.GetIntURLParam(r, "itemID")
		if e != nil {
			e.Handle(w)
			return
		}

		item := models.Item{}

		e = request.DecodeAndPatchValidate(r, &item, itemID)
		if e != nil {
			e.Handle(w)
			return
		}

		// make sure the patch request does not change the item's hunt from hunt
		// specified by the URL
		if item.HuntID != 0 && item.HuntID != huntID {
			e = response.NewError("hunt_id: the item's hunt_id can not be modified", http.StatusBadRequest)
			e.Handle(w)
			return
		}

		e = UpdateItem(env, &item)
		if e != nil {
			e.Handle(w)
		}

		return
	})
}
