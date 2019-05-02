// +build unit

package teams_test

import (
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/cljohnson4343/scavenge/teams"
)

var routes = map[string]string{
	"get_teams":       `/`,
	"get_team":        `/%d`,
	"get_points":      `/%d/points/`,
	"get_players":     `/%d/players/`,
	"post_player":     `/%d/players/`,
	"delete_player":   `/%d/players/43`,
	"delete_team":     `/43`,
	"post_team":       `/`,
	"patch_team":      `/43`,
	"get_locations":   `/%d/locations/`,
	"post_location":   `/%d/locations/`,
	"delete_location": `/%d/locations/43`,
	"get_media":       `/%d/media/`,
	"post_media":      `/%d/media/`,
	"delete_media":    `/%d/media/43`,
	"post_populate":   `/populate/`,
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

func TestGenerateGetTeams(t *testing.T) {
	perm := teams.GeneratePermission("get_teams", 1)
	cases := generateCases("get_teams", 1)

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := perm.Authorized(c.req)
			if got != c.expected {
				t.Errorf(
					"epected %v got %v\n\tregex: %s\n\turl: %s\n",
					c.expected,
					got,
					perm.URLRegex,
					c.req.URL.Path,
				)
			}
		})
	}
}

func TestGenerateGetTeam(t *testing.T) {
	perm := teams.GeneratePermission("get_team", 1)
	cases := append(generateCases("get_team", 1), &testCase{
		name:     "wrong_id_get_team",
		req:      getReq(43, "get_team"),
		expected: false,
	})

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := perm.Authorized(c.req)

			if got != c.expected {
				t.Fatalf("expected %v got %v", c.expected, got)
			}
		})
	}
}

func TestGenerateGetPoints(t *testing.T) {
	perm := teams.GeneratePermission("get_points", 1)
	cases := append(generateCases("get_points", 1), &testCase{
		name:     "wrong_id_get_points",
		req:      getReq(43, "get_points"),
		expected: false,
	})

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := perm.Authorized(c.req)

			if got != c.expected {
				t.Fatalf("expected %v got %v", c.expected, got)
			}
		})
	}
}

func TestGenerateGetPlayers(t *testing.T) {
	perm := teams.GeneratePermission("get_players", 1)
	cases := append(generateCases("get_players", 1), &testCase{
		name:     "wrong_id_get_players",
		req:      getReq(43, "get_players"),
		expected: false,
	})

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := perm.Authorized(c.req)

			if got != c.expected {
				t.Fatalf("expected %v got %v", c.expected, got)
			}
		})
	}
}
