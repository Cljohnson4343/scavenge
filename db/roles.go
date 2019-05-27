package db

import (
	"net/http"

	"github.com/cljohnson4343/scavenge/response"
)

// TODO rework roles/permissions so that there isn't so much redundant
// data storage. Each role, team_owner_*, only differs by entity id
// but a whole new set of perms are stored with only the entity id
// varying.

// PermissionDB is a representation of a permissions table row
type PermissionDB struct {
	ID       int    `json:"permissionID"`
	URLRegex string `json:"urlRegex"`
	Method   string `json:"method"`
}

// RoleDB is a representation of a roles table row
type RoleDB struct {
	ID          int             `json:"roleID"`
	EntityID    int             `json:"entityID"`
	Name        string          `json:"roleName"`
	UserID      int             `json:"userID"`
	Permissions []*PermissionDB `json:"permissions"`
}

var roleInsertScript = `
	SELECT ins_sel_role($1, $2, $3);
`

var permissionInsertScript = `
	SELECT ins_sel_perm($1, $2, $3);
	`

// AddRoles stores the given roles in the db
func AddRoles(roles []*RoleDB) *response.Error {

	// TODO look into using a sql stored procedure to handle the whole
	// role insertion instead of breaking it up in a transaction.

	tx, err := db.Begin()
	if err != nil {
		return response.NewErrorf(
			http.StatusInternalServerError,
			"error beginning a transaction: %v",
			err,
		)
	}

	roleInsStmt := tx.Stmt(stmtMap["roleInsert"])
	permInsStmt := tx.Stmt(stmtMap["permissionInsert"])

	for _, r := range roles {
		err := roleInsStmt.QueryRow(r.Name, r.UserID, r.EntityID).Scan(&r.ID)
		if err != nil {
			if rollbackErr := tx.Rollback(); rollbackErr != nil {
				return response.NewErrorf(
					http.StatusInternalServerError,
					"error rolling back tx: %v",
					err,
				)
			}
			return response.NewErrorf(
				http.StatusInternalServerError,
				"error inserting role %s: %v:",
				r.Name,
				err,
			)
		}

		for _, p := range r.Permissions {
			err := permInsStmt.QueryRow(r.ID, p.URLRegex, p.Method).Scan(&p.ID)
			if err != nil {
				if rollbackErr := tx.Rollback(); rollbackErr != nil {
					return response.NewErrorf(
						http.StatusInternalServerError,
						"error rolling back tx: %v",
						err,
					)
				}
				return response.NewErrorf(
					http.StatusInternalServerError,
					"error inserting permission w/ regex %s for role %s: %v:",
					p.URLRegex,
					r.Name,
					err,
				)
			}
		}
	}

	if err = tx.Commit(); err != nil {
		return response.NewErrorf(
			http.StatusInternalServerError,
			"error committing transaction for inserting roles: %v",
			err,
		)
	}

	return nil
}

var rolesForUserScript = `
	WITH roles_for_user AS (
		SELECT id, name 
		FROM users_roles ur 
		INNER JOIN roles r ON ur.user_id = $1 AND r.id = ur.role_id
		)
	SELECT r.id, r.name, COALESCE(p.id, 0), COALESCE(p.url_regex, ''), COALESCE(p.method, '')
	FROM roles_for_user r 
	LEFT OUTER JOIN permissions p ON r.id = p.role_id
	ORDER BY r.id;
	`

// RolesForUser returns the roles for a user
func RolesForUser(userID int) ([]*RoleDB, *response.Error) {
	rows, err := stmtMap["rolesForUser"].Query(userID)
	if err != nil {
		return nil, response.NewErrorf(
			http.StatusInternalServerError,
			"error getting roles for user %d: %v",
			userID,
			err,
		)
	}
	defer rows.Close()

	// TODO look back over this and think about benchmarking/optimizing
	// I think alot of efficiency can be gained by reducing the dynamic allocations

	roles := make([]*RoleDB, 0)
	e := response.NewNilError()
	role := new(RoleDB)
	prevRoleID := 0
	for rows.Next() {
		var id int
		var name string
		p := PermissionDB{}
		err = rows.Scan(&id, &name, &p.ID, &p.URLRegex, &p.Method)
		if err != nil {
			e.Addf(http.StatusInternalServerError, "error getting row: %v", err)
		}

		if id != prevRoleID && prevRoleID != 0 {
			role.UserID = userID
			roles = append(roles, role)
			role = new(RoleDB)
		}

		role.ID = id
		role.Name = name
		role.Permissions = append(role.Permissions, &p)
		prevRoleID = id
	}

	if err = rows.Err(); err != nil {
		e.Addf(http.StatusInternalServerError, "error getting role: %v", err)
	}

	if prevRoleID != 0 {
		roles = append(roles, role)
	}

	return roles, e.GetError()
}

var permissionsForUserScript = `
	SELECT p.url_regex, p.method, p.id
	FROM users_roles ur 
	INNER JOIN permissions p ON ur.user_id = $1 AND ur.role_id = p.role_id; 
	`

// PermissionsForUser returns an array of user permissions
func PermissionsForUser(userID int) ([]PermissionDB, *response.Error) {
	rows, err := stmtMap["permissionsForUser"].Query(userID)
	if err != nil {
		return nil, response.NewErrorf(
			http.StatusInternalServerError,
			"error getting permissions for user %d: %v",
			userID,
			err,
		)
	}
	defer rows.Close()

	perms := make([]PermissionDB, 0, 16)
	e := response.NewNilError()

	for rows.Next() {
		p := PermissionDB{}
		err = rows.Scan(&p.URLRegex, &p.Method, &p.ID)
		if err != nil {
			e.Addf(
				http.StatusInternalServerError,
				"error getting row for user %d: %v",
				userID,
				err,
			)
		}

		perms = append(perms, p)
	}

	if err = rows.Err(); err != nil {
		e.Addf(
			http.StatusInternalServerError,
			"error getting row for user %d: %v",
			userID,
			err,
		)
	}

	return perms, e.GetError()
}

var roleRemoveScript = `
	DELETE FROM users_roles
	WHERE user_id = $1 AND role_id = $2; 
	`

// RemoveRole removes the given role from the given user
// without recursively removing children roles
func RemoveRole(roleID, userID int) *response.Error {
	res, err := stmtMap["roleRemove"].Exec(userID, roleID)
	if err != nil {
		return response.NewErrorf(
			http.StatusInternalServerError,
			"error removing role %d from user %d: %v",
			roleID,
			userID,
			err,
		)
	}

	numRow, err := res.RowsAffected()
	if err != nil {
		return response.NewErrorf(
			http.StatusInternalServerError,
			"error removing role %d from user %d: %v",
			roleID,
			userID,
			err,
		)
	}

	if numRow < 1 {
		return response.NewErrorf(
			http.StatusInternalServerError,
			"user %d does not have a role %d",
			userID,
			roleID,
		)
	}

	return nil
}

var rolesDeleteByRegexScript = `
	DELETE FROM roles
	WHERE name ~ $1;
`

// DeleteRolesByRegex Deletes all roles whose names match the given regex
func DeleteRolesByRegex(regex string) *response.Error {
	res, err := stmtMap["rolesDeleteByRegex"].Exec(regex)
	if err != nil {
		return response.NewErrorf(
			http.StatusInternalServerError,
			"error deleting roles whose names match %s: %v",
			regex,
			err,
		)
	}

	_, err = res.RowsAffected()
	if err != nil {
		return response.NewErrorf(
			http.StatusInternalServerError,
			"error deleting roles whose names match %s: %v",
			regex,
			err,
		)
	}

	return nil
}
