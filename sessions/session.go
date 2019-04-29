package sessions

import (
	"fmt"
	"net/http"
	"time"

	"github.com/cljohnson4343/scavenge/response"

	"github.com/cljohnson4343/scavenge/db"
	"github.com/google/uuid"
)

var sessionDuration time.Duration

// SessionCookieName should not be used outside of sessions except in testing
var SessionCookieName = `scavenge_session`

func init() {
	var err error
	sessionDuration, err = time.ParseDuration(fmt.Sprintf("%dh", 24*365))
	if err != nil {
		panic(err)
	}
}

// Session represents a user session and is associated with one user
type Session struct {
	db.SessionDB
}

// New returns a new user session that has been stored in the db
func New(userID int) (*Session, *response.Error) {
	key := uuid.New()
	expiration := time.Now().Add(sessionDuration)

	s := Session{db.SessionDB{Key: key, Expires: expiration, UserID: userID}}
	e := s.Insert()
	if e != nil {
		return nil, e
	}

	return &s, nil
}

// Cookie creates a cookie for the given session and returns it
func (s *Session) Cookie() *http.Cookie {
	secs := s.Expires.Sub(time.Now()).Seconds()
	c := http.Cookie{
		Name:     SessionCookieName,
		Value:    s.Key.String(),
		Secure:   false,
		HttpOnly: false,
		MaxAge:   int(secs),
		Path:     "/",
	}

	return &c
}

// GetCookie returns the session cookie for the user agents
func GetCookie(r *http.Request) *http.Cookie {
	cookies := r.Cookies()
	var cookie *http.Cookie
	for _, c := range cookies {
		if c.Name == SessionCookieName {
			cookie = c
		}
	}

	return cookie
}

// GetCurrent returns the user agents current session.
func GetCurrent(cookie *http.Cookie) (*Session, *response.Error) {
	key, err := uuid.Parse(cookie.Value)
	if err != nil {
		return nil, response.NewErrorf(http.StatusInternalServerError,
			"sessions.GetCurrent: error parsing the cookie value: %v", err)
	}

	s, e := db.GetSession(key)
	if e != nil {
		return nil, e
	}

	if s.Expires.Before(time.Now()) {
		return nil, response.NewError(http.StatusBadRequest, "session expired")
	}

	return &Session{*s}, nil
}

// RemoveCookie removes the current session cookie from the user agent and deletes
// the associated session from the db
func RemoveCookie(w http.ResponseWriter, cookie *http.Cookie) *response.Error {
	key, err := uuid.Parse(cookie.Value)
	if err != nil {
		return response.NewErrorf(http.StatusInternalServerError,
			"sessions.DeleteCurrent: error parsing cookie.Value: %v", err)
	}

	e := db.DeleteSession(key)
	if e != nil {
		return e
	}

	cookie.MaxAge = -1

	return nil
}
