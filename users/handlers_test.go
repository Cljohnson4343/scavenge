// +build api_test

package users_test

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	c "github.com/cljohnson4343/scavenge/config"
	"github.com/cljohnson4343/scavenge/db"
	"github.com/cljohnson4343/scavenge/response"
	"github.com/cljohnson4343/scavenge/sessions"
	"github.com/cljohnson4343/scavenge/users"
)

var env *c.Env

func TestMain(m *testing.M) {
	d := db.InitDB("../db/db_info_test.json")
	defer db.Shutdown(d)

	env = c.CreateEnv(d)

	response.SetDevMode(true)

	os.Exit(m.Run())
}

type loginRequest struct {
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Username  string `json:"username"`
	Email     string `json:"email"`
	ID        int    `json:"id"`
	ImageURL  string `json:"image_url"`
}

func TestLoginHandler(t *testing.T) {

	tables := []struct {
		testName   string
		reqData    loginRequest
		statusCode int
	}{
		{
			testName: `new user`,
			reqData: loginRequest{
				FirstName: "cj",
				LastName:  "johnson",
				Username:  "cj43",
				Email:     "cj43@gmail.com",
				ImageURL:  "amazon.cdn.com",
			},
			statusCode: http.StatusOK,
		},
		{
			testName: `existing user`,
			reqData: loginRequest{
				FirstName: "cj",
				LastName:  "johnson",
				Username:  "cj43",
				Email:     "cj43@gmail.com",
				ImageURL:  "amazon.cdn.com",
				ID:        1,
			},
			statusCode: http.StatusOK,
		},
		{
			testName: `existing user without providing user_id`,
			reqData: loginRequest{
				FirstName: "cj",
				LastName:  "johnson",
				Username:  "cj43",
				Email:     "cj43@gmail.com",
				ImageURL:  "amazon.cdn.com",
			},
			statusCode: http.StatusBadRequest,
		},
		{
			testName: `request missing first name`,
			reqData: loginRequest{
				LastName: "johnson",
				Username: "cj43",
				Email:    "cj43@gmail.com",
				ImageURL: "amazon.cdn.com",
				ID:       1,
			},
			statusCode: http.StatusBadRequest,
		},
		{
			testName: `request missing last name`,
			reqData: loginRequest{
				FirstName: "cj",
				Username:  "cj43",
				Email:     "cj43@gmail.com",
				ImageURL:  "amazon.cdn.com",
				ID:        1,
			},
			statusCode: http.StatusBadRequest,
		},
		{
			testName: `request missing username`,
			reqData: loginRequest{
				FirstName: "cj",
				LastName:  "johnson",
				Email:     "cj43@gmail.com",
				ImageURL:  "amazon.cdn.com",
				ID:        1,
			},
			statusCode: http.StatusBadRequest,
		},
		{
			testName: `request missing email`,
			reqData: loginRequest{
				FirstName: "cj",
				LastName:  "johnson",
				Username:  "cj43",
				ImageURL:  "amazon.cdn.com",
				ID:        1,
			},
			statusCode: http.StatusBadRequest,
		},
	}

	for _, table := range tables {
		t.Run(table.testName, func(t *testing.T) {
			bodyJSON, err := json.Marshal(&table.reqData)
			if err != nil {
				t.Errorf("failed to marshal req body: %v", err)
			}

			bodyReader := bytes.NewReader(bodyJSON)
			req, err := http.NewRequest("POST", "/", bodyReader)
			if err != nil {
				t.Errorf("failed to create request: %v", err)
			}

			rr := httptest.NewRecorder()
			loginHandler := users.GetLoginHandler(env)
			loginHandler.ServeHTTP(rr, req)

			res := rr.Result()

			resBody, err := ioutil.ReadAll(res.Body)
			if err != nil {
				t.Errorf("error reading response body: %v", err)
			}

			// Make sure login was successful
			if res.StatusCode != table.statusCode {
				t.Errorf("expected a return code %d but got %d: %s",
					table.statusCode,
					res.StatusCode,
					resBody,
				)
			}

			if table.statusCode == http.StatusOK {
				// Make sure a new session cookie was included in the response
				cookies := res.Cookies()
				if len(cookies) != 1 {
					t.Error("failed to return a cookie")
				}

				// Make sure new cookie's name was set correctly
				var cookie *http.Cookie
				for i := range cookies {
					if cookies[i].Name == sessions.SessionCookieName {
						cookie = cookies[i]
					}
				}
				if cookie == nil {
					t.Error("failed to find the cookie for new session")
				}
			}
		})
	}
}

func TestLogoutHandler(t *testing.T) {
	cases := []struct {
		name       string
		userData   loginRequest
		statusCode int
	}{
		{
			name: "logged in user",
			userData: loginRequest{
				FirstName: "tj",
				LastName:  "rrrrson",
				Email:     "rrrrr43@gmail.com",
				Username:  "tj43",
				ImageURL:  "amazon.cdn.com",
			},
			statusCode: http.StatusOK,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			// login to get a valid cookie
			loginReqBody, err := json.Marshal(&c.userData)
			if err != nil {
				t.Errorf("error marshaling login request data: %v", err)
			}

			loginReq, err := http.NewRequest("POST", "/", bytes.NewReader(loginReqBody))
			if err != nil {
				t.Errorf("error getting a new request: %v", err)
			}

			rr := httptest.NewRecorder()
			login := users.GetLoginHandler(env)
			login.ServeHTTP(rr, loginReq)
			loginRes := rr.Result()
			if code := loginRes.StatusCode; code != http.StatusOK {
				loginResBody, err := ioutil.ReadAll(loginRes.Body)
				if err != nil {
					loginResBody = []byte("")
				}

				t.Errorf("expected login status code %d got %d: %s", http.StatusOK, code, loginResBody)
			}

			cookies := loginRes.Cookies()
			var cookie *http.Cookie
			for i := range cookies {
				if cookies[i].Name == sessions.SessionCookieName {
					cookie = cookies[i]
				}
			}

			if cookie == nil {
				t.Error("expicted a cookie")
				t.FailNow()
			}

			logoutReq, err := http.NewRequest("POST", "/", nil)
			if err != nil {
				t.Errorf("error getting a logout request: %v", err)
			}

			logoutReq.AddCookie(cookie)
			rr = httptest.NewRecorder()
			logout := users.GetLogoutHandler(env)
			logout.ServeHTTP(rr, logoutReq)

			res := rr.Result()
			resBody, err := ioutil.ReadAll(res.Body)
			if err != nil {
				t.Errorf("error reading response body: %v", err)
			}

			if c.statusCode != res.StatusCode {
				t.Errorf("expected a status code %d got %d: %s", c.statusCode, res.StatusCode, resBody)
			}

			if len(res.Cookies()) != 0 {
				t.Error("expected all cookies to be deleted upon logout")
			}
		})
	}
}
