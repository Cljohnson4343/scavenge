package db

import (
	"fmt"
	"net/http"
	"time"

	"github.com/asaskevich/govalidator"
	"github.com/cljohnson4343/scavenge/response"

	"github.com/google/uuid"
)

// SessionDB is the representation of a User's session.
type SessionDB struct {
	// Key is the UUID that identifies a session
	//
	// required: true
	Key uuid.UUID `json:"session_key" valid:"uuid"`

	// Expires is the expiration date for the session. If
	// this date has already past, then the session is not
	// valid.
	//
	// required: true
	Expires time.Time `json:"expires" valid:"timeNotPast"`

	// CreatedAt is the time stamp for the session creation
	//
	// required: true
	CreatedAt time.Time `json:"created_at" valid:"-"`

	// UserID is the id of the user associated with this
	// session.
	//
	// required: true
	UserID int `json:"user_id" valid:"int"`
}

// Validate validates the given session
func (s *SessionDB) Validate(r *http.Request) *response.Error {
	_, err := govalidator.ValidateStruct(s)
	if err != nil {
		return response.NewError(fmt.Sprintf("error validating session: %v", err), http.StatusBadRequest)
	}

	return nil
}

var sessionInsertScript = `
	INSERT INTO user_sessions(session_key, expires, user_id)
	VALUES ($1, $2, $3)
	RETURNING user_sessions(created_at);
	`

// Insert inserts the given session into the db. The created_at time stamp will be written
// back to the given session.
func (s *SessionDB) Insert() *response.Error {
	err := stmtMap["sessionInsert"].QueryRow(s.Key, s.Expires, s.UserID).Scan(&s.CreatedAt)
	if err != nil {
		return response.NewError(fmt.Sprintf("error inserting session %s: %v",
			s.Key.String(), err), http.StatusInternalServerError)
	}

	return nil
}
