package db

import (
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
	Key uuid.UUID `json:"sessionKey" valid:"uuid"`

	// Expires is the expiration date for the session. If
	// this date has already past, then the session is not
	// valid.
	//
	// required: true
	Expires time.Time `json:"expires" valid:"timeNotPast"`

	// CreatedAt is the time stamp for the session creation
	//
	// required: true
	CreatedAt time.Time `json:"createdAt" valid:"-"`

	// UserID is the id of the user associated with this
	// session.
	//
	// required: true
	UserID int `json:"userID" valid:"int"`
}

// Validate validates the given session
func (s *SessionDB) Validate(r *http.Request) *response.Error {
	_, err := govalidator.ValidateStruct(s)
	if err != nil {
		return response.NewErrorf(http.StatusBadRequest, "error validating session: %v", err)
	}

	return nil
}

var sessionInsertScript = `
	INSERT INTO users_sessions(session_key, expires, user_id)
	VALUES ($1, $2, $3)
	RETURNING created_at;
	`

// Insert inserts the given session into the db. The created_at time stamp will be written
// back to the given session.
func (s *SessionDB) Insert() *response.Error {
	err := stmtMap["sessionInsert"].QueryRow(s.Key, s.Expires, s.UserID).Scan(&s.CreatedAt)
	if err != nil {
		return response.NewErrorf(http.StatusInternalServerError, "error inserting session %s: %v", s.Key.String(), err)
	}

	return nil
}

var sessionGetForUserScript = `
	SELECT session_key, expires, created_at, user_id
	FROM users_sessions u
	WHERE  u.user_id = $1;`

// GetSessionsForUser returns the sessions for the given user
func GetSessionsForUser(userID int) ([]*SessionDB, *response.Error) {
	rows, err := stmtMap["sessionGetForUser"].Query(userID)
	if err != nil {
		return nil, response.NewErrorf(http.StatusInternalServerError,
			"GetSessionsForUser: error getting sessions for user %d: %v", userID, err)
	}
	defer rows.Close()

	e := response.NewNilError()
	sesses := make([]*SessionDB, 0)
	for rows.Next() {
		s := SessionDB{}
		err = rows.Scan(&s.Key, &s.Expires, &s.CreatedAt, &s.UserID)
		if err != nil {
			e.Addf(http.StatusInternalServerError,
				"GerSessionsForUser: error getting session for user %d: %v", userID, err)
			break
		}

		sesses = append(sesses, &s)
	}

	if err = rows.Err(); err != nil {
		e.Addf(http.StatusInternalServerError,
			"GerSessionsForUser: error getting session for user %d: %v", userID, err)
	}

	return sesses, e.GetError()
}

var sessionDeleteScript = `
	DELETE FROM users_sessions
	WHERE session_key = $1;`

// DeleteSession deletes the session with the given key.
func DeleteSession(key uuid.UUID) *response.Error {
	res, err := stmtMap["sessionDelete"].Exec(key)
	if err != nil {
		return response.NewErrorf(http.StatusInternalServerError,
			"db.DeleteSession: error deleting session with key %s: %v", key.String(), err)
	}

	numRows, err := res.RowsAffected()
	if err != nil {
		return response.NewErrorf(http.StatusInternalServerError,
			"db.DeleteSession: error deleting session with key %s: %v", key.String(), err)
	}

	if numRows < 1 {
		return response.NewErrorf(http.StatusBadRequest,
			"error deleting session with key %s: %v", key.String(), err)
	}

	return nil
}

var sessionGetScript = `
	SELECT expires, created_at, user_id
	FROM users_sessions
	WHERE session_key = $1;`

// GetSession returns the session with the given key
func GetSession(key uuid.UUID) (*SessionDB, *response.Error) {
	s := SessionDB{}
	err := stmtMap["sessionGet"].QueryRow(key).Scan(&s.Expires, &s.CreatedAt, &s.UserID)
	if err != nil {
		return nil, response.NewErrorf(http.StatusInternalServerError,
			"GetSession: error getting session %s: %v", key.String(), err)
	}

	s.Key = key
	return &s, nil
}
