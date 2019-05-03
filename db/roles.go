package db

import (
	"net/http"
	"time"

	"github.com/cljohnson4343/scavenge/response"
)

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
	CreatedAt   time.Time       `json:"created_at"`
	Permissions []*PermissionDB `json:"permissions"`
}

var roleInsertScript = `
	WITH ins_role AS (
		INSERT INTO roles(name)
		VALUES ($1)
		RETURNING id, created_at
	)
	INSERT INTO users_roles(user_id, role_id)
	VALUES ($2, (SELECT id FROM ins_role))
	RETURNING (SELECT id FROM ins_role), (SELECT created_at FROM ins_role);
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
		err := roleInsStmt.QueryRow(r.Name, r.UserID).Scan(&r.CreatedAt, &r.ID)
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
