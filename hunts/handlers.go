package hunts

import (
	"bytes"
	"encoding/json"
	"net/http"
	"os"

	"github.com/cljohnson4343/scavenge/config"
	"github.com/cljohnson4343/scavenge/hunts/models"
	"github.com/cljohnson4343/scavenge/request"
	"github.com/cljohnson4343/scavenge/response"
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
func getHuntsHandler(env *config.Env) http.HandlerFunc {
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
func getHuntHandler(env *config.Env) http.HandlerFunc {
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
func createHuntHandler(env *config.Env) http.HandlerFunc {
	return (func(w http.ResponseWriter, r *http.Request) {
		hunt := Hunt{}
		e := request.DecodeAndValidate(r, &hunt)
		if e != nil {
			e.Handle(w)
			return
		}

		e = InsertHunt(r.Context(), &hunt)
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
func deleteHuntHandler(env *config.Env) http.HandlerFunc {
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
func patchHuntHandler(env *config.Env) http.HandlerFunc {
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
			e := response.NewError(http.StatusBadRequest, "cannot patch a hunt that doesn't exist")
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
func getItemsHandler(env *config.Env) http.HandlerFunc {
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
func deleteItemHandler(env *config.Env) http.HandlerFunc {
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
func createItemHandler(env *config.Env) http.HandlerFunc {
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

		item.HuntID = huntID
		e = InsertItem(&item)
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
func patchItemHandler(env *config.Env) http.HandlerFunc {
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
			e = response.NewError(http.StatusBadRequest, "hunt_id: the item's hunt_id can not be modified")
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

// populateDBHandler fills the db with the hunts in 'test_data.json'
func populateDBHandler(env *config.Env) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		file, err := os.Open("./hunts/test_data.json")
		if err != nil {
			panic(err)
		}
		defer file.Close()

		hunts := make([]Hunt, 0)

		url := `http://localhost:4343/api/v0/hunts/`

		err = json.NewDecoder(file).Decode(&hunts)
		if err != nil {
			e := response.NewErrorf(http.StatusInternalServerError, "error decoding json data: %s", err.Error())

			e.Handle(w)
			return
		}

		for _, hunt := range hunts {
			b, err := json.Marshal(hunt)
			if err != nil {
				e := response.NewErrorf(http.StatusInternalServerError, "error decoding json data: %s", err.Error())

				e.Handle(w)
				return
			}

			buf := bytes.NewBuffer(b)

			res, err := http.Post(url, "application/json", buf)
			if err != nil {
				e := response.NewErrorf(http.StatusInternalServerError, "error decoding json data: %s", err.Error())

				e.Handle(w)
				return
			}
			res.Body.Close()

		}

		w.Write([]byte("Success!!!"))
	}
}
