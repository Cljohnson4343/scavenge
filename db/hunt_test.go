package db

import (
	"encoding/json"
	"testing"
	"time"
)

func TestHuntDB(t *testing.T) {
	jsonReq := []byte(`
	{
		"name": "Huntsville Hunt",
		"max_teams": 4,
		"id": 1,
		"start": "2019-04-21T10:00:00Z",
		"end": "2019-04-21T13:00:00Z",
		"teams": [
			{
				"name": "Home Team"
			},
			{
				"name": "Second Team"
			}
		],
		"items": [
			{
				"name": "Santa Clause",
				"points": 10,
				"is_done": false
			},
			{
				"name": "Rudolph",
				"points": 10,
				"is_done": true
			},
			{
				"name": "Reindeer on the Roof",
				"points": 40,
				"is_done": false
			},
			{
				"name": "Mrs. Clause",
				"points": 30,
				"is_done": true
        },
        {
            "name": "Snow Globe",
            "points": 40,
            "is_done": true
        }
		],
		"location_name": "Huntsville, AL",
		"latitude": 34.730705,
		"longitude": -86.59481
	}`)

	hunt := HuntDB{}
	err := json.Unmarshal(jsonReq, &hunt)
	if err != nil {
		t.Fatalf("error unmarshaling json: %s\n", err.Error())
	}

	h := HuntDB{ID: 43, Name: "Chris Johnson", MaxTeams: 43, Latitude: 43.43, Longitude: 4343.4343, LocationName: "HuntsVegas", StartTime: time.Now(), EndTime: time.Now(), CreatedAt: time.Now()}
	hJSON, err := json.Marshal(&h)
	if err != nil {
		t.Fatalf("error marshaling: %s\n", err.Error())
	}

	huntJSON, err := json.Marshal(&hunt)
	if err != nil {
		t.Fatalf("error marshaling: %s\n", err.Error())
	}

	t.Logf("the resulting json for hunt: %s\n\nh: %s\n", string(huntJSON), string(hJSON))
}
