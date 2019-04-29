// +build apiTest

package teams_test

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/cljohnson4343/scavenge/apitest"
	c "github.com/cljohnson4343/scavenge/config"
	"github.com/cljohnson4343/scavenge/db"
	"github.com/cljohnson4343/scavenge/hunts"
	"github.com/cljohnson4343/scavenge/response"
	"github.com/cljohnson4343/scavenge/teams"
	"github.com/cljohnson4343/scavenge/users"
)

var env *c.Env
var hunt hunts.Hunt
var sessionCookie *http.Cookie
var newUser *users.User

func TestMain(m *testing.M) {
	d := db.InitDB("../db/db_info_test.json")
	defer db.Shutdown(d)
	env = c.CreateEnv(d)
	response.SetDevMode(true)

	// Login in user to get a valid user session cookie
	newUser = &users.User{
		db.UserDB{
			FirstName: "Fernando",
			LastName:  "Sucre",
			Username:  "sucre_43",
			Email:     "sucre433@gmail.com",
		},
	}
	apitest.CreateUser(newUser, env)
	sessionCookie = apitest.Login(newUser, env)

	// Create hunt to use for tests
	hunt.HuntDB = db.HuntDB{
		Name:         "Teams Test Hunt 43",
		MaxTeams:     43,
		StartTime:    time.Now().AddDate(0, 0, 1),
		EndTime:      time.Now().AddDate(0, 0, 2),
		LocationName: "Fake Location",
		Latitude:     34.730705,
		Longitude:    -86.59481,
	}
	apitest.CreateHunt(&hunt, env, sessionCookie)

	os.Exit(m.Run())
}

func TestCreateTeamHandler(t *testing.T) {
	cases := []struct {
		name       string
		team       teams.Team
		statusCode int
	}{
		{
			name: "add new team",
			team: teams.Team{
				TeamDB: db.TeamDB{
					HuntID: hunt.ID,
					Name:   "team 1",
				},
			},
			statusCode: http.StatusOK,
		},
		{
			name: "add team with same name as another team in same hunt",
			team: teams.Team{
				TeamDB: db.TeamDB{
					HuntID: hunt.ID,
					Name:   "team 1",
				},
			},
			statusCode: http.StatusBadRequest,
		},
		{
			name: "add team with ID",
			team: teams.Team{
				TeamDB: db.TeamDB{
					HuntID: hunt.ID,
					Name:   "team 2",
					ID:     1,
				},
			},
			statusCode: http.StatusBadRequest,
		},
		{
			name: "add team without hunt id",
			team: teams.Team{
				TeamDB: db.TeamDB{
					Name: "team 3",
				},
			},
			statusCode: http.StatusBadRequest,
		},
		{
			name: "add team without name",
			team: teams.Team{
				TeamDB: db.TeamDB{
					HuntID: hunt.ID,
				},
			},
			statusCode: http.StatusBadRequest,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			reqBody, err := json.Marshal(&c.team)
			if err != nil {
				t.Fatalf("error marshalling req data: %v", err)
			}
			req, err := http.NewRequest("POST", "/", bytes.NewReader(reqBody))
			if err != nil {
				t.Fatalf("error getting new request: %v", err)
			}

			rr := httptest.NewRecorder()
			handler := teams.Routes(env)
			handler.ServeHTTP(rr, req)

			res := rr.Result()

			if res.StatusCode != c.statusCode {
				resBody, err := ioutil.ReadAll(res.Body)
				if err != nil {
					t.Fatalf("error reading res body: %v", err)
				}

				t.Fatalf("expected code %d got %d: %s", c.statusCode, res.StatusCode, resBody)
			}

			if c.statusCode == http.StatusOK {
				err = json.NewDecoder(res.Body).Decode(&c.team)
				if err != nil {
					t.Fatalf("error decoding the res body: %v", err)
				}

				if c.team.ID == 0 {
					t.Error("expected id to be returned")
				}
			}
		})
	}
}
