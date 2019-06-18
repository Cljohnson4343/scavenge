// +build integration

package db_test

import (
	"fmt"
	"net/http"
	"os"
	"testing"

	"github.com/cljohnson4343/scavenge/apitest"
	"github.com/cljohnson4343/scavenge/config"
	"github.com/cljohnson4343/scavenge/db"
	"github.com/cljohnson4343/scavenge/response"
	"github.com/cljohnson4343/scavenge/roles"
	"github.com/cljohnson4343/scavenge/users"
)

var env *config.Env
var sessionCookie *http.Cookie
var addRolesUsers map[string]*users.User = map[string]*users.User{
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
	if err := config.Read(""); err != nil {
		fmt.Printf("unable to read config: %v\n", err)
		os.Exit(1)
	}

	d := db.InitDB("testing")
	defer db.Shutdown(d)

	env = config.CreateEnv(d)
	response.SetDevMode(true)

	// Login in user to get a valid user session cookie
	for _, v := range addRolesUsers {
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
			userID:   addRolesUsers["team_owner"].ID,
			numRoles: 5,
		},
		{
			name:     "team editor",
			role:     "team_editor",
			userID:   addRolesUsers["team_editor"].ID,
			numRoles: 4,
		},
		{
			name:     "team member",
			role:     "team_member",
			userID:   addRolesUsers["team_member"].ID,
			numRoles: 3,
		},
		{
			name:     "hunt owner",
			role:     "hunt_owner",
			userID:   addRolesUsers["hunt_owner"].ID,
			numRoles: 5,
		},
		{
			name:     "hunt editor",
			role:     "hunt_editor",
			userID:   addRolesUsers["hunt_editor"].ID,
			numRoles: 4,
		},
		{
			name:     "hunt member",
			role:     "hunt_member",
			userID:   addRolesUsers["hunt_member"].ID,
			numRoles: 3,
		},
		{
			name:     "user",
			role:     "user",
			userID:   addRolesUsers["user"].ID,
			numRoles: 2,
		},
		{
			name:     "user owner",
			role:     "user_owner",
			userID:   addRolesUsers["user_owner"].ID,
			numRoles: 3,
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
				t.Fatalf("error getting roles for user %d: \n%s", c.userID, e.JSON())
			}
			if len(got) != c.numRoles {
				t.Fatalf("expected %d roles got %d", c.numRoles, len(got))
			}
		})
	}
}

func TestRemoveRole(t *testing.T) {
	cases := []struct {
		name     string
		role     string
		numRoles int
	}{
		{
			name:     "team owner",
			role:     "team_owner",
			numRoles: 3,
		},
		{
			name:     "team editor",
			role:     "team_editor",
			numRoles: 2,
		},
		{
			name:     "team member",
			role:     "team_member",
			numRoles: 1,
		},
		{
			name:     "hunt owner",
			role:     "hunt_owner",
			numRoles: 3,
		},
		{
			name:     "hunt editor",
			role:     "hunt_editor",
			numRoles: 2,
		},
		{
			name:     "hunt member",
			role:     "hunt_member",
			numRoles: 1,
		},
		{
			name:     "user",
			role:     "user",
			numRoles: 1,
		},
		{
			name:     "user owner",
			role:     "user_owner",
			numRoles: 1,
		},
	}

	user := &users.User{
		UserDB: db.UserDB{
			FirstName: "remove",
			LastName:  "role",
			Username:  "remove_role",
			Email:     "remove_role43@gmail.com",
		},
	}
	apitest.CreateUser(user, env)

	userRoles, e := db.RolesForUser(user.ID)
	if e != nil {
		t.Fatalf("error getting roles for user: %s", e.JSON())
	}
	for _, r := range userRoles {
		e = db.RemoveRole(r.ID, user.ID)
		if e != nil {
			t.Fatalf("error removing role from user: %s", e.JSON())
		}
	}

	entityID := 2323

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			role := roles.New(c.role, entityID)
			e := role.AddTo(user.ID)
			if e != nil {
				t.Fatalf("error adding roles to user: %s", e.JSON())
			}

			roleDBs, e := db.RolesForUser(user.ID)
			if e != nil {
				t.Fatalf("error getting roles for user: %s", e.JSON())
			}

			if len(roleDBs) != c.numRoles {
				t.Fatalf("expected to get %d roles got %d", c.numRoles, len(roleDBs))
			}

			for _, r := range roleDBs {
				e = roles.RemoveRole(r.ID, user.ID)
				if e != nil {
					t.Fatalf("error removing role %d from user %d: %s", r.ID, user.ID, e.JSON())
				}
			}

			roleDBs, e = db.RolesForUser(user.ID)
			if e != nil {
				t.Fatalf("error getting roles for user: %s", e.JSON())
			}

			if len(roleDBs) != 0 {
				t.Fatalf("expected all roles were removed got %d roles", len(roleDBs))
			}
		})
	}
}
