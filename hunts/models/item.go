// Package models provides all data models for package hunts
package models

import "github.com/cljohnson4343/scavenge/db"

// Item is the data representation of a scavenger hunt item
//
// swagger:model item
type Item struct {
	db.ItemDB
}
