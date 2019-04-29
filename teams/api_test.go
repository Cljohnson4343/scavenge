// +build apiTest

package teams_test

import (
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/cljohnson4343/scavenge/apitest"
	c "github.com/cljohnson4343/scavenge/config"
	"github.com/cljohnson4343/scavenge/db"
	"github.com/cljohnson4343/scavenge/hunts"
	"github.com/cljohnson4343/scavenge/response"
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

/*
func TestCreateTeamHandler(t *testing.T) {
	cases := []struct {
		name        string
		newTeamData teams.Team
		statusCode  int
	}{
		{
			name:        "add new team",
			newTeamData: teams.Team{
				HuntID: hunt.ID,
				name: ,
			},
			statusCode:  http.StatusOK,
		},
	}

}

*/
