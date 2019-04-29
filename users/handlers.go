package users

import (
	"net/http"

	c "github.com/cljohnson4343/scavenge/config"
	"github.com/cljohnson4343/scavenge/db"
	"github.com/cljohnson4343/scavenge/request"
	"github.com/cljohnson4343/scavenge/response"
	"github.com/cljohnson4343/scavenge/sessions"
	"github.com/go-chi/render"
)

// GetLogoutHandler logs out the given user. Requires that the userID be part of
// the ctx provided by the req.
//
// swagger:route POST /users/logout logout user GetLogoutHandler
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
func GetLogoutHandler(env *c.Env) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cookie := sessions.GetCookie(r)
		if cookie == nil {
			e := response.NewError(http.StatusBadRequest, "not logged in")
			e.Handle(w)
			return
		}

		e := sessions.RemoveCookie(w, cookie)
		if e != nil {
			e.Handle(w)
		}
	}
}

// GetLoginHandler logs in the given user. If no user id is provided in req body
// and no user exists with the given user info, then a new user will be created.
//
// swagger:route POST /users/login login user GetLoginHandler
//
// Consumes:
// 	- application/json
//
// Produces:
//	- No response content
//
// Schemes: http, https
//
// Responses:
// 	200:
//  400:
func GetLoginHandler(env *c.Env) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		u := User{}
		e := request.DecodeAndValidate(r, &u)
		if e != nil {
			e.Handle(w)
			return
		}

		// If login provided a userID verify user exists
		if u.ID != 0 {
			existing, e := db.GetUser(u.ID)
			if e != nil {
				e.Handle(w)
			}

			if existing == nil {
				e := response.NewErrorf(http.StatusBadRequest,
					"getLoginHandler: unable to login user %d", u.ID)
				e.Handle(w)
				return
			}

		} else {
			// create new user
			e = u.Insert()
			if e != nil {
				e.Handle(w)
				return
			}
		}

		// create session and add a session cookie to user agent
		sess, e := sessions.New(u.ID)
		if e != nil {
			e.Handle(w)
			return
		}

		cookie := sess.Cookie()

		http.SetCookie(w, cookie)
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

// GetDeleteUserHandler deletes the user with the given id.
//
// swagger:route DELETE /users/{userID} delete user GetDeleteUserHandler
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
func GetDeleteUserHandler(env *c.Env) http.HandlerFunc {
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

		e = u.Update(env, userID)
		if e != nil {
			e.Handle(w)
			return
		}

	}
}

// GetCreateUserHandler creates the given user.
//
// swagger:route POST /users/ creates user GetCreateUserHandler
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
func GetCreateUserHandler(env *c.Env) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		u := User{}
		e := request.DecodeAndValidate(r, &u)
		if e != nil {
			e.Handle(w)
			return
		}

		// create new user
		e = u.Insert()
		if e != nil {
			e.Handle(w)
			return
		}

		render.JSON(w, r, &u)
	}
}
