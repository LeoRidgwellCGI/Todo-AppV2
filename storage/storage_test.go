package storage

import (
	"context"
	"os"
	"testing"
	"time"
)

// setupTestFile creates a temporary file with the given data and returns its name.
func setupTestFile(t *testing.T, data string) string {
	tmpfile, err := os.CreateTemp("", "testdata*.json")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	if _, err := tmpfile.Write([]byte(data)); err != nil {
		t.Fatalf("Failed to write temp file: %v", err)
	}
	tmpfile.Close()
	return tmpfile.Name()
}

// TestStorage_CreateItem tests the CreateItem function.
func TestStorage_CreateItem(t *testing.T) {
	ctx := context.Background()
	itemsList = Items{}
	itemsDatafile = setupTestFile(t, "{}")
	defer os.Remove(itemsDatafile)

	item, err := CreateItem(ctx, "Test description", "not_started")
	if err != nil {
		t.Fatalf("CreateItem failed: %v", err)
	}
	if item.Description != "Test description" || item.Status != "not_started" {
		t.Errorf("CreateItem returned wrong item: %+v", item)
	}
}

// TestStorage_CreateItem_EmptyDescription tests CreateItem with an empty description.
func TestStorage_CreateItem_EmptyDescription(t *testing.T) {
	ctx := context.Background()
	itemsList = Items{}
	itemsDatafile = setupTestFile(t, "{}")
	defer os.Remove(itemsDatafile)

	_, err := CreateItem(ctx, "", "not_started")
	if err == nil {
		t.Error("Expected error for empty description")
	}
}

// TestStorage_CreateItem_InvalidStatus tests CreateItem with an invalid status.
func TestStorage_CreateItem_InvalidStatus(t *testing.T) {
	ctx := context.Background()
	itemsList = Items{}
	itemsDatafile = setupTestFile(t, "{}")
	defer os.Remove(itemsDatafile)

	_, err := CreateItem(ctx, "desc", "invalid_status")
	if err == nil {
		t.Error("Expected error for invalid status")
	}
}

// TestStorage_UpdateItem tests the UpdateItem function.
func TestStorage_UpdateItem(t *testing.T) {
	ctx := context.Background()
	itemsList = Items{}
	itemsDatafile = setupTestFile(t, "{}")
	defer os.Remove(itemsDatafile)

	item, _ := CreateItem(ctx, "desc", "not_started")
	item.Description = "updated"
	item.Status = "completed"
	updated, err := UpdateItem(ctx, item)
	if err != nil {
		t.Fatalf("UpdateItem failed: %v", err)
	}
	if updated.Description != "updated" || updated.Status != "completed" {
		t.Errorf("UpdateItem did not update fields")
	}
}

// TestStorage_UpdateItem_InvalidID tests UpdateItem with an invalid ID.
func TestStorage_UpdateItem_InvalidID(t *testing.T) {
	ctx := context.Background()
	itemsList = Items{}
	itemsDatafile = setupTestFile(t, "{}")
	defer os.Remove(itemsDatafile)

	item := Item{ID: 0, Description: "desc", Status: "not_started"}
	_, err := UpdateItem(ctx, item)
	if err == nil {
		t.Error("Expected error for invalid ID")
	}
}

// TestStorage_DeleteItem tests the DeleteItem function.
func TestStorage_DeleteItem(t *testing.T) {
	ctx := context.Background()
	itemsList = Items{}
	itemsDatafile = setupTestFile(t, "{}")
	defer os.Remove(itemsDatafile)

	item, _ := CreateItem(ctx, "desc", "not_started")
	err := DeleteItem(ctx, item.ID)
	if err != nil {
		t.Fatalf("DeleteItem failed: %v", err)
	}
	if _, exists := itemsList[item.ID]; exists {
		t.Error("DeleteItem did not remove item")
	}
}

// TestStorage_DeleteItem_InvalidID tests DeleteItem with an invalid ID.
func TestStorage_DeleteItem_InvalidID(t *testing.T) {
	ctx := context.Background()
	itemsList = Items{}
	itemsDatafile = setupTestFile(t, "{}")
	defer os.Remove(itemsDatafile)

	err := DeleteItem(ctx, 0)
	if err == nil {
		t.Error("Expected error for invalid ID")
	}
}

// TestStorage_GetItemByID tests the GetItemByID function.
func TestStorage_GetItemByID(t *testing.T) {
	ctx := context.Background()
	itemsList = Items{}
	itemsDatafile = setupTestFile(t, "{}")
	defer os.Remove(itemsDatafile)

	item, _ := CreateItem(ctx, "desc", "not_started")
	got, err := GetItemByID(item.ID)
	if err != nil {
		t.Fatalf("GetItemByID failed: %v", err)
	}
	if got.ID != item.ID {
		t.Errorf("GetItemByID returned wrong item")
	}
}

// TestStorage_GetItemByID_NotFound tests GetItemByID with a non-existent ID.
func TestStorage_GetItemByID_NotFound(t *testing.T) {
	itemsList = Items{}
	_, err := GetItemByID(999)
	if err == nil {
		t.Error("Expected error for not found")
	}
}

// TestStorage_GetAllItems tests the GetAllItems function.
func TestStorage_GetAllItems(t *testing.T) {
	ctx := context.Background()
	itemsList = Items{}
	itemsDatafile = setupTestFile(t, "{}")
	defer os.Remove(itemsDatafile)

	_, _ = CreateItem(ctx, "desc1", "not_started")
	_, _ = CreateItem(ctx, "desc2", "completed")
	all, err := GetAllItems()
	if err != nil {
		t.Fatalf("GetAllItems failed: %v", err)
	}
	if len(all) != 2 {
		t.Errorf("GetAllItems returned wrong count")
	}
}

// TestStorage_SaveAndLoad tests the Save and Load functions.
func TestStorage_SaveAndLoad(t *testing.T) {
	ctx := context.Background()
	itemsList = Items{}
	datafile := setupTestFile(t, "{}")
	defer os.Remove(datafile)
	itemsDatafile = datafile

	item, _ := CreateItem(ctx, "desc", "not_started")
	if err := Save(ctx, datafile); err != nil {
		t.Fatalf("Save failed: %v", err)
	}
	loaded, err := Load(ctx, datafile)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}
	if len(loaded) != 1 {
		t.Errorf("Load did not load correct items")
	}
	if loaded[item.ID].Description != "desc" {
		t.Errorf("Loaded item mismatch")
	}
}

// TestStorage_Open tests the Open function.
func TestStorage_Open(t *testing.T) {
	ctx := context.Background()
	itemsList = Items{}
	datafile := setupTestFile(t, "{}")
	defer os.Remove(datafile)

	err := Open(ctx, datafile)
	if err != nil {
		t.Fatalf("Open failed: %v", err)
	}
	if itemsDatafile != datafile {
		t.Errorf("Open did not set datafile")
	}
}

// TestStorage_ListItem_NoItems tests ListItem when there are no items.
func TestStorage_ListItem_NoItems(t *testing.T) {
	itemsList = Items{}
	err := ListItem(0)
	if err == nil {
		t.Error("Expected error for no items to list")
	}
}

// TestStorage_HighestKey tests the highestKey function.
func TestStorage_HighestKey(t *testing.T) {
	keys := []int{1, 2, 5, 3}
	if highestKey(keys) != 5 {
		t.Errorf("highestKey failed")
	}
}

// TestStorage_CollectKeys tests the collectKeys function.
func TestStorage_CollectKeys(t *testing.T) {
	items := Items{
		1: Item{ID: 1},
		2: Item{ID: 2},
	}
	keys := collectKeys(items)
	if len(keys) != 2 {
		t.Errorf("collectKeys failed")
	}
}

// TestStorage_NewItem tests the newItem function.
func TestStorage_NewItem(t *testing.T) {
	now := time.Now()
	item := newItem(1, "desc", "not_started")
	if item.ID != 1 || item.Description != "desc" || item.Status != "not_started" {
		t.Errorf("newItem failed")
	}
	if item.Created.Before(now) {
		t.Errorf("newItem Created time incorrect")
	}
}
