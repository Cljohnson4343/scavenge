package users

import (
	"fmt"

	"github.com/cljohnson4343/scavenge/db"
)

type linker interface {
	Link()
}

// Notification is a wrapper for HuntInvitations that adds the links for
// manipulating the hunt invitation
type Notification struct {
	db.HuntInvitationDB

	Links linker `json:"links"`
}

// TODO this smells: the hunt invitation notification is using hard coded
// url addresses that it should not know about. Get rid of this
// obscure dependency

// GetNotification returns a notification from the given HuntInvitation
func GetNotification(invite *db.HuntInvitationDB) *Notification {
	return &Notification{
		HuntInvitationDB: *invite,
		Links: &huntInvitationLink{
			Accept: accept{
				Path: fmt.Sprintf(
					"/hunts/%d/invitations/%d/accept",
					invite.HuntID,
					invite.ID,
				),
				Method: "POST",
			},
			Decline: decline{
				Path: fmt.Sprintf(
					"/hunts/%d/invitations/%d/decline",
					invite.HuntID,
					invite.ID,
				),
				Method: "POST",
			},
			Delete: delete{
				Path: fmt.Sprintf(
					"/hunts/%d/invitations/%d/decline",
					invite.HuntID,
					invite.ID,
				),
				Method: "POST",
			},
		},
	}
}

type huntInvitationLink struct {
	Accept  accept  `json:"accept"`
	Decline decline `json:"decline"`
	Delete  delete  `json:"delete"`
}

func (h *huntInvitationLink) Link() {
	return
}

type accept struct {
	Path   string `json:"path"`
	Method string `json:"method"`
	Data   string `json:"data"`
}

type decline struct {
	Path   string `json:"path"`
	Method string `json:"method"`
}

type delete struct {
	Path   string `json:"path"`
	Method string `json:"method"`
}
