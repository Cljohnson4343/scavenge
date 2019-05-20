// +build integration

package teams_test

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
	"github.com/cljohnson4343/scavenge/hunts/models"
	"github.com/cljohnson4343/scavenge/response"
	"github.com/cljohnson4343/scavenge/roles"
	"github.com/cljohnson4343/scavenge/routes"
	"github.com/cljohnson4343/scavenge/teams"
	"github.com/cljohnson4343/scavenge/users"
)

var env *config.Env
var hunt hunts.Hunt
var hunt2 hunts.Hunt
var sessionCookie *http.Cookie
var newUser *users.User

func TestMain(m *testing.M) {
	d := db.InitDB("../db/db_info_test.json")
	defer db.Shutdown(d)
	env = config.CreateEnv(d)
	response.SetDevMode(true)

	// Login in user to get a valid user session cookie
	newUser = &users.User{
		UserDB: db.UserDB{
			FirstName: "Fernando",
			LastName:  "Sucre",
			Username:  "sucre_43",
			Email:     "sucre433@gmail.com",
		},
	}
	apitest.CreateUser(newUser, env)
	sessionCookie = apitest.Login(newUser, env)

	// Create hunts to use for tests
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

	hunt2.HuntDB = db.HuntDB{
		Name:         "Teams Second Hunt43",
		MaxTeams:     4,
		StartTime:    time.Now().AddDate(0, 0, 1),
		EndTime:      time.Now().AddDate(0, 0, 2),
		LocationName: "Fake Location",
		Latitude:     34.730705,
		Longitude:    -86.59481,
	}
	apitest.CreateHunt(&hunt2, env, sessionCookie)

	os.Exit(m.Run())
}

func TestCreateTeamHandler(t *testing.T) {
	// Login in user to get a valid user session cookie
	newUser := &users.User{
		UserDB: db.UserDB{
			FirstName: "Test Create Team",
			LastName:  "Test Create Team",
			Username:  "test_create_team",
			Email:     "test_create_team@gmail.com",
		},
	}
	apitest.CreateUser(newUser, env)
	cookie := apitest.Login(newUser, env)
	huntEditor := roles.New("hunt_editor", hunt.ID)
	e := huntEditor.AddTo(newUser.ID)
	if e != nil {
		t.Fatalf("error adding hunt editor role to user: %s", e.JSON())
	}

	cases := []struct {
		name string
		team teams.Team
		code int
	}{
		{
			name: "add new team",
			team: teams.Team{
				TeamDB: db.TeamDB{
					HuntID: hunt.ID,
					Name:   "team 1",
				},
			},
			code: http.StatusOK,
		},
		{
			name: "add team with same name as another team in same hunt",
			team: teams.Team{
				TeamDB: db.TeamDB{
					HuntID: hunt.ID,
					Name:   "team 1",
				},
			},
			code: http.StatusBadRequest,
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
			code: http.StatusBadRequest,
		},
		{
			name: "add team without hunt id",
			team: teams.Team{
				TeamDB: db.TeamDB{
					Name: "team 3",
				},
			},
			code: http.StatusBadRequest,
		},
		{
			name: "add team without name",
			team: teams.Team{
				TeamDB: db.TeamDB{
					HuntID: hunt.ID,
				},
			},
			code: http.StatusBadRequest,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			reqBody, err := json.Marshal(&c.team)
			if err != nil {
				t.Fatalf("error marshalling req data: %v", err)
			}
			req, err := http.NewRequest("POST", config.BaseAPIURL+"teams/", bytes.NewReader(reqBody))
			if err != nil {
				t.Fatalf("error getting new request: %v", err)
			}
			req.AddCookie(cookie)

			rr := httptest.NewRecorder()
			handler := routes.Routes(env)
			handler.ServeHTTP(rr, req)

			res := rr.Result()

			if res.StatusCode != c.code {
				resBody, err := ioutil.ReadAll(res.Body)
				if err != nil {
					t.Fatalf("error reading res body: %v", err)
				}

				t.Fatalf("expected code %d got %d: %s", c.code, res.StatusCode, resBody)
			}

			if c.code == http.StatusOK {
				err = json.NewDecoder(res.Body).Decode(&c.team)
				if err != nil {
					t.Fatalf("error decoding the res body: %v", err)
				}

				if c.team.ID == 0 {
					t.Error("expected id to be returned")
				}

				got, e := roles.UserHasRole("team_owner", c.team.ID, newUser.ID)
				if e != nil {
					t.Fatalf(
						"error determining if user %d has role: %s",
						newUser.ID,
						e.JSON(),
					)
				}

				if !got {
					t.Fatal("expected to have a team_owner role for team.")
				}
			}
		})
	}
}

func TestGetTeamsHandler(t *testing.T) {
	getTeamsUser := &users.User{
		UserDB: db.UserDB{
			FirstName: "GetTeamsHandler",
			LastName:  "GetTeamsHandler",
			Username:  "GetTeamsHandler",
			Email:     "GetTeamsHandler@gmail.com",
		},
	}
	apitest.CreateUser(getTeamsUser, env)
	sessionCookie := apitest.Login(getTeamsUser, env)

	adminRole := roles.New("admin", 0)
	e := adminRole.AddTo(getTeamsUser.ID)
	if e != nil {
		t.Fatalf("error adding role to user: %s", e.JSON())
	}

	// Create hunts to use for tests
	getTeamsHunt := hunts.Hunt{
		HuntDB: db.HuntDB{
			Name:         "GetTeamsHunt 1",
			MaxTeams:     43,
			StartTime:    time.Now().AddDate(0, 0, 1),
			EndTime:      time.Now().AddDate(0, 0, 2),
			LocationName: "Fake Location",
			Latitude:     34.730705,
			Longitude:    -86.59481,
		},
	}
	apitest.CreateHunt(&getTeamsHunt, env, sessionCookie)

	cases := []struct {
		name  string
		code  int
		teams []teams.Team
	}{
		{
			name: "get teams",
			code: http.StatusOK,
			teams: []teams.Team{
				teams.Team{
					TeamDB: db.TeamDB{
						Name:   "first get teams",
						HuntID: getTeamsHunt.ID,
					},
				},
				teams.Team{
					TeamDB: db.TeamDB{
						Name:   "second get teams",
						HuntID: getTeamsHunt.ID,
					},
				},
				teams.Team{
					TeamDB: db.TeamDB{
						Name:   "third get teams",
						HuntID: getTeamsHunt.ID,
					},
				},
				teams.Team{
					TeamDB: db.TeamDB{
						Name:   "fourth get teams",
						HuntID: getTeamsHunt.ID,
					},
				},
				teams.Team{
					TeamDB: db.TeamDB{
						Name:   "fifth get teams",
						HuntID: getTeamsHunt.ID,
					},
				},
				teams.Team{
					TeamDB: db.TeamDB{
						Name:   "sixth get teams",
						HuntID: getTeamsHunt.ID,
					},
				},
				teams.Team{
					TeamDB: db.TeamDB{
						Name:   "seventh get teams",
						HuntID: getTeamsHunt.ID,
					},
				},
				teams.Team{
					TeamDB: db.TeamDB{
						Name:   "eighth get teams",
						HuntID: getTeamsHunt.ID,
					},
				},
			},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			for _, t := range c.teams {
				// use logged in user's session to create each team
				apitest.CreateTeam(&t, env, sessionCookie)
			}

			req, err := http.NewRequest("GET", config.BaseAPIURL+"teams/", nil)
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
					t.Errorf("error reading res body: %v", err)
				}
				t.Fatalf("expected code %d got %d: %s", c.code, res.StatusCode, resBody)
			}

			gotTeams := make([]teams.Team, 0)
			err = json.NewDecoder(res.Body).Decode(&gotTeams)
			if err != nil {
				t.Fatalf("error decoding res body: %v", err)
			}

			if len(gotTeams) <= len(c.teams) {
				t.Fatalf("expected at least %d teams to be returned but got %d", len(c.teams), len(gotTeams))
			}
		})
	}
}

func TestGetTeamHandler(t *testing.T) {
	getTeamUser := &users.User{
		UserDB: db.UserDB{
			FirstName: "GetTeamHandler",
			LastName:  "GetTeamHandler",
			Username:  "GetTeamHandler",
			Email:     "GetTeamHandler@gmail.com",
		},
	}
	apitest.CreateUser(getTeamUser, env)
	sessionCookie := apitest.Login(getTeamUser, env)
	// Create hunts to use for tests
	getTeamHunt := hunts.Hunt{
		HuntDB: db.HuntDB{
			Name:         "GetTeamHunt 1",
			MaxTeams:     43,
			StartTime:    time.Now().AddDate(0, 0, 1),
			EndTime:      time.Now().AddDate(0, 0, 2),
			LocationName: "Fake Location",
			Latitude:     34.730705,
			Longitude:    -86.59481,
		},
	}
	apitest.CreateHunt(&getTeamHunt, env, sessionCookie)
	expected := teams.Team{
		TeamDB: db.TeamDB{
			Name:   "GetTeamHunt",
			HuntID: getTeamHunt.ID,
		},
	}
	apitest.CreateTeam(&expected, env, sessionCookie)

	cases := []struct {
		name   string
		code   int
		teamID int
	}{
		{
			name:   "valid team",
			code:   http.StatusOK,
			teamID: expected.ID,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			req, err := http.NewRequest(
				"GET",
				config.BaseAPIURL+fmt.Sprintf("teams/%d", c.teamID),
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
					t.Errorf("error reading res body: %v", err)
				}
				t.Fatalf("expected code %d got %d: %s", c.code, res.StatusCode, resBody)
			}

			if c.code == http.StatusOK {
				got := teams.Team{}
				err = json.NewDecoder(res.Body).Decode(&got)
				if err != nil {
					t.Fatalf("error decoding team: %v", err)
				}

				compareTeams(t, &expected, &got)
			}
		})
	}
}

func TestDeleteTeamHandler(t *testing.T) {
	deleteTeamUser := &users.User{
		UserDB: db.UserDB{
			FirstName: "DeleteTeamHandler",
			LastName:  "DeleteTeamHandler",
			Username:  "DeleteTeamHandler",
			Email:     "DeleteTeamHandler@gmail.com",
		},
	}
	apitest.CreateUser(deleteTeamUser, env)
	sessionCookie := apitest.Login(deleteTeamUser, env)
	// Create hunts to use for tests
	deleteTeamHunt := hunts.Hunt{
		HuntDB: db.HuntDB{
			Name:         "DeleteTeamHunt 1",
			MaxTeams:     43,
			StartTime:    time.Now().AddDate(0, 0, 1),
			EndTime:      time.Now().AddDate(0, 0, 2),
			LocationName: "Fake Location",
			Latitude:     34.730705,
			Longitude:    -86.59481,
		},
	}
	apitest.CreateHunt(&deleteTeamHunt, env, sessionCookie)
	expected := teams.Team{
		TeamDB: db.TeamDB{
			Name:   "GetTeamHunt",
			HuntID: deleteTeamHunt.ID,
		},
	}
	apitest.CreateTeam(&expected, env, sessionCookie)

	team := teams.Team{
		TeamDB: db.TeamDB{
			Name:   "fox river eight",
			HuntID: deleteTeamHunt.ID,
		},
	}
	apitest.CreateTeam(&team, env, sessionCookie)

	cases := []struct {
		name string
		code int
		id   int
	}{
		{
			name: "valid team",
			code: http.StatusOK,
			id:   team.ID,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			req, err := http.NewRequest(
				"DELETE",
				config.BaseAPIURL+fmt.Sprintf("teams/%d", c.id),
				nil,
			)
			if err != nil {
				t.Fatalf("error creating new request: %v", err)
			}
			req.AddCookie(sessionCookie)

			rr := httptest.NewRecorder()
			router := routes.Routes(env)
			router.ServeHTTP(rr, req)

			res := rr.Result()

			if res.StatusCode != c.code {
				resBody, err := ioutil.ReadAll(res.Body)
				if err != nil {
					t.Errorf("error reading the res body: %v", err)
				}

				t.Fatalf("expoected code %d got %d: %s", c.code, res.StatusCode, resBody)
			}
		})
	}
}

/*
func TestPatchTeamHandler(t *testing.T) {
	team := teams.Team{
		TeamDB: db.TeamDB{
			Name:   "Breaking Bad",
			HuntID: hunt.ID,
		},
	}
	apitest.CreateTeam(&team, env, sessionCookie)

	cases := []struct {
		name   string
		code   int
		update teams.Team
		id     int
	}{
		{
			name: "update valid team",
			id:   team.ID,
			code: http.StatusOK,
			update: teams.Team{
				TeamDB: db.TeamDB{
					Name: "patched team name",
				},
			},
		},
		{
			name: "update hunt id",
			id:   team.ID,
			code: http.StatusBadRequest,
			update: teams.Team{
				TeamDB: db.TeamDB{
					HuntID: 43,
				},
			},
		},
		{
			name: "update id",
			id:   team.ID,
			code: http.StatusBadRequest,
			update: teams.Team{
				TeamDB: db.TeamDB{
					ID: 43,
				},
			},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			reqBody, err := json.Marshal(&c.update)
			if err != nil {
				t.Fatalf("error marshalling update data: %v", err)
			}

			req, err := http.NewRequest(
				"PATCH",
				fmt.Sprintf("/%d", c.id),
				bytes.NewReader(reqBody),
			)
			if err != nil {
				t.Fatalf("error getting new request: %v", err)
			}

			rr := httptest.NewRecorder()
			handler := teams.Routes(env)
			handler.ServeHTTP(rr, req)
			res := rr.Result()

			if res.StatusCode != c.code {
				resBody, err := ioutil.ReadAll(res.Body)
				if err != nil {
					t.Errorf("error reading res body: %v", err)
				}

				t.Fatalf("expected code %d got %d: %s", c.code, res.StatusCode, resBody)
			}
		})
	}

}
*/

func TestCreateMediaHandler(t *testing.T) {
	team := teams.Team{
		TeamDB: db.TeamDB{
			Name:   "Create media team",
			HuntID: hunt.ID,
		},
	}
	apitest.CreateTeam(&team, env, sessionCookie)

	item := models.Item{
		ItemDB: db.ItemDB{
			HuntID: hunt.ID,
			Name:   "easter egg",
			Points: 43,
		},
	}
	apitest.CreateItem(&item, env, sessionCookie)

	duplicatTime := time.Now().AddDate(0, 0, -2)
	cases := []struct {
		name  string
		code  int
		media db.MediaMetaDB
	}{
		{
			name: "valid team",
			code: http.StatusOK,
			media: db.MediaMetaDB{
				TeamID: team.ID,
				URL:    "amazon.com/cdn/media",
				Location: db.LocationDB{
					TimeStamp: time.Now().AddDate(0, 0, -1),
					Latitude:  34.730705,
					Longitude: -86.59481,
					TeamID:    team.ID,
				},
			},
		},
		{
			name: "valid team with item id",
			code: http.StatusOK,
			media: db.MediaMetaDB{
				TeamID: team.ID,
				ItemID: item.ID,
				URL:    "amazon.com/cdn/media",
				Location: db.LocationDB{
					TimeStamp: duplicatTime,
					Latitude:  34.730705,
					Longitude: -86.59481,
					TeamID:    team.ID,
				},
			},
		},
		{
			name: "same location and timestamp",
			code: http.StatusBadRequest,
			media: db.MediaMetaDB{
				TeamID: team.ID,
				ItemID: item.ID,
				URL:    "amazon.com/cdn/media",
				Location: db.LocationDB{
					TimeStamp: duplicatTime,
					Latitude:  34.730705,
					Longitude: -86.59481,
					TeamID:    team.ID,
				},
			},
		},
		{
			name: "with media id",
			code: http.StatusBadRequest,
			media: db.MediaMetaDB{
				TeamID: team.ID,
				ID:     43,
				ItemID: item.ID,
				URL:    "amazon.com/cdn/media",
				Location: db.LocationDB{
					TimeStamp: time.Now().AddDate(0, 0, -3),
					Latitude:  34.730705,
					Longitude: -86.59481,
					TeamID:    0,
				},
			},
		},
		{
			name: "with location id",
			code: http.StatusBadRequest,
			media: db.MediaMetaDB{
				TeamID: team.ID,
				ID:     43,
				ItemID: item.ID,
				URL:    "amazon.com/cdn/media",
				Location: db.LocationDB{
					ID:        43,
					TimeStamp: time.Now().AddDate(0, 0, -3),
					Latitude:  34.730705,
					Longitude: -86.59481,
					TeamID:    0,
				},
			},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			reqBody, err := json.Marshal(&c.media)
			if err != nil {
				t.Fatalf("error marshalling request: %v", err)
			}

			req, err := http.NewRequest(
				"POST",
				config.BaseAPIURL+fmt.Sprintf("teams/%d/media/", team.ID),
				bytes.NewReader(reqBody),
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
					t.Errorf("error reading res body: %v", err)
				}

				t.Fatalf(
					"expected code %d got %d: %s",
					c.code,
					res.StatusCode,
					resBody,
				)
			}

			if c.code == http.StatusOK {
				err = json.NewDecoder(res.Body).Decode(&c.media)
				if err != nil {
					t.Fatalf("error decoding res body: %v", err)
				}

				if c.media.ID == 0 {
					t.Errorf("expected media id to be returned")
				}

				if c.media.Location.ID == 0 {
					t.Errorf("expected location id to be returned")
				}
			}
		})
	}
}

func compareTeams(t *testing.T, expected *teams.Team, got *teams.Team) {
	if got.ID != expected.ID {
		t.Errorf("expected id to be %d got %d", expected.ID, got.ID)
	}

	if got.HuntID != expected.HuntID {
		t.Errorf("expected hunt id to be %d got %d", expected.HuntID, got.HuntID)
	}

	if got.Name != expected.Name {
		t.Errorf("expected name to be %s got %s", expected.Name, got.Name)
	}
}

func TestGetMediaForTeamHandler(t *testing.T) {
	team := teams.Team{
		TeamDB: db.TeamDB{
			Name:   "Get media team",
			HuntID: hunt.ID,
		},
	}
	apitest.CreateTeam(&team, env, sessionCookie)

	media := []db.MediaMetaDB{
		{
			TeamID: team.ID,
			URL:    "amazon.com/cdn/media",
			Location: db.LocationDB{
				TimeStamp: time.Now().AddDate(0, 0, -1),
				Latitude:  34.730705,
				Longitude: -86.59481,
				TeamID:    team.ID,
			},
		},
		{
			TeamID: team.ID,
			URL:    "amazon.com/cdn/media",
			Location: db.LocationDB{
				TimeStamp: time.Now().AddDate(0, -1, 0),
				Latitude:  34.730705,
				Longitude: -86.59481,
				TeamID:    team.ID,
			},
		},
		{
			TeamID: team.ID,
			URL:    "amazon.com/cdn/media",
			Location: db.LocationDB{
				TimeStamp: time.Now().AddDate(0, 0, -3),
				Latitude:  34.730705,
				Longitude: -86.59481,
				TeamID:    team.ID,
			},
		},
		{
			TeamID: team.ID,
			URL:    "amazon.com/cdn/media",
			Location: db.LocationDB{
				TimeStamp: time.Now().AddDate(0, 0, -4),
				Latitude:  34.730705,
				Longitude: -86.59481,
				TeamID:    team.ID,
			},
		},
		{
			TeamID: team.ID,
			URL:    "amazon.com/cdn/media",
			Location: db.LocationDB{
				TimeStamp: time.Now().AddDate(0, 0, -5),
				Latitude:  34.730705,
				Longitude: -86.59481,
				TeamID:    team.ID,
			},
		},
		{
			TeamID: team.ID,
			URL:    "amazon.com/cdn/media",
			Location: db.LocationDB{
				TimeStamp: time.Now().AddDate(0, 0, -6),
				Latitude:  34.730705,
				Longitude: -86.59481,
				TeamID:    team.ID,
			},
		},
		{
			TeamID: team.ID,
			URL:    "amazon.com/cdn/media",
			Location: db.LocationDB{
				TimeStamp: time.Now().AddDate(0, 0, -7),
				Latitude:  34.730705,
				Longitude: -86.59481,
				TeamID:    team.ID,
			},
		},
		{
			TeamID: team.ID,
			URL:    "amazon.com/cdn/media",
			Location: db.LocationDB{
				TimeStamp: time.Now().AddDate(0, 0, -8),
				Latitude:  34.730705,
				Longitude: -86.59481,
				TeamID:    team.ID,
			},
		},
		{
			TeamID: team.ID,
			URL:    "amazon.com/cdn/media",
			Location: db.LocationDB{
				TimeStamp: time.Now().AddDate(0, 0, -9),
				Latitude:  34.730705,
				Longitude: -86.59481,
				TeamID:    team.ID,
			},
		},
		{
			TeamID: team.ID,
			URL:    "amazon.com/cdn/media",
			Location: db.LocationDB{
				TimeStamp: time.Now().AddDate(0, 0, -10),
				Latitude:  34.730705,
				Longitude: -86.59481,
				TeamID:    team.ID,
			},
		},
		{
			TeamID: team.ID,
			URL:    "amazon.com/cdn/media",
			Location: db.LocationDB{
				TimeStamp: time.Now().AddDate(0, 0, -11),
				Latitude:  34.730705,
				Longitude: -86.59481,
				TeamID:    team.ID,
			},
		},
		{
			TeamID: team.ID,
			URL:    "amazon.com/cdn/media",
			Location: db.LocationDB{
				TimeStamp: time.Now().AddDate(0, 0, -12),
				Latitude:  34.730705,
				Longitude: -86.59481,
				TeamID:    team.ID,
			},
		},
		{
			TeamID: team.ID,
			URL:    "amazon.com/cdn/media",
			Location: db.LocationDB{
				TimeStamp: time.Now().AddDate(0, 0, -13),
				Latitude:  34.730705,
				Longitude: -86.59481,
				TeamID:    team.ID,
			},
		},
		{
			TeamID: team.ID,
			URL:    "amazon.com/cdn/media",
			Location: db.LocationDB{
				TimeStamp: time.Now().AddDate(0, 0, -14),
				Latitude:  34.730705,
				Longitude: -86.59481,
				TeamID:    team.ID,
			},
		},
	}
	for _, m := range media {
		apitest.CreateMedia(&m, env, sessionCookie)
	}

	cases := []struct {
		name          string
		code          int
		teamID        int
		returnedMedia int
	}{
		{
			name:          "valid team",
			code:          http.StatusOK,
			teamID:        team.ID,
			returnedMedia: len(media),
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			req, err := http.NewRequest(
				"GET",
				config.BaseAPIURL+fmt.Sprintf("teams/%d/media/", c.teamID),
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
					t.Errorf("error reading res body: %v", err)
				}

				t.Fatalf("expected code %d got %d: %s", c.code, res.StatusCode, resBody)
			}

			got := make([]db.MediaMetaDB, 0, len(media))
			err = json.NewDecoder(res.Body).Decode(&got)
			if err != nil {
				t.Fatalf("error decoding the response body: %v", err)
			}

			if c.code == http.StatusOK {
				if len(got) != c.returnedMedia {
					t.Errorf(
						"expected to recieve %d media entities got %d",
						c.returnedMedia,
						len(got),
					)
				}
			}
		})
	}
}

func TestDeleteMediaHandler(t *testing.T) {
	team := teams.Team{
		TeamDB: db.TeamDB{
			Name:   "Delete media team",
			HuntID: hunt.ID,
		},
	}
	apitest.CreateTeam(&team, env, sessionCookie)

	media := db.MediaMetaDB{
		TeamID: team.ID,
		URL:    "amazon.com/cdn/media",
		Location: db.LocationDB{
			TimeStamp: time.Now().AddDate(0, 0, -1),
			Latitude:  34.730705,
			Longitude: -86.59481,
			TeamID:    team.ID,
		},
	}
	apitest.CreateMedia(&media, env, sessionCookie)

	cases := []struct {
		name    string
		code    int
		mediaID int
		teamID  int
	}{
		{
			name:    "valid team and media",
			code:    http.StatusOK,
			mediaID: media.ID,
			teamID:  team.ID,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			req, err := http.NewRequest(
				"DELETE",
				config.BaseAPIURL+fmt.Sprintf("teams/%d/media/%d", c.teamID, c.mediaID),
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
					t.Errorf("error reading response body: %v", err)
				}

				t.Fatalf("expected code %d got %d: %s", c.code, res.StatusCode, resBody)
			}
		})
	}
}

func TestCreateLocationHandler(t *testing.T) {
	team := teams.Team{
		TeamDB: db.TeamDB{
			Name:   "create location team",
			HuntID: hunt.ID,
		},
	}
	apitest.CreateTeam(&team, env, sessionCookie)

	cases := []struct {
		name     string
		code     int
		location db.LocationDB
	}{
		{
			name: "valid team",
			code: http.StatusOK,
			location: db.LocationDB{
				TimeStamp: time.Now().AddDate(0, 0, -1),
				Latitude:  34.730705,
				Longitude: -86.59481,
				TeamID:    team.ID,
			},
		},
		{
			name: "time stamp in the future",
			code: http.StatusBadRequest,
			location: db.LocationDB{
				TimeStamp: time.Now().AddDate(0, 0, 1),
				Latitude:  34.730705,
				Longitude: -86.59481,
				TeamID:    team.ID,
			},
		},
		{
			name: "no longitude",
			code: http.StatusBadRequest,
			location: db.LocationDB{
				TimeStamp: time.Now().AddDate(0, 0, -2),
				Latitude:  34.730705,
				TeamID:    team.ID,
			},
		},
		{
			name: "no latitude",
			code: http.StatusBadRequest,
			location: db.LocationDB{
				TimeStamp: time.Now().AddDate(0, 0, -4),
				Longitude: -86.59481,
				TeamID:    team.ID,
			},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			reqBody, err := json.Marshal(&c.location)
			if err != nil {
				t.Fatalf("error marshalling the request data: %v", err)
			}

			req, err := http.NewRequest(
				"POST",
				config.BaseAPIURL+fmt.Sprintf("teams/%d/locations/", c.location.TeamID),
				bytes.NewReader(reqBody),
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
					t.Errorf("error reading response body: %v", err)
				}

				t.Fatalf("expected code %d got %d: %s", c.code, res.StatusCode, resBody)
			}

			if c.code == http.StatusOK {
				err = json.NewDecoder(res.Body).Decode(&c.location)
				if err != nil {
					t.Fatalf("error decoding response: %v", err)
				}

				if c.location.ID == 0 {
					t.Errorf("expected location id to be returned")
				}
			}
		})
	}
}

func TestDeleteLocationHandler(t *testing.T) {
	deleteLocationUser := &users.User{
		UserDB: db.UserDB{
			FirstName: "DeleteLocationHandler",
			LastName:  "DeleteLocationHandler",
			Username:  "DeleteLocationHandler",
			Email:     "DeleteLocationHandler@gmail.com",
		},
	}
	apitest.CreateUser(deleteLocationUser, env)
	sessionCookie := apitest.Login(deleteLocationUser, env)
	adminRole := roles.New("admin", 0)
	e := adminRole.AddTo(deleteLocationUser.ID)
	if e != nil {
		t.Fatalf("error adding admin role to user: %s", e.JSON())
	}

	// Create hunts to use for tests
	deleteLocationHunt := hunts.Hunt{
		HuntDB: db.HuntDB{
			Name:         "DeleteLocationHunt 1",
			MaxTeams:     43,
			StartTime:    time.Now().AddDate(0, 0, 1),
			EndTime:      time.Now().AddDate(0, 0, 2),
			LocationName: "Fake Location",
			Latitude:     34.730705,
			Longitude:    -86.59481,
		},
	}
	apitest.CreateHunt(&deleteLocationHunt, env, sessionCookie)

	team := teams.Team{
		TeamDB: db.TeamDB{
			Name:   "delete location team",
			HuntID: deleteLocationHunt.ID,
		},
	}
	apitest.CreateTeam(&team, env, sessionCookie)

	location := db.LocationDB{
		TimeStamp: time.Now().AddDate(0, 0, -1),
		Latitude:  34.730705,
		Longitude: -86.59481,
		TeamID:    team.ID,
	}
	apitest.CreateLocation(&location, env, sessionCookie)

	cases := []struct {
		name       string
		code       int
		teamID     int
		locationID int
	}{
		{
			name:       "valid team and location",
			code:       http.StatusOK,
			teamID:     team.ID,
			locationID: location.ID,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			req, err := http.NewRequest(
				"DELETE",
				config.BaseAPIURL+fmt.Sprintf("teams/%d/locations/%d", c.teamID, c.locationID),
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
					t.Errorf("error reading the response body: %v", err)
				}
				t.Fatalf("expected code %d got %d: %s", c.code, res.StatusCode, resBody)
			}
		})
	}
}

func TestGetLocationsForTeamHandler(t *testing.T) {
	team := teams.Team{
		TeamDB: db.TeamDB{
			Name:   "Get locations for team",
			HuntID: hunt.ID,
		},
	}
	apitest.CreateTeam(&team, env, sessionCookie)

	locations := []db.LocationDB{
		{
			TimeStamp: time.Now().AddDate(0, 0, -1),
			Latitude:  34.730705,
			Longitude: -86.59481,
			TeamID:    team.ID,
		},
		{
			TimeStamp: time.Now().AddDate(0, -1, 0),
			Latitude:  34.730705,
			Longitude: -86.59481,
			TeamID:    team.ID,
		},
		{
			TimeStamp: time.Now().AddDate(0, 0, -3),
			Latitude:  34.730705,
			Longitude: -86.59481,
			TeamID:    team.ID,
		},
		{
			TimeStamp: time.Now().AddDate(0, 0, -4),
			Latitude:  34.730705,
			Longitude: -86.59481,
			TeamID:    team.ID,
		},
		{
			TimeStamp: time.Now().AddDate(0, 0, -5),
			Latitude:  34.730705,
			Longitude: -86.59481,
			TeamID:    team.ID,
		},
		{
			TimeStamp: time.Now().AddDate(0, 0, -6),
			Latitude:  34.730705,
			Longitude: -86.59481,
			TeamID:    team.ID,
		},
		{
			TimeStamp: time.Now().AddDate(0, 0, -7),
			Latitude:  34.730705,
			Longitude: -86.59481,
			TeamID:    team.ID,
		},
		{
			TimeStamp: time.Now().AddDate(0, 0, -8),
			Latitude:  34.730705,
			Longitude: -86.59481,
			TeamID:    team.ID,
		},
		{
			TimeStamp: time.Now().AddDate(0, 0, -9),
			Latitude:  34.730705,
			Longitude: -86.59481,
			TeamID:    team.ID,
		},
		{
			TimeStamp: time.Now().AddDate(0, 0, -10),
			Latitude:  34.730705,
			Longitude: -86.59481,
			TeamID:    team.ID,
		},
		{
			TimeStamp: time.Now().AddDate(0, 0, -11),
			Latitude:  34.730705,
			Longitude: -86.59481,
			TeamID:    team.ID,
		},
		{
			TimeStamp: time.Now().AddDate(0, 0, -12),
			Latitude:  34.730705,
			Longitude: -86.59481,
			TeamID:    team.ID,
		},
		{
			TimeStamp: time.Now().AddDate(0, 0, -13),
			Latitude:  34.730705,
			Longitude: -86.59481,
			TeamID:    team.ID,
		},
		{
			TimeStamp: time.Now().AddDate(0, 0, -14),
			Latitude:  34.730705,
			Longitude: -86.59481,
			TeamID:    team.ID,
		},
	}
	for _, l := range locations {
		apitest.CreateLocation(&l, env, sessionCookie)
	}

	cases := []struct {
		name              string
		code              int
		teamID            int
		returnedLocations int
	}{
		{
			name:              "valid team",
			code:              http.StatusOK,
			teamID:            team.ID,
			returnedLocations: len(locations),
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			req, err := http.NewRequest(
				"GET",
				config.BaseAPIURL+fmt.Sprintf("teams/%d/locations/", c.teamID),
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
					t.Errorf("error reading the response body: %v", err)
				}
				t.Fatalf("expected code %d got %d: %s", c.code, res.StatusCode, resBody)
			}

			got := make([]db.LocationDB, 0, len(locations))
			err = json.NewDecoder(res.Body).Decode(&got)
			if err != nil {
				t.Fatalf("error decoding the response body: %v", err)
			}

			if c.returnedLocations != len(got) {
				t.Errorf(
					"expected %d locations to be returned but got %d",
					c.returnedLocations,
					len(got),
				)
			}
		})
	}
}

func TestGetAddPlayerHandler(t *testing.T) {
	team := teams.Team{
		TeamDB: db.TeamDB{
			Name:   "Add player handler",
			HuntID: hunt.ID,
		},
	}
	apitest.CreateTeam(&team, env, sessionCookie)

	team2 := teams.Team{
		TeamDB: db.TeamDB{
			Name:   "Another add player handler",
			HuntID: hunt.ID,
		},
	}
	apitest.CreateTeam(&team2, env, sessionCookie)

	diffHuntTeam := teams.Team{
		TeamDB: db.TeamDB{
			Name:   "Different hunt",
			HuntID: hunt2.ID,
		},
	}
	apitest.CreateTeam(&diffHuntTeam, env, sessionCookie)

	type addPlayerData struct {
		PlayerID int `json:"userID"`
	}

	cases := []struct {
		name   string
		code   int
		addReq addPlayerData
		teamID int
	}{
		{
			name:   "valid user and team",
			code:   http.StatusOK,
			teamID: team.ID,
			addReq: addPlayerData{
				PlayerID: newUser.ID,
			},
		},
		{
			name:   "add user to second team in same hunt",
			code:   http.StatusBadRequest,
			teamID: team2.ID,
			addReq: addPlayerData{
				PlayerID: newUser.ID,
			},
		},
		{
			name:   "add user to different hunt",
			code:   http.StatusOK,
			teamID: diffHuntTeam.ID,
			addReq: addPlayerData{
				PlayerID: newUser.ID,
			},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			reqBody, err := json.Marshal(&c.addReq)
			if err != nil {
				t.Fatalf("error marshalling req data: %v", err)
			}

			req, err := http.NewRequest(
				"POST",
				config.BaseAPIURL+fmt.Sprintf("teams/%d/players/", c.teamID),
				bytes.NewReader(reqBody),
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
					t.Errorf("error reading the response body: %v", err)
				}
				t.Fatalf(
					"expected code %d got %d: %s",
					c.code,
					res.StatusCode,
					resBody,
				)
			}
		})
	}
}

func TestGetRemovePlayerHandler(t *testing.T) {
	removePlayerUser := &users.User{
		UserDB: db.UserDB{
			FirstName: "RemovePlayerHandler",
			LastName:  "RemovePlayerHandler",
			Username:  "RemovePlayerHandler",
			Email:     "RemovePlayerHandler@gmail.com",
		},
	}
	apitest.CreateUser(removePlayerUser, env)
	sessionCookie := apitest.Login(removePlayerUser, env)
	adminRole := roles.New("admin", 0)
	e := adminRole.AddTo(removePlayerUser.ID)
	if e != nil {
		t.Fatalf("error adding admin role to user: %s", e.JSON())
	}
	removePlayerHunt := hunts.Hunt{
		HuntDB: db.HuntDB{
			Name:         "Remove players Hunt 43",
			MaxTeams:     43,
			StartTime:    time.Now().AddDate(0, 0, 1),
			EndTime:      time.Now().AddDate(0, 0, 2),
			LocationName: "Fake Location",
			Latitude:     34.730705,
			Longitude:    -86.59481,
		},
	}
	apitest.CreateHunt(&removePlayerHunt, env, sessionCookie)

	team := teams.Team{
		TeamDB: db.TeamDB{
			Name:   "Remove player handler",
			HuntID: removePlayerHunt.ID,
		},
	}
	apitest.CreateTeam(&team, env, sessionCookie)

	playerID := newUser.ID
	apitest.AddPlayer(playerID, team.ID, env, sessionCookie)

	cases := []struct {
		name     string
		code     int
		playerID int
		teamID   int
	}{
		{
			name:     "valid team and player",
			code:     http.StatusOK,
			playerID: playerID,
			teamID:   team.ID,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			req, err := http.NewRequest(
				"DELETE",
				config.BaseAPIURL+fmt.Sprintf("teams/%d/players/%d", c.teamID, c.playerID),
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
					t.Errorf("error reading the res body: %v", err)
				}
				t.Fatalf("expected code %d got %d: %s", c.code, res.StatusCode, resBody)
			}
		})
	}
}

func TestGetTeamPlayersHandler(t *testing.T) {
	getTeamPlayersHunt := hunts.Hunt{
		HuntDB: db.HuntDB{
			Name:         "GetTeamPlayers Hunt 43",
			MaxTeams:     43,
			StartTime:    time.Now().AddDate(0, 0, 1),
			EndTime:      time.Now().AddDate(0, 0, 2),
			LocationName: "Fake Location",
			Latitude:     34.730705,
			Longitude:    -86.59481,
		},
	}
	apitest.CreateHunt(&getTeamPlayersHunt, env, sessionCookie)

	team := teams.Team{
		TeamDB: db.TeamDB{
			Name:   "GetTeamPlayers handler",
			HuntID: getTeamPlayersHunt.ID,
		},
	}
	apitest.CreateTeam(&team, env, sessionCookie)

	secondUser := &users.User{
		UserDB: db.UserDB{
			FirstName: "Michael",
			LastName:  "Jordan",
			Username:  "theGreatest23",
			Email:     "mj_23@gmail.com",
		},
	}
	apitest.CreateUser(secondUser, env)
	apitest.AddPlayer(secondUser.ID, team.ID, env, sessionCookie)
	apitest.AddPlayer(newUser.ID, team.ID, env, sessionCookie)

	cases := []struct {
		name            string
		code            int
		teamID          int
		returnedPlayers int
	}{
		{
			name:            "valid team",
			code:            http.StatusOK,
			teamID:          team.ID,
			returnedPlayers: 2,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			req, err := http.NewRequest(
				"GET",
				config.BaseAPIURL+fmt.Sprintf("teams/%d/players/", c.teamID),
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
					t.Errorf("error reading res body: %v", err)
				}

				t.Fatalf(
					"expected code %d got %d: %s",
					c.code,
					res.StatusCode,
					resBody,
				)
			}

			if c.code == http.StatusOK {
				got := make([]*users.User, 0)
				err = json.NewDecoder(res.Body).Decode(&got)
				if err != nil {
					t.Fatalf("error decoding the res body: %v", err)
				}

				if len(got) != c.returnedPlayers {
					t.Fatalf(
						"expected %d players returned but got %d",
						c.returnedPlayers,
						len(got),
					)
				}
			}
		})
	}
}

func TestGetTeamPointsHandler(t *testing.T) {
	pointsHunt := hunts.Hunt{
		HuntDB: db.HuntDB{
			Name:         "points Hunt 43",
			MaxTeams:     43,
			StartTime:    time.Now().AddDate(0, 0, 1),
			EndTime:      time.Now().AddDate(0, 0, 2),
			LocationName: "Fake Location",
			Latitude:     34.730705,
			Longitude:    -86.59481,
		},
	}
	apitest.CreateHunt(&pointsHunt, env, sessionCookie)

	team := teams.Team{
		TeamDB: db.TeamDB{
			Name:   "team points team",
			HuntID: pointsHunt.ID,
		},
	}
	apitest.CreateTeam(&team, env, sessionCookie)

	items := []*models.Item{
		{
			ItemDB: db.ItemDB{
				Name:   "Santa Clause",
				Points: 10,
				HuntID: pointsHunt.ID,
			},
		},
		{
			ItemDB: db.ItemDB{
				Name:   "Rudolph",
				Points: 10,
				HuntID: pointsHunt.ID,
			},
		},
		{
			ItemDB: db.ItemDB{
				Name:   "Reindeer on the Roof",
				Points: 40,
				HuntID: pointsHunt.ID,
			},
		},
		{
			ItemDB: db.ItemDB{
				Name:   "Mrs. Clause",
				Points: 30,
				HuntID: pointsHunt.ID,
			},
		},
		{
			ItemDB: db.ItemDB{
				Name:   "Snow Globe",
				Points: 40,
				HuntID: pointsHunt.ID,
			},
		},
	}
	for _, i := range items {
		apitest.CreateItem(i, env, sessionCookie)
	}

	media := []db.MediaMetaDB{
		{
			TeamID: team.ID,
			URL:    "amazon.com/cdn/media",
			ItemID: items[0].ID,
			Location: db.LocationDB{
				TimeStamp: time.Now().AddDate(0, -1, 0),
				Latitude:  34.730705,
				Longitude: -86.59481,
				TeamID:    team.ID,
			},
		},
		{
			TeamID: team.ID,
			URL:    "amazon.com/cdn/media",
			ItemID: items[1].ID,
			Location: db.LocationDB{
				TimeStamp: time.Now().AddDate(0, 0, -3),
				Latitude:  34.730705,
				Longitude: -86.59481,
				TeamID:    team.ID,
			},
		},
		{
			TeamID: team.ID,
			URL:    "amazon.com/cdn/media",
			ItemID: items[2].ID,
			Location: db.LocationDB{
				TimeStamp: time.Now().AddDate(0, 0, -4),
				Latitude:  34.730705,
				Longitude: -86.59481,
				TeamID:    team.ID,
			},
		},
		{
			TeamID: team.ID,
			URL:    "amazon.com/cdn/media",
			ItemID: items[3].ID,
			Location: db.LocationDB{
				TimeStamp: time.Now().AddDate(0, 0, -5),
				Latitude:  34.730705,
				Longitude: -86.59481,
				TeamID:    team.ID,
			},
		},
		{
			TeamID: team.ID,
			URL:    "amazon.com/cdn/media",
			ItemID: items[4].ID,
			Location: db.LocationDB{
				TimeStamp: time.Now().AddDate(0, 0, -6),
				Latitude:  34.730705,
				Longitude: -86.59481,
				TeamID:    team.ID,
			},
		},
	}
	for _, m := range media {
		apitest.CreateMedia(&m, env, sessionCookie)
	}

	cases := []struct {
		name     string
		code     int
		teamID   int
		expected int
	}{
		{
			name:     "valid team",
			code:     http.StatusOK,
			teamID:   team.ID,
			expected: 130,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			req, err := http.NewRequest(
				"GET",
				config.BaseAPIURL+fmt.Sprintf("teams/%d/points/", c.teamID),
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
					t.Errorf("error reading res body: %v", err)
				}

				t.Fatalf(
					"expected code %d got %d: %s",
					c.code,
					res.StatusCode,
					resBody,
				)
			}

			if c.code == http.StatusOK {
				resData := struct {
					Points int `json:"points"`
				}{}
				err = json.NewDecoder(res.Body).Decode(&resData)
				if err != nil {
					t.Fatalf("error decoding res body: %v", err)
				}

				if c.expected != resData.Points {
					t.Errorf("expected %d points but got %d", c.expected, resData.Points)
				}
			}
		})
	}
}
