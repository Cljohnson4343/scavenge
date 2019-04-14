// Package models provides all data models for package hunts
package models

// Item is the data representation of a scavenger hunt item
//
// swagger:model item
type Item struct {
	// the name of the item
	//
	// maximum length: 255
	// required: true
	Name string `json:"name"`

	// the amount of points this item is worth
	//
	// minimum: 1
	Points uint `json:"points,omitempty"`

	// whether or not this item has been found
	//
	// required true
	IsDone bool `json:"is_done"`
}
