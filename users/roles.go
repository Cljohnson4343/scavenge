package users

import (
	"fmt"
	"strings"

	"github.com/cljohnson4343/scavenge/roles"
)

var permissionMap = map[string]string{
	"get_user":    `^/%d$`,
	"post_login":  `^/login/$`,
	"post_logout": `^/logout/$`,
	"post_user":   `^/$`,
	"delete_user": `^/%d$`,
	"patch_user":  `^/%d$`,
}

// GeneratePermission generates permission for the given route and user id
func GeneratePermission(perm string, id int) *roles.Permission {
	var regex string
	if strings.Contains(permissionMap[perm], "%") {
		regex = fmt.Sprintf(permissionMap[perm], id)
	} else {
		regex = permissionMap[perm]
	}
	permission := roles.Permission{
		Method:   strings.Split(perm, "_")[0],
		URLRegex: regex,
	}

	return &permission
}
