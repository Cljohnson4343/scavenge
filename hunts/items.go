package hunts

import (
	"github.com/cljohnson4343/scavenge/config"
	"github.com/cljohnson4343/scavenge/db"
	"github.com/cljohnson4343/scavenge/hunts/models"
	"github.com/cljohnson4343/scavenge/response"
)

// GetItems returns the items for the given hunt
func GetItems(huntID int) ([]*models.Item, *response.Error) {
	itemDBs, e := db.GetItemsWithHuntID(huntID)
	if itemDBs == nil {
		return nil, e
	}

	items := make([]*models.Item, 0, len(itemDBs))
	for _, itemDB := range itemDBs {
		item := models.Item{ItemDB: *itemDB}
		items = append(items, &item)
	}

	return items, e
}

// GetItemsForHunt returns all the items for the hunt with the given id
func GetItemsForHunt(huntID int) ([]*models.Item, *response.Error) {
	itemDBs, e := db.GetItemsWithHuntID(huntID)
	if itemDBs == nil {
		return nil, e
	}

	items := make([]*models.Item, 0, len(itemDBs))

	for _, v := range itemDBs {
		item := models.Item{ItemDB: *v}
		items = append(items, &item)
	}

	return items, e
}

// InsertItem inserts an Item into the db
func InsertItem(item *models.Item) *response.Error {
	return item.Insert()
}

// DeleteItem deletes the item with the given itemID AND huntID
func DeleteItem(env *config.Env, huntID, itemID int) *response.Error {
	return db.DeleteItem(itemID, huntID)
}

// UpdateItem executes a partial update of the item with the given id. NOTE:
// item_id and hunt_id are not eligible to be changed
func UpdateItem(env *config.Env, item *models.Item) *response.Error {
	return item.Update(env)
}
