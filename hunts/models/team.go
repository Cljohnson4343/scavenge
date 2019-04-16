package models

// A Team is a representation of a Team
//
// swagger:model Team
type Team struct {

	// The id of the Hunt
	//
	// required: true
	HuntID int `json:"hunt_id"`

	// The id of the team
	//
	// required: true
	ID int `json:"id"`

	// the name of the team
	//
	// maximum length: 255
	// required: true
	Name string `json:"name"`
}
