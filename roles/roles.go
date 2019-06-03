package roles

import (
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"github.com/cljohnson4343/scavenge/db"
	"github.com/cljohnson4343/scavenge/response"
)

// Role is a structure that maps permissions to users
type Role struct {
	Name        string
	Permissions []*Permission
	Child       *Role
	EntityID    int
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

// AddTo adds the role to the given user
func (r *Role) AddTo(userID int) *response.Error {
	return db.AddRoles(r.RoleDBs(userID))
}

// RemoveRole removes the role from the user without recursively
// removing children roles
func RemoveRole(roleID, userID int) *response.Error {
	return db.RemoveRole(roleID, userID)
}

// DeleteRolesForTeam deletes all the roles and permissions for the given team
func DeleteRolesForTeam(teamID int) *response.Error {
	regex := fmt.Sprintf("team_[a-zA-Z]+_%d", teamID)

	return db.DeleteRolesByRegex(regex)
}

// DeleteRolesForHunt deletes all the roles and permissions for the given hunt
// this includes all roles and permissions for teams in the hunt
func DeleteRolesForHunt(huntID int, teams []*db.TeamDB) *response.Error {
	e := response.NewNilError()
	for _, t := range teams {
		teamErr := DeleteRolesForTeam(t.ID)
		if teamErr != nil {
			e.AddError(teamErr)
		}
	}

	regex := fmt.Sprintf("hunt_[a-zA-Z]+_%d", huntID)
	huntErr := db.DeleteRolesByRegex(regex)
	if huntErr != nil {
		e.AddError(huntErr)
	}

	return e.GetError()
}

// DeleteRolesForUser deletes all the roles and permissions associated with the given user
func DeleteRolesForUser(userID int) *response.Error {
	regex := fmt.Sprintf("user_[a-zA-Z]+_%d", userID)

	return db.DeleteRolesByRegex(regex)
}

// RoleDBs returns a slice of all the roles (in their RoleDB form)
// the given role is comprised of
func (r *Role) RoleDBs(userID int) []*db.RoleDB {
	roleDB := db.RoleDB{
		Name:        r.Name,
		UserID:      userID,
		EntityID:    r.EntityID,
		Permissions: make([]*db.PermissionDB, 0, len(r.Permissions)),
	}

	for _, p := range r.Permissions {
		roleDB.Permissions = append(roleDB.Permissions, &p.PermissionDB)
	}

	if r.Child == nil {
		return append(make([]*db.RoleDB, 0), &roleDB)
	}

	return append(r.Child.RoleDBs(userID), &roleDB)
}

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
	db.PermissionDB
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
	if strings.Contains(PermToRoleEndpoint[perm].FormattedRegex, "%") {
		regex = fmt.Sprintf(PermToRoleEndpoint[perm].FormattedRegex, entityID)
	} else {
		regex = PermToRoleEndpoint[perm].FormattedRegex
	}
	permission := Permission{
		db.PermissionDB{
			Method:   strings.Split(perm, "_")[0],
			URLRegex: regex,
		},
	}

	return &permission
}

type roleEndPoint struct {
	FormattedRegex string
	Route          string
	Role           string
}

// PermToRoleEndpoint maps permissions to FormattedRegex, Route, and Role
var PermToRoleEndpoint = map[string]roleEndPoint{
	// team endpoints
	"get_teams": roleEndPoint{
		FormattedRegex: `/teams/$`,
		Route:          `/teams/`,
		Role:           `admin`,
	},
	"get_team": roleEndPoint{
		FormattedRegex: `/teams/\d+$`,
		Route:          `/teams/%d`,
		Role:           `user`,
	},
	"get_points": roleEndPoint{
		FormattedRegex: `/teams/\d+/points/$`,
		Route:          `/teams/%d/points/`,
		Role:           `user`,
	},
	"get_players": roleEndPoint{
		FormattedRegex: `/teams/\d+/players/$`,
		Route:          `/teams/%d/players/`,
		Role:           `user`,
	},
	"post_player": roleEndPoint{
		FormattedRegex: `/teams/%d/players/$`,
		Route:          `/teams/%d/players/`,
		Role:           `team_editor`,
	},
	"delete_player": roleEndPoint{
		FormattedRegex: `/teams/%d/players/\d+$`,
		Route:          `/teams/%d/players/43`,
		Role:           `team_editor`,
	},
	"delete_team": roleEndPoint{
		FormattedRegex: `/teams/%d$`,
		Route:          `/teams/%d`,
		Role:           `team_owner`,
	},
	"post_team": roleEndPoint{
		FormattedRegex: `/teams/$`,
		Route:          `/teams/`,
		Role:           `hunt_editor`,
	},
	"patch_team": roleEndPoint{
		FormattedRegex: `/teams/%d$`,
		Route:          `/teams/%d`,
		Role:           `team_editor`,
	},
	"get_locations": roleEndPoint{
		FormattedRegex: `/teams/\d+/locations/$`,
		Route:          `/teams/%d/locations/`,
		Role:           `user`,
	},
	"post_location": roleEndPoint{
		FormattedRegex: `/teams/%d/locations/$`,
		Route:          `/teams/%d/locations/`,
		Role:           `team_member`,
	},
	"delete_location": roleEndPoint{
		FormattedRegex: `/teams/\d+/locations/\d+$`,
		Route:          `/teams/%d/locations/43`,
		Role:           `admin`,
	},
	"get_media": roleEndPoint{
		FormattedRegex: `/teams/\d+/media/$`,
		Route:          `/teams/%d/media/`,
		Role:           `user`,
	},
	"post_media": roleEndPoint{
		FormattedRegex: `/teams/%d/media/$`,
		Route:          `/teams/%d/media/`,
		Role:           `team_member`,
	},
	"delete_media": roleEndPoint{
		FormattedRegex: `/teams/%d/media/\d+$`,
		Route:          `/teams/%d/media/43`,
		Role:           `team_member`,
	},
	"post_teams_populate": roleEndPoint{
		FormattedRegex: `/teams/populate/$`,
		Route:          `/teams/populate/`,
		Role:           `admin`,
	},

	// user endpoints
	"get_user": roleEndPoint{
		FormattedRegex: `/users/\d+$`,
		Route:          `/users/%d`,
		Role:           `user`,
	},
	"post_login": roleEndPoint{
		FormattedRegex: `/users/login/$`,
		Route:          `/users/login/`,
		Role:           `user`,
	},
	"post_logout": roleEndPoint{
		FormattedRegex: `/users/logout/$`,
		Route:          `/users/logout/`,
		Role:           `user`,
	},
	"post_user": roleEndPoint{
		FormattedRegex: `/users/$`,
		Route:          `/users/`,
		Role:           `public`,
	},
	"delete_user": roleEndPoint{
		FormattedRegex: `/users/%d$`,
		Route:          `/users/%d`,
		Role:           `user_owner`,
	},
	"patch_user": roleEndPoint{
		FormattedRegex: `/users/%d$`,
		Route:          `/users/%d`,
		Role:           `user_owner`,
	},
	"delete_notification": roleEndPoint{
		FormattedRegex: `/users/%d/notifications/\d+$`,
		Route:          `/users/%d/notifications/43`,
		Role:           `user_owner`,
	},
	"get_notifications": roleEndPoint{
		FormattedRegex: `/users/%d/notifications/$`,
		Route:          `/users/%d/notifications/`,
		Role:           `user_owner`,
	},

	// hunt endpoints
	"get_hunts": roleEndPoint{
		FormattedRegex: `/hunts/$`,
		Route:          `/hunts/`,
		Role:           `user`,
	},
	"get_hunt": roleEndPoint{
		FormattedRegex: `/hunts/\d+$`,
		Route:          `/hunts/%d`,
		Role:           `user`,
	},
	"post_hunt": roleEndPoint{
		FormattedRegex: `/hunts/$`,
		Route:          `/hunts/`,
		Role:           `user`,
	},
	"delete_hunt": roleEndPoint{
		FormattedRegex: `/hunts/%d$`,
		Route:          `/hunts/%d`,
		Role:           `hunt_owner`,
	},
	"patch_hunt": roleEndPoint{
		FormattedRegex: `/hunts/%d$`,
		Route:          `/hunts/%d`,
		Role:           `hunt_editor`,
	},
	"post_hunts_populate": roleEndPoint{
		FormattedRegex: `/hunts/populate/$`,
		Route:          `/hunts/populate/`,
		Role:           `admin`,
	},
	"get_items": roleEndPoint{
		FormattedRegex: `/hunts/\d+/items/$`,
		Route:          `/hunts/%d/items/`,
		Role:           `user`,
	},
	"delete_item": roleEndPoint{
		FormattedRegex: `/hunts/%d/items/\d+$`,
		Route:          `/hunts/%d/items/43`,
		Role:           `hunt_editor`,
	},
	"post_item": roleEndPoint{
		FormattedRegex: `/hunts/%d/items/$`,
		Route:          `/hunts/%d/items/`,
		Role:           `hunt_editor`,
	},
	"patch_item": roleEndPoint{
		FormattedRegex: `/hunts/%d/items/\d+$`,
		Route:          `/hunts/%d/items/43`,
		Role:           `hunt_editor`,
	},
	"delete_invitation": roleEndPoint{
		FormattedRegex: `/hunts/%d/invitations/\d+$`,
		Route:          `/hunts/%d/invitations/43`,
		Role:           `hunt_editor`,
	},
	"post_invitation": roleEndPoint{
		FormattedRegex: `/hunts/%d/invitations/$`,
		Route:          `/hunts/%d/invitations/`,
		Role:           `hunt_member`,
	},
	"get_invitations": roleEndPoint{
		FormattedRegex: `/hunts/\d+/invitations/$`,
		Route:          `/hunts/%d/invitations/`,
		Role:           `user`,
	},
	"delete_hunt_players": roleEndPoint{
		FormattedRegex: `/hunts/%d/players/\d+$`,
		Route:          `/hunts/%d/players/43`,
		Role:           `hunt_member`,
	},
	"get_hunt_players": roleEndPoint{
		FormattedRegex: `/hunts/\d+/players/$`,
		Route:          `/hunts/%d/players/`,
		Role:           `user`,
	},
	"post_hunt_player": roleEndPoint{
		FormattedRegex: `/hunts/%d/players/$`,
		Route:          `/hunts/%d/players/`,
		Role:           `hunt_editor`,
	},
	"post_accept_hunt_invite": roleEndPoint{
		FormattedRegex: `/hunts/\d+/invitations/\d+/accept$`,
		Route:          `/hunts/43/invitations/43/accept`,
		Role:           `user`,
	},
	"post_decline_hunt_invite": roleEndPoint{
		FormattedRegex: `/hunts/\d+/invitations/\d+/decline$`,
		Route:          `/hunts/43/invitations/43/decline`,
		Role:           `user`,
	},
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
	"admin":       genAdminRole,
}

func getRoleName(role string, entityID int) string {
	if role == "user" {
		return "user"
	}

	if role == "admin" {
		return "admin"
	}
	return fmt.Sprintf("%s_%d", role, entityID)
}

func genRole(name string, id int) *Role {
	role := Role{
		Name:        getRoleName(name, id),
		Permissions: make([]*Permission, 0),
		EntityID:    id,
	}
	// create role specific permissions
	for k, v := range PermToRoleEndpoint {
		if v.Role == name {
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

func genAdminRole(id int) *Role {
	return genRole("admin", id)
}

// UserHasRole returns whether or not the given user has a particular role
func UserHasRole(role string, entityID int, userID int) (bool, *response.Error) {
	userRoles, e := db.RolesForUser(userID)
	if e != nil {
		return false, e
	}

	roleName := getRoleName(role, entityID)
	for _, r := range userRoles {
		if r.Name == roleName {
			return true, nil
		}
	}

	return false, nil
}
