package roles

import (
	"fmt"
	"net/http"
	"regexp"
	"strings"
)

// Role is a structure that maps permissions to users
type Role struct {
	Name        string
	Permissions []*Permission
	Child       *Role
}

// New returns a new role
func New(role string, entityID int) *Role {
	return roleToGenerator[role](entityID)
}

// Add adds the childRole to the given role. This is how inheritence relationships
// are modeled
func (r *Role) Add(childRole *Role) {
	r.Child = childRole
}

/*
// AddTo adds the role to the given user
func (r *Role) AddTo(userID int) {

}
*/

// Authorized returns whether or not the role contains a permission that is
// authorized for the given req
func (r *Role) Authorized(req *http.Request) bool {
	for _, p := range r.Permissions {
		if p.Authorized(req) {
			return true
		}
	}

	if r.Child != nil {
		return r.Child.Authorized(req)
	}

	return false
}

// Permission is the primitive that is responsible for an endpoint authorization
type Permission struct {
	URLRegex string
	Method   string
}

// Authorized returns whether or not the permission is authorized for the
// given request
func (p *Permission) Authorized(r *http.Request) bool {
	if strings.ToLower(r.Method) != strings.ToLower(p.Method) {
		return false
	}

	regEx := regexp.MustCompile(p.URLRegex)
	return regEx.MatchString(r.URL.Path)
}

// GeneratePermission generates permission for the given route and entity id
func GeneratePermission(perm string, entityID int) *Permission {
	var regex string
	if strings.Contains(permToFormattedRegex[perm], "%") {
		regex = fmt.Sprintf(permToFormattedRegex[perm], entityID)
	} else {
		regex = permToFormattedRegex[perm]
	}
	permission := Permission{
		Method:   strings.Split(perm, "_")[0],
		URLRegex: regex,
	}

	return &permission
}

var permToFormattedRegex = map[string]string{
	// team endpoints
	"get_teams":           `^teams/$`,
	"get_team":            `^teams/%d$`,
	"get_points":          `^teams/%d/points/$`,
	"get_players":         `^teams/%d/players/$`,
	"post_player":         `^teams/%d/players/$`,
	"delete_player":       `^teams/%d/players/\d+$`,
	"delete_team":         `^teams/%d$`,
	"post_team":           `^teams/$`,
	"patch_team":          `^teams/%d$`,
	"get_locations":       `^teams/%d/locations/$`,
	"post_location":       `^teams/%d/locations/$`,
	"delete_location":     `^teams/%d/locations/\d+$`,
	"get_media":           `^teams/%d/media/$`,
	"post_media":          `^teams/%d/media/$`,
	"delete_media":        `^teams/%d/media/\d+$`,
	"post_teams_populate": `^teams/populate/$`,

	// user endpoints
	"get_user":    `^users/%d$`,
	"post_login":  `^users/login/$`,
	"post_logout": `^users/logout/$`,
	"post_user":   `^users/$`,
	"delete_user": `^users/%d$`,
	"patch_user":  `^users/%d$`,

	// hunt endpoints
	"get_hunts":           `^hunts/$`,
	"get_hunt":            `^hunts/%d$`,
	"post_hunt":           `^hunts/$`,
	"delete_hunt":         `^hunts/%d$`,
	"patch_hunt":          `^hunts/%d$`,
	"post_hunts_populate": `^hunts/populate/$`,
	"get_items":           `^hunts/%d/items/$`,
	"delete_item":         `^hunts/%d/items/\d+$`,
	"post_item":           `^hunts/%d/items/$`,
	"patch_item":          `^hunts/%d/items/\d+$`,
}

// PermToRole maps a permissions key to a role
var PermToRole = map[string]string{
	// team endpoints
	"get_teams":           `admin`,
	"get_team":            `user`,
	"get_points":          `hunt_member`,
	"get_players":         `user`,
	"post_player":         ``,
	"delete_player":       `team_editor`,
	"delete_team":         `team_owner`,
	"post_team":           ``,
	"patch_team":          `team_editor`,
	"get_locations":       `hunt_member`,
	"post_location":       `team_member`,
	"delete_location":     `admin`,
	"get_media":           `hunt_member`,
	"post_media":          `team_member`,
	"delete_media":        `team_member`,
	"post_teams_populate": `admin`,

	// user endpoints
	"get_user":    `public`,
	"post_login":  `user`,
	"post_logout": `user`,
	"post_user":   `public`,
	"delete_user": `user_owner`,
	"patch_user":  `user_owner`,

	// hunt endpoints
	"get_hunts":           `user`,
	"get_hunt":            `user`,
	"post_hunt":           `user`,
	"delete_hunt":         `hunt_owner`,
	"patch_hunt":          `hunt_editor`,
	"post_hunts_populate": `admin`,
	"get_items":           `user`,
	"delete_item":         `hunt_editor`,
	"post_item":           `hunt_editor`,
	"patch_item":          `hunt_editor`,
}

// the team role relationships look like: Owner -> Editor -> Member -> HuntMember -> User
var roleToGenerator = map[string]func(int) *Role{
	//	"admin":       genAdminRole,
	"hunt_owner":  genHuntOwnerRole,
	"hunt_editor": genHuntEditorRole,
	"hunt_member": genHuntMemberRole,
	"team_owner":  genTeamOwnerRole,
	"team_editor": genTeamEditorRole,
	"team_member": genTeamMemberRole,
	"user":        genUserRole,
	"user_owner":  genUserOwnerRole,
}

func genRole(name string, id int) *Role {
	role := Role{
		Name:        fmt.Sprintf("%s_%d", name, id),
		Permissions: make([]*Permission, 0),
	}
	// create role specific permissions
	for k, v := range PermToRole {
		if v == name {
			role.Permissions = append(role.Permissions, GeneratePermission(k, id))
		}
	}

	return &role
}

func genHuntOwnerRole(id int) *Role {
	owner := genRole("hunt_owner", id)
	owner.Add(genHuntEditorRole(id))

	return owner
}

func genHuntEditorRole(id int) *Role {
	editor := genRole("hunt_editor", id)
	editor.Add(genHuntMemberRole(id))

	return editor
}

func genHuntMemberRole(id int) *Role {
	editor := genRole("hunt_member", id)

	return editor
}
func genTeamOwnerRole(id int) *Role {
	owner := genRole("team_owner", id)
	owner.Add(genTeamEditorRole(id))

	return owner
}

func genTeamEditorRole(id int) *Role {
	editor := genRole("team_editor", id)
	editor.Add(genTeamMemberRole(id))

	return editor
}

func genTeamMemberRole(id int) *Role {
	return genRole("team_member", id)
}

func genUserRole(id int) *Role {
	return genRole("user", id)
}

func genUserOwnerRole(id int) *Role {
	return genRole("user_owner", id)
}
