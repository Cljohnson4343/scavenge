package db

import (
	"net/http"
	"time"

	"github.com/asaskevich/govalidator"
	"github.com/cljohnson4343/scavenge/response"
	"github.com/lib/pq"
)

// HuntInvitationDB is info associated with a notification
type HuntInvitationDB struct {

	// The id of the invitation
	//
	// required: false
	ID int `json:"huntInvitationID" valid:"int,optional"`

	// The email address the invitation is/was sent to
	//
	// required: true
	Email string `json:"email" valid:"email"`

	// The id of the hunt
	//
	// required: true
	HuntID int `json:"huntID" valid:"int"`

	// The id of the user that sent the invite
	//
	// required: true
	InviterID int `json:"inviterID" valid:"int"`

	// The time the invitation was sent
	//
	// required: true
	// swagger:strfmt date
	InvitedAt time.Time `json:"invitedAt" valid:"-"`
}

// Validate validates the struct
func (i *HuntInvitationDB) Validate(r *http.Request) *response.Error {
	_, err := govalidator.ValidateStruct(i)
	if err != nil {
		return response.NewErrorf(
			http.StatusBadRequest,
			"error validating hunt invitation: %v",
			err,
		)
	}

	return nil
}

var huntInvitationsByUserIDScript = `
	SELECT hi.id, hi.email, hi.hunt_id, hi.inviter_id, hi.invited_at 
	FROM hunt_invitations hi INNER JOIN users u
	ON u.id = $1 AND u.email = hi.email;
	`

// GetHuntInvitationsByUserID returns all hunt invitations for the user with the
// given id. A result with both media meta objects and an error is possible
func GetHuntInvitationsByUserID(userID int) ([]*HuntInvitationDB, *response.Error) {
	rows, err := stmtMap["huntInvitationByUserID"].Query(userID)
	if err != nil {
		return nil, response.NewErrorf(
			http.StatusInternalServerError,
			"error getting all hunt invitations for user %d: %v",
			userID,
			err,
		)
	}
	defer rows.Close()

	e := response.NewNilError()
	invitations := make([]*HuntInvitationDB, 0)

	for rows.Next() {
		i := HuntInvitationDB{}

		err = rows.Scan(
			&i.ID,
			&i.Email,
			&i.HuntID,
			&i.InviterID,
			&i.InvitedAt,
		)
		if err != nil {
			e.Addf(
				http.StatusInternalServerError,
				"error getting hunt invitation for user %d: %v",
				userID,
				err,
			)
			break
		}
		invitations = append(invitations, &i)
	}

	if err = rows.Err(); err != nil {
		e.Addf(
			http.StatusInternalServerError,
			"error getting hunt invitation for user %d: %v",
			userID,
			err,
		)
	}

	return invitations, e.GetError()
}

var huntInvitationInsertScript = `
	INSERT INTO hunt_invitations(email, hunt_id, inviter_id)
	VALUES ($1, $2, $3)
	RETURNING id, invited_at;
	`

// Insert inserts the given data for a hunt invitation into the db. The
// id and the invitedAt timestamp are written back to the HuntInvitationDB
// struct
func (i *HuntInvitationDB) Insert() *response.Error {
	err := stmtMap["huntInvitationInsert"].QueryRow(
		i.Email,
		i.HuntID,
		i.InviterID,
	).Scan(&i.ID, &i.InvitedAt)

	if err != nil {
		return i.ParseError(err, "insert")
	}

	return nil
}

var huntInvitationDeleteScript = `
	WITH email_for_user AS (
		SELECT email
		FROM users u 
		WHERE u.id = $2
	)
	DELETE FROM hunt_invitations hi
	USING email_for_user efu
	WHERE hi.id = $1 AND efu.email = hi.email;`

// DeleteHuntInvitation deletes the row from the hunt_invitations table
func DeleteHuntInvitation(huntInvitationID, userID int) *response.Error {
	res, err := stmtMap["huntInvitationDelete"].Exec(huntInvitationID, userID)
	if err != nil {
		return response.NewErrorf(
			http.StatusInternalServerError,
			"error deleting hunt invitation %d: %v",
			huntInvitationID,
			err,
		)
	}

	numRows, err := res.RowsAffected()
	if err != nil {
		return response.NewErrorf(
			http.StatusInternalServerError,
			"error deleting hunt invitation %d: %v",
			huntInvitationID,
			err,
		)
	}

	if numRows < 1 {
		return response.NewErrorf(
			http.StatusBadRequest,
			"error deleting hunt invitation with id %d: no invitation with that id for user %d exists",
			huntInvitationID,
			userID,
		)
	}

	return nil
}

// ParseError maps a pq driver error to a response.Error
func (i *HuntInvitationDB) ParseError(err error, op string) *response.Error {
	pqErr, ok := err.(*pq.Error)
	if ok {
		if pqErr.Constraint != "" {
			switch pqErr.Constraint {
			case "hunt_invitations_hunt_id_fkey":
				return response.NewErrorf(
					http.StatusBadRequest,
					"huntID: hunt %d does not exist",
					i.HuntID,
				)
			case "media_location_id_fkey":
				return response.NewErrorf(
					http.StatusBadRequest,
					"inviterID: user %d does not exist",
					i.InviterID,
				)
			}
		}
	}

	return response.NewErrorf(
		http.StatusInternalServerError,
		"HuntInvitation: db error on hunt_invitations operation: %s",
		err.Error(),
	)
}