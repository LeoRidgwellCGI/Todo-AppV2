package storage

import (
	"context"
	"errors"
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

type Item struct {
	ID          string    `json:"id"`
	Description string    `json:"description"`
	Status      string    `json:"status"`
	Created     time.Time `json:"created"`
}

type Items map[int64]Item

func createTask(ctx context.Context, description string, optionalStatus string) (string, error) {
	// Validate inputs
	if !validateDescription(description) {
		return "", errors.New("description cannot be empty")
	}
	status := StatusNotStarted
	if optionalStatus != "" {
		if !validateStatus(optionalStatus) {
			return "", errors.New("status is not valid, must be one of: " + strings.Join([]string{StatusNotStarted, StatusStarted, StatusCompleted}, ", "))
		}
		status = optionalStatus
	}

	item := newItem(description, status)
	// TODO: Store the record using storage
	//record := Record{item: item}

	logging.Log().InfoContext(ctx, "Added new item", "id", item.ID, "description", item.Description, "status", item.Status)
	return item.ID, nil
}

func newItem(description string, status string) Item {
	id := logging.GenerateID()
	item := Item{
		ID:          id,
		Description: description,
		Status:      status,
		Created:     time.Now().UTC(),
	}
	return item
}

func updateDescription(ctx context.Context, index int64, description string) error {
	return nil
}

func updateStatus(ctx context.Context, index int64, status string) error {
	return nil
}

func updateTask(ctx context.Context, item Item) (Item, error) {
	return newItem("test", "test"), nil
}

func deleteTask(ctx context.Context, index int64) error {
	return nil
}
func listTask(index int64) {
	return
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
