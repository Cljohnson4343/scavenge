package users

import (
	"net/http"

	"github.com/go-chi/render"

	"github.com/cljohnson4343/scavenge/db"

	"github.com/cljohnson4343/scavenge/request"

	"github.com/cljohnson4343/scavenge/sessions"

	c "github.com/cljohnson4343/scavenge/config"
)

// swagger:route POST /users/login login user getLoginHandler
//
// Logs in the given user.
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
func getLoginHandler(env *c.Env) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		u := User{}
		e := request.DecodeAndValidate(r, &u)
		if e != nil {
			e.Handle(w)
			return
		}

		e = u.Insert()
		if e != nil {
			e.Handle(w)
			return
		}

		// create session and add a session cookie to user agent
		sess := sessions.New(u.ID)
		c := sess.Cookie()
		http.SetCookie(w, c)

		return
	}
}

// swagger:route GET /users/{userID} get user getSelectUserHandler
//
// Gets the user with the given id.
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
func getSelectUserHandler(env *c.Env) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, e := request.GetIntURLParam(r, "userID")
		if e != nil {
			e.Handle(w)
			return
		}

		u, e := db.GetUser(userID)
		if e != nil {
			e.Handle(w)
			return
		}

		render.JSON(w, r, &u)
	}
}

// swagger:route DELETE /users/{userID} delete user getDeleteUserHandler
//
// Deletes the user with the given id.
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
func getDeleteUserHandler(env *c.Env) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, e := request.GetIntURLParam(r, "userID")
		if e != nil {
			e.Handle(w)
			return
		}

		e = db.DeleteUser(userID)
		if e != nil {
			e.Handle(w)
			return
		}
	}
}

// swagger:route PATCH /users/{userID} patch user getUpdateUserHandler
//
// Updates the db user with the given id using the given user.
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
func getUpdateUserHandler(env *c.Env) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, e := request.GetIntURLParam(r, "userID")
		if e != nil {
			e.Handle(w)
			return
		}

		u := db.UserDB{}
		e = request.DecodeAndPatchValidate(r, &u, userID)
		if e != nil {
			e.Handle(w)
			return
		}

	}
}
