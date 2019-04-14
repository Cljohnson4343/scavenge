package models

// A Team is a representation of a Team
//
// swagger:model Team
type Team struct {

	// the name of the team
	//
	// maximum length: 255
	// required: true
	Name string `json:"name"`
}
