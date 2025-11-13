package actor

import (
	"context"
	"todo-app/storage"
)

const (
	CreateCmd  string = "CreateCmd"
	UpdateCmd  string = "UpdateCmd"
	DeleteCmd  string = "DeleteCmd"
	ListAllCmd string = "ListAllCmd"
	ListCmd    string = "ListCmd"
)

type Command struct {
	Type        string
	ID          int
	Description string
	Status      string
	ResultChan  chan Response
}

type Response struct {
	Error error
	Item  storage.Item
	Items storage.Items
}

type Actor struct {
	cmdChan chan Command
}

// NewActor creates and starts a new Actor instance.
func NewActor(ctx context.Context) *Actor {
	actor := &Actor{
		cmdChan: make(chan Command),
	}
	go actor.run(ctx)
	return actor
}

// run processes incoming commands sequentially.
func (a *Actor) run(ctx context.Context) {
	for cmd := range a.cmdChan {
		switch cmd.Type {
		case CreateCmd:
			// reload storage to ensure we have the latest data
			reloadStorage(ctx)

			// create the item
			item, err := storage.CreateItem(ctx, cmd.Description, cmd.Status)

			// send back result
			if err != nil {
				cmd.ResultChan <- Response{Error: err}
			} else {
				cmd.ResultChan <- Response{Item: item}
			}

		case UpdateCmd:
			// reload storage to ensure we have the latest data
			reloadStorage(ctx)

			// update the item
			item := storage.Item{ID: cmd.ID, Description: cmd.Description, Status: cmd.Status}
			updated, err := storage.UpdateItem(ctx, item)

			// send back result
			if err != nil {
				cmd.ResultChan <- Response{Error: err}
			} else {
				cmd.ResultChan <- Response{Item: updated}
			}

		case DeleteCmd:
			// reload storage to ensure we have the latest data
			reloadStorage(ctx)

			// delete the item
			err := storage.DeleteItem(ctx, cmd.ID)
			// send back result
			cmd.ResultChan <- Response{Error: err}
		case ListAllCmd:
			// reload storage to ensure we have the latest data
			reloadStorage(ctx)

			// get all items
			items, err := storage.GetAllItems()

			// send back result
			if err != nil {
				cmd.ResultChan <- Response{Error: err}
			} else {
				cmd.ResultChan <- Response{Items: items}
			}
		case ListCmd:
			// reload storage to ensure we have the latest data
			reloadStorage(ctx)

			// get the item by ID
			item, err := storage.GetItemByID(cmd.ID)

			// send back result
			if err != nil {
				cmd.ResultChan <- Response{Error: err}
			} else {
				cmd.ResultChan <- Response{Item: item}
			}
		}
	}
}

// Create creates a new item with the given description and status.
func (a *Actor) Create(ctx context.Context, description string, status string) (storage.Item, error) {
	resultChan := make(chan Response)
	a.cmdChan <- Command{Type: CreateCmd, Description: description, Status: status, ResultChan: resultChan}
	result := <-resultChan
	if result.Error != nil {
		return storage.Item{}, result.Error
	}
	return result.Item, nil
}

// Update updates an existing item with the given ID, description, and status.
func (a *Actor) Update(ctx context.Context, id int, description string, status string) (storage.Item, error) {
	resultChan := make(chan Response)
	a.cmdChan <- Command{Type: UpdateCmd, ID: id, Description: description, Status: status, ResultChan: resultChan}
	result := <-resultChan
	if result.Error != nil {
		return storage.Item{}, result.Error
	}
	return result.Item, nil
}

// Delete deletes the item with the given ID.
func (a *Actor) Delete(ctx context.Context, id int) error {
	resultChan := make(chan Response)
	a.cmdChan <- Command{Type: DeleteCmd, ID: id, ResultChan: resultChan}
	result := <-resultChan
	if result.Error != nil {
		return result.Error
	}
	return nil
}

// ListAll returns all items.
func (a *Actor) ListAll(ctx context.Context) (storage.Items, error) {
	resultChan := make(chan Response)
	a.cmdChan <- Command{Type: ListAllCmd, ResultChan: resultChan}
	result := <-resultChan
	if result.Error != nil {
		return storage.Items{}, result.Error
	}
	return result.Items, nil
}

// List returns the item with the given ID.
func (a *Actor) List(ctx context.Context, id int) (storage.Item, error) {
	resultChan := make(chan Response)
	a.cmdChan <- Command{Type: ListCmd, ID: id, ResultChan: resultChan}
	result := <-resultChan
	if result.Error != nil {
		return storage.Item{}, result.Error
	}
	return result.Item, nil
}

// Helper to reload storage before every read
func reloadStorage(ctx context.Context) {
	if storageFile := storage.GetDataFile(); storageFile != "" {
		_ = storage.Open(ctx, storageFile)
	}
}
