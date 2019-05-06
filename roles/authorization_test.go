// +build integration

package roles_test

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"strings"
	"testing"

	"github.com/cljohnson4343/scavenge/apitest"
	c "github.com/cljohnson4343/scavenge/config"
	"github.com/cljohnson4343/scavenge/db"
	"github.com/cljohnson4343/scavenge/response"
	"github.com/cljohnson4343/scavenge/roles"
	"github.com/cljohnson4343/scavenge/routes"
	"github.com/cljohnson4343/scavenge/users"
	"github.com/go-chi/chi"
)

var env *c.Env
var sessionCookie *http.Cookie
var newUser = users.User{
	UserDB: db.UserDB{
		FirstName: "authorization",
		LastName:  "tests",
		Username:  "cj4343",
		Email:     "cj_4343@gmail.com",
	},
}
var router *chi.Mux

func TestMain(m *testing.M) {
	d := db.InitDB("../db/db_info_test.json")
	defer db.Shutdown(d)

	env = c.CreateEnv(d)
	response.SetDevMode(true)

	router = routes.Routes(env, false)

	// Login in user to get a valid user session cookie
	apitest.CreateUser(&newUser, env)
	sessionCookie = apitest.Login(&newUser, env)

	os.Exit(m.Run())
}

func TestRequireAuth(t *testing.T) {
	validEntityID := 43
	invalidEntityID := 23
	requests := append(
		getRoutes(validEntityID, sessionCookie),
		getRoutes(invalidEntityID, sessionCookie)...,
	)

	cases := []struct {
		role string
	}{
		{
			role: "team_owner",
		},
		{
			role: "team_editor",
		},
		{
			role: "team_member",
		},
		{
			role: "hunt_owner",
		},
		{
			role: "hunt_editor",
		},
		{
			role: "hunt_member",
		},
		{
			role: "user_owner",
		},
	}

	for _, c := range cases {
		role := roles.New(c.role, validEntityID)
		e := role.AddTo(newUser.ID)
		if e != nil {
			t.Fatalf("error adding role %s to user %d", c.role, newUser.ID)
		}

		for _, req := range requests {
			t.Run(c.role+req.URL.Path, func(t *testing.T) {
				rr := httptest.NewRecorder()
				router.ServeHTTP(rr, req)
				res := rr.Result()

				expected := expectAuthorized(role.Name, req)
				var got bool
				if res.StatusCode == http.StatusUnauthorized {
					got = false
				} else {
					got = true
				}

				if got != expected {
					resBody, err := ioutil.ReadAll(res.Body)
					if err != nil {
						t.Fatalf("error reading res body: %v", err)
					}

					t.Fatalf(
						"\nexpected authorized: %v\ngot authorized: %v\nreturn code: %d\nres body: %s\nreq method: %s\nreq path: %s\n",
						expected,
						got,
						res.StatusCode,
						resBody,
						req.Method,
						req.URL.Path,
					)
				}
			})
		}
		roleDBs, e := db.RolesForUser(newUser.ID)
		if e != nil {
			t.Fatalf("error getting roles for user: %s", e.JSON())
		}

		for _, r := range roleDBs {
			e := roles.RemoveRole(r.ID, newUser.ID)
			if e != nil {
				t.Fatalf("error removing role from user: %s", e.JSON())
			}
		}
	}
}

func getRoutes(entityID int, cookie *http.Cookie) []*http.Request {
	requests := make([]*http.Request, 0, len(roles.PermToRoutes))
	for k, v := range roles.PermToRoutes {
		var url string
		if strings.Contains(v, "%") {
			url = fmt.Sprintf(v, entityID)
		} else {
			url = v
		}

		req, err := http.NewRequest(
			strings.ToUpper(strings.Split(k, "_")[0]),
			"/api/v0"+url,
			nil,
		)
		if err != nil {
			panic(fmt.Sprintf("error getting new request: %v", err))
		}

		req.AddCookie(cookie)

		requests = append(requests, req)
	}

	return requests
}

func expectAuthorized(roleWID string, req *http.Request) bool {
	roleSplit := strings.Split(roleWID, "_")
	entityID, err := strconv.Atoi(roleSplit[len(roleSplit)-1])
	if err != nil {
		panic("unable to convert string to int")
	}

	var permKey string
	found := false

	for perm, fmtRoute := range roles.PermToRoutes {
		var route string
		if strings.Contains(fmtRoute, "%") {
			route = fmt.Sprintf(fmtRoute, entityID)
		} else {
			route = fmtRoute
		}

		route = "/api/v0" + route

		method := strings.Split(perm, "_")[0]

		if method == strings.ToLower(req.Method) &&
			route == req.URL.Path {
			permKey = perm
			found = true
		}
	}

	if !found {
		return false
	}

	roleStr := strings.Join(roleSplit[:len(roleSplit)-1], "_")
	switch roleStr {
	case "team_owner":
		if roles.PermToRole[permKey] == roleStr ||
			roles.PermToRole[permKey] == "team_editor" ||
			roles.PermToRole[permKey] == "team_member" {
			return true
		}
	case "team_editor":
		if roles.PermToRole[permKey] == roleStr ||
			roles.PermToRole[permKey] == "team_member" {
			return true
		}
	case "team_member":
		if roles.PermToRole[permKey] == roleStr {
			return true
		}
	case "hunt_owner":
		if roles.PermToRole[permKey] == roleStr ||
			roles.PermToRole[permKey] == "hunt_editor" ||
			roles.PermToRole[permKey] == "hunt_member" {
			return true
		}
	case "hunt_editor":
		if roles.PermToRole[permKey] == roleStr ||
			roles.PermToRole[permKey] == "hunt_member" {
			return true
		}
	case "hunt_member":
		if roles.PermToRole[permKey] == roleStr {
			return true
		}
	case "user":
		if roles.PermToRole[permKey] == roleStr {
			return true
		}
	case "user_owner":
		if roles.PermToRole[permKey] == roleStr {
			return true
		}
	}

	return false
}
