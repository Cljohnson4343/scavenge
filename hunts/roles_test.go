// +build unit

package hunts_test

import (
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/cljohnson4343/scavenge/hunts"
)

var routes = map[string]string{
	"get_hunts":     `/`,
	"get_hunt":      `/%d`,
	"post_hunt":     `/`,
	"delete_hunt":   `/%d`,
	"patch_hunt":    `/%d`,
	"post_populate": `/populate/`,
	"get_items":     `/%d/items/`,
	"delete_item":   `/%d/items/43`,
	"post_item":     `/%d/items/`,
	"patch_item":    `/%d/items/43`,
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

func TestGenerateGetHunts(t *testing.T) {
	perm := hunts.GeneratePermission("get_hunts", 1)
	cases := generateCases("get_hunts", 1)

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

func TestGenerateGetHunt(t *testing.T) {
	perm := hunts.GeneratePermission("get_hunt", 1)
	cases := append(generateCases("get_hunt", 1), &testCase{
		name:     "wrong_id_get_hunt",
		req:      getReq(43, "get_hunt"),
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

func TestGeneratePostHunt(t *testing.T) {
	perm := hunts.GeneratePermission("post_hunt", 1)
	cases := generateCases("post_hunt", 1)

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

func TestGenerateDeleteHunt(t *testing.T) {
	perm := hunts.GeneratePermission("delete_hunt", 1)
	cases := append(generateCases("delete_hunt", 1), &testCase{
		name:     "wrong_id_delete_hunt",
		req:      getReq(43, "delete_hunt"),
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

func TestGeneratePatchHunt(t *testing.T) {
	perm := hunts.GeneratePermission("patch_hunt", 1)
	cases := append(generateCases("patch_hunt", 1), &testCase{
		name:     "wrong_id_patch_hunt",
		req:      getReq(43, "patch_hunt"),
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

func TestGeneratePostPopulate(t *testing.T) {
	perm := hunts.GeneratePermission("post_populate", 1)
	cases := generateCases("post_populate", 1)

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

func TestGenerateGetItems(t *testing.T) {
	perm := hunts.GeneratePermission("get_items", 1)
	cases := append(generateCases("get_items", 1), &testCase{
		name:     "wrong_id_get_items",
		req:      getReq(43, "get_items"),
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

func TestGenerateDeleteItem(t *testing.T) {
	perm := hunts.GeneratePermission("delete_item", 1)
	cases := append(generateCases("delete_item", 1), &testCase{
		name:     "wrong_id_delete_item",
		req:      getReq(43, "delete_item"),
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

func TestGeneratePostItem(t *testing.T) {
	perm := hunts.GeneratePermission("post_item", 1)
	cases := append(generateCases("post_item", 1), &testCase{
		name:     "wrong_id_post_item",
		req:      getReq(43, "post_item"),
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

func TestGeneratePatchItem(t *testing.T) {
	perm := hunts.GeneratePermission("patch_item", 1)
	cases := append(generateCases("patch_item", 1), &testCase{
		name:     "wrong_id_patch_item",
		req:      getReq(43, "patch_item"),
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
