// +build unit

package users_test

import (
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/cljohnson4343/scavenge/users"
)

var routes = map[string]string{
	"get_user":    `/%d`,
	"post_login":  `/login/`,
	"post_logout": `/logout/`,
	"post_user":   `/`,
	"delete_user": `/%d`,
	"patch_user":  `/%d`,
}

type testCase struct {
	name     string
	expected bool
	req      *http.Request
}

func getReq(id int, key string) *http.Request {
	var url string
	if strings.Contains(routes[key], "%") {
		url = fmt.Sprintf(routes[key], id)
	} else {
		url = routes[key]
	}

	req, err := http.NewRequest(strings.Split(key, "_")[0], url, nil)
	if err != nil {
		panic(err.Error())
	}

	return req
}

func generateCases(perm string, id int) []*testCase {
	cases := make([]*testCase, 0, len(routes))

	for k, _ := range routes {
		c := testCase{
			name:     k,
			expected: false,
			req:      getReq(id, k),
		}
		if k == perm {
			c.expected = true
		}

		cases = append(cases, &c)
	}

	return cases
}

func TestGenerateGetUser(t *testing.T) {
	perm := users.GeneratePermission("get_user", 1)
	cases := append(generateCases("get_user", 1), &testCase{
		name:     "wrong_id_get_user",
		req:      getReq(43, "get_user"),
		expected: false,
	})

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := perm.Authorized(c.req)
			if got != c.expected {
				t.Errorf(
					"expected %v got %v\n\tregex: %s\n\turl: %s\n",
					c.expected,
					got,
					perm.URLRegex,
					c.req.URL.Path,
				)
			}
		})
	}
}

func TestGeneratePostLogin(t *testing.T) {
	perm := users.GeneratePermission("post_login", 1)
	cases := generateCases("post_login", 1)

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := perm.Authorized(c.req)

			if got != c.expected {
				t.Fatalf("expected %v got %v", c.expected, got)
			}
		})
	}
}

func TestGeneratePostLogout(t *testing.T) {
	perm := users.GeneratePermission("post_logout", 1)
	cases := generateCases("post_logout", 1)

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := perm.Authorized(c.req)
			if got != c.expected {
				t.Errorf(
					"expected %v got %v\n\tregex: %s\n\turl: %s\n",
					c.expected,
					got,
					perm.URLRegex,
					c.req.URL.Path,
				)
			}
		})
	}
}

func TestGeneratePostUser(t *testing.T) {
	perm := users.GeneratePermission("post_user", 1)
	cases := generateCases("post_user", 1)

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := perm.Authorized(c.req)
			if got != c.expected {
				t.Errorf(
					"expected %v got %v\n\tregex: %s\n\turl: %s\n",
					c.expected,
					got,
					perm.URLRegex,
					c.req.URL.Path,
				)
			}
		})
	}
}

func TestGenerateDeleteUser(t *testing.T) {
	perm := users.GeneratePermission("delete_user", 1)
	cases := append(generateCases("delete_user", 1), &testCase{
		name:     "wrong_id_delete_user",
		req:      getReq(43, "delete_user"),
		expected: false,
	})

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := perm.Authorized(c.req)
			if got != c.expected {
				t.Errorf(
					"expected %v got %v\n\tregex: %s\n\turl: %s\n",
					c.expected,
					got,
					perm.URLRegex,
					c.req.URL.Path,
				)
			}
		})
	}
}

func TestGeneratePatchUser(t *testing.T) {
	perm := users.GeneratePermission("patch_user", 1)
	cases := append(generateCases("patch_user", 1), &testCase{
		name:     "wrong_id_patch_user",
		req:      getReq(43, "patch_user"),
		expected: false,
	})

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := perm.Authorized(c.req)
			if got != c.expected {
				t.Errorf(
					"expected %v got %v\n\tregex: %s\n\turl: %s\n",
					c.expected,
					got,
					perm.URLRegex,
					c.req.URL.Path,
				)
			}
		})
	}
}
