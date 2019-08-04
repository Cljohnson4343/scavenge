package teams

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/cljohnson4343/scavenge/config"
	"github.com/cljohnson4343/scavenge/db"
	"github.com/cljohnson4343/scavenge/request"
	"github.com/cljohnson4343/scavenge/response"
	"github.com/cljohnson4343/scavenge/s3"
	"github.com/cljohnson4343/scavenge/users"
	"github.com/go-chi/render"
	"github.com/google/uuid"
)

// swagger:route GET /teams/ teams getTeamsHandler
//
// Lists all the teams.
//
// Consumes:
// 	- application/json
//
// Produces:
//	- application/json
//
// Schemes: http, https
//
// Responses:
// 	200:
// 	400:
//  500:
func getTeamsHandler(env *config.Env) http.HandlerFunc {
	return (func(w http.ResponseWriter, r *http.Request) {
		teams, e := GetTeams()
		if e != nil {
			e.Handle(w)
		}

		render.JSON(w, r, teams)
		return
	})
}

// swagger:route GET /teams/{teamID} team getTeamHandler
//
// Gets the team with given id.
//
// Consumes:
// 	- application/json
//
// Produces:
//	- application/json
//
// Schemes: http, https
//
// Responses:
// 	200:
// 	400:
// 	404:
func getTeamHandler(env *config.Env) http.HandlerFunc {
	return (func(w http.ResponseWriter, r *http.Request) {
		teamID, e := request.GetIntURLParam(r, "teamID")
		if e != nil {
			e.Handle(w)
			return
		}

		team, e := GetTeam(teamID)
		if e != nil {
			e.Handle(w)
			return
		}

		render.JSON(w, r, team)
		return
	})
}

// swagger:route DELETE /teams/{teamID} delete team deleteTeamHandler
//
// Deletes the given team.
//
// Consumes:
// 	- application/json
//
// Produces:
//	- application/json
//
// Schemes: http, https
//
// Responses:
// 	200:
//  400:
func deleteTeamHandler(env *config.Env) http.HandlerFunc {
	return (func(w http.ResponseWriter, r *http.Request) {
		teamID, e := request.GetIntURLParam(r, "teamID")
		if e != nil {
			e.Handle(w)
			return
		}

		e = DeleteTeam(teamID)
		if e != nil {
			e.Handle(w)
		}

		return
	})
}

// swagger:route POST /teams/ team create createTeamHandler
//
// Creates the given team.
//
// Consumes:
// 	- application/json
//
// Produces:
//	- application/json
//
// Schemes: http, https
//
// Responses:
// 	200:
//  400:
//  500:
func createTeamHandler(env *config.Env) http.HandlerFunc {
	return (func(w http.ResponseWriter, r *http.Request) {
		team := Team{}
		e := request.DecodeAndValidate(r, &team)
		if e != nil {
			e.Handle(w)
			return
		}

		userID, e := users.GetUserID(r.Context())
		if e != nil {
			e.Handle(w)
			return
		}

		e = InsertTeam(userID, &team)
		if e != nil {
			e.Handle(w)
			return
		}

		render.JSON(w, r, &team)
		return
	})
}

// swagger:route PATCH /teams/{teamID} team patchTeamHandler
//
// Partial update on the team with the given id.
// The data that will be updated will be retrieved from
// the request body. All valid keys from the request body
// will update the corresponding team's value with that
// key's value. To update the name of the team send
// body: {"name": "New Team Name"}. NOTE that the id and
// the hunt_id are not eligible to be changed.
//
// Consumes:
// 	- application/json
//
// Produces:
//	- application/json
//
// Schemes: http, https
//
// Responses:
// 	200:
// 	400:
func patchTeamHandler(env *config.Env) http.HandlerFunc {
	return (func(w http.ResponseWriter, r *http.Request) {
		teamID, e := request.GetIntURLParam(r, "teamID")
		if e != nil {
			e.Handle(w)
			return
		}

		team := Team{}

		e = request.DecodeAndPatchValidate(r, &team, teamID)
		if e != nil {
			e.Handle(w)
			return
		}

		e = UpdateTeam(env, &team)
		if e != nil {
			e.Handle(w)
		}

		return
	})
}

// swagger:route GET /teams/{teamID}/locations/ locations getLocationsForTeamHandler
//
// Lists all the locations for a team.
//
// Consumes:
// 	- application/json
//
// Produces:
//	- application/json
//
// Schemes: http, https
//
// Responses:
// 	200:
// 	400:
//  500:
func getLocationsForTeamHandler(env *config.Env) http.HandlerFunc {
	return (func(w http.ResponseWriter, r *http.Request) {
		e := response.NewNilError()
		teamID, e := request.GetIntURLParam(r, "teamID")
		if e != nil {
			e.Handle(w)
			return
		}

		locationDBs, e := db.GetLocationsForTeam(teamID)
		if e != nil {
			e.Handle(w)
		}

		render.JSON(w, r, locationDBs)
		return
	})
}

// swagger:route POST /teams/{teamID}/locations/ location create createLocationHandler
//
// Creates the given location.
//
// Consumes:
// 	- application/json
//
// Produces:
//	- application/json
//
// Schemes: http, https
//
// Responses:
// 	200:
//  400:
//  500:
func createLocationHandler(env *config.Env) http.HandlerFunc {
	return (func(w http.ResponseWriter, r *http.Request) {
		teamID, e := request.GetIntURLParam(r, "teamID")
		if e != nil {
			e.Handle(w)
			return
		}

		location := db.LocationDB{}
		location.TeamID = teamID

		e = request.DecodeAndValidate(r, &location)
		if e != nil {
			e.Handle(w)
			return
		}

		e = location.Insert(teamID)
		if e != nil {
			e.Handle(w)
			return
		}

		render.JSON(w, r, &location)
		return
	})
}

// swagger:route DELETE /teams/{teamID}/locations/{locationID} delete location deleteLocationHandler
//
// Deletes the given location.
//
// Consumes:
// 	- application/json
//
// Produces:
//	- application/json
//
// Schemes: http, https
//
// Responses:
// 	200:
//  400:
func deleteLocationHandler(env *config.Env) http.HandlerFunc {
	return (func(w http.ResponseWriter, r *http.Request) {
		teamID, e := request.GetIntURLParam(r, "teamID")
		if e != nil {
			e.Handle(w)
			return
		}

		locationID, e := request.GetIntURLParam(r, "locationID")
		if e != nil {
			e.Handle(w)
			return
		}

		e = db.DeleteLocation(locationID, teamID)
		if e != nil {
			e.Handle(w)
		}

		return
	})
}

// swagger:route GET /teams/{teamID}/media/ media getMediaForTeamHandler
//
// Lists all the info for the media files associated with a team.
//
// Consumes:
// 	- application/json
//
// Produces:
//	- application/json
//
// Schemes: http, https
//
// Responses:
// 	200:
// 	400:
//  500:
func getMediaForTeamHandler(env *config.Env) http.HandlerFunc {
	return (func(w http.ResponseWriter, r *http.Request) {
		e := response.NewNilError()
		teamID, e := request.GetIntURLParam(r, "teamID")
		if e != nil {
			e.Handle(w)
			return
		}

		mediaMetaDBs, e := db.GetMediaMetasForTeam(teamID)
		if e != nil {
			e.Handle(w)
		}

		render.JSON(w, r, mediaMetaDBs)
		return
	})
}

// swagger:route POST /teams/{teamID}/media/ media create createMediaHandler
//
// Stores the given media info.
//
// Consumes:
// 	- application/json
//
// Produces:
//	- application/json
//
// Schemes: http, https
//
// Responses:
// 	200:
//  400:
//  500:
func createMediaHandler(env *config.Env) http.HandlerFunc {
	return (func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("File Upload Endpoint Hit")

		teamID, e := request.GetIntURLParam(r, "teamID")
		if e != nil {
			e.Handle(w)
			return
		}

		// Parse our multipart form, 10 << 20 specifies a maximum
		// upload of 10 MB files.
		err := r.ParseMultipartForm(10 << 20)
		if err != nil {
			e := response.NewErrorf(http.StatusBadRequest, "Invalid file: %v", err)
			e.Handle(w)
			return
		}

		file, handler, err := r.FormFile("file")
		if err != nil {
			e := response.NewErrorf(
				http.StatusBadRequest,
				"error retrieving 'file' from request: %v",
				err,
			)
			e.Handle(w)
			return
		}
		defer file.Close()

		key := uuid.New()
		url, e := s3.Upload(
			key.String()+s3.GetExt(handler.Filename),
			file,
		)
		if e != nil {
			e.Handle(w)
			return
		}

		media := db.MediaMetaDB{}
		media.TeamID = teamID
		media.URL = url

		jsonBody, _, err := r.FormFile("json")
		if err != nil {
			e := response.NewErrorf(
				http.StatusBadRequest,
				"error retrieving 'json' from request: %v",
				err,
			)
			e.Handle(w)
			return
		}
		defer jsonBody.Close()

		err = json.NewDecoder(jsonBody).Decode(&media)
		if err != nil {
			e := response.NewErrorf(
				http.StatusBadRequest,
				"error reading 'json' from request: %v",
				err,
			)
			e.Handle(w)
			return
		}

		e = media.Insert(teamID)
		if e != nil {
			e.Handle(w)
			return
		}

		render.JSON(w, r, &media)
		return
	})
}

// swagger:route DELETE /teams/{teamID}/media/{mediaID} delete media deleteMediaHandler
//
// Deletes the given media.
//
// Consumes:
// 	- application/json
//
// Produces:
//	- application/json
//
// Schemes: http, https
//
// Responses:
// 	200:
//  400:
func deleteMediaHandler(env *config.Env) http.HandlerFunc {
	return (func(w http.ResponseWriter, r *http.Request) {
		teamID, e := request.GetIntURLParam(r, "teamID")
		if e != nil {
			e.Handle(w)
			return
		}

		mediaID, e := request.GetIntURLParam(r, "mediaID")
		if e != nil {
			e.Handle(w)
			return
		}

		url, e := db.DeleteMedia(mediaID, teamID)
		if e != nil {
			e.Handle(w)
			return
		}

		e = s3.Delete(s3.GetKey(url))
		if e != nil {
			e.Handle(w)
			return
		}

		return
	})
}

// populateMediaDBHandler fills the db with the media in 'media_data.json'
func populateMediaDBHandler(env *config.Env) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		file, err := os.Open("./teams/media_data.json")
		if err != nil {
			panic(err)
		}
		defer file.Close()

		metas := make([]db.MediaMetaDB, 0)

		url := `http://localhost:4343/api/v0/teams/1/media/`

		err = json.NewDecoder(file).Decode(&metas)
		if err != nil {
			e := response.NewErrorf(
				http.StatusInternalServerError,
				"error decoding json data: %s",
				err.Error(),
			)
			e.Handle(w)
			return
		}

		for _, m := range metas {
			b, err := json.Marshal(m)
			if err != nil {
				e := response.NewErrorf(
					http.StatusInternalServerError,
					"error decoding json data: %s",
					err.Error(),
				)
				e.Handle(w)
				return
			}

			buf := bytes.NewBuffer(b)

			res, err := http.Post(url, "application/json", buf)
			if err != nil {
				e := response.NewErrorf(
					http.StatusInternalServerError,
					"error decoding json data: %s",
					err.Error(),
				)
				e.Handle(w)
				return
			}
			defer res.Body.Close()
		}

		w.Write([]byte("Success!!!"))
	}
}

// swagger:route GET /teams/{teamID}/points/ points getTeamPointsHandler
//
// Gets the point total for team.
//
// Consumes:
// 	- application/json
//
// Produces:
//	- application/json
//
// Schemes: http, https
//
// Responses:
// 	200:
// 	400:
// 	404:
func getTeamPointsHandler(env *config.Env) http.HandlerFunc {
	return (func(w http.ResponseWriter, r *http.Request) {
		teamID, e := request.GetIntURLParam(r, "teamID")
		if e != nil {
			e.Handle(w)
			return
		}

		pointTotal, e := db.GetTeamPoints(teamID)
		if e != nil {
			e.Handle(w)
			return
		}

		type pts struct {
			Points int `json:"points"`
		}
		pt := pts{pointTotal}

		render.JSON(w, r, pt)
		return
	})
}

// swagger:route GET /teams/{teamID}/players team players getTeamPlayersHandler
//
// Gets the players on this team.
//
// Consumes:
// 	- application/json
//
// Produces:
//	- application/json
//
// Schemes: http, https
//
// Responses:
// 	200:
// 	400:
// 	404:
func getTeamPlayersHandler(env *config.Env) http.HandlerFunc {
	return (func(w http.ResponseWriter, r *http.Request) {
		teamID, e := request.GetIntURLParam(r, "teamID")
		if e != nil {
			e.Handle(w)
			return
		}

		players, e := db.GetUsersForTeam(teamID)
		if e != nil {
			e.Handle(w)
			return
		}

		render.JSON(w, r, players)
		return
	})
}

// swagger:route POST /teams/{teamID}/players/ add player team getAddPlayerHandler
//
// Gets the handler to add a player to a team.
//
// Consumes:
// 	- application/json
//
// Produces:
//	- application/json
//
// Schemes: http, https
//
// Responses:
// 	200:
// 	400:
// 	404:
func getAddPlayerHandler(env *config.Env) http.HandlerFunc {
	return (func(w http.ResponseWriter, r *http.Request) {
		teamID, e := request.GetIntURLParam(r, "teamID")
		if e != nil {
			e.Handle(w)
			return
		}

		reqBody := struct {
			PlayerID int `json:"userID"`
		}{}

		err := json.NewDecoder(r.Body).Decode(&reqBody)
		if err != nil {
			e := response.NewErrorf(
				http.StatusInternalServerError,
				"error decoding request: %v",
				err,
			)
			e.Handle(w)
			return
		}

		e = AddPlayer(teamID, reqBody.PlayerID)
		if e != nil {
			e.Handle(w)
			return
		}
		// TODO add roles
	})
}

// swagger:route DELETE /teams/{teamID}/players/{playerID} remove players getRemovePlayerHandler
//
// Gets the handler to remove players from a team.
//
// Consumes:
// 	- application/json
//
// Produces:
//	- application/json
//
// Schemes: http, https
//
// Responses:
// 	200:
// 	400:
// 	404:
func getRemovePlayerHandler(env *config.Env) http.HandlerFunc {
	return (func(w http.ResponseWriter, r *http.Request) {
		teamID, e := request.GetIntURLParam(r, "teamID")
		if e != nil {
			e.Handle(w)
			return
		}

		playerID, e := request.GetIntURLParam(r, "playerID")
		if e != nil {
			e.Handle(w)
			return
		}

		e = db.TeamRemovePlayer(teamID, playerID)
		if e != nil {
			e.Handle(w)
			return
		}

	})
}
