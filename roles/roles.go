package roles

import (
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"github.com/cljohnson4343/scavenge/db"
	"github.com/cljohnson4343/scavenge/response"
	"github.com/cljohnson4343/scavenge/users"
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
func DeleteRolesForHunt(huntID int) *response.Error {
	e := response.NewNilError()

	teams, getErr := db.GetTeamsWithHuntID(huntID)
	if getErr != nil {
		e.AddError(getErr)
	}

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

// RoleDBs returns a slice of all the roles (in their RoleDB form)
// the given role is comprised of
func (r *Role) RoleDBs(userID int) []*db.RoleDB {
	roleDB := db.RoleDB{
		Name:        r.Name,
		UserID:      userID,
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

// RequireAuth checks to make sure the requesting user agent has
// authorization to make the request
func RequireAuth(fn http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		userID, e := users.GetUserID(req.Context())
		if e != nil {
			e.Handle(w)
			return
		}

		perms, e := db.PermissionsForUser(userID)
		if e != nil {
			e.Handle(w)
			return
		}

		for _, p := range perms {
			perm := Permission{PermissionDB: p}
			if perm.Authorized(req) {
				return
			}
		}

		e = response.NewErrorf(
			http.StatusUnauthorized,
			"User %d is not authorized to access %s %s",
			userID,
			req.Method,
			req.URL.Path,
		)
		e.Handle(w)
	})
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
	if strings.Contains(permToFormattedRegex[perm], "%") {
		regex = fmt.Sprintf(permToFormattedRegex[perm], entityID)
	} else {
		regex = permToFormattedRegex[perm]
	}
	permission := Permission{
		db.PermissionDB{
			Method:   strings.Split(perm, "_")[0],
			URLRegex: regex,
		},
	}

	return &permission
}

// TODO clean this up so that there is only one map from perm to a struct
// containing the other info
var permToFormattedRegex = map[string]string{
	// team endpoints
	"get_teams":           `/teams/$`,
	"get_team":            `/teams/%d$`,
	"get_points":          `/teams/%d/points/$`,
	"get_players":         `/teams/%d/players/$`,
	"post_player":         `/teams/%d/players/$`,
	"delete_player":       `/teams/%d/players/\d+$`,
	"delete_team":         `/teams/%d$`,
	"post_team":           `/teams/$`,
	"patch_team":          `/teams/%d$`,
	"get_locations":       `/teams/%d/locations/$`,
	"post_location":       `/teams/%d/locations/$`,
	"delete_location":     `/teams/%d/locations/\d+$`,
	"get_media":           `/teams/%d/media/$`,
	"post_media":          `/teams/%d/media/$`,
	"delete_media":        `/teams/%d/media/\d+$`,
	"post_teams_populate": `/teams/populate/$`,

	// user endpoints
	"get_user":    `/users/%d$`,
	"post_login":  `/users/login/$`,
	"post_logout": `/users/logout/$`,
	"post_user":   `/users/$`,
	"delete_user": `/users/%d$`,
	"patch_user":  `/users/%d$`,

	// hunt endpoints
	"get_hunts":           `/hunts/$`,
	"get_hunt":            `/hunts/%d$`,
	"post_hunt":           `/hunts/$`,
	"delete_hunt":         `/hunts/%d$`,
	"patch_hunt":          `/hunts/%d$`,
	"post_hunts_populate": `/hunts/populate/$`,
	"get_items":           `/hunts/%d/items/$`,
	"delete_item":         `/hunts/%d/items/\d+$`,
	"post_item":           `/hunts/%d/items/$`,
	"patch_item":          `/hunts/%d/items/\d+$`,
}

// PermToRole maps a permissions key to a role
var PermToRole = map[string]string{
	// team endpoints
	"get_teams":           `admin`,
	"get_team":            `user`,
	"get_points":          `hunt_member`,
	"get_players":         `user`,
	"post_player":         `team_editor`,
	"delete_player":       `team_editor`,
	"delete_team":         `team_owner`,
	"post_team":           `hunt_editor`,
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

// PermToRoutes maps permission to the endpoint route
var PermToRoutes = map[string]string{
	// teams routes
	"get_teams":           `/teams/`,
	"get_team":            `/teams/%d`,
	"get_points":          `/teams/%d/points/`,
	"get_players":         `/teams/%d/players/`,
	"post_player":         `/teams/%d/players/`,
	"delete_player":       `/teams/%d/players/43`,
	"delete_team":         `/teams/%d`,
	"post_team":           `/teams/`,
	"patch_team":          `/teams/%d`,
	"get_locations":       `/teams/%d/locations/`,
	"post_location":       `/teams/%d/locations/`,
	"delete_location":     `/teams/%d/locations/43`,
	"get_media":           `/teams/%d/media/`,
	"post_media":          `/teams/%d/media/`,
	"delete_media":        `/teams/%d/media/43`,
	"post_teams_populate": `/teams/populate/`,

	// hunts routes
	"get_hunts":           `/hunts/`,
	"get_hunt":            `/hunts/%d`,
	"post_hunt":           `/hunts/`,
	"delete_hunt":         `/hunts/%d`,
	"patch_hunt":          `/hunts/%d`,
	"post_hunts_populate": `/hunts/populate/`,
	"get_items":           `/hunts/%d/items/`,
	"delete_item":         `/hunts/%d/items/43`,
	"post_item":           `/hunts/%d/items/`,
	"patch_item":          `/hunts/%d/items/43`,

	// users routes
	"get_user":    `/users/%d`,
	"post_login":  `/users/login/`,
	"post_logout": `/users/logout/`,
	"post_user":   `/users/`,
	"delete_user": `/users/%d`,
	"patch_user":  `/users/%d`,
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
