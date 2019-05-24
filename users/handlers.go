package users

import (
	"net/http"

	"github.com/cljohnson4343/scavenge/config"
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
func GetLogoutHandler(env *config.Env) http.HandlerFunc {
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
func GetLoginHandler(env *config.Env) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		u := User{}
		e := request.Decode(r, &u)
		if e != nil {
			e.Handle(w)
			return
		}

		// If login provided a userID verify user exists
		if u.ID != 0 {
			existing, e := db.GetUser(u.ID)
			if e != nil {
				e.Handle(w)
				return
			}

			if existing == nil {
				e := response.NewErrorf(http.StatusBadRequest,
					"getLoginHandler: unable to login user %d", u.ID)
				e.Handle(w)
				return
			}
		} else if u.Username != "" {
			existing, e := db.GetUserByUsername(u.Username)
			if e != nil {
				e.Handle(w)
				return
			}

			if existing == nil {
				e := response.NewErrorf(http.StatusBadRequest,
					"getLoginHandler: unable to login user %s", u.Username)
				e.Handle(w)
				return
			}

			u.ID = existing.ID
		} else {
			e := response.NewErrorf(http.StatusBadRequest,
				"getLoginHandler: must provide either userID or username")
			e.Handle(w)
			return
		}

		// create session and add a session cookie to user agent
		sess, e := sessions.New(u.ID)
		if e != nil {
			e.Handle(w)
			return
		}

		cookie := sess.Cookie()
		http.SetCookie(w, cookie)
		render.JSON(w, r, &u)

		return
	}
}

// swagger:route GET /users/ get current user getCurrentUserHandler
//
// Gets the user that is using this session.
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
func getCurrentUserHandler(env *config.Env) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		sess, e := sessions.GetCurrent(sessions.GetCookie(r))
		if e != nil {
			e.Handle(w)
			return
		}

		u, e := db.GetUser(sess.UserID)
		if e != nil {
			e.Handle(w)
			return
		}

		render.JSON(w, r, &u)
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
func getSelectUserHandler(env *config.Env) http.HandlerFunc {
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
func GetDeleteUserHandler(env *config.Env) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, e := request.GetIntURLParam(r, "userID")
		if e != nil {
			e.Handle(w)
			return
		}

		e = DeleteUser(userID)
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
func getUpdateUserHandler(env *config.Env) http.HandlerFunc {
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

		render.JSON(w, r, &u)
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
func GetCreateUserHandler(env *config.Env) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		u := User{}
		e := request.DecodeAndValidate(r, &u)
		if e != nil {
			e.Handle(w)
			return
		}

		// create new user
		e = InsertUser(&u)
		if e != nil {
			e.Handle(w)
			return
		}

		// create session and add a session cookie to user agent
		sess, e := sessions.New(u.ID)
		if e != nil {
			e.Handle(w)
			return
		}

		cookie := sess.Cookie()
		http.SetCookie(w, cookie)
		render.JSON(w, r, &u)
	}
}

// swagger:route GET /users/{userID}/notifications/ get user notifications
//
// Gets the notifications for the user with the given id.
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
func getNotificationsHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, e := request.GetIntURLParam(r, "userID")
		if e != nil {
			e.Handle(w)
			return
		}

		invitations, e := db.GetHuntInvitationsByUserID(userID)
		if e != nil {
			e.Handle(w)
			return
		}

		render.JSON(w, r, &invitations)
	}
}

// DeleteNotificationHandler deletes the notification with the given id
//
// swagger:route DELETE /users/{userID}/notifications/{notificationID}
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
func DeleteNotificationHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, e := request.GetIntURLParam(r, "userID")
		if e != nil {
			e.Handle(w)
			return
		}

		notificationID, e := request.GetIntURLParam(r, "notificationID")
		if e != nil {
			e.Handle(w)
			return
		}

		e = db.DeleteHuntInvitation(notificationID, userID)
		if e != nil {
			e.Handle(w)
			return
		}

		return
	}
}
