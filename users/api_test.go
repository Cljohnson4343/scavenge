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
	"time"

	"github.com/cljohnson4343/scavenge/apitest"
	c "github.com/cljohnson4343/scavenge/config"
	"github.com/cljohnson4343/scavenge/db"
	"github.com/cljohnson4343/scavenge/response"
	"github.com/cljohnson4343/scavenge/sessions"
	"github.com/cljohnson4343/scavenge/users"
)

var env *c.Env
var sessionCookie *http.Cookie
var newUser = users.User{
	UserDB: db.UserDB{
		FirstName: "Lincoln",
		LastName:  "Burrows",
		Username:  "sink43",
		Email:     "linkthesink@gmail.com",
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

func TestLoginHandler(t *testing.T) {
	tables := []struct {
		testName   string
		user       users.User
		statusCode int
	}{
		{
			testName: `new user`,
			user: users.User{
				UserDB: db.UserDB{
					FirstName: "cj",
					LastName:  "johnson",
					Username:  "cj43",
					Email:     "cj43@gmail.com",
					ImageURL:  "amazon.cdn.com",
				},
			},
			statusCode: http.StatusOK,
		},
		{
			testName: `existing user`,
			user: users.User{
				UserDB: db.UserDB{
					FirstName: "cj",
					LastName:  "johnson",
					Username:  "cj43",
					Email:     "cj43@gmail.com",
					ImageURL:  "amazon.cdn.com",
					ID:        1,
				},
			},
			statusCode: http.StatusOK,
		},
		{
			testName: `existing user without providing user_id`,
			user: users.User{
				UserDB: db.UserDB{
					FirstName: "cj",
					LastName:  "johnson",
					Username:  "cj43",
					Email:     "cj43@gmail.com",
					ImageURL:  "amazon.cdn.com",
				},
			},
			statusCode: http.StatusBadRequest,
		},
		{
			testName: `request missing first name`,
			user: users.User{
				UserDB: db.UserDB{
					LastName: "johnson",
					Username: "cj43",
					Email:    "cj43@gmail.com",
					ImageURL: "amazon.cdn.com",
					ID:       1,
				},
			},
			statusCode: http.StatusBadRequest,
		},
		{
			testName: `request missing last name`,
			user: users.User{
				UserDB: db.UserDB{
					FirstName: "cj",
					Username:  "cj43",
					Email:     "cj43@gmail.com",
					ImageURL:  "amazon.cdn.com",
					ID:        1,
				},
			},
			statusCode: http.StatusBadRequest,
		},
		{
			testName: `request missing username`,
			user: users.User{
				UserDB: db.UserDB{
					FirstName: "cj",
					LastName:  "johnson",
					Email:     "cj43@gmail.com",
					ImageURL:  "amazon.cdn.com",
					ID:        1,
				},
			},
			statusCode: http.StatusBadRequest,
		},
		{
			testName: `request missing email`,
			user: users.User{
				UserDB: db.UserDB{
					FirstName: "cj",
					LastName:  "johnson",
					Username:  "cj43",
					ImageURL:  "amazon.cdn.com",
					ID:        1,
				},
			},
			statusCode: http.StatusBadRequest,
		},
	}

	for _, table := range tables {
		t.Run(table.testName, func(t *testing.T) {
			bodyJSON, err := json.Marshal(&table.user)
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

	// login to get a valid cookie
	user := users.User{
		UserDB: db.UserDB{
			FirstName: "tj",
			LastName:  "rrrrson",
			Email:     "rrrrr43@gmail.com",
			Username:  "tj43",
			ImageURL:  "amazon.cdn.com",
		},
	}
	apitest.CreateUser(&user, env)
	cookie := apitest.Login(&user, env)

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			req, err := http.NewRequest("POST", "/logout/", nil)
			if err != nil {
				t.Errorf("error getting a logout request: %v", err)
			}

			if c.withCookie {
				req.AddCookie(cookie)
			}

			res := serveAndReturnResponse(users.Routes(env), req)
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
		user       users.User
		statusCode int
	}{
		{
			name: `new user`,
			user: users.User{
				UserDB: db.UserDB{
					FirstName: "Create",
					LastName:  "User",
					Username:  "create_user_43",
					Email:     "create433@gmail.com",
					ImageURL:  "amazon.cdn.com",
				},
			},
			statusCode: http.StatusOK,
		},
		{
			name: `provide a user id`,
			user: users.User{
				UserDB: db.UserDB{
					FirstName: "cj1",
					LastName:  "johnson1",
					Username:  "cj431",
					Email:     "cj43@gmail.com",
					ImageURL:  "amazon.cdn.com",
					ID:        1,
				},
			},
			statusCode: http.StatusBadRequest,
		},
		{
			name: `duplicate username`,
			user: users.User{
				UserDB: db.UserDB{
					FirstName: "rj",
					LastName:  "mohnson",
					Username:  "create_user_43",
					Email:     "rj43@gmail.com",
					ImageURL:  "amazon.cdn.com",
				},
			},
			statusCode: http.StatusBadRequest,
		},
		{
			name: `duplicate email`,
			user: users.User{
				UserDB: db.UserDB{
					FirstName: "rj",
					LastName:  "mohnson",
					Username:  "rj43",
					Email:     "create433@gmail.com",
					ImageURL:  "amazon.cdn.com",
				},
			},
			statusCode: http.StatusBadRequest,
		},
		{
			name: `request missing first name`,
			user: users.User{
				UserDB: db.UserDB{
					LastName: "johnson",
					Username: "cj43",
					Email:    "cj43@gmail.com",
					ImageURL: "amazon.cdn.com",
				},
			},
			statusCode: http.StatusBadRequest,
		},
		{
			name: `request missing last name`,
			user: users.User{
				UserDB: db.UserDB{
					FirstName: "cj",
					Username:  "cj43",
					Email:     "cj43@gmail.com",
					ImageURL:  "amazon.cdn.com",
				},
			},
			statusCode: http.StatusBadRequest,
		},
		{
			name: `request missing username`,
			user: users.User{
				UserDB: db.UserDB{
					FirstName: "cj",
					LastName:  "johnson",
					Email:     "cj43@gmail.com",
					ImageURL:  "amazon.cdn.com",
				},
			},
			statusCode: http.StatusBadRequest,
		},
		{
			name: `request missing email`,
			user: users.User{
				UserDB: db.UserDB{
					FirstName: "cj",
					LastName:  "johnson",
					Username:  "cj43",
					ImageURL:  "amazon.cdn.com",
				},
			},
			statusCode: http.StatusBadRequest,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			bodyBuf, err := json.Marshal(&c.user)
			if err != nil {
				t.Errorf("error marshalling user data: %v", err)
				t.FailNow()
			}

			req, err := http.NewRequest("POST", "/", bytes.NewReader(bodyBuf))
			if err != nil {
				t.Errorf("error getting a new request: %v", err)
			}

			res := serveAndReturnResponse(users.Routes(env), req)
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

				compareSharedFields(t, &nu, &c.user)
			}
		})
	}
}

func TestDeleteUserHandler(t *testing.T) {
	cases := []struct {
		name        string
		user        users.User
		withNewUser bool
		statusCode  int
	}{
		{
			name: `delete existing user`,
			user: users.User{
				UserDB: db.UserDB{
					FirstName: "Delete",
					LastName:  "User II",
					Username:  "delete_user_43",
					Email:     "delete433@gmail.com",
					ImageURL:  "amazon.cdn.com",
				},
			},
			statusCode:  http.StatusOK,
			withNewUser: true,
		},
		{
			name:        `non-existing user`,
			user:        users.User{},
			statusCode:  http.StatusBadRequest,
			withNewUser: false,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			userID := 0

			if c.withNewUser {
				apitest.CreateUser(&c.user, env)
				userID = c.user.ID
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
		user        users.User
		withNewUser bool
		statusCode  int
	}{
		{
			name: `select existing user`,
			user: users.User{
				UserDB: db.UserDB{
					FirstName: "select",
					LastName:  "user III",
					Username:  "select_user_43",
					Email:     "select433@gmail.com",
					ImageURL:  "amazon.cdn.com",
				},
			},
			statusCode:  http.StatusOK,
			withNewUser: true,
		},
		{
			name:        `select non-existing user`,
			user:        users.User{},
			statusCode:  http.StatusBadRequest,
			withNewUser: false,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			userID := 0

			if c.withNewUser {
				apitest.CreateUser(&c.user, env)
				userID = c.user.ID
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

				compareSharedFields(t, &nu, &c.user)
			}
		})
	}
}

func TestUpdateUserHandler(t *testing.T) {
	cases := []struct {
		name           string
		updateUserJSON string
		statusCode     int
		userID         int
	}{
		{
			name:           `update user first name`,
			updateUserJSON: `{"first_name": "New First Name"}`,
			statusCode:     http.StatusOK,
			userID:         newUser.ID,
		},
		{
			name:           `update user last name`,
			updateUserJSON: `{"last_name": "New Last Name"}`,
			statusCode:     http.StatusOK,
			userID:         newUser.ID,
		},
		{
			name:           `update user username`,
			updateUserJSON: `{"username": "New Username"}`,
			statusCode:     http.StatusOK,
			userID:         newUser.ID,
		},
		{
			name:           `update user email`,
			updateUserJSON: `{"email": "new_email433@gmail.com"}`,
			statusCode:     http.StatusOK,
			userID:         newUser.ID,
		},
		{
			name:           `update user image url`,
			updateUserJSON: `{"image_url": "aws.cdn.com"}`,
			statusCode:     http.StatusOK,
			userID:         newUser.ID,
		},
		{
			name:           `update joined at time`,
			updateUserJSON: fmt.Sprintf("{\"joined_at\": %v}", time.Now()),
			statusCode:     http.StatusBadRequest,
			userID:         newUser.ID,
		},
		{
			name:           `update last visit time`,
			updateUserJSON: fmt.Sprintf("{\"last_visit\": %v}", time.Now()),
			statusCode:     http.StatusBadRequest,
			userID:         newUser.ID,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			req, err := http.NewRequest(
				"PATCH",
				fmt.Sprintf("/%d", c.userID),
				strings.NewReader(c.updateUserJSON))
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

func compareSharedFields(t *testing.T, got *users.User, expected *users.User) {
	if got.ImageURL != expected.ImageURL {
		t.Errorf("expected user ImageURL to be %s got %s", expected.ImageURL, got.ImageURL)
	}

	if got.Email != expected.Email {
		t.Errorf("expected user Email to be %s got %s", expected.Email, got.Email)
	}

	if got.FirstName != expected.FirstName {
		t.Errorf("expected user FirstName to be %s got %s", expected.FirstName, got.FirstName)
	}

	if got.Username != expected.Username {
		t.Errorf("expected user Username to be %s got %s", expected.Username, got.Username)
	}

	if got.LastName != expected.LastName {
		t.Errorf("expected user LastName to be %s got %s", expected.LastName, got.LastName)
	}
}
