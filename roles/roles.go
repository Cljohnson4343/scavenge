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

// Add adds the childRole to the given role. This is how inheritence relationships
// are modeled
func (r *Role) Add(childRole *Role) {
	r.Child = childRole
}

// Authorized returns whether or not the role contains a permission that is
// authorized for the given req
func (r *Role) Authorized(req *http.Request) bool {
	for _, p := range r.Permissions {
		if p.Authorized(req) {
			return true
		}
	}

	if r.Child != nil {
		return r.Child.Authorized(req)
	}

	return false
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
