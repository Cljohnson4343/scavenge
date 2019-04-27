package users

import (
	"context"
	"net/http"

	"github.com/cljohnson4343/scavenge/response"

	"github.com/cljohnson4343/scavenge/db"
)

type userIDKeyType string

var userIDKey userIDKeyType = "userID"

// WithUserID returns a context that has the given userID added to its Values
func WithUserID(ctx context.Context, userID int) context.Context {
	return context.WithValue(ctx, userIDKey, userID)
}

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
