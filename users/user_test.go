// +build integration

package users_test

/*
import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/cljohnson4343/scavenge/sessions"

	"github.com/cljohnson4343/scavenge/users"

	c "github.com/cljohnson4343/scavenge/config"
	"github.com/cljohnson4343/scavenge/db"
	"github.com/cljohnson4343/scavenge/response"
)

var env *c.Env
var sessionCookie *http.Cookie

var newUser = db.UserDB{
	FirstName: "Lincoln",
	LastName:  "Burrows",
	Username:  "sink43",
	Email:     "linkthesink@gmail.com",
}

func TestMain(m *testing.M) {
	d := db.InitDB("../db/db_info_test.json")
	defer db.Shutdown(d)

	env = c.CreateEnv(d)
	response.SetDevMode(true)

	reqBody, err := json.Marshal(&newUser)
	if err != nil {
		panic(err)
	}

	// Login in user to get a valid user session cookie
	req, err := http.NewRequest("POST", "/login/", bytes.NewReader(reqBody))
	if err != nil {
		panic(err)
	}
	rr := httptest.NewRecorder()
	handler := users.Routes(env)
	handler.ServeHTTP(rr, req)
	res := rr.Result()
	if res.StatusCode != http.StatusOK {
		resBody, err := ioutil.ReadAll(res.Body)
		if err != nil {
			panic(err)
		}
		panic(fmt.Sprintf("error logging in: %s", resBody))
	}

	err = json.NewDecoder(res.Body).Decode(&newUser)
	if err != nil {
		panic(fmt.Sprintf("error decoding the res body: %v", err))
	}

	if newUser.ID == 0 {
		panic("expected user's id to be returned")
	}

	cookies := res.Cookies()
	for _, c := range cookies {
		if c.Name == sessions.SessionCookieName {
			sessionCookie = c
		}
	}

	if sessionCookie == nil {
		panic("expected a cookie on login")
	}

	os.Exit(m.Run())
}

func TestRequireUser(t *testing.T) {
	cases := []struct {
		name       string
		statusCode int
		withCookie bool
	}{
		{
			name:       "with valid session cookie",
			statusCode: http.StatusOK,
			withCookie: true,
		},
		{
			name:       "without session cookie",
			statusCode: http.StatusUnauthorized,
			withCookie: false,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			req, err := http.NewRequest("GET", "/", nil)
			if err != nil {
				t.Fatalf("error getting new request: %v", err)
			}

			if c.withCookie {
				req.AddCookie(sessionCookie)
			}

			rr := httptest.NewRecorder()
			hasContext := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				ctx := r.Context()

				userID, e := users.GetUserID(ctx)
				if e != nil {
					e.Handle(w)
					return
				}

				if userID != newUser.ID {
					t.Fatal("expected context's user id to be the same as the logged in user's")
				}
			})

			handler := users.WithUser(hasContext)
			handler.ServeHTTP(rr, req)

			res := rr.Result()
			resBody, err := ioutil.ReadAll(res.Body)
			if err != nil {
				t.Fatalf("error reading response body: %v", err)
			}

			if res.StatusCode != c.statusCode {
				t.Fatalf("expected code %d got %d: %s", c.statusCode, res.StatusCode, resBody)
			}
		})
	}
}
*/
