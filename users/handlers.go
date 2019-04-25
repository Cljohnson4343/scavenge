package users

import (
	"net/http"

	"github.com/cljohnson4343/scavenge/request"

	"github.com/cljohnson4343/scavenge/sessions"

	c "github.com/cljohnson4343/scavenge/config"
)

func getLoginHandler(env *c.Env) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		u := User{}
		e := request.DecodeAndValidate(r, &u)
		if e != nil {
			e.Handle(w)
			return
		}

		e = u.Insert()
		if e != nil {
			e.Handle(w)
			return
		}

		// create session and add a session cookie to user agent
		sess := sessions.New(u.ID)
		c := sess.Cookie()
		http.SetCookie(w, c)

		return
	}
}
