package users

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/cljohnson4343/scavenge/response"
	"github.com/cljohnson4343/scavenge/sessions"

	c "github.com/cljohnson4343/scavenge/config"
)

// LoginInfoTemp temporary struct for early users development
type LoginInfoTemp struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

//
func getLoginHandler(env *c.Env) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l := LoginInfoTemp{}
		err := json.NewDecoder(r.Body).Decode(&l)
		if err != nil {
			e := response.NewError(fmt.Sprintf("error logging in: %v", err), http.StatusBadRequest)
			e.Handle(w)
			return
		}

		//
		// create user
		//

		// create session and add a session cookie to user agent
		session := sessions.New(43)
		c := session.Cookie()
		http.SetCookie(w, c)

		return
	}
}
