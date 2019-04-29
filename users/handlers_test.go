// +build apiTest

package users_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
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

type newUserReq struct {
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
		reqData    newUserReq
		statusCode int
	}{
		{
			testName: `new user`,
			reqData: newUserReq{
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
			reqData: newUserReq{
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
			reqData: newUserReq{
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
			reqData: newUserReq{
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
			reqData: newUserReq{
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
			reqData: newUserReq{
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
			reqData: newUserReq{
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

			res := serveAndReturnResponse(users.GetLoginHandler(env), req)

			resBody := getBody(t, res)

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

				cookie := getSessionCookie(cookies)
				if cookie == nil {
					t.Errorf("expected a cookie")
				}
			}
		})
	}
}

func TestLogoutHandler(t *testing.T) {
	cases := []struct {
		name       string
		statusCode int
		withCookie bool
	}{
		{
			name:       "logged in user",
			statusCode: http.StatusOK,
			withCookie: true,
		},
		{
			name:       "logged out user",
			statusCode: http.StatusBadRequest,
			withCookie: false,
		},
	}

	userInfo := newUserReq{
		FirstName: "tj",
		LastName:  "rrrrson",
		Email:     "rrrrr43@gmail.com",
		Username:  "tj43",
		ImageURL:  "amazon.cdn.com",
	}

	// login to get a valid cookie
	reqBody, err := json.Marshal(&userInfo)
	if err != nil {
		t.Errorf("error marshaling login request data: %v", err)
	}

	req, err := http.NewRequest("POST", "/", bytes.NewReader(reqBody))
	if err != nil {
		t.Errorf("error getting a new request: %v", err)
	}

	rr := httptest.NewRecorder()
	login := users.GetLoginHandler(env)
	login.ServeHTTP(rr, req)
	res := rr.Result()
	resBody := getBody(t, res)

	if code := res.StatusCode; code != http.StatusOK {
		t.Errorf("expected login status code %d got %d: %s", http.StatusOK, code, resBody)
	}

	cookie := getSessionCookie(res.Cookies())
	if cookie == nil {
		t.Error("expicted a cookie")
		t.FailNow()
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			req, err := http.NewRequest("POST", "/", nil)
			if err != nil {
				t.Errorf("error getting a logout request: %v", err)
			}

			if c.withCookie {
				req.AddCookie(cookie)
			}

			res := serveAndReturnResponse(users.GetLogoutHandler(env), req)

			resBody, err := ioutil.ReadAll(res.Body)
			if err != nil {
				t.Errorf("error reading response body: %v", err)
			}

			if c.statusCode != res.StatusCode {
				t.Errorf("expected a status code %d got %d: %s", c.statusCode, res.StatusCode, resBody)
			}

			if c.statusCode == http.StatusOK {
				if len(res.Cookies()) != 0 {
					t.Error("expected all cookies to be deleted upon logout")
				}
			}
		})
	}
}

func TestCreateUserHandler(t *testing.T) {
	cases := []struct {
		name       string
		reqData    newUserReq
		statusCode int
	}{
		{
			name: `new user`,
			reqData: newUserReq{
				FirstName: "Create",
				LastName:  "User",
				Username:  "create_user_43",
				Email:     "create433@gmail.com",
				ImageURL:  "amazon.cdn.com",
			},
			statusCode: http.StatusOK,
		},
		{
			name: `provide a user id`,
			reqData: newUserReq{
				FirstName: "cj1",
				LastName:  "johnson1",
				Username:  "cj431",
				Email:     "cj43@gmail.com",
				ImageURL:  "amazon.cdn.com",
				ID:        1,
			},
			statusCode: http.StatusBadRequest,
		},
		{
			name: `duplicate username`,
			reqData: newUserReq{
				FirstName: "rj",
				LastName:  "mohnson",
				Username:  "create_user_43",
				Email:     "rj43@gmail.com",
				ImageURL:  "amazon.cdn.com",
			},
			statusCode: http.StatusBadRequest,
		},
		{
			name: `duplicate email`,
			reqData: newUserReq{
				FirstName: "rj",
				LastName:  "mohnson",
				Username:  "rj43",
				Email:     "create433@gmail.com",
				ImageURL:  "amazon.cdn.com",
			},
			statusCode: http.StatusBadRequest,
		},
		{
			name: `request missing first name`,
			reqData: newUserReq{
				LastName: "johnson",
				Username: "cj43",
				Email:    "cj43@gmail.com",
				ImageURL: "amazon.cdn.com",
			},
			statusCode: http.StatusBadRequest,
		},
		{
			name: `request missing last name`,
			reqData: newUserReq{
				FirstName: "cj",
				Username:  "cj43",
				Email:     "cj43@gmail.com",
				ImageURL:  "amazon.cdn.com",
			},
			statusCode: http.StatusBadRequest,
		},
		{
			name: `request missing username`,
			reqData: newUserReq{
				FirstName: "cj",
				LastName:  "johnson",
				Email:     "cj43@gmail.com",
				ImageURL:  "amazon.cdn.com",
			},
			statusCode: http.StatusBadRequest,
		},
		{
			name: `request missing email`,
			reqData: newUserReq{
				FirstName: "cj",
				LastName:  "johnson",
				Username:  "cj43",
				ImageURL:  "amazon.cdn.com",
			},
			statusCode: http.StatusBadRequest,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			bodyBuf, err := json.Marshal(&c.reqData)
			if err != nil {
				t.Errorf("error marshalling user data: %v", err)
				t.FailNow()
			}

			req, err := http.NewRequest("POST", "/", bytes.NewReader(bodyBuf))
			if err != nil {
				t.Errorf("error getting a new request: %v", err)
			}

			res := serveAndReturnResponse(users.GetCreateUserHandler(env), req)
			resBody := getBody(t, res)

			if c.statusCode != res.StatusCode {
				t.Errorf("expected code %d got %d: %s", c.statusCode, res.StatusCode, resBody)
			}

			if c.statusCode == http.StatusOK {
				nu := users.User{}
				err := json.NewDecoder(strings.NewReader(resBody)).Decode(&nu)
				if err != nil {
					t.Errorf("error decoding response: %v", err)
				}

				if nu.ID == 0 {
					t.Error("expected new user ID to be returned")
				}

				if nu.LastVisit.IsZero() {
					t.Error("expected new user LastVisit to be returned")
				}

				if nu.JoinedAt.IsZero() {
					t.Error("expected new user JoinedAt to be returned")
				}

				if nu.ImageURL != c.reqData.ImageURL {
					t.Error("expected new user ImageURL to be the same")
				}

				if nu.Email != c.reqData.Email {
					t.Error("expected new user Email to be the same")
				}

				if nu.FirstName != c.reqData.FirstName {
					t.Error("expected new user FirstName to be the same")
				}

				if nu.Username != c.reqData.Username {
					t.Error("expected new user Username to be the same")
				}

				if nu.LastName != c.reqData.LastName {
					t.Error("expected new user LastName to be the same")
				}
			}
		})
	}
}

func TestDeleteUserHandler(t *testing.T) {
	cases := []struct {
		name        string
		newUserData newUserReq
		withNewUser bool
		statusCode  int
	}{
		{
			name: `delete existing user`,
			newUserData: newUserReq{
				FirstName: "Delete",
				LastName:  "User II",
				Username:  "delete_user_43",
				Email:     "delete433@gmail.com",
				ImageURL:  "amazon.cdn.com",
			},
			statusCode:  http.StatusOK,
			withNewUser: true,
		},
		{
			name:        `non-existing user`,
			newUserData: newUserReq{},
			statusCode:  http.StatusBadRequest,
			withNewUser: false,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			userID := 0

			if c.withNewUser {
				bodyBuf, err := json.Marshal(&c.newUserData)
				if err != nil {
					t.Errorf("error marshalling user data: %v", err)
					t.FailNow()
				}

				req, err := http.NewRequest("POST", "/", bytes.NewReader(bodyBuf))
				if err != nil {
					t.Errorf("error getting a new request: %v", err)
					t.FailNow()
				}

				res := serveAndReturnResponse(users.GetCreateUserHandler(env), req)
				resBody := getBody(t, res)

				if res.StatusCode != http.StatusOK {
					t.Errorf("error creating user: %s", resBody)
					t.FailNow()
				}

				resStruct := struct {
					ID int `json:"id"`
				}{}

				err = json.Unmarshal([]byte(resBody), &resStruct)
				if err != nil {
					t.Errorf("error unmarshalling response string: %v", err)
				}

				userID = resStruct.ID
			}

			req, err := http.NewRequest("DELETE", fmt.Sprintf("/%d", userID), nil)
			if err != nil {
				t.Errorf("error getting new request: %v", err)
				t.FailNow()
			}
			res := serveAndReturnResponse(users.Routes(env), req)
			resBody := getBody(t, res)

			if res.StatusCode != c.statusCode {
				t.Errorf("expected status code %d got %d: %s", c.statusCode, res.StatusCode, resBody)
			}
		})
	}
}

func TestSelectUserHandler(t *testing.T) {
	cases := []struct {
		name        string
		newUserData newUserReq
		withNewUser bool
		statusCode  int
	}{
		{
			name: `select existing user`,
			newUserData: newUserReq{
				FirstName: "select",
				LastName:  "user III",
				Username:  "select_user_43",
				Email:     "select433@gmail.com",
				ImageURL:  "amazon.cdn.com",
			},
			statusCode:  http.StatusOK,
			withNewUser: true,
		},
		{
			name:        `select non-existing user`,
			newUserData: newUserReq{},
			statusCode:  http.StatusBadRequest,
			withNewUser: false,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			userID := 0

			if c.withNewUser {
				bodyBuf, err := json.Marshal(&c.newUserData)
				if err != nil {
					t.Errorf("error marshalling user data: %v", err)
					t.FailNow()
				}

				req, err := http.NewRequest("POST", "/", bytes.NewReader(bodyBuf))
				if err != nil {
					t.Errorf("error getting a new request: %v", err)
					t.FailNow()
				}

				res := serveAndReturnResponse(users.Routes(env), req)
				resBody := getBody(t, res)

				if res.StatusCode != http.StatusOK {
					t.Errorf("error creating user: %s", resBody)
					t.FailNow()
				}

				resStruct := struct {
					ID int `json:"id"`
				}{}

				err = json.Unmarshal([]byte(resBody), &resStruct)
				if err != nil {
					t.Errorf("error unmarshalling response string: %v", err)
				}

				userID = resStruct.ID
			}

			req, err := http.NewRequest("GET", fmt.Sprintf("/%d", userID), nil)
			if err != nil {
				t.Errorf("error getting new request: %v", err)
				t.FailNow()
			}
			res := serveAndReturnResponse(users.Routes(env), req)
			resBody := getBody(t, res)

			if res.StatusCode != c.statusCode {
				t.Errorf("expected status code %d got %d: %s", c.statusCode, res.StatusCode, resBody)
			}

			if c.statusCode == http.StatusOK {
				nu := users.User{}
				err := json.NewDecoder(strings.NewReader(resBody)).Decode(&nu)
				if err != nil {
					t.Errorf("error decoding response: %v", err)
				}

				if nu.ID == 0 {
					t.Error("expected user ID to be returned")
				}

				if nu.LastVisit.IsZero() {
					t.Error("expected user LastVisit to be returned")
				}

				if nu.JoinedAt.IsZero() {
					t.Error("expected user JoinedAt to be returned")
				}

				if nu.ImageURL != c.newUserData.ImageURL {
					t.Error("expected user ImageURL to be the same")
				}

				if nu.Email != c.newUserData.Email {
					t.Error("expected user Email to be the same")
				}

				if nu.FirstName != c.newUserData.FirstName {
					t.Error("expected user FirstName to be the same")
				}

				if nu.Username != c.newUserData.Username {
					t.Error("expected user Username to be the same")
				}

				if nu.LastName != c.newUserData.LastName {
					t.Error("expected user LastName to be the same")
				}
			}

		})
	}
}
func getBody(t *testing.T, res *http.Response) string {
	bodyBuf, err := ioutil.ReadAll(res.Body)
	if err != nil {
		t.Errorf("error getting response body: %v", err)
	}

	return string(bodyBuf)
}

func getSessionCookie(cookies []*http.Cookie) *http.Cookie {
	// Make sure new cookie's name was set correctly
	var cookie *http.Cookie
	for i := range cookies {
		if cookies[i].Name == sessions.SessionCookieName {
			cookie = cookies[i]
		}
	}

	return cookie
}

func serveAndReturnResponse(fn http.Handler, req *http.Request) *http.Response {
	rr := httptest.NewRecorder()
	fn.ServeHTTP(rr, req)
	return rr.Result()
}