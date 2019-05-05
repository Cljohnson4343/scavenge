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
var newUsers map[string]*users.User = map[string]*users.User{
	"team_owner": &users.User{
		UserDB: db.UserDB{
			FirstName: "team",
			LastName:  "owner",
			Username:  "team_owner",
			Email:     "team_owner43@gmail.com",
		},
	},
	"team_editor": &users.User{
		UserDB: db.UserDB{
			FirstName: "team",
			LastName:  "editor",
			Username:  "team_editor",
			Email:     "team_editor@gmail.com",
		},
	},
	"team_member": &users.User{
		UserDB: db.UserDB{
			FirstName: "team",
			LastName:  "member",
			Username:  "team_member",
			Email:     "team_member@gmail.com",
		},
	},
	"hunt_owner": &users.User{
		UserDB: db.UserDB{
			FirstName: "hunt",
			LastName:  "owner",
			Username:  "hunt_owner",
			Email:     "hunt_owner43@gmail.com",
		},
	},
	"hunt_editor": &users.User{
		UserDB: db.UserDB{
			FirstName: "hunt",
			LastName:  "editor",
			Username:  "hunt_editor",
			Email:     "hunt_editor@gmail.com",
		},
	},
	"hunt_member": &users.User{
		UserDB: db.UserDB{
			FirstName: "hunt",
			LastName:  "member",
			Username:  "hunt_member",
			Email:     "hunt_member@gmail.com",
		},
	},
	"user": &users.User{
		UserDB: db.UserDB{
			FirstName: "user",
			LastName:  "user",
			Username:  "user",
			Email:     "user@gmail.com",
		},
	},
	"user_owner": &users.User{
		UserDB: db.UserDB{
			FirstName: "user",
			LastName:  "owner",
			Username:  "user_owner",
			Email:     "user_owner@gmail.com",
		},
	},
}

func TestMain(m *testing.M) {
	d := db.InitDB("../db/db_info_test.json")
	defer db.Shutdown(d)

	env = c.CreateEnv(d)
	response.SetDevMode(true)

	// Login in user to get a valid user session cookie
	for _, v := range newUsers {
		apitest.CreateUser(v, env)
	}

	os.Exit(m.Run())
}

// TODO add failure test cases, especially for non-existent users
func TestAddRoles(t *testing.T) {
	cases := []struct {
		name     string
		role     string
		numRoles int
		userID   int
	}{
		{
			name:     "team owner",
			role:     "team_owner",
			userID:   newUsers["team_owner"].ID,
			numRoles: 3,
		},
		{
			name:     "team editor",
			role:     "team_editor",
			userID:   newUsers["team_editor"].ID,
			numRoles: 2,
			//
		},
		{
			name:     "team member",
			role:     "team_member",
			userID:   newUsers["team_member"].ID,
			numRoles: 1,
		},
		{
			name:     "hunt owner",
			role:     "hunt_owner",
			userID:   newUsers["hunt_owner"].ID,
			numRoles: 3,
		},
		{
			name:     "hunt editor",
			role:     "hunt_editor",
			userID:   newUsers["hunt_editor"].ID,
			numRoles: 2,
		},
		{
			name:     "hunt member",
			role:     "hunt_member",
			userID:   newUsers["hunt_member"].ID,
			numRoles: 1,
		},
		{
			name:     "user",
			role:     "user",
			userID:   newUsers["user"].ID,
			numRoles: 1,
		},
		{
			name:     "user owner",
			role:     "user_owner",
			userID:   newUsers["user_owner"].ID,
			numRoles: 1,
		},
	}

	entityID := 43

	// the is
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			role := roles.New(c.role, entityID)
			roleDBs := role.RoleDBs(c.userID)
			e := db.AddRoles(roleDBs)
			if e != nil {
				t.Fatalf(
					"expected no errors for user %d but got: \n%s",
					c.userID,
					e.JSON(),
				)
			}

			got, e := db.RolesForUser(c.userID)
			if e != nil {
				t.Fatalf("error getting roles: \n%s", e.JSON())
			}
			if len(got) != c.numRoles {
				t.Fatalf("expected %d roles got %d", c.numRoles, len(got))
			}

			numPerms := 0
			for _, r := range roleDBs {
				numPerms += len(r.Permissions)
			}

			perms, e := db.PermissionsForUser(c.userID)
			if e != nil {
				t.Fatalf("error getting permissions: \n%s", e.JSON())
			}

			if numPerms != len(perms) {
				t.Fatalf("expected %d permissions got %d", numPerms, len(perms))
			}
		})
	}
}
