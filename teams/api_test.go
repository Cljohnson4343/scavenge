// +build apiTest

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
	c "github.com/cljohnson4343/scavenge/config"
	"github.com/cljohnson4343/scavenge/db"
	"github.com/cljohnson4343/scavenge/hunts"
	"github.com/cljohnson4343/scavenge/hunts/models"
	"github.com/cljohnson4343/scavenge/response"
	"github.com/cljohnson4343/scavenge/teams"
	"github.com/cljohnson4343/scavenge/users"
)

var env *c.Env
var hunt hunts.Hunt
var hunt2 hunts.Hunt
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
			req, err := http.NewRequest("POST", "/", bytes.NewReader(reqBody))
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
			}
		})
	}
}

func TestGetTeamsHandler(t *testing.T) {
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
						HuntID: hunt.ID,
					},
				},
				teams.Team{
					TeamDB: db.TeamDB{
						Name:   "second get teams",
						HuntID: hunt.ID,
					},
				},
				teams.Team{
					TeamDB: db.TeamDB{
						Name:   "third get teams",
						HuntID: hunt.ID,
					},
				},
				teams.Team{
					TeamDB: db.TeamDB{
						Name:   "fourth get teams",
						HuntID: hunt.ID,
					},
				},
				teams.Team{
					TeamDB: db.TeamDB{
						Name:   "fifth get teams",
						HuntID: hunt.ID,
					},
				},
				teams.Team{
					TeamDB: db.TeamDB{
						Name:   "sixth get teams",
						HuntID: hunt.ID,
					},
				},
				teams.Team{
					TeamDB: db.TeamDB{
						Name:   "seventh get teams",
						HuntID: hunt.ID,
					},
				},
				teams.Team{
					TeamDB: db.TeamDB{
						Name:   "eighth get teams",
						HuntID: hunt.ID,
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

			req, err := http.NewRequest("GET", "/", nil)
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
	expected := teams.Team{
		TeamDB: db.TeamDB{
			Name:   "Get this team",
			HuntID: hunt.ID,
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
		{
			name:   "invalid team",
			code:   http.StatusBadRequest,
			teamID: 0,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			req, err := http.NewRequest("GET", fmt.Sprintf("/%d", c.teamID), nil)
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
	team := teams.Team{
		TeamDB: db.TeamDB{
			Name:   "fox river eight",
			HuntID: hunt.ID,
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
		{
			name: "invalid team",
			code: http.StatusBadRequest,
			id:   0,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			req, err := http.NewRequest("DELETE", fmt.Sprintf("/%d", c.id), nil)
			if err != nil {
				t.Fatalf("error creating new request: %v", err)
			}

			rr := httptest.NewRecorder()
			handler := teams.Routes(env)
			handler.ServeHTTP(rr, req)

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
		{
			name:   "update invalid team",
			id:     0,
			code:   http.StatusBadRequest,
			update: teams.Team{},
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
			name: "invalid media team id",
			code: http.StatusBadRequest,
			media: db.MediaMetaDB{
				TeamID: 0,
				ItemID: item.ID,
				URL:    "amazon.com/cdn/media",
				Location: db.LocationDB{
					TimeStamp: time.Now().AddDate(0, 0, -3),
					Latitude:  34.730705,
					Longitude: -86.59481,
					TeamID:    team.ID,
				},
			},
		},
		{
			name: "invalid team id for location",
			code: http.StatusBadRequest,
			media: db.MediaMetaDB{
				TeamID: team.ID,
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
				fmt.Sprintf("/%d/media/", team.ID),
				bytes.NewReader(reqBody),
			)
			if err != nil {
				t.Fatalf("error getting new request: %v", err)
			}
			req.AddCookie(sessionCookie)

			rr := httptest.NewRecorder()
			handler := teams.Routes(env)
			handler.ServeHTTP(rr, req)

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
		{
			name:          "invalid team",
			code:          http.StatusOK,
			teamID:        0,
			returnedMedia: 0,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			req, err := http.NewRequest(
				"GET",
				fmt.Sprintf("/%d/media/", c.teamID),
				nil,
			)
			if err != nil {
				t.Fatalf("error getting new request: %v", err)
			}
			req.AddCookie(sessionCookie)

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
			name:    "invalid team and valid media",
			code:    http.StatusBadRequest,
			mediaID: media.ID,
			teamID:  4343,
		},
		{
			name:    "valid team and invalid media",
			code:    http.StatusBadRequest,
			mediaID: 434343,
			teamID:  team.ID,
		},
		{
			name:    "invalid team and invalid media",
			code:    http.StatusBadRequest,
			mediaID: 434343,
			teamID:  43434343,
		},
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
				fmt.Sprintf("/%d/media/%d", c.teamID, c.mediaID),
				nil,
			)
			if err != nil {
				t.Fatalf("error getting new request: %v", err)
			}
			req.AddCookie(sessionCookie)

			rr := httptest.NewRecorder()
			handler := teams.Routes(env)
			handler.ServeHTTP(rr, req)
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
			name: "without team id",
			code: http.StatusBadRequest,
			location: db.LocationDB{
				TimeStamp: time.Now().AddDate(0, 0, -1),
				Latitude:  34.730705,
				Longitude: -86.59481,
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
				fmt.Sprintf("/%d/locations/", c.location.TeamID),
				bytes.NewReader(reqBody),
			)
			if err != nil {
				t.Fatalf("error getting new request: %v", err)
			}
			req.AddCookie(sessionCookie)

			rr := httptest.NewRecorder()
			handler := teams.Routes(env)
			handler.ServeHTTP(rr, req)
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
	team := teams.Team{
		TeamDB: db.TeamDB{
			Name:   "delete location team",
			HuntID: hunt.ID,
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
			name:       "invalid team and valid location",
			code:       http.StatusBadRequest,
			teamID:     4343,
			locationID: location.ID,
		},
		{
			name:       "invalid team and location",
			code:       http.StatusBadRequest,
			teamID:     434343,
			locationID: 4343,
		},
		{
			name:       "valid team and invalid location",
			code:       http.StatusBadRequest,
			teamID:     team.ID,
			locationID: 4343,
		},
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
				fmt.Sprintf("/%d/locations/%d", c.teamID, c.locationID),
				nil,
			)
			if err != nil {
				t.Fatalf("error getting new request: %v", err)
			}
			req.AddCookie(sessionCookie)
			rr := httptest.NewRecorder()
			handler := teams.Routes(env)
			handler.ServeHTTP(rr, req)
			res := rr.Result()

			if res.StatusCode != c.code {
				resBody, err := ioutil.ReadAll(res.Body)
				if err != nil {
					t.Errorf("error reading the response body: %v", err)
				}
				t.Fatalf("expected code %d got %d: %v", c.code, res.StatusCode, resBody)
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
		{
			name:              "invalid team",
			code:              http.StatusOK,
			teamID:            434343,
			returnedLocations: 0,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			req, err := http.NewRequest(
				"GET",
				fmt.Sprintf("/%d/locations/", c.teamID),
				nil,
			)
			if err != nil {
				t.Fatalf("error getting new request: %v", err)
			}
			req.AddCookie(sessionCookie)

			rr := httptest.NewRecorder()
			handler := teams.Routes(env)
			handler.ServeHTTP(rr, req)
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
		PlayerID int `json:"id"`
	}

	cases := []struct {
		name   string
		code   int
		addReq addPlayerData
		teamID int
	}{
		{
			name:   "invalid user and team",
			code:   http.StatusBadRequest,
			teamID: 0,
			addReq: addPlayerData{
				PlayerID: 0,
			},
		},
		{
			name:   "invalid user and valid team",
			code:   http.StatusBadRequest,
			teamID: team.ID,
			addReq: addPlayerData{
				PlayerID: 0,
			},
		},
		{
			name:   "valid user and invalid team",
			code:   http.StatusBadRequest,
			teamID: 0,
			addReq: addPlayerData{
				PlayerID: newUser.ID,
			},
		},
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
				fmt.Sprintf("/%d/players/", c.teamID),
				bytes.NewReader(reqBody),
			)
			if err != nil {
				t.Fatalf("error getting new request: %v", err)
			}
			req.AddCookie(sessionCookie)

			rr := httptest.NewRecorder()
			handler := teams.Routes(env)
			handler.ServeHTTP(rr, req)
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
			name:     "invalid team and player",
			code:     http.StatusBadRequest,
			playerID: 0,
			teamID:   0,
		},
		{
			name:     "invalid team and valid player",
			code:     http.StatusBadRequest,
			playerID: playerID,
			teamID:   0,
		},
		{
			name:     "valid team and invalid player",
			code:     http.StatusBadRequest,
			playerID: 0,
			teamID:   team.ID,
		},
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
				fmt.Sprintf("/%d/players/%d", c.teamID, c.playerID),
				nil,
			)
			if err != nil {
				t.Fatalf("error getting new request: %v", err)
			}
			req.AddCookie(sessionCookie)

			rr := httptest.NewRecorder()
			handler := teams.Routes(env)
			handler.ServeHTTP(rr, req)
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
		db.UserDB{
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
		{
			name:            "invalid team",
			code:            http.StatusOK,
			teamID:          0,
			returnedPlayers: 0,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			req, err := http.NewRequest(
				"GET",
				fmt.Sprintf("/%d/players/", c.teamID),
				nil,
			)
			if err != nil {
				t.Fatalf("error getting new request: %v", err)
			}
			req.AddCookie(sessionCookie)

			rr := httptest.NewRecorder()
			handler := teams.Routes(env)
			handler.ServeHTTP(rr, req)
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
		{
			name:     "invalid team",
			code:     http.StatusOK,
			teamID:   0,
			expected: 0,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			req, err := http.NewRequest(
				"GET",
				fmt.Sprintf("/%d/points/", c.teamID),
				nil,
			)
			if err != nil {
				t.Fatalf("error getting new request: %v", err)
			}
			req.AddCookie(sessionCookie)
			rr := httptest.NewRecorder()
			handler := teams.Routes(env)
			handler.ServeHTTP(rr, req)
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
