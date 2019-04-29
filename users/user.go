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
		return 0, response.NewError(http.StatusInternalServerError,
			"GetUserID: the given context does not contain a userID of type int")
	}

	return id, nil
}

// User represents a user
type User struct {
	db.UserDB
}

// RequireUser is middleware that checks to make sure the user agent has a valid user
// session. The userID for a valid user session is then added to the context that
// is past down to the given handler
func RequireUser(fn http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie := sessions.GetCookie(r)
		if cookie == nil {
			e := response.NewErrorf(http.StatusUnauthorized, "must be logged in")
			e.Handle(w)
			return
		}

		s, e := sessions.GetCurrent(cookie)
		if e != nil {
			e.Handle(w)
			return
		}

		ctx := context.WithValue(r.Context(), userIDKey, s.UserID)
		fn.ServeHTTP(w, r.WithContext(ctx))
	})
}
