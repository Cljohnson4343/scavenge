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
	"time"

	"github.com/go-chi/chi"

	"github.com/cljohnson4343/scavenge/apitest"
	"github.com/cljohnson4343/scavenge/config"
	"github.com/cljohnson4343/scavenge/db"
	"github.com/cljohnson4343/scavenge/hunts"
	"github.com/cljohnson4343/scavenge/response"
	"github.com/cljohnson4343/scavenge/roles"
	"github.com/cljohnson4343/scavenge/teams"
	"github.com/cljohnson4343/scavenge/users"
)

var env *config.Env
var sessionCookie *http.Cookie
var newUser = users.User{
	UserDB: db.UserDB{
		FirstName: "authorization",
		LastName:  "tests",
		Username:  "cj4343",
		Email:     "cj_4343@gmail.com",
	},
}

func TestMain(m *testing.M) {
	d := db.InitDB("../db/db_info_test.json")
	defer db.Shutdown(d)

	env = config.CreateEnv(d)
	response.SetDevMode(true)

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

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	router := chi.NewMux()
	router.Use(users.WithUser)
	router.Use(roles.RequireAuth)
	router.Mount("/", handler)

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
			config.BaseAPIURL+url,
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

		route = config.BaseAPIURL + route

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

func TestDeleteRolesForTeam(t *testing.T) {
	var teamOwner = users.User{
		UserDB: db.UserDB{
			FirstName: "team",
			LastName:  "owner",
			Username:  "delete_roles_team_owner",
			Email:     "delete_roles_team_owner@gmail.com",
		},
	}
	apitest.CreateUser(&teamOwner, env)
	var teamEditor = users.User{
		UserDB: db.UserDB{
			FirstName: "team",
			LastName:  "editor",
			Username:  "delete_roles_team_editor",
			Email:     "delete_roles_team_editor@gmail.com",
		},
	}
	apitest.CreateUser(&teamEditor, env)
	var teamMember = users.User{
		UserDB: db.UserDB{
			FirstName: "team",
			LastName:  "member",
			Username:  "delete_roles_team_member",
			Email:     "delete_roles_team_member@gmail.com",
		},
	}
	apitest.CreateUser(&teamMember, env)

	cases := []struct {
		name     string
		role     string
		userID   int
		teamID   int
		numRoles int
	}{
		{
			name:     "delete team owner's roles",
			role:     "team_owner",
			userID:   teamOwner.ID,
			teamID:   3,
			numRoles: 3,
		},
		{
			name:     "delete team editor's roles",
			role:     "team_editor",
			userID:   teamEditor.ID,
			teamID:   3,
			numRoles: 2,
		},
		{
			name:     "delete team member's roles",
			role:     "team_member",
			userID:   teamEditor.ID,
			teamID:   3,
			numRoles: 1,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			role := roles.New(c.role, c.teamID)
			e := role.AddTo(c.userID)
			if e != nil {
				t.Fatalf("error adding role to user %d: %s", c.userID, e.JSON())
			}

			roleDBs, e := db.RolesForUser(c.userID)
			if len(roleDBs) != c.numRoles {
				t.Fatalf("expected %d roles got %d", c.numRoles, len(roleDBs))
			}

			e = roles.DeleteRolesForTeam(c.teamID)
			if e != nil {
				t.Fatalf("error deleting roles for team %d: %s", c.teamID, e.JSON())
			}

			roleDBs, e = db.RolesForUser(c.userID)
			if len(roleDBs) != 0 {
				t.Fatalf("expected %d roles got %d", 0, len(roleDBs))
			}

			perms, e := db.PermissionsForUser(c.userID)
			if e != nil {
				t.Fatalf("error getting permissions for user %d: %s", c.userID, e.JSON())
			}

			if len(perms) != 0 {
				t.Fatalf("expected %d permissions got %d", 0, len(perms))
			}
		})
	}
}

func TestDeleteRolesForHunt(t *testing.T) {
	var huntOwner = users.User{
		UserDB: db.UserDB{
			FirstName: "hunt",
			LastName:  "owner",
			Username:  "delete_roles_hunt_owner",
			Email:     "delete_roles_hunt_owner@gmail.com",
		},
	}
	apitest.CreateUser(&huntOwner, env)
	var huntEditor = users.User{
		UserDB: db.UserDB{
			FirstName: "hunt",
			LastName:  "editor",
			Username:  "delete_roles_hunt_editor",
			Email:     "delete_roles_hunt_editor@gmail.com",
		},
	}
	apitest.CreateUser(&huntEditor, env)

	cases := []struct {
		name     string
		huntRole string
		user     *users.User
		numRoles int
	}{
		{
			name:     "delete hunt owner's roles",
			huntRole: "hunt_owner",
			user:     &huntOwner,
			numRoles: 6,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			sessionCookie := apitest.Login(c.user, env)

			// Create hunt and add hunt role to user
			hunt := hunts.Hunt{
				HuntDB: db.HuntDB{
					Name:         fmt.Sprintf("delete roles for %s", c.huntRole),
					MaxTeams:     43,
					StartTime:    time.Now().AddDate(0, 0, 1),
					EndTime:      time.Now().AddDate(0, 0, 2),
					LocationName: "Fake Location",
					Latitude:     34.730705,
					Longitude:    -86.59481,
				},
			}
			apitest.CreateHunt(&hunt, env, sessionCookie)

			huntRole := roles.New(c.huntRole, hunt.ID)
			e := huntRole.AddTo(c.user.ID)
			if e != nil {
				t.Fatalf("error adding role to user %d: %s", c.user.ID, e.JSON())
			}

			// Create team and add it to hunt
			team := teams.Team{
				TeamDB: db.TeamDB{
					HuntID: hunt.ID,
					Name:   c.name,
				},
			}
			apitest.CreateTeam(&team, env, sessionCookie)

			roleDBs, e := db.RolesForUser(c.user.ID)
			if len(roleDBs) != c.numRoles {
				t.Fatalf("expected %d roles got %d", c.numRoles, len(roleDBs))
			}

			e = roles.DeleteRolesForHunt(hunt.ID)
			if e != nil {
				t.Fatalf("error deleting roles for team %d: %s", hunt.ID, e.JSON())
			}

			roleDBs, e = db.RolesForUser(c.user.ID)
			if len(roleDBs) != 0 {
				t.Fatalf("expected %d roles got %d", 0, len(roleDBs))
			}

			perms, e := db.PermissionsForUser(c.user.ID)
			if e != nil {
				t.Fatalf("error getting permissions for user %d: %s", c.user.ID, e.JSON())
			}

			if len(perms) != 0 {
				t.Fatalf("expected %d permissions got %d", 0, len(perms))
			}

		})
	}
}
