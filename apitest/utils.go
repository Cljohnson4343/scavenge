package apitest

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"time"

	c "github.com/cljohnson4343/scavenge/config"
	"github.com/cljohnson4343/scavenge/hunts"
	"github.com/cljohnson4343/scavenge/sessions"
	"github.com/cljohnson4343/scavenge/users"
)

// CreateUser creates the given user. Panics on any errors.
func CreateUser(u *users.User, env *c.Env) {
	reqBody, err := json.Marshal(u)
	if err != nil {
		panic(err)
	}

	req, err := http.NewRequest("POST", "/", bytes.NewReader(reqBody))
	if err != nil {
		panic(err)
	}
	rr := httptest.NewRecorder()
	handler := users.Routes(env)
	handler.ServeHTTP(rr, req)
	res := rr.Result()
	if res.StatusCode != http.StatusOK {
		resBody, err := ioutil.ReadAll(res.Body)
		if err != nil {
			panic(err)
		}
		panic(fmt.Sprintf("error creating user: %s", resBody))
	}

	err = json.NewDecoder(res.Body).Decode(u)
	if err != nil {
		panic(fmt.Sprintf("error decoding the res body: %v", err))
	}

	if u.ID == 0 {
		panic("expected user's id to be returned")
	}
}

// Login logs the given user in. Panics on all errors
func Login(u *users.User, env *c.Env) *http.Cookie {
	// reset fields that shouldn't be included
	u.LastVisit = time.Time{}
	u.JoinedAt = time.Time{}

	reqBody, err := json.Marshal(u)
	if err != nil {
		panic(err)
	}

	req, err := http.NewRequest("POST", "/login/", bytes.NewReader(reqBody))
	if err != nil {
		panic(err)
	}
	rr := httptest.NewRecorder()
	handler := users.Routes(env)
	handler.ServeHTTP(rr, req)
	res := rr.Result()
	if res.StatusCode != http.StatusOK {
		resBody, err := ioutil.ReadAll(res.Body)
		if err != nil {
			panic(err)
		}
		panic(fmt.Sprintf("error logging in: %s", resBody))
	}

	err = json.NewDecoder(res.Body).Decode(u)
	if err != nil {
		panic(fmt.Sprintf("error decoding the res body: %v", err))
	}

	var cookie *http.Cookie
	cookies := res.Cookies()
	for _, c := range cookies {
		if c.Name == sessions.SessionCookieName {
			cookie = c
		}
	}

	if cookie == nil {
		panic("expected a cookie on login")
	}

	return cookie
}

// CreateHunt creates the given hunt. Panics on all errors.
func CreateHunt(h *hunts.Hunt, env *c.Env, cookie *http.Cookie) {
	reqBody, err := json.Marshal(h)
	if err != nil {
		panic(fmt.Sprintf("error marshalling hunt data: %v", err))
	}

	req, err := http.NewRequest("POST", "/", bytes.NewReader(reqBody))
	if err != nil {
		panic(fmt.Sprintf("error getting new request: %v", err))
	}
	req.AddCookie(cookie)

	rr := httptest.NewRecorder()
	huntHandler := users.RequireUser(hunts.Routes(env))
	huntHandler.ServeHTTP(rr, req)

	res := rr.Result()

	if res.StatusCode != http.StatusOK {
		resBody, err := ioutil.ReadAll(res.Body)
		if err != nil {
			panic(fmt.Sprintf("error reading res: %v", err))
		}

		panic(fmt.Sprintf("error creating new hunt: %s", resBody))
	}

	err = json.NewDecoder(res.Body).Decode(h)
	if err != nil {
		panic(fmt.Sprintf("error decoding res body: %v", err))
	}

	if h.ID == 0 {
		panic("expected hunt id to be returned")
	}
}
