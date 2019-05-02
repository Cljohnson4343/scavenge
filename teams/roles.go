package teams

import (
	"fmt"
	"strings"

	"github.com/cljohnson4343/scavenge/roles"
)

var permToFormattedRegex = map[string]string{
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
	if strings.Contains(permToFormattedRegex[perm], "%") {
		regex = fmt.Sprintf(permToFormattedRegex[perm], id)
	} else {
		regex = permToFormattedRegex[perm]
	}
	permission := roles.Permission{
		Method:   strings.Split(perm, "_")[0],
		URLRegex: regex,
	}

	return &permission
}

// PermToRole maps a permissions key to a role
var PermToRole = map[string]string{
	"get_teams":       `admin`,
	"get_team":        `hunt_invitee`,
	"get_points":      `hunt_member`,
	"get_players":     `hunt_invitee`,
	"post_player":     ``,
	"delete_player":   `team_editor`,
	"delete_team":     `team_owner`,
	"post_team":       ``,
	"patch_team":      `team_editor`,
	"get_locations":   `hunt_member`,
	"post_location":   `team_member`,
	"delete_location": `admin`,
	"get_media":       `hunt_member`,
	"post_media":      `team_member`,
	"delete_media":    `team_member`,
	"post_populate":   `admin`,
}

// the team role relationships look like: Owner -> Editor -> Member -> HuntMember -> User
var roleToGenerator = map[string]func(int) *roles.Role{
	"team_owner":  genTeamOwnerRole,
	"team_editor": genTeamEditorRole,
	"team_member": genTeamMemberRole,
}

// GenerateRole creates a role for the given entity id.
func GenerateRole(role string, id int) *roles.Role {
	return roleToGenerator[role](id)
}

func genTeamOwnerRole(id int) *roles.Role {
	owner := roles.Role{
		Name:        fmt.Sprintf("team_owner_%d", id),
		Permissions: make([]*roles.Permission, 0),
	}
	// create owner specific permissions
	for k, v := range PermToRole {
		if v == "team_owner" {
			perm := GeneratePermission(k, id)

			owner.Permissions = append(owner.Permissions, perm)
		}
	}

	// create editor role and add to owner role
	owner.Add(genTeamEditorRole(id))

	return &owner
}

func genTeamEditorRole(id int) *roles.Role {
	editor := roles.Role{
		Name:        fmt.Sprintf("team_editor_%d", id),
		Permissions: make([]*roles.Permission, 0),
	}
	// create owner specific permissions
	for k, v := range PermToRole {
		if v == "team_editor" {
			perm := GeneratePermission(k, id)

			editor.Permissions = append(editor.Permissions, perm)
		}
	}

	// create hunt member role and add to member role
	editor.Add(genTeamMemberRole(id))

	return &editor
}

func genTeamMemberRole(id int) *roles.Role {
	member := roles.Role{
		Name:        fmt.Sprintf("team_member%d", id),
		Permissions: make([]*roles.Permission, 0),
	}
	// create owner specific permissions
	for k, v := range PermToRole {
		if v == "team_member" {
			perm := GeneratePermission(k, id)

			member.Permissions = append(member.Permissions, perm)
		}
	}

	// create hunt member role and add to member role

	// TODO make sure team_member inherits hunt_member
	return &member
}

/*
func genHuntMemberRole(id int) *roles.Role {
	perms := make([]*roles.Permission, 0)
	// create member specific permissions
	for k, v := range PermToRole {
		if v == "hunt_member" {
			perm := GeneratePermission(k, id)
			perms = append(perms, perm)
		}
	}

}
*/
