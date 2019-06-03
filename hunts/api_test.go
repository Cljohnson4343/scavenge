// +build integration

package hunts_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/cljohnson4343/scavenge/apitest"
	"github.com/cljohnson4343/scavenge/config"
	"github.com/cljohnson4343/scavenge/db"
	"github.com/cljohnson4343/scavenge/hunts"
	"github.com/cljohnson4343/scavenge/response"
	"github.com/cljohnson4343/scavenge/roles"
	"github.com/cljohnson4343/scavenge/routes"
	"github.com/cljohnson4343/scavenge/users"
)

var env *config.Env
var sessionCookie *http.Cookie
var newUser *users.User

func TestMain(m *testing.M) {
	d := db.InitDB("../db/db_info_test.json")
	defer db.Shutdown(d)
	env = config.CreateEnv(d)
	response.SetDevMode(true)

	// Login in user to get a valid user session cookie
	newUser = &users.User{
		db.UserDB{
			FirstName: "Hunts API Tests",
			LastName:  "Hunts API Tests",
			Username:  "hunts_api_tests",
			Email:     "hunts_api_tests@gmail.com",
		},
	}
	apitest.CreateUser(newUser, env)
	sessionCookie = apitest.Login(newUser, env)
	// TODO get rid of role assignment when it is handled by user creation
	userRole := roles.New("user", 0)
	e := userRole.AddTo(newUser.ID)
	if e != nil {
		panic(fmt.Sprintf("error adding role to user: %s", e.JSON()))
	}

	os.Exit(m.Run())
}

func TestCreateHuntHandler(t *testing.T) {
	cases := []struct {
		name string
		code int
		hunt hunts.Hunt
		user *users.User
	}{
		{
			name: "valid hunt and user",
			code: http.StatusOK,
			hunt: hunts.Hunt{
				HuntDB: db.HuntDB{
					Name:         "CreateHuntHandler 1 hunt",
					MaxTeams:     43,
					StartTime:    time.Now().AddDate(0, 0, 1),
					EndTime:      time.Now().AddDate(0, 0, 2),
					LocationName: "Fake Location",
					Latitude:     34.730705,
					Longitude:    -86.59481,
				},
			},
			user: newUser,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			reqBody, err := json.Marshal(&c.hunt)
			if err != nil {
				t.Fatalf("error marshalling hunt: %v", err)
			}

			req, err := http.NewRequest(
				"POST",
				config.BaseAPIURL+"hunts/",
				bytes.NewReader(reqBody),
			)
			if err != err {
				t.Fatalf("error getting new request: %v", err)
			}
			req.AddCookie(sessionCookie)

			rr := httptest.NewRecorder()
			router := routes.Routes(env)
			router.ServeHTTP(rr, req)
			res := rr.Result()

			if res.StatusCode != c.code {
				resBody, err := ioutil.ReadAll(res.Body)
				if err != nil {
					t.Fatalf("error reading res body: %v", err)
				}
				t.Fatalf("expected code %d got %d: %s", c.code, res.StatusCode, resBody)
			}

			if c.code == http.StatusOK {

				userRoles, e := db.RolesForUser(c.user.ID)
				if e != nil {
					t.Fatalf("error getting roles for user: %s", e.JSON())
				}

				if len(userRoles) == 0 {
					t.Fatalf("expected hunt creator to have roles assigned")
				}

				teams, e := db.TeamsForHunt(c.hunt.ID)
				if e != nil {
					t.Fatalf("error getting teams for hunt: %s", e.JSON())
				}

				e = roles.DeleteRolesForHunt(c.hunt.ID, teams)
				if e != nil {
					t.Fatalf("error deleting roles for newly created hunt: %s", e.JSON())
				}
			}
		})
	}
}

func TestDeleteHuntHandler(t *testing.T) {
	// Create and log in user to get a valid user session cookie
	newUser := &users.User{
		db.UserDB{
			FirstName: "TestDeleteHuntHandler",
			LastName:  "TestDeleteHuntHandler",
			Username:  "TestDeleteHuntHandler",
			Email:     "TestDeleteHuntHandler@gmail.com",
		},
	}
	apitest.CreateUser(newUser, env)
	sessionCookie = apitest.Login(newUser, env)
	// TODO get rid of role assignment when it is handled by user creation
	userRole := roles.New("user", 0)
	e := userRole.AddTo(newUser.ID)
	if e != nil {
		panic(fmt.Sprintf("error adding role to user: %s", e.JSON()))
	}

	hunt := hunts.Hunt{
		HuntDB: db.HuntDB{
			Name:         "DeleteHuntHandler 1 hunt",
			MaxTeams:     43,
			StartTime:    time.Now().AddDate(0, 0, 1),
			EndTime:      time.Now().AddDate(0, 0, 2),
			LocationName: "Fake Location",
			Latitude:     34.730705,
			Longitude:    -86.59481,
		},
	}
	apitest.CreateHunt(&hunt, env, sessionCookie)

	cases := []struct {
		name string
		code int
		user *users.User
		hunt *hunts.Hunt
	}{
		{
			name: "valid team and user",
			code: http.StatusOK,
			user: newUser,
			hunt: &hunt,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			req, err := http.NewRequest(
				"DELETE",
				config.BaseAPIURL+fmt.Sprintf("hunts/%d", c.hunt.ID),
				nil,
			)
			if err != nil {
				t.Fatalf("error getting new request: %v", err)
			}
			req.AddCookie(sessionCookie)

			rr := httptest.NewRecorder()
			router := routes.Routes(env)
			router.ServeHTTP(rr, req)
			res := rr.Result()

			if res.StatusCode != c.code {
				resBody, err := ioutil.ReadAll(res.Body)
				if err != nil {
					t.Fatalf("error reading res body: %v", err)
				}
				t.Fatalf("expected code %d got %d: %s", c.code, res.StatusCode, resBody)
			}

			if c.code == http.StatusOK {
				userRoles, e := db.RolesForUser(c.user.ID)
				if e != nil {
					t.Fatalf("error getting roles for user: %s", e.JSON())
				}

				if len(userRoles) != 2 {
					t.Fatalf(
						"expected hunt roles to be deleted with the hunt got %d",
						len(userRoles),
					)
				}
			}
		})
	}
}

func TestGetHuntsHandler(t *testing.T) {

}

func TestCreateHuntInvitationHandler(t *testing.T) {
	cases := []struct {
		name       string
		code       int
		hunt       hunts.Hunt
		user       *users.User
		invitation db.HuntInvitationDB
	}{
		{
			name: "valid case",
			code: http.StatusOK,
			hunt: hunts.Hunt{
				HuntDB: db.HuntDB{
					Name:         "CreateHuntInvitationHandler 1 hunt",
					MaxTeams:     43,
					StartTime:    time.Now().AddDate(0, 0, 1),
					EndTime:      time.Now().AddDate(0, 0, 2),
					LocationName: "Fake Location",
					Latitude:     34.730705,
					Longitude:    -86.59481,
				},
			},
			user: newUser,
			invitation: db.HuntInvitationDB{
				Email: newUser.Email,
			},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			reqBody, err := json.Marshal(&c.invitation)
			if err != nil {
				t.Fatalf("error marshalling invitation: %v", err)
			}

			apitest.CreateHunt(&c.hunt, env, sessionCookie)

			req, err := http.NewRequest(
				"POST",
				config.BaseAPIURL+fmt.Sprintf("hunts/%d/invitations/", c.hunt.ID),
				bytes.NewReader(reqBody),
			)
			if err != err {
				t.Fatalf("error getting new request: %v", err)
			}
			req.AddCookie(sessionCookie)

			rr := httptest.NewRecorder()
			router := routes.Routes(env)
			router.ServeHTTP(rr, req)
			res := rr.Result()

			if res.StatusCode != c.code {
				resBody, err := ioutil.ReadAll(res.Body)
				if err != nil {
					t.Fatalf("error reading res body: %v", err)
				}
				t.Fatalf("expected code %d got %d: %s", c.code, res.StatusCode, resBody)
			}

			if c.code == http.StatusOK {

				invitations, e := db.GetHuntInvitationsByUserID(c.user.ID)
				if e != nil {
					t.Fatalf("error getting invitations for user: %s", e.JSON())
				}

				if len(invitations) == 0 {
					t.Fatalf("expected user to have invitations")
				}

				e = db.DeleteHuntInvitation(invitations[0].ID)
				if e != nil {
					t.Fatalf("error deleting newly created invitation: %s", e.JSON())
				}

				teams, e := db.TeamsForHunt(c.hunt.ID)
				if e != nil {
					t.Fatalf("error getting teams for hunt: %s", e.JSON())
				}

				e = roles.DeleteRolesForHunt(c.hunt.ID, teams)
				if e != nil {
					t.Fatalf("error deleting roles for newly created hunt: %s", e.JSON())
				}
			}
		})
	}
}
