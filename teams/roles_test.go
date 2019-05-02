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
	"delete_team":     `/%d`,
	"post_team":       `/`,
	"patch_team":      `/%d`,
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

func TestGeneratePostPlayer(t *testing.T) {
	perm := teams.GeneratePermission("post_player", 1)
	cases := append(generateCases("post_player", 1), &testCase{
		name:     "wrong_id_post_player",
		req:      getReq(43, "post_player"),
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

func TestGenerateDeletePlayer(t *testing.T) {
	perm := teams.GeneratePermission("delete_player", 1)
	cases := append(generateCases("delete_player", 1), &testCase{
		name:     "wrong_id_delete_player",
		req:      getReq(43, "delete_player"),
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

func TestGenerateDeleteTeam(t *testing.T) {
	perm := teams.GeneratePermission("delete_team", 1)
	cases := append(generateCases("delete_team", 1), &testCase{
		name:     "wrong_id_delete_team",
		req:      getReq(43, "delete_team"),
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
func TestGeneratePostTeam(t *testing.T) {
	perm := teams.GeneratePermission("post_team", 1)
	cases := generateCases("post_team", 1)

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := perm.Authorized(c.req)

			if got != c.expected {
				t.Fatalf("expected %v got %v", c.expected, got)
			}
		})
	}
}

func TestGeneratePatchTeam(t *testing.T) {
	perm := teams.GeneratePermission("patch_team", 1)
	cases := append(generateCases("patch_team", 1), &testCase{
		name:     "wrong_id_patch_team",
		req:      getReq(43, "patch_team"),
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

func TestGenerateGetLocations(t *testing.T) {
	perm := teams.GeneratePermission("get_locations", 1)
	cases := append(generateCases("get_locations", 1), &testCase{
		name:     "wrong_id_get_locations",
		req:      getReq(43, "get_locations"),
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

func TestGeneratePostLocation(t *testing.T) {
	perm := teams.GeneratePermission("post_location", 1)
	cases := append(generateCases("post_location", 1), &testCase{
		name:     "wrong_id_post_location",
		req:      getReq(43, "post_location"),
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

func TestGenerateDeleteLocation(t *testing.T) {
	perm := teams.GeneratePermission("delete_location", 1)
	cases := append(generateCases("delete_location", 1), &testCase{
		name:     "wrong_id_delete_location",
		req:      getReq(43, "delete_location"),
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

func TestGenerateGetMedia(t *testing.T) {
	perm := teams.GeneratePermission("get_media", 1)
	cases := append(generateCases("get_media", 1), &testCase{
		name:     "wrong_id_get_media",
		req:      getReq(43, "get_media"),
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
func TestGeneratePostMedia(t *testing.T) {
	perm := teams.GeneratePermission("post_media", 1)
	cases := append(generateCases("post_media", 1), &testCase{
		name:     "wrong_id_post_media",
		req:      getReq(43, "post_media"),
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

func TestGenerateDeleteMedia(t *testing.T) {
	perm := teams.GeneratePermission("delete_media", 1)
	cases := append(generateCases("delete_media", 1), &testCase{
		name:     "wrong_id_delete_media",
		req:      getReq(43, "delete_media"),
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

func TestGeneratePostPopulate(t *testing.T) {
	perm := teams.GeneratePermission("post_populate", 1)
	cases := generateCases("post_populate", 1)

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := perm.Authorized(c.req)

			if got != c.expected {
				t.Fatalf("expected %v got %v", c.expected, got)
			}
		})
	}
}
