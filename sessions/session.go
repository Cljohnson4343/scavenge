package sessions

import (
	"fmt"
	"net/http"
	"time"

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

// New returns a new user session
func New(userID int) *Session {
	key := uuid.New()
	expiration := time.Now().Add(sessionDuration)

	s := Session{db.SessionDB{Key: key, Expires: expiration, UserID: userID}}

	return &s
}

// Cookie creates a cookie for the given session and returns it
func (s *Session) Cookie() *http.Cookie {
	c := http.Cookie{
		Name:     sessionCookieName,
		Value:    s.Key.String(),
		Expires:  s.Expires,
		Secure:   true,
		HttpOnly: true,
		MaxAge:   0,
	}

	return &c
}
