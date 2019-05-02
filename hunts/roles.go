package hunts

import (
	"fmt"
	"strings"

	"github.com/cljohnson4343/scavenge/roles"
)

var permissionMap = map[string]string{
	"get_hunts":     `^/$`,
	"get_hunt":      `^/%d$`,
	"post_hunt":     `^/$`,
	"delete_hunt":   `^/%d$`,
	"patch_hunt":    `^/%d$`,
	"post_populate": `^/populate/$`,
	"get_items":     `^/%d/items/$`,
	"delete_item":   `^/%d/items/\d+$`,
	"post_item":     `^/%d/items/$`,
	"patch_item":    `^/%d/items/\d+$`,
}

// GeneratePermission generates permission for the given route and hunt id
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
