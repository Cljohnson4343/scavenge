package db

import (
	"net/http"
	"time"

	"github.com/asaskevich/govalidator"

	"github.com/cljohnson4343/scavenge/response"
)

// UserDB is a representation of a user stored in the DB
type UserDB struct {
	// ID is the user id
	//
	// required: true
	ID int `json:"id" valid:"int,optional"`

	// FirstName is the user's first name
	//
	// required: true
	// maximum length: 64
	// minimum length: 1
	FirstName string `json:"first_name" valid:"stringlength(1|64)"`

	// LastName is the user's last name
	//
	// required: true
	// maximum length: 64
	// minimum length: 1
	LastName string `json:"last_name" valid:"stringlength(1|64)"`

	// Username is the user's unique username
	//
	// required: true
	// maximum length: 64
	// minimum length: 1
	Username string `json:"username" valid:"stringlength(1|64)"`

	// JoinedAt is the time stamp for the user's join date
	//
	// required: true
	JoinedAt time.Time `json:"joined_at" valid:"isZeroTime~joined_at: should not be included,optional"`

	// LastVisit is the time of the user's last visit
	//
	// required: true
	LastVisit time.Time `json:"last_visit" valid:"isZeroTime~last_visit: should not be included,optional"`

	// ImageURL is the url of the user's profile pic
	//
	// required: false
	ImageURL string `json:"image_url" valid:"url,optional"`

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

var userInsertScript = `
	INSERT INTO users(first_name, last_name, username, image_url, email)
	VALUES ($1, $2, $3, NULLIF($4, ''), $5)
	RETURNING id, joined_at, last_visit;
	`

// Insert inserts the given userDB. The ID, JoinedAt, and LastVisit fields are written back
// to the given userDB
func (u *UserDB) Insert() *response.Error {
	err := stmtMap["userInsert"].QueryRow(u.FirstName, u.LastName, u.Username, u.ImageURL, u.Email).Scan(&u.ID, &u.JoinedAt, &u.LastVisit)
	if err != nil {
		return response.NewErrorf(http.StatusInternalServerError, "error inserting user %s: %v", u.Username, err)
	}

	return nil
}
