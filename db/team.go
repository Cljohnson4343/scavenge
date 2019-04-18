package db

// A TeamDB is a representation of a row in the teams table
//
// swagger:model TeamDB
type TeamDB struct {

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
