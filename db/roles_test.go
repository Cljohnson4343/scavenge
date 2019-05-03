// +build apiTest

package db_test

import (
	"net/http"
	"os"
	"testing"

	"github.com/cljohnson4343/scavenge/apitest"
	c "github.com/cljohnson4343/scavenge/config"
	"github.com/cljohnson4343/scavenge/db"
	"github.com/cljohnson4343/scavenge/response"
	"github.com/cljohnson4343/scavenge/roles"
	"github.com/cljohnson4343/scavenge/users"
)

var env *c.Env
var sessionCookie *http.Cookie
var newUser = users.User{
	UserDB: db.UserDB{
		FirstName: "Rowdy",
		LastName:  "Johnson",
		Username:  "pretty_girl",
		Email:     "pretty_girl43@gmail.com",
	},
}

func TestMain(m *testing.M) {
	d := db.InitDB("../db/db_info_test.json")
	defer db.Shutdown(d)

	env = c.CreateEnv(d)
	response.SetDevMode(true)

	// Login in user to get a valid user session cookie
	apitest.CreateUser(&newUser, env)
	sessionCookie = apitest.Login(&newUser, env)

	os.Exit(m.Run())
}

// TODO add failure test cases, especially for non-existent users
func TestAddRoles(t *testing.T) {
	cases := []struct {
		name string
		role string
	}{
		{
			name: "team owner",
			role: "team_owner",
		},
		{
			name: "team editor",
			role: "team_editor",
		},
		{
			name: "team member",
			role: "team_member",
		},
		{
			name: "hunt owner",
			role: "hunt_owner",
		},
		{
			name: "hunt editor",
			role: "hunt_editor",
		},
		{
			name: "hunt member",
			role: "hunt_member",
		},
		{
			name: "user",
			role: "user",
		},
		{
			name: "user owner",
			role: "user_owner",
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			role := roles.New(c.role, 43)
			got := db.AddRoles(role.RoleDBs(newUser.ID))
			if got != nil {
				t.Errorf("expected no errors but got: \n%s", got.JSON())
			}
		})
	}
}
