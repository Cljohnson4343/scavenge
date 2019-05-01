package apitest

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"time"

	"github.com/cljohnson4343/scavenge/db"

	c "github.com/cljohnson4343/scavenge/config"
	"github.com/cljohnson4343/scavenge/hunts"
	"github.com/cljohnson4343/scavenge/hunts/models"
	"github.com/cljohnson4343/scavenge/sessions"
	"github.com/cljohnson4343/scavenge/teams"
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
	huntHandler := hunts.Routes(env)
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

// CreateTeam creates a team. panics on any error.
func CreateTeam(t *teams.Team, env *c.Env, cookie *http.Cookie) {
	reqBody, err := json.Marshal(t)
	if err != nil {
		panic(fmt.Sprintf("error marshalling team data: %v", err))
	}

	req, err := http.NewRequest("POST", "/", bytes.NewReader(reqBody))
	if err != nil {
		panic(fmt.Sprintf("error getting new request: %v", err))
	}
	req.AddCookie(cookie)

	rr := httptest.NewRecorder()
	teamsHandler := teams.Routes(env)
	teamsHandler.ServeHTTP(rr, req)

	res := rr.Result()

	if res.StatusCode != http.StatusOK {
		resBody, err := ioutil.ReadAll(res.Body)
		if err != nil {
			panic(fmt.Sprintf("error reading res: %v", err))
		}

		panic(fmt.Sprintf("error creating new team: %s", resBody))
	}

	err = json.NewDecoder(res.Body).Decode(t)
	if err != nil {
		panic(fmt.Sprintf("error decoding res body: %v", err))
	}

	if t.ID == 0 {
		panic("expected team id to be returned")
	}
}

// CreateItem creates an item. panics on any errors.
func CreateItem(i *models.Item, env *c.Env, cookie *http.Cookie) {
	reqBody, err := json.Marshal(i)
	if err != nil {
		panic(fmt.Sprintf("error marshalling request data: %v", err))
	}

	req, err := http.NewRequest(
		"POST",
		fmt.Sprintf("/%d/items/", i.HuntID),
		bytes.NewReader(reqBody),
	)
	if err != nil {
		panic(fmt.Sprintf("error getting new request: %v", err))
	}
	req.AddCookie(cookie)

	rr := httptest.NewRecorder()
	handler := hunts.Routes(env)
	handler.ServeHTTP(rr, req)

	res := rr.Result()
	if res.StatusCode != http.StatusOK {
		resBody, err := ioutil.ReadAll(res.Body)
		if err != nil {
			panic(fmt.Sprintf("error reading res body: %v", err))
		}

		panic(fmt.Sprintf("expected code 200 got %d: %s", res.StatusCode, resBody))
	}

	err = json.NewDecoder(res.Body).Decode(i)
	if err != nil {
		panic(err.Error())
	}

	if i.ID == 0 {
		panic("expected item id to be returned")
	}
}

// CreateMedia creates the given media. panics on any erors
func CreateMedia(m *db.MediaMetaDB, env *c.Env, cookie *http.Cookie) {
	reqBody, err := json.Marshal(m)
	if err != nil {
		panic(fmt.Sprintf("error marshalling req body: %v", err))
	}

	req, err := http.NewRequest(
		"POST",
		fmt.Sprintf("/%d/media/", m.TeamID),
		bytes.NewReader(reqBody),
	)
	if err != nil {
		panic(fmt.Sprintf("error getting new request: %v", err))
	}
	req.AddCookie(cookie)

	rr := httptest.NewRecorder()
	handler := teams.Routes(env)
	handler.ServeHTTP(rr, req)

	res := rr.Result()

	if res.StatusCode != http.StatusOK {
		resBody, _ := ioutil.ReadAll(res.Body)
		panic(fmt.Sprintf(
			"expected return code of 200 got %d: %s",
			res.StatusCode,
			resBody),
		)
	}

	err = json.NewDecoder(res.Body).Decode(m)
	if err != nil {
		panic(fmt.Sprintf("error decoding response body: %v", err))
	}

	if m.ID == 0 {
		panic("expected media id to be returned")
	}

}

// CreateLocation creates a location. panics on any errors.
func CreateLocation(l *db.LocationDB, env *c.Env, cookie *http.Cookie) {
	reqBody, err := json.Marshal(l)
	if err != nil {
		panic(fmt.Sprintf("error marshalling req data: %v", err))
	}

	req, err := http.NewRequest(
		"POST",
		fmt.Sprintf("/%d/locations/", l.TeamID),
		bytes.NewReader(reqBody),
	)
	if err != nil {
		panic(fmt.Sprintf("error getting new request: %v", err))
	}
	req.AddCookie(cookie)

	rr := httptest.NewRecorder()
	handler := teams.Routes(env)
	handler.ServeHTTP(rr, req)
	res := rr.Result()

	if res.StatusCode != http.StatusOK {
		resBody, err := ioutil.ReadAll(res.Body)
		if err != nil {
			panic(fmt.Sprintf("error reading response: %v", err))
		}
		panic(fmt.Sprintf(
			"expected code %d got %d: %s",
			http.StatusOK,
			res.StatusCode,
			resBody),
		)
	}

	err = json.NewDecoder(res.Body).Decode(l)
	if err != nil {
		panic(fmt.Sprintf("error decoding response: %v", err))
	}

	if l.ID == 0 {
		panic("expected id to be returned")
	}
}

// AddPlayer adds the player with teh given user_id to the given team. Panics
// on all errors.
func AddPlayer(playerID, teamID int, env *c.Env, cookie *http.Cookie) {
	reqData := struct {
		PlayerID int `json:"id"`
	}{PlayerID: playerID}

	reqBody, err := json.Marshal(&reqData)
	if err != nil {
		panic(fmt.Sprintf("error marshalling req data: %v", err))
	}

	req, err := http.NewRequest(
		"POST",
		fmt.Sprintf("/%d/players/", teamID),
		bytes.NewReader(reqBody),
	)
	if err != nil {
		panic(fmt.Sprintf("error getting new request: %v", err))
	}

	rr := httptest.NewRecorder()
	handler := teams.Routes(env)
	handler.ServeHTTP(rr, req)

	res := rr.Result()

	if res.StatusCode != http.StatusOK {
		resBody, err := ioutil.ReadAll(res.Body)
		if err != nil {
			panic(fmt.Sprintf("error reading res body: %v", err))
		}
		panic(fmt.Sprintf(
			"expected code %d got %d: %s",
			http.StatusOK,
			res.StatusCode,
			resBody),
		)
	}
}
