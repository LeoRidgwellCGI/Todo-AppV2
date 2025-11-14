package storage

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"slices"
	"strings"
	"time"
)

var (
	itemsList     Items = Items{}
	itemsDatafile string
)

type Item struct {
	ID          int       `json:"id"`
	Description string    `json:"description"`
	Status      string    `json:"status"`
	Created     time.Time `json:"created"`
}

type Items map[int]Item

// newItem creates a new Item with the given parameters.
func newItem(id int, description string, status string) Item {
	item := Item{
		ID:          id,
		Description: description,
		Status:      status,
		Created:     time.Now().UTC(),
	}
	return item
}

// GetDataFile returns the current datafile path used by storage
func GetDataFile() string {
	return itemsDatafile
}

// Save writes the current items list to the specified json file.
func Save(ctx context.Context, datafile string) error {
	if data, err := json.Marshal(itemsList); err != nil {
		fmt.Printf("Save failed converting todo list to json, error: %s \n", err)
		slog.ErrorContext(ctx, "Save failed converting todo list to json", "error", err)
		return err
	} else {
		if destination, err := openFileWriteTruncate(datafile); err != nil {
			fmt.Printf("Save failed getting file, error: %s, datafile: %s \n", err, datafile)
			slog.ErrorContext(ctx, "Save failed getting file", "error", err, "datafile", datafile)
			return err
		} else {
			defer destination.Close()
			if _, err := destination.Write(data); err != nil {
				fmt.Printf("Save to file failed, error: %s, datafile: %s \n", err, datafile)
				slog.ErrorContext(ctx, "Save to file failed", "error", err, "datafile", datafile)
				return err
			}
		}
	}
	fmt.Printf("Saved data to file, datafile: %s \n", datafile)
	slog.InfoContext(ctx, "Saved data to file", "datafile", datafile)
	return nil
}

// Load reads the items list from the specified json file.
func Load(ctx context.Context, datafile string) (Items, error) {
	destination, err := openFileReadWrite(datafile)
	if err != nil {
		fmt.Printf("Load failed listing file, error: %s, datafile: %s\n", err, datafile)
		slog.ErrorContext(ctx, "Load failed listing file", "error", err, "datafile", datafile)
		return Items{}, err
	}
	if destination != nil {
		defer destination.Close()
	}
	return loadItem(ctx, destination)
}

// Open initializes the storage by loading items from the specified data file.
func Open(ctx context.Context, datafile string) error {
	// load existing
	items, err := Load(ctx, datafile)
	if err != nil {
		fmt.Printf("Open file failed, error: %s, datafile: %s\n", err, datafile)
		slog.ErrorContext(ctx, "Open file failed", "error", err, "datafile", datafile)
		return err
	}

	// set global items list
	itemsList = items
	itemsDatafile = datafile

	// log loaded items count
	fmt.Printf("Opened file and loaded items, count: %d, datafile: %s \n", len(itemsList), datafile)
	slog.InfoContext(ctx, "Opened file and loaded items", "count", len(itemsList), "datafile", datafile)
	return nil
}

// CreateItem creates a new item with the given description and adds it to the items list.
func CreateItem(ctx context.Context, description string, status string) (Item, error) {
	// Validate inputs
	if description == "" {
		return Item{}, errors.New("description cannot be empty")
	}
	if status != "" {
		if status != "not_started" && status != "in_progress" && status != "completed" {
			return Item{}, errors.New("invalid status value")
		}
	} else {
		status = "not_started"
	}

	// Determine next key
	itemKeys := collectKeys(itemsList)
	nextKey := highestKey(itemKeys) + 1
	item := newItem(nextKey, description, status)
	itemsList[nextKey] = item

	// Commit to file
	commitFile(ctx)

	// Log creation
	slog.InfoContext(ctx, "Created new item", "ID", item.ID, "Description", item.Description, "Status:", item.Status)
	fmt.Printf("Created new item, ID: %d, Description: %s, Status: %s \n", item.ID, item.Description, item.Status)

	// return new item ID
	return item, nil
}

// UpdateItem updates an existing item in the items list.
func UpdateItem(ctx context.Context, item Item) (Item, error) {
	// Validate inputs
	if item.ID <= 0 {
		return Item{}, errors.New("invalid item ID")
	}
	if item.Description == "" {
		return Item{}, errors.New("description cannot be empty")
	}
	if item.Status != "not_started" && item.Status != "in_progress" && item.Status != "completed" {
		return Item{}, errors.New("invalid status value")
	}

	// Update the item
	fmt.Printf("Updating item %d:\n", item.ID)

	// check item exists
	current, exists := itemsList[item.ID]
	if !exists {
		return Item{}, errors.New("item not found")
	}

	// update item
	itemsList[item.ID] = item

	// Commit to file
	commitFile(ctx)

	// Log update
	slog.InfoContext(ctx, "Updated item", "ID", item.ID, "Old Description", current.Description, "New Description", item.Description, "Old Status", current.Status, "New Status", item.Status)
	fmt.Printf("Updated item, ID: %d, Old Description: %s, New Description: %s, Old Status: %s, New Status: %s \n", item.ID, current.Description, item.Description, current.Status, item.Status)

	// return updated item
	return item, nil
}

// DeleteItem deletes an item from the items list by its ID.
func DeleteItem(ctx context.Context, index int) error {
	// validate inputs
	if index <= 0 {
		return errors.New("invalid item ID")
	}

	// Delete the item
	fmt.Printf("Deleting item %d:\n", index)

	// check item exists
	_, exists := itemsList[index]
	if !exists {
		return errors.New("item not found")
	}

	// delete item
	delete(itemsList, index)

	// Commit to file
	commitFile(ctx)

	// Log deletion
	slog.InfoContext(ctx, "Deleted item", "ID", index)
	fmt.Printf("Deleted item, ID: %d \n", index)

	// return nil error
	return nil
}

// ListItem lists items; if index is 0, lists all items, otherwise lists the item with the given ID.
func ListItem(index int) error {
	// List items
	fmt.Printf("Listing items:\n")

	// print header
	fmt.Printf("%s\t%s\t\t%s\n", "ID", "Status", "Description")
	fmt.Printf("%s\t%s\t%s\n", strings.Repeat("-", 1), strings.Repeat("-", 12), strings.Repeat("-", 120))

	// reference current items list
	if len(itemsList) > 0 {
		if listItem, ok := itemsList[index]; ok {
			fmt.Printf("%d\t%s\t%s\t[%s]\n", listItem.ID, listItem.Status, listItem.Description, listItem.Created.Format(time.RFC822))
		} else {
			itemKeys := collectKeys(itemsList)
			slices.Sort(itemKeys)
			for _, i := range itemKeys {
				listItem := itemsList[i]
				fmt.Printf("%d\t%s\t%s\t[%s]\n", listItem.ID, listItem.Status, listItem.Description, listItem.Created.Format(time.RFC822))
			}
		}
	} else {
		// no items to list
		return errors.New("no items to list")
	}
	return nil
}

func GetItemByID(id int) (Item, error) {
	// validate inputs
	if id <= 0 {
		return Item{}, errors.New("invalid item ID")
	}
	// retrieve item by ID
	if len(itemsList) > 0 {
		item, ok := itemsList[id]
		if ok {
			return item, nil
		} else {
			return Item{}, errors.New("item not found")
		}
	} else {
		return Item{}, errors.New("no items available")
	}
}

// GetAllItems returns all items in the items list.
func GetAllItems() (Items, error) {
	if len(itemsList) > 0 {
		return itemsList, nil
	}
	return Items{}, errors.New("no items available")
}

// commitFile saves the current items list to the data file if it is open.
func commitFile(ctx context.Context) {
	if itemsList != nil {
		Save(ctx, itemsDatafile)
	}
}

// openFileReadWrite opens (or creates) a file for reading and writing.
func openFileReadWrite(fileName string) (*os.File, error) {
	// open the file for reading and writing, creating it if it does not exist
	// with permissions rw-r--r--
	if fi, err := os.OpenFile(fileName, os.O_RDWR|os.O_CREATE, 0644); err != nil {
		slog.Error(fmt.Sprintf("%s\n", "Failed to open data file for reading and writing"))
		slog.Error(err.Error())
		return &os.File{}, err
	} else {
		return fi, nil
	}
}

// OpenFileWriteTruncate opens (or creates) a file for writing, truncating it if it exists.
func openFileWriteTruncate(fileName string) (*os.File, error) {
	// open the file for writing, truncating it if it exists, or creating it if it does not exist
	// with permissions rw-r--r--
	// truncate mode so we start fresh
	if fi, err := os.OpenFile(fileName, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644); err != nil {
		slog.Error(fmt.Sprintf("%s\n", "Failed to open data file for writing and truncating"))
		slog.Error(err.Error())
		return &os.File{}, err
	} else {
		return fi, nil
	}
}

// loadItem reads and unmarshals the items from the given reader.
func loadItem(ctx context.Context, destination io.Reader) (Items, error) {
	// read all data from the reader
	if item, err := io.ReadAll(destination); err != nil {
		fmt.Println(err)
		fmt.Printf("Load item failed, error: %s \n", err)
		slog.ErrorContext(ctx, "Load item failed", "error", err)
		return Items{}, err
	} else if len(item) == 0 {
		// not neccessarily an error
		fmt.Printf("No data to load, returning empty item list \n")
		return Items{}, nil
	} else {
		// unmarshal json data
		data := []byte(string(item))
		itemsList := Items{}
		err := json.Unmarshal(data, &itemsList)
		if err != nil {
			fmt.Println(err)
			slog.ErrorContext(ctx, "Load item from json failed", "error", err)
			return Items{}, err
		}
		return itemsList, nil
	}
}

// collectKeys collects the keys from the Items map and returns them as a slice of ints.
func collectKeys(data Items) []int {
	keys := make([]int, 0, len(data))
	for i := range data {
		// map keys are int64, convert to int
		i32 := int(i)
		keys = append(keys, i32)
	}
	return keys
}

// highestKey returns the highest key from a slice of ints.
func highestKey(keys []int) int {
	key := 0
	for _, i := range keys {
		if i > key {
			key = i
		}
	}
	return key
}
