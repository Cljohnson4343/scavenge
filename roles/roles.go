package roles

import (
	"net/http"
	"regexp"
	"strings"
)

// Role is a structure that maps permissions to users
type Role struct {
	Name        string
	Permissions []*Permission
	Child       *Role
}

// Permission is the primitive that is responsible for an endpoint authorization
type Permission struct {
	URLRegex string
	Method   string
}

// Authorized returns whether or not the permission is authorized for the
// given request
func (p *Permission) Authorized(r *http.Request) bool {
	if strings.ToLower(r.Method) != strings.ToLower(p.Method) {
		return false
	}

	regEx := regexp.MustCompile(p.URLRegex)
	return regEx.MatchString(r.URL.Path)
}
