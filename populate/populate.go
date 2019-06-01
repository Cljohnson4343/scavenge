package populate

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/cljohnson4343/scavenge/db"

	h "github.com/cljohnson4343/scavenge/hunts"
	"github.com/cljohnson4343/scavenge/response"
	u "github.com/cljohnson4343/scavenge/users"
)

func readResource(resource string, data interface{}) {
	file, err := os.Open(fmt.Sprintf("./populate/%s.json", resource))
	if err != nil {
		panic(fmt.Sprintf("error opening %s.json file: %v", resource, err))
	}
	defer file.Close()

	err = json.NewDecoder(file).Decode(&data)
	if err != nil {
		panic(fmt.Sprintf("error decoding %s.json: %v", resource, err))
	}
}

func addPlayers(hunts []*h.Hunt, users []*u.User) {
	players := make([]*db.PlayerDB, 0)
	for _, p := range users {
		players = append(players, &db.PlayerDB{UserDB: (*p).UserDB})
	}

	for i, v := range hunts {
		v.Players = make([]*db.PlayerDB, len(users), len(users))
		copy(v.Players, players)
		v.CreatorID = i % len(users)
		v.CreatorUsername = users[v.CreatorID].Username
	}
}

// Populate fills the database with dummy test data
func Populate(populateFlag bool) {
	if !populateFlag {
		return
	}

	users := make([]*u.User, 0)
	readResource("users", &users)
	e := response.NewNilError()
	for _, v := range users {
		userErr := u.InsertUser(v)
		if userErr != nil {
			e.AddError(userErr)
		}
	}

	hunts := make([]*h.Hunt, 0)
	readResource("hunts", &hunts)
	addPlayers(hunts, users)
	for _, v := range hunts {
		huntErr := h.InsertHunt(v.CreatorID, v)
		if huntErr != nil {
			e.AddError(huntErr)
		}
	}

	if error := e.GetError(); error != nil {
		log.Printf("error populating db: %s", error.JSON())
	}
}
