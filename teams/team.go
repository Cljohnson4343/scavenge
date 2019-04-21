package teams

import "github.com/cljohnson4343/scavenge/db"

// A Team is a representation of a Team
//
// swagger:model Team
type Team struct {
	db.TeamDB `valid:"-"`
}

// A PartialTeam is a Team that can be paritally validated
type PartialTeam struct {
	db.PartialTeamDB `valid:"-"`
}
