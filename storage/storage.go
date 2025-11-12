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

const (
	StatusNotStarted = "Not Started"
	StatusStarted    = "Started"
	StatusCompleted  = "Completed"
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

// Save writes the current items list to the specified json file.
func Save(ctx context.Context, datafile string) error {
	if data, err := json.Marshal(itemsList); err != nil {
		fmt.Printf("Save failed converting todo list to json, error: %s \n", err)
		slog.ErrorContext(ctx, "Save failed converting todo list to json", "error", err)
		return err
	} else {
		if destination, err := OpenFileWriteTruncate(datafile); err != nil {
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
	destination, err := OpenFileReadWrite(datafile)
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
func CreateItem(ctx context.Context, description string) (int, error) {
	// Validate inputs
	if description == "" {
		return 0, errors.New("description cannot be empty")
	}

	// Determine next key
	itemKeys := collectKeys(itemsList)
	nextKey := highestKey(itemKeys) + 1
	item := newItem(nextKey, description, StatusNotStarted)
	itemsList[nextKey] = item

	// Commit to file
	CommitFile(ctx)

	// Log creation
	slog.InfoContext(ctx, "Created new item", "ID", item.ID, "Description", item.Description, "Status:", item.Status)
	fmt.Printf("Created new item, ID: %d, Description: %s, Status: %s \n", item.ID, item.Description, item.Status)

	// return new item ID
	return nextKey, nil
}

// UpdateDescription updates the description of the item with the given index.
func UpdateDescription(ctx context.Context, index int, description string) error {
	// Validate inputs
	if description == "" {
		return errors.New("description cannot be empty")
	}

	// Update the item
	fmt.Printf("Updating item %d description:\n", index)

	// check item exists
	item, exists := itemsList[index]
	if !exists {
		return errors.New("item not found")
	}

	// update description
	item.Description = description
	itemsList[index] = item

	// Commit to file
	CommitFile(ctx)

	// Log update
	slog.InfoContext(ctx, "Updated item description", "ID", item.ID, "New Description", item.Description)
	fmt.Printf("Updated item description, ID: %d, New Description: %s \n", item.ID, item.Description)

	// return nil error
	return nil
}

// UpdateStatus updates the status of the item with the given index.
func UpdateStatus(ctx context.Context, index int, status string) error {
	// Validate inputs
	validStatuses := []string{StatusNotStarted, StatusStarted, StatusCompleted}
	if !slices.Contains(validStatuses, status) {
		return errors.New("invalid status value")
	}

	// Update the item
	fmt.Printf("Updating item %d status:\n", index)

	// check item exists
	item, exists := itemsList[index]
	if !exists {
		return errors.New("item not found")
	}

	// update status
	item.Status = status
	itemsList[index] = item

	// Commit to file
	CommitFile(ctx)

	// Log update
	slog.InfoContext(ctx, "Updated item status", "ID", item.ID, "New Status", item.Status)
	fmt.Printf("Updated item status, ID: %d, New Status: %s \n", item.ID, item.Status)

	// return nil error
	return nil
}

func UpdateItem(ctx context.Context, item Item) (Item, error) {
	// Validate inputs
	if item.ID <= 0 {
		return Item{}, errors.New("invalid item ID")
	}
	if item.Description == "" {
		return Item{}, errors.New("description cannot be empty")
	}
	validStatuses := []string{StatusNotStarted, StatusStarted, StatusCompleted}
	if !slices.Contains(validStatuses, item.Status) {
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
	current.Description = item.Description
	current.Status = item.Status
	itemsList[item.ID] = current

	// Commit to file
	CommitFile(ctx)

	// Log update
	slog.InfoContext(ctx, "Updated item", "ID", item.ID, "New Description", item.Description, "New Status", item.Status)
	fmt.Printf("Updated item, ID: %d, New Description: %s, New Status: %s \n", item.ID, item.Description, item.Status)

	// return updated item
	return item, nil
}

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
	CommitFile(ctx)

	// Log deletion
	slog.InfoContext(ctx, "Deleted item", "ID", index)
	fmt.Printf("Deleted item, ID: %d \n", index)

	// return nil error
	return nil
}
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

// OpenFileReadWrite opens (or creates) a file for reading and writing.
func OpenFileReadWrite(fileName string) (*os.File, error) {
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
func OpenFileWriteTruncate(fileName string) (*os.File, error) {
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

// CommitFile saves the current items list to the data file if it is open.
func CommitFile(ctx context.Context) {
	if itemsList != nil {
		Save(ctx, itemsDatafile)
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
