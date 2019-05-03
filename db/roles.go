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
