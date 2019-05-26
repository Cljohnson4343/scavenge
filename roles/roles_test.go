// +build unit

package roles_test

import (
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/cljohnson4343/scavenge/roles"
)

type testCase struct {
	name     string
	expected bool
	req      *http.Request
}

func getReq(id int, key string) *http.Request {
	var url string
	if strings.Contains(roles.PermToRoleEndpoint[key].Route, "%") {
		url = fmt.Sprintf(roles.PermToRoleEndpoint[key].Route, id)
	} else {
		url = roles.PermToRoleEndpoint[key].Route
	}

	req, err := http.NewRequest(strings.Split(key, "_")[0], url, nil)
	if err != nil {
		panic(err.Error())
	}

	return req
}

func generatePermissionCases(perm string, id int) []*testCase {
	cases := make([]*testCase, 0, len(roles.PermToRoleEndpoint))

	for k, _ := range roles.PermToRoleEndpoint {
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

func testGeneratePermission(t *testing.T, perm string, optionalCase *testCase) {
	permission := roles.GeneratePermission(perm, 1)
	cases := generatePermissionCases(perm, 1)
	if optionalCase != nil {
		cases = append(cases, optionalCase)
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := permission.Authorized(c.req)
			if got != c.expected {
				t.Errorf(
					"expected %v got %v\n\tregex: %s\n\turl: %s\n",
					c.expected,
					got,
					permission.URLRegex,
					c.req.URL.Path,
				)
			}
		})
	}
}

func TestGenerateGetTeams(t *testing.T) {
	testGeneratePermission(t, "get_teams", nil)
}

func TestGenerateGetTeam(t *testing.T) {
	testGeneratePermission(t, "get_team", nil)
}

func TestGenerateGetPoints(t *testing.T) {
	testGeneratePermission(t, "get_points", nil)
}

func TestGenerateGetPlayers(t *testing.T) {
	testGeneratePermission(t, "get_players", nil)
}

func TestGeneratePostPlayer(t *testing.T) {
	testGeneratePermission(t, "post_player", nil)
}

func TestGenerateDeletePlayer(t *testing.T) {
	testGeneratePermission(t, "delete_player", nil)
}

func TestGenerateDeleteTeam(t *testing.T) {
	testGeneratePermission(t, "delete_team", nil)
}

func TestGeneratePostTeam(t *testing.T) {
	testGeneratePermission(t, "post_team", nil)
}

func TestGeneratePatchTeam(t *testing.T) {
	testGeneratePermission(t, "patch_team", nil)
}

func TestGenerateGetLocations(t *testing.T) {
	testGeneratePermission(t, "get_locations", nil)
}

func TestGeneratePostLocation(t *testing.T) {
	testGeneratePermission(t, "post_location", nil)
}

func TestGenerateDeleteLocation(t *testing.T) {
	testGeneratePermission(t, "delete_location", nil)
}

func TestGenerateGetMedia(t *testing.T) {
	testGeneratePermission(t, "get_media", nil)
}

func TestGeneratePostMedia(t *testing.T) {
	testGeneratePermission(t, "post_media", nil)
}

func TestGenerateDeleteMedia(t *testing.T) {
	testGeneratePermission(t, "delete_media", nil)
}

func TestGeneratePostTeamsPopulate(t *testing.T) {
	testGeneratePermission(t, "post_teams_populate", nil)
}

func TestGenerateGetUser(t *testing.T) {
	testGeneratePermission(t, "get_user", nil)
}

func TestGeneratePostLogin(t *testing.T) {
	testGeneratePermission(t, "post_login", nil)
}

func TestGeneratePostLogout(t *testing.T) {
	testGeneratePermission(t, "post_logout", nil)
}

func TestGeneratePostUser(t *testing.T) {
	testGeneratePermission(t, "post_user", nil)
}

func TestGenerateDeleteUser(t *testing.T) {
	testGeneratePermission(t, "delete_user", nil)
}

func TestGeneratePatchUser(t *testing.T) {
	testGeneratePermission(t, "patch_user", nil)
}

func TestGenerateGetHunts(t *testing.T) {
	testGeneratePermission(t, "get_hunts", nil)
}

func TestGenerateGetHunt(t *testing.T) {
	testGeneratePermission(t, "get_hunt", nil)
}

func TestGeneratePostHunt(t *testing.T) {
	testGeneratePermission(t, "post_hunt", nil)
}

func TestGenerateDeleteHunt(t *testing.T) {
	testGeneratePermission(t, "delete_hunt", nil)
}

func TestGeneratePatchHunt(t *testing.T) {
	testGeneratePermission(t, "patch_hunt", nil)
}

func TestGeneratePostHuntsPopulate(t *testing.T) {
	testGeneratePermission(t, "post_hunts_populate", nil)
}

func TestGenerateGetItems(t *testing.T) {
	testGeneratePermission(t, "get_items", nil)
}

func TestGenerateDeleteItem(t *testing.T) {
	testGeneratePermission(t, "delete_item", nil)
}

func TestGeneratePostItem(t *testing.T) {
	testGeneratePermission(t, "post_item", nil)
}

func TestGeneratePatchItem(t *testing.T) {
	testGeneratePermission(t, "patch_item", nil)
}

//
// role testing
//

func generateRoleCases(role string, id int) []*testCase {
	cases := make([]*testCase, 0, len(roles.PermToRoleEndpoint))

	for k, _ := range roles.PermToRoleEndpoint {
		c := testCase{
			name:     k,
			expected: getExpected(role, k),
			req:      getReq(id, k),
		}

		cases = append(cases, &c)
	}

	return cases
}

func getExpected(role, permKey string) bool {
	switch role {
	case "team_owner":
		if roles.PermToRoleEndpoint[permKey].Role == role ||
			roles.PermToRoleEndpoint[permKey].Role == "team_editor" ||
			roles.PermToRoleEndpoint[permKey].Role == "team_member" {
			return true
		}
	case "team_editor":
		if roles.PermToRoleEndpoint[permKey].Role == role ||
			roles.PermToRoleEndpoint[permKey].Role == "team_member" {
			return true
		}
	case "team_member":
		if roles.PermToRoleEndpoint[permKey].Role == role {
			return true
		}
	case "hunt_owner":
		if roles.PermToRoleEndpoint[permKey].Role == role ||
			roles.PermToRoleEndpoint[permKey].Role == "hunt_editor" ||
			roles.PermToRoleEndpoint[permKey].Role == "hunt_member" {
			return true
		}
	case "hunt_editor":
		if roles.PermToRoleEndpoint[permKey].Role == role ||
			roles.PermToRoleEndpoint[permKey].Role == "hunt_member" {
			return true
		}
	case "hunt_member":
		if roles.PermToRoleEndpoint[permKey].Role == role {
			return true
		}
	case "user":
		if roles.PermToRoleEndpoint[permKey].Role == role {
			return true
		}
	case "user_owner":
		if roles.PermToRoleEndpoint[permKey].Role == role {
			return true
		}
	case "admin":
		if roles.PermToRoleEndpoint[permKey].Role == role {
			return true
		}
	}

	return false
}

func testGenerateRole(t *testing.T, name string) {
	role := roles.New(name, 1)
	cases := generateRoleCases(name, 1)

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := role.Authorized(c.req)

			if got != c.expected {
				t.Errorf("expected %v got %v for request: %s", c.expected, got, c.req.URL.Path)
			}
		})
	}
}

func TestGenerateTeamOwnerRole(t *testing.T) {
	testGenerateRole(t, "team_owner")
}

func TestGenerateTeamEditorRole(t *testing.T) {
	testGenerateRole(t, "team_editor")
}

func TestGenerateTeamMemberRole(t *testing.T) {
	testGenerateRole(t, "team_member")
}

func TestGenerateHuntOwnerRole(t *testing.T) {
	testGenerateRole(t, "hunt_owner")
}

func TestGenerateHuntEditorRole(t *testing.T) {
	testGenerateRole(t, "hunt_editor")
}

func TestGenerateHuntMemberRole(t *testing.T) {
	testGenerateRole(t, "hunt_member")
}

func TestGenerateUserRole(t *testing.T) {
	testGenerateRole(t, "user")
}

func TestGenerateUserOwnerRole(t *testing.T) {
	testGenerateRole(t, "user_owner")
}

func TestRoleDBs(t *testing.T) {
	cases := []struct {
		name           string
		expectedLength int
		role           string
	}{
		{
			name:           "team owner",
			role:           "team_owner",
			expectedLength: 3,
		},
		{
			name:           "team editor",
			role:           "team_editor",
			expectedLength: 2,
		},
		{
			name:           "team member",
			role:           "team_member",
			expectedLength: 1,
		},
		{
			name:           "hunt owner",
			role:           "hunt_owner",
			expectedLength: 3,
		},
		{
			name:           "hunt editor",
			role:           "hunt_editor",
			expectedLength: 2,
		},
		{
			name:           "hunt member",
			role:           "hunt_member",
			expectedLength: 1,
		},
		{
			name:           "user",
			role:           "user",
			expectedLength: 1,
		},
		{
			name:           "user owner",
			role:           "user_owner",
			expectedLength: 1,
		},
	}

	entityID := 43
	userID := 1

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			role := roles.New(c.role, entityID)

			roleDBs := role.RoleDBs(userID)

			if len(roleDBs) != c.expectedLength {
				t.Errorf(
					"expected %d roles to be returned but got %d",
					c.expectedLength,
					len(roleDBs),
				)
			}
		})
	}
}
