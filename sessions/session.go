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
	secs := 60 * 60 * 24 * 365
	c := http.Cookie{
		Name:     SessionCookieName,
		Value:    s.Key.String(),
		Expires:  s.Expires,
		Secure:   false,
		HttpOnly: false,
		MaxAge:   secs,
		Path:     "/",
	}

	return &c
}

func getCurrentCookie(r *http.Request) (*http.Cookie, *response.Error) {
	cookie, err := r.Cookie(SessionCookieName)
	if err != nil {
		return nil, response.NewErrorf(http.StatusInternalServerError,
			"sessions.getCurrentCookie: error getting cookie: %v", err)
	}

	if cookie == nil {
		return nil, response.NewError(http.StatusInternalServerError,
			"sessions.getCurrentCookie: no cookie found")
	}

	return cookie, nil
}

// DeleteCurrent deletes the user agents current session.
func DeleteCurrent(r *http.Request) *response.Error {
	cookie, e := getCurrentCookie(r)
	if e != nil {
		return e
	}

	key, err := uuid.Parse(cookie.Value)
	if err != nil {
		return response.NewErrorf(http.StatusInternalServerError,
			"sessions.DeleteCurrent: error parsing cookie.Value: %v", err)
	}

	e = db.DeleteSession(key)
	if e != nil {
		return e
	}

	return nil
}

// GetCurrent returns the user agents current session.
func GetCurrent(r *http.Request) (*Session, *response.Error) {
	cookie, e := getCurrentCookie(r)
	if e != nil {
		return nil, e
	}

	// TODO think about how to handle this case. i.e. redirect extend cookie etc.
	if cookie.Expires.Before(time.Now()) {
		return nil, response.NewError(http.StatusBadRequest,
			"sessions.GetCurrent: the current session is expired")
	}

	key, err := uuid.Parse(cookie.Value)
	if err != nil {
		return nil, response.NewErrorf(http.StatusInternalServerError,
			"sessions.GetCurrent: error parsing the cookie value: %v", err)
	}

	s, e := db.GetSession(key)
	if e != nil {
		return nil, e
	}

	return &Session{*s}, nil
}

// RemoveCookie removes the current session cookie from the user agent
func RemoveCookie(w http.ResponseWriter, r *http.Request) *response.Error {
	cookie, e := getCurrentCookie(r)
	if e != nil {
		return e
	}

	cookie.MaxAge = -1

	return nil
}
