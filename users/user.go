package users

import (
	"context"
	"net/http"

	"github.com/cljohnson4343/scavenge/roles"

	"github.com/cljohnson4343/scavenge/db"
	"github.com/cljohnson4343/scavenge/response"
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

// ContextWithUser returns a context with the userID stored as a value
func ContextWithUser(ctx context.Context, userID int) context.Context {
	return context.WithValue(ctx, userIDKey, userID)
}

// User represents a user
type User struct {
	db.UserDB
}

// InsertUser inserts the given user into the db and assigns user roles
func InsertUser(u *User) *response.Error {
	e := u.Insert()
	if e != nil {
		return e
	}

	userRole := roles.New("user", 0)
	e = userRole.AddTo(u.ID)
	if e != nil {
		return e
	}

	userOwnerRole := roles.New("user_owner", u.ID)
	e = userOwnerRole.AddTo(u.ID)
	if e != nil {
		return e
	}

	return nil
}

// DeleteUser deletes the given user from the db as well as all associated roles
func DeleteUser(userID int) *response.Error {
	e := db.DeleteUser(userID)
	if e != nil {
		return nil
	}

	return roles.DeleteRolesForUser(userID)
}
