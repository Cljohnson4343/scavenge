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
var sessionCookieName = `scavenge_session`

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
	c := http.Cookie{
		Name:     sessionCookieName,
		Value:    s.Key.String(),
		Expires:  s.Expires,
		Secure:   false,
		HttpOnly: true,
		MaxAge:   int(time.Until(s.Expires).Seconds()),
		Path:     "/",
	}

	return &c
}

func getCurrentCookie(r *http.Request) (*http.Cookie, *response.Error) {
	cookies := r.Cookies()
	for _, c := range cookies {
		fmt.Println(c)
	}

	cookie, err := r.Cookie(sessionCookieName)
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

// RemoveCookie removes the current session cookie from the user agent
func RemoveCookie(w http.ResponseWriter, r *http.Request) *response.Error {
	cookie, e := getCurrentCookie(r)
	if e != nil {
		return e
	}

	cookie.MaxAge = -1

	return nil
}
