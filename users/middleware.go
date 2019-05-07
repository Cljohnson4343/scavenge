package users

import (
	"net/http"

	"github.com/cljohnson4343/scavenge/db"
	"github.com/cljohnson4343/scavenge/response"
	"github.com/cljohnson4343/scavenge/roles"
	"github.com/cljohnson4343/scavenge/sessions"
)

// WithUser is middleware that checks to see if the user agent is using a valid
// session. If so, the userID is stored in the context that is passed to the
// next handler.
func WithUser(fn http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie := sessions.GetCookie(r)
		if cookie == nil {
			fn.ServeHTTP(w, r)
			return
		}

		// TODO think about changing this so that a db lookup is not required. i.e.
		// jwt or a cache for sessions. As it stands each request starts with a
		// sessions db lookup and then an roles.authorization lookup
		s, e := sessions.GetCurrent(cookie)
		if e != nil {
			e.Handle(w)
			return
		}

		ctx := ContextWithUser(r.Context(), s.UserID)
		fn.ServeHTTP(w, r.WithContext(ctx))
	})
}

// RequireAuth checks to make sure the requesting user agent has
// authorization to make the request
func RequireAuth(fn http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		userID, e := GetUserID(req.Context())
		if e != nil {
			e.Handle(w)
			return
		}

		perms, e := db.PermissionsForUser(userID)
		if e != nil {
			e.Handle(w)
			return
		}

		for _, p := range perms {
			perm := roles.Permission{PermissionDB: p}
			if perm.Authorized(req) {
				fn.ServeHTTP(w, req)
				return
			}
		}

		e = response.NewErrorf(
			http.StatusUnauthorized,
			"User %d is not authorized to access %s %s",
			userID,
			req.Method,
			req.URL.Path,
		)
		e.Handle(w)
	})
}
