package teams

import (
	"fmt"
	"strings"

	"github.com/cljohnson4343/scavenge/roles"
)

var permissionMap = map[string]string{
	"get_teams":       `^/$`,
	"get_team":        `^/%d$`,
	"get_points":      `^/%d/points/$`,
	"get_players":     `^/%d/players/$`,
	"post_player":     `^/%d/players/$`,
	"delete_player":   `^/%d/players/\d+$`,
	"delete_team":     `^/%d$`,
	"post_team":       `^/$`,
	"patch_team":      `^/%d$`,
	"get_locations":   `^/%d/locations/$`,
	"post_location":   `^/%d/locations/$`,
	"delete_location": `^/%d/locations/\d+$`,
	"get_media":       `^/%d/media/$`,
	"post_media":      `^/%d/media/$`,
	"delete_media":    `^/%d/media/\d+$`,
	"post_populate":   `^/populate/$`,
}

// GeneratePermission generates permission for the given route and team id
func GeneratePermission(perm string, id int) *roles.Permission {
	var regex string
	if strings.Contains(permissionMap[perm], "%") {
		regex = fmt.Sprintf(permissionMap[perm], id)
	} else {
		regex = permissionMap[perm]
	}
	permission := roles.Permission{
		Method:   strings.Split(perm, "_")[0],
		URLRegex: regex,
	}

	return &permission
}

/*
// CreateRole reates a role for the given entity id
func CreateRole(role string, id int) *roles.Role {

}
*/
