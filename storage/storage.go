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

	"todo-app/logging"
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
	ID          string    `json:"id"`
	Description string    `json:"description"`
	Status      string    `json:"status"`
	Created     time.Time `json:"created"`
	Line        int       `json:"line"`
}

type Items map[int]Item

func newItem(description string, status string, line int) Item {
	id := logging.GenerateID()
	item := Item{
		ID:          id,
		Description: description,
		Status:      status,
		Created:     time.Now().UTC(),
		Line:        line,
	}
	return item
}

// save list back to json file
func Save(ctx context.Context, datafile string) error {
	if data, err := json.Marshal(itemsList); err != nil {
		fmt.Printf("Save failed converting todo list to json, error: %s \n", err)
		logging.Log().ErrorContext(ctx, "Save failed converting todo list to json", "error", err)
		return err
	} else {
		if destination, err := OpenFileWriteTruncate(datafile); err != nil {
			fmt.Printf("Save failed getting file, error: %s, datafile: %s \n", err, datafile)
			logging.Log().ErrorContext(ctx, "Save failed getting file", "error", err, "datafile", datafile)
			return err
		} else {
			defer destination.Close()
			if _, err := destination.Write(data); err != nil {
				fmt.Printf("Save to file failed, error: %s, datafile: %s \n", err, datafile)
				logging.Log().ErrorContext(ctx, "Save to file failed", "error", err, "datafile", datafile)
				return err
			}
		}
	}
	fmt.Printf("Saved data to file, datafile: %s \n", datafile)
	logging.Log().InfoContext(ctx, "Saved data to file", "datafile", datafile)
	return nil
}

// load list from json file
func Load(ctx context.Context, datafile string) (Items, error) {
	destination, err := OpenFileReadWrite(datafile)
	if err != nil {
		fmt.Printf("Load failed listing file, error: %s, datafile: %s\n", err, datafile)
		logging.Log().ErrorContext(ctx, "Load failed listing file", "error", err, "datafile", datafile)
		return Items{}, err
	}
	if destination != nil {
		defer destination.Close()
	}
	return loadItem(ctx, destination)
}

func loadItem(ctx context.Context, destination io.Reader) (Items, error) {
	if item, err := io.ReadAll(destination); err != nil {
		fmt.Println(err)
		fmt.Printf("Load item failed, error: %s \n", err)
		logging.Log().ErrorContext(ctx, "Load item failed", "error", err)
		return Items{}, err
	} else if len(item) == 0 {
		// not neccessarily an error
		fmt.Printf("No data to load, returning empty item list \n")
		return Items{}, nil
	} else {
		data := []byte(string(item))
		itemsList := Items{}
		err := json.Unmarshal(data, &itemsList)
		if err != nil {
			fmt.Println(err)
			logging.Log().ErrorContext(ctx, "Load item from json failed", "error", err)
			return Items{}, err
		}
		return itemsList, nil
	}
}

func Open(ctx context.Context, datafile string) error {
	items, err := Load(ctx, datafile)
	if err != nil {
		fmt.Printf("Open file failed, error: %s, datafile: %s\n", err, datafile)
		logging.Log().ErrorContext(ctx, "Open file failed", "error", err, "datafile", datafile)
		return err
	}
	itemsList = items
	itemsDatafile = datafile
	fmt.Printf("Opened file and loaded items, count: %d, datafile: %s \n", len(itemsList), datafile)
	logging.Log().InfoContext(ctx, "Opened file and loaded items", "count", len(itemsList), "datafile", datafile)
	return nil
}

func CreateItem(ctx context.Context, description string, optionalStatus string) (int, error) {
	// Validate inputs
	if !validateDescription(description) {
		return 0, errors.New("description cannot be empty")
	}
	status := StatusNotStarted
	if optionalStatus != "" {
		if !validateStatus(optionalStatus) {
			return 0, errors.New("status is not valid, must be one of: " + strings.Join([]string{StatusNotStarted, StatusStarted, StatusCompleted}, ", "))
		}
		status = optionalStatus
	}
	// Determine next key
	itemKeys := collectKeys(itemsList)
	nextKey := highestKey(itemKeys) + 1
	item := newItem(description, status, nextKey)
	itemsList[nextKey] = item

	// Log creation
	logging.Log().InfoContext(ctx, "Added new item", "ID", item.ID, "Description", item.Description, "Status:", item.Status)
	return nextKey, nil
}

func UpdateDescription(ctx context.Context, index int, description string) error {
	return nil
}

func UpdateStatus(ctx context.Context, index int, status string) error {
	return nil
}

func UpdateItem(ctx context.Context, item Item) (Item, error) {
	return newItem("test", "test", 0), nil
}

func DeleteItem(ctx context.Context, index int) error {
	return nil
}
func ListItem(index int) {
	return
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

// IsDataFileOpen checks if the data file is currently open.
func IsDataFileOpen() bool {
	return (itemsList != nil)
}

// CommitFile saves the current items list to the data file if it is open.
func CommitFile(ctx context.Context) {
	if IsDataFileOpen() {
		Save(ctx, itemsDatafile)
	}
}

// getItems returns the current list of items
func getItems() Items {
	return itemsList
}

// resetItems clears the current list of items
func resetItems() {
	itemsList = Items{}
}

// validateDescription checks if the provided description is valid (non-empty).
func validateDescription(description string) bool {
	return description != ""
}

// validateStatus checks if the provided status is one of the valid statuses.
func validateStatus(status string) bool {
	validStatuses := []string{StatusNotStarted, StatusStarted, StatusCompleted}
	return slices.Contains(validStatuses, status)
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
