package db

import (
	"database/sql"
	"net/http"
	"time"

	"github.com/asaskevich/govalidator"
	"github.com/cljohnson4343/scavenge/pgsql"
	"github.com/cljohnson4343/scavenge/request"
	"github.com/cljohnson4343/scavenge/response"
	"github.com/lib/pq"
)

var userTbl = "users"

// UserDB is a representation of a user stored in the DB
type UserDB struct {
	// ID is the user id
	//
	// required: true
	ID int `json:"userID" valid:"int,optional"`

	// FirstName is the user's first name
	//
	// required: true
	// maximum length: 64
	// minimum length: 1
	FirstName string `json:"firstName" valid:"stringlength(1|64)"`

	// LastName is the user's last name
	//
	// required: true
	// maximum length: 64
	// minimum length: 1
	LastName string `json:"lastName" valid:"stringlength(1|64)"`

	// Username is the user's unique username
	//
	// required: true
	// maximum length: 64
	// minimum length: 1
	Username string `json:"username" valid:"stringlength(1|64)"`

	// JoinedAt is the time stamp for the user's join date
	//
	// required: true
	JoinedAt time.Time `json:"joinedAt" valid:"isZeroTime~joined_at: should not be included,optional"`

	// LastVisit is the time of the user's last visit
	//
	// required: true
	LastVisit time.Time `json:"lastVisit" valid:"isZeroTime~last_visit: should not be included,optional"`

	// ImageURL is the url of the user's profile pic
	//
	// required: false
	ImageURL string `json:"imageUrl" valid:"url,optional"`

	// Email is the email for the user
	//
	// required: true
	Email string `json:"email" valid:"email"`
}

// Validate validates a userDB
func (u *UserDB) Validate(r *http.Request) *response.Error {
	_, err := govalidator.ValidateStruct(u)
	if err != nil {
		return response.NewErrorf(http.StatusBadRequest, "error validating user %s: %v", u.Username, err)
	}
	return nil
}

// PatchValidate validates only the non-zero valued fields for a UserDB
func (u *UserDB) PatchValidate(r *http.Request, entityID int) *response.Error {
	tblColMap := u.GetTableColumnMap()
	e := response.NewNilError()

	id, ok := tblColMap[userTbl]["id"]
	if !ok {
		tblColMap[userTbl]["id"] = entityID
		u.ID = entityID
		id = entityID
	}

	if id != entityID {
		e.Add(http.StatusBadRequest, "patch id: check to make sure the given id matches the URL id")
		delete(tblColMap[userTbl], "id")
	}

	if _, ok := tblColMap[userTbl]["joined_at"]; ok {
		e.Add(http.StatusBadRequest, "patch joined_at: patch does not support changing joined_at field")
		delete(tblColMap[userTbl], "joined_at")
	}

	if _, ok := tblColMap[userTbl]["last_visit"]; ok {
		e.Add(http.StatusBadRequest, "patch last_visit: patch does not support changing last_visit field")
		delete(tblColMap[userTbl], "last_visit")
	}

	userErr := request.PatchValidate(tblColMap[userTbl], u)
	if userErr != nil {
		e.AddError(userErr)
	}

	return e.GetError()
}

// GetTableColumnMap returns the non-zero valued fields for the given UserDB
func (u *UserDB) GetTableColumnMap() pgsql.TableColumnMap {
	z := UserDB{}

	tblColMap := make(pgsql.TableColumnMap)
	tblColMap[userTbl] = make(pgsql.ColumnMap)

	if z.ID != u.ID {
		tblColMap[userTbl]["id"] = u.ID
	}

	if z.Email != u.Email {
		tblColMap[userTbl]["email"] = u.Email
	}

	if z.FirstName != u.FirstName {
		tblColMap[userTbl]["first_name"] = u.FirstName
	}

	if z.LastName != u.LastName {
		tblColMap[userTbl]["last_name"] = u.LastName
	}

	if z.Username != u.Username {
		tblColMap[userTbl]["username"] = u.Username
	}

	if !u.JoinedAt.IsZero() {
		tblColMap[userTbl]["joined_at"] = u.JoinedAt
	}

	if !u.LastVisit.IsZero() {
		tblColMap[userTbl]["last_visit"] = u.LastVisit
	}

	if z.ImageURL != u.ImageURL {
		tblColMap[userTbl]["image_url"] = u.ImageURL
	}

	return tblColMap
}

var userInsertScript = `
	INSERT INTO users(first_name, last_name, username, image_url, email)
	VALUES ($1, $2, $3, NULLIF($4, ''), $5)
	RETURNING id, joined_at, last_visit;
	`

// Insert inserts the given userDB. The ID, JoinedAt, and LastVisit fields are written back
// to the given userDB
func (u *UserDB) Insert() *response.Error {
	err := stmtMap["userInsert"].QueryRow(
		u.FirstName,
		u.LastName,
		u.Username,
		u.ImageURL,
		u.Email).Scan(&u.ID, &u.JoinedAt, &u.LastVisit)
	if err != nil {
		return u.ParseError(err, "insert")
	}

	return nil
}

var userGetScript = `
	SELECT 
		id, 
		first_name, 
		last_name, 
		username, 
		joined_at, 
		last_visit, 
		COALESCE(image_url, ''), 
		email
	FROM users
	WHERE id = $1;`

// GetUser returns the user with the given id
func GetUser(userID int) (*UserDB, *response.Error) {
	u := UserDB{}
	err := stmtMap["userGet"].QueryRow(userID).Scan(&u.ID, &u.FirstName, &u.LastName,
		&u.Username, &u.JoinedAt, &u.LastVisit, &u.ImageURL, &u.Email)
	if err != nil {
		// check to see if user doesn't exist
		if err == sql.ErrNoRows {
			return nil, response.NewErrorf(http.StatusBadRequest, "Get user: there is no user with id %d", userID)
		}

		return nil, response.NewErrorf(http.StatusInternalServerError, "Get user: %v", err)
	}

	return &u, nil
}

var userDeleteScript = `
	DELETE FROM users
	WHERE id = $1;`

// DeleteUser deletes the user with the given id
func DeleteUser(userID int) *response.Error {
	res, err := stmtMap["userDelete"].Exec(userID)
	if err != nil {
		return response.NewErrorf(http.StatusInternalServerError, "delete user: %v", err)
	}

	numRows, err := res.RowsAffected()
	if err != nil {
		return response.NewErrorf(http.StatusInternalServerError, "delete user: %v", err)
	}

	if numRows < 1 {
		return response.NewErrorf(http.StatusBadRequest,
			"delete user: there was no user with the id %d", userID)
	}

	return nil
}

// Update updates the db with the given UserDB
func (u *UserDB) Update(ex pgsql.Executioner, userID int) *response.Error {
	return update(u, ex, userID)
}

// ParseError maps a pq error to a response.Error with the information that the client
// needs to know.
func (u *UserDB) ParseError(err error, op string) *response.Error {
	pqErr, ok := err.(*pq.Error)
	if ok {
		if pqErr.Constraint != "" {
			switch pqErr.Constraint {
			case "users_unique_lower_email_idx":
				return response.NewErrorf(
					http.StatusBadRequest,
					"email: %s is not a valid email",
					u.Email,
				)
			case "users_unique_username_idx":
				return response.NewErrorf(
					http.StatusBadRequest,
					"username: %s is not a valid username",
					u.Username,
				)
			}
		}
	}

	return response.NewErrorf(
		http.StatusInternalServerError,
		"error performing operation %s: %s",
		op,
		err.Error(),
	)
}
