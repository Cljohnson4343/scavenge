package users

import (
	"context"
	"net/http"

	"github.com/cljohnson4343/scavenge/db"
	"github.com/cljohnson4343/scavenge/response"
	"github.com/cljohnson4343/scavenge/sessions"
)

type userIDKeyType string

var userIDKey userIDKeyType = "userID"

// GetUserID gets the userID from the given context
func GetUserID(ctx context.Context) (int, *response.Error) {
	id, ok := ctx.Value(userIDKey).(int)
	if !ok {
		return 0, response.NewError(
			http.StatusUnauthorized,
			"GetUserID: the given context does not contain a userID of type int",
		)
	}

	return id, nil
}

// User represents a user
type User struct {
	db.UserDB
}

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

		ctx := context.WithValue(r.Context(), userIDKey, s.UserID)
		fn.ServeHTTP(w, r.WithContext(ctx))
	})
}
