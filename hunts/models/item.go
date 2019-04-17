// Package models provides all data models for package hunts
package models

// Item is the data representation of a scavenger hunt item
//
// swagger:model item
type Item struct {

	// The id of the Hunt
	//
	// required: true
	HuntID int `json:"hunt_id"`

	// The id of the item
	//
	// required: true
	ID int `json:"id"`

	// the name of the item
	//
	// maximum length: 255
	// required: true
	Name string `json:"name"`

	// the amount of points this item is worth
	//
	// minimum: 1
	Points uint `json:"points,omitempty"`
}
