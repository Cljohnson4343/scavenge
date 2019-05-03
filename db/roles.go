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
	ID       int    `json:"id"`
	URLRegex string `json:"url_regex"`
	Method   string `json:"method"`
}

// RoleDB is a representation of a roles table row
type RoleDB struct {
	ID          int             `json:"id"`
	Name        string          `json:"name"`
	UserID      int             `json:"user_id"`
	Permissions []*PermissionDB `json:"permissions"`
}

var roleInsertScript = `
	SELECT ins_sel_role($1, $2);
`

var permissionInsertScript = `
	INSERT INTO permissions(url_regex, method, role_id)
	VALUES ($1, $2, $3)
	RETURNING id;
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
		err := roleInsStmt.QueryRow(r.Name, r.UserID).Scan(&r.ID)
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
			err := permInsStmt.QueryRow(p.URLRegex, p.Method, r.ID).Scan(&p.ID)
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
	SELECT r.id, r.name, p.id, p.url_regex, p.method
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

	roles := make([]*RoleDB, 0)
	e := response.NewNilError()
	role := RoleDB{}
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
			roles = append(roles, &role)
			role = RoleDB{}
		}

		role.ID = id
		role.Name = name
		role.Permissions = append(role.Permissions, &p)
		prevRoleID = id
	}

	if err = rows.Err(); err != nil {
		e.Addf(http.StatusInternalServerError, "error getting role: %v", err)
	}

	roles = append(roles, &role)

	return roles, e.GetError()
}
