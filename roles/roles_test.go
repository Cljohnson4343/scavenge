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
	if strings.Contains(roles.PermToRoutes[key], "%") {
		url = fmt.Sprintf(roles.PermToRoutes[key], id)
	} else {
		url = roles.PermToRoutes[key]
	}

	req, err := http.NewRequest(strings.Split(key, "_")[0], url, nil)
	if err != nil {
		panic(err.Error())
	}

	return req
}

func generatePermissionCases(perm string, id int) []*testCase {
	cases := make([]*testCase, 0, len(roles.PermToRoutes))

	for k, _ := range roles.PermToRoutes {
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
					"epected %v got %v\n\tregex: %s\n\turl: %s\n",
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
	optional := &testCase{
		name:     "wrong_id_get_team",
		req:      getReq(43, "get_team"),
		expected: false,
	}
	testGeneratePermission(t, "get_team", optional)
}

func TestGenerateGetPoints(t *testing.T) {
	optional := &testCase{
		name:     "wrong_id_get_points",
		req:      getReq(43, "get_points"),
		expected: false,
	}

	testGeneratePermission(t, "get_points", optional)
}

func TestGenerateGetPlayers(t *testing.T) {
	optional := &testCase{
		name:     "wrong_id_get_players",
		req:      getReq(43, "get_players"),
		expected: false,
	}
	testGeneratePermission(t, "get_players", optional)
}

func TestGeneratePostPlayer(t *testing.T) {
	optional := &testCase{
		name:     "wrong_id_post_player",
		req:      getReq(43, "post_player"),
		expected: false,
	}
	testGeneratePermission(t, "post_player", optional)
}

func TestGenerateDeletePlayer(t *testing.T) {
	optional := &testCase{
		name:     "wrong_id_delete_player",
		req:      getReq(43, "delete_player"),
		expected: false,
	}

	testGeneratePermission(t, "delete_player", optional)
}

func TestGenerateDeleteTeam(t *testing.T) {
	optional := &testCase{
		name:     "wrong_id_delete_team",
		req:      getReq(43, "delete_team"),
		expected: false,
	}
	testGeneratePermission(t, "delete_team", optional)
}

func TestGeneratePostTeam(t *testing.T) {
	testGeneratePermission(t, "post_team", nil)
}

func TestGeneratePatchTeam(t *testing.T) {
	optional := &testCase{
		name:     "wrong_id_patch_team",
		req:      getReq(43, "patch_team"),
		expected: false,
	}
	testGeneratePermission(t, "patch_team", optional)
}

func TestGenerateGetLocations(t *testing.T) {
	optional := &testCase{
		name:     "wrong_id_get_locations",
		req:      getReq(43, "get_locations"),
		expected: false,
	}
	testGeneratePermission(t, "get_locations", optional)
}

func TestGeneratePostLocation(t *testing.T) {
	optional := &testCase{
		name:     "wrong_id_post_location",
		req:      getReq(43, "post_location"),
		expected: false,
	}
	testGeneratePermission(t, "post_location", optional)
}

func TestGenerateDeleteLocation(t *testing.T) {
	optional := &testCase{
		name:     "wrong_id_delete_location",
		req:      getReq(43, "delete_location"),
		expected: false,
	}
	testGeneratePermission(t, "delete_location", optional)
}

func TestGenerateGetMedia(t *testing.T) {
	optional := &testCase{
		name:     "wrong_id_get_media",
		req:      getReq(43, "get_media"),
		expected: false,
	}
	testGeneratePermission(t, "get_media", optional)
}

func TestGeneratePostMedia(t *testing.T) {
	optional := &testCase{
		name:     "wrong_id_post_media",
		req:      getReq(43, "post_media"),
		expected: false,
	}
	testGeneratePermission(t, "post_media", optional)
}

func TestGenerateDeleteMedia(t *testing.T) {
	optional := &testCase{
		name:     "wrong_id_delete_media",
		req:      getReq(43, "delete_media"),
		expected: false,
	}
	testGeneratePermission(t, "delete_media", optional)
}

func TestGeneratePostTeamsPopulate(t *testing.T) {
	testGeneratePermission(t, "post_teams_populate", nil)
}

func TestGenerateGetUser(t *testing.T) {
	optional := &testCase{
		name:     "wrong_id_get_user",
		req:      getReq(43, "get_user"),
		expected: false,
	}
	testGeneratePermission(t, "get_user", optional)
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
	optional := &testCase{
		name:     "wrong_id_delete_user",
		req:      getReq(43, "delete_user"),
		expected: false,
	}
	testGeneratePermission(t, "delete_user", optional)
}

func TestGeneratePatchUser(t *testing.T) {
	optional := &testCase{
		name:     "wrong_id_patch_user",
		req:      getReq(43, "patch_user"),
		expected: false,
	}
	testGeneratePermission(t, "patch_user", optional)
}

func TestGenerateGetHunts(t *testing.T) {
	testGeneratePermission(t, "get_hunts", nil)
}

func TestGenerateGetHunt(t *testing.T) {
	optional := &testCase{
		name:     "wrong_id_get_hunt",
		req:      getReq(43, "get_hunt"),
		expected: false,
	}
	testGeneratePermission(t, "get_hunt", optional)
}

func TestGeneratePostHunt(t *testing.T) {
	testGeneratePermission(t, "post_hunt", nil)
}

func TestGenerateDeleteHunt(t *testing.T) {
	optional := &testCase{
		name:     "wrong_id_delete_hunt",
		req:      getReq(43, "delete_hunt"),
		expected: false,
	}
	testGeneratePermission(t, "delete_hunt", optional)
}

func TestGeneratePatchHunt(t *testing.T) {
	optional := &testCase{
		name:     "wrong_id_patch_hunt",
		req:      getReq(43, "patch_hunt"),
		expected: false,
	}
	testGeneratePermission(t, "patch_hunt", optional)
}

func TestGeneratePostHuntsPopulate(t *testing.T) {
	testGeneratePermission(t, "post_hunts_populate", nil)
}

func TestGenerateGetItems(t *testing.T) {
	optional := &testCase{
		name:     "wrong_id_get_items",
		req:      getReq(43, "get_items"),
		expected: false,
	}
	testGeneratePermission(t, "get_items", optional)
}

func TestGenerateDeleteItem(t *testing.T) {
	optional := &testCase{
		name:     "wrong_id_delete_item",
		req:      getReq(43, "delete_item"),
		expected: false,
	}
	testGeneratePermission(t, "delete_item", optional)
}

func TestGeneratePostItem(t *testing.T) {
	optional := &testCase{
		name:     "wrong_id_post_item",
		req:      getReq(43, "post_item"),
		expected: false,
	}
	testGeneratePermission(t, "post_item", optional)
}

func TestGeneratePatchItem(t *testing.T) {
	optional := &testCase{
		name:     "wrong_id_patch_item",
		req:      getReq(43, "patch_item"),
		expected: false,
	}
	testGeneratePermission(t, "patch_item", optional)
}

//
// role testing
//

func generateRoleCases(role string, id int) []*testCase {
	cases := make([]*testCase, 0, len(roles.PermToRoutes))

	for k, _ := range roles.PermToRoutes {
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
		if roles.PermToRole[permKey] == role ||
			roles.PermToRole[permKey] == "team_editor" ||
			roles.PermToRole[permKey] == "team_member" {
			return true
		}
	case "team_editor":
		if roles.PermToRole[permKey] == role ||
			roles.PermToRole[permKey] == "team_member" {
			return true
		}
	case "team_member":
		if roles.PermToRole[permKey] == role {
			return true
		}
	case "hunt_owner":
		if roles.PermToRole[permKey] == role ||
			roles.PermToRole[permKey] == "hunt_editor" ||
			roles.PermToRole[permKey] == "hunt_member" {
			return true
		}
	case "hunt_editor":
		if roles.PermToRole[permKey] == role ||
			roles.PermToRole[permKey] == "hunt_member" {
			return true
		}
	case "hunt_member":
		if roles.PermToRole[permKey] == role {
			return true
		}
	case "user":
		if roles.PermToRole[permKey] == role {
			return true
		}
	case "user_owner":
		if roles.PermToRole[permKey] == role {
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
				t.Errorf("expected %v got %v", c.expected, got)
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
