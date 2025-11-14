package actor

import (
	"context"
	"os"
	"sync"
	"testing"
	"time"
	"todo-app/storage"
)

// setupTestStorage initializes a temporary storage file for testing.
func setupTestStorage(t *testing.T) (string, func()) {
	tmpFile := "test_todos_" + time.Now().Format("20060102150405") + ".json"
	ctx := context.Background()

	// Initialize storage with temp file
	err := storage.Open(ctx, tmpFile)
	if err != nil {
		t.Fatalf("Failed to open test storage: %v", err)
	}

	cleanup := func() {
		_ = os.Remove(tmpFile)
	}

	return tmpFile, cleanup
}

// TestActor_NewActor tests the creation of a new Actor instance.
func TestActor_NewActor(t *testing.T) {
	_, cleanup := setupTestStorage(t)
	defer cleanup()

	ctx := context.Background()
	actor := NewActor(ctx)

	if actor == nil {
		t.Fatal("NewActor returned nil")
	}
	if actor.cmdChan == nil {
		t.Error("Actor command channel is nil")
	}
}

// TestActor_Create tests the Create method.
func TestActor_Create(t *testing.T) {
	_, cleanup := setupTestStorage(t)
	defer cleanup()

	ctx := context.Background()
	actor := NewActor(ctx)

	item, err := actor.Create(ctx, "Test Item", "not_started")
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	if item.Description != "Test Item" {
		t.Errorf("Expected description 'Test Item', got '%s'", item.Description)
	}
	if item.Status != "not_started" {
		t.Errorf("Expected status 'not_started', got '%s'", item.Status)
	}
	if item.ID <= 0 {
		t.Errorf("Expected positive ID, got %d", item.ID)
	}
}

// TestActor_CreateMultiple tests creating multiple items.
func TestActor_CreateMultiple(t *testing.T) {
	_, cleanup := setupTestStorage(t)
	defer cleanup()

	ctx := context.Background()
	actor := NewActor(ctx)

	items := []struct {
		desc   string
		status string
	}{
		{"Item 1", "not_started"},
		{"Item 2", "in_progress"},
		{"Item 3", "completed"},
	}

	for _, item := range items {
		created, err := actor.Create(ctx, item.desc, item.status)
		if err != nil {
			t.Fatalf("Create failed for '%s': %v", item.desc, err)
		}
		if created.Description != item.desc {
			t.Errorf("Expected description '%s', got '%s'", item.desc, created.Description)
		}
	}
}

// TestActor_List tests retrieving a single item by ID.
func TestActor_List(t *testing.T) {
	_, cleanup := setupTestStorage(t)
	defer cleanup()

	ctx := context.Background()
	actor := NewActor(ctx)

	// Create an item first
	created, err := actor.Create(ctx, "Test Item", "not_started")
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	// Retrieve the item
	retrieved, err := actor.List(ctx, created.ID)
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}

	if retrieved.ID != created.ID {
		t.Errorf("Expected ID %d, got %d", created.ID, retrieved.ID)
	}
	if retrieved.Description != created.Description {
		t.Errorf("Expected description '%s', got '%s'", created.Description, retrieved.Description)
	}
}

// TestActor_List_NotFound tests retrieving a non-existent item.
func TestActor_List_NotFound(t *testing.T) {
	_, cleanup := setupTestStorage(t)
	defer cleanup()

	ctx := context.Background()
	actor := NewActor(ctx)

	_, err := actor.List(ctx, 999)
	if err == nil {
		t.Error("Expected error for non-existent item, got nil")
	}
}

// TestActor_ListAll tests retrieving all items.
func TestActor_ListAll(t *testing.T) {
	_, cleanup := setupTestStorage(t)
	defer cleanup()

	ctx := context.Background()
	actor := NewActor(ctx)

	// Create multiple items
	for i := 1; i <= 3; i++ {
		_, err := actor.Create(ctx, "Item "+string(rune('0'+i)), "not_started")
		if err != nil {
			t.Fatalf("Create failed: %v", err)
		}
	}

	// List all items
	items, err := actor.ListAll(ctx)
	if err != nil {
		t.Fatalf("ListAll failed: %v", err)
	}

	if len(items) != 3 {
		t.Errorf("Expected 3 items, got %d", len(items))
	}
}

// TestActor_Update tests updating an existing item.
func TestActor_Update(t *testing.T) {
	_, cleanup := setupTestStorage(t)
	defer cleanup()

	ctx := context.Background()
	actor := NewActor(ctx)

	// Create an item
	created, err := actor.Create(ctx, "Original Description", "not_started")
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	// Update the item
	updated, err := actor.Update(ctx, created.ID, "Updated Description", "in_progress")
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}

	if updated.ID != created.ID {
		t.Errorf("Expected ID %d, got %d", created.ID, updated.ID)
	}
	if updated.Description != "Updated Description" {
		t.Errorf("Expected description 'Updated Description', got '%s'", updated.Description)
	}
	if updated.Status != "in_progress" {
		t.Errorf("Expected status 'in_progress', got '%s'", updated.Status)
	}
}

// TestActor_Update_NotFound tests updating a non-existent item.
func TestActor_Update_NotFound(t *testing.T) {
	_, cleanup := setupTestStorage(t)
	defer cleanup()

	ctx := context.Background()
	actor := NewActor(ctx)

	_, err := actor.Update(ctx, 999, "Updated Description", "in_progress")
	if err == nil {
		t.Error("Expected error for non-existent item, got nil")
	}
}

// TestActor_Delete tests deleting an item.
func TestActor_Delete(t *testing.T) {
	_, cleanup := setupTestStorage(t)
	defer cleanup()

	ctx := context.Background()
	actor := NewActor(ctx)

	// Create an item
	created, err := actor.Create(ctx, "To Be Deleted", "not_started")
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	// Delete the item
	err = actor.Delete(ctx, created.ID)
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	// Verify it's gone
	_, err = actor.List(ctx, created.ID)
	if err == nil {
		t.Error("Expected error after delete, got nil")
	}
}

// TestActor_Delete_NotFound tests deleting a non-existent item.
func TestActor_Delete_NotFound(t *testing.T) {
	_, cleanup := setupTestStorage(t)
	defer cleanup()

	ctx := context.Background()
	actor := NewActor(ctx)

	err := actor.Delete(ctx, 999)
	if err == nil {
		t.Error("Expected error for non-existent item, got nil")
	}
}

// TestActor_EmptyList tests listing when no items exist.
func TestActor_EmptyList(t *testing.T) {
	_, cleanup := setupTestStorage(t)
	defer cleanup()

	ctx := context.Background()
	actor := NewActor(ctx)

	items, _ := actor.ListAll(ctx)

	if len(items) != 0 {
		t.Errorf("Expected 0 items, got %d", len(items))
	}
}

// TestActor_InvalidStatus tests creating with invalid status.
func TestActor_InvalidStatus(t *testing.T) {
	_, cleanup := setupTestStorage(t)
	defer cleanup()

	ctx := context.Background()
	actor := NewActor(ctx)

	_, err := actor.Create(ctx, "Test Item", "invalid_status")
	if err == nil {
		t.Error("Expected error for invalid status, got nil")
	}
}

// TestActor_EmptyDescription tests creating with empty description.
func TestActor_EmptyDescription(t *testing.T) {
	_, cleanup := setupTestStorage(t)
	defer cleanup()

	ctx := context.Background()
	actor := NewActor(ctx)

	_, err := actor.Create(ctx, "", "not_started")
	if err == nil {
		t.Error("Expected error for empty description, got nil")
	}
}

// TestActor_Concurrency_Create tests concurrent create operations.
func TestActor_Concurrency_Create(t *testing.T) {
	_, cleanup := setupTestStorage(t)
	defer cleanup()

	ctx := context.Background()
	actor := NewActor(ctx)

	const numGoroutines = 20
	var wg sync.WaitGroup
	errChan := make(chan error, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			_, err := actor.Create(ctx, "Concurrent Item", "not_started")
			if err != nil {
				errChan <- err
			}
		}(i)
	}

	wg.Wait()
	close(errChan)

	for err := range errChan {
		t.Errorf("Concurrent create failed: %v", err)
	}

	// Verify all items were created
	items, err := actor.ListAll(ctx)
	if err != nil {
		t.Fatalf("ListAll failed: %v", err)
	}
	if len(items) != numGoroutines {
		t.Errorf("Expected %d items, got %d", numGoroutines, len(items))
	}
}

// TestActor_Concurrency_Read tests concurrent read operations.
func TestActor_Concurrency_Read(t *testing.T) {
	_, cleanup := setupTestStorage(t)
	defer cleanup()

	ctx := context.Background()
	actor := NewActor(ctx)

	// Create an item to read
	created, err := actor.Create(ctx, "Read Test Item", "not_started")
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	const numGoroutines = 50
	var wg sync.WaitGroup
	errChan := make(chan error, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := actor.List(ctx, created.ID)
			if err != nil {
				errChan <- err
			}
		}()
	}

	wg.Wait()
	close(errChan)

	for err := range errChan {
		t.Errorf("Concurrent read failed: %v", err)
	}
}

// TestActor_Concurrency_Update tests concurrent update operations.
func TestActor_Concurrency_Update(t *testing.T) {
	_, cleanup := setupTestStorage(t)
	defer cleanup()

	ctx := context.Background()
	actor := NewActor(ctx)

	// Create an item to update
	created, err := actor.Create(ctx, "Original", "not_started")
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	const numGoroutines = 10
	var wg sync.WaitGroup
	errChan := make(chan error, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			_, err := actor.Update(ctx, created.ID, "Updated", "in_progress")
			if err != nil {
				errChan <- err
			}
		}(i)
	}

	wg.Wait()
	close(errChan)

	for err := range errChan {
		t.Errorf("Concurrent update failed: %v", err)
	}

	// Verify the item still exists
	updated, err := actor.List(ctx, created.ID)
	if err != nil {
		t.Fatalf("List after updates failed: %v", err)
	}
	if updated.Description != "Updated" {
		t.Errorf("Expected description 'Updated', got '%s'", updated.Description)
	}
}

// TestActor_Concurrency_MixedOperations tests concurrent mixed operations.
func TestActor_Concurrency_MixedOperations(t *testing.T) {
	_, cleanup := setupTestStorage(t)
	defer cleanup()

	ctx := context.Background()
	actor := NewActor(ctx)

	// Create some initial items
	for i := 1; i <= 5; i++ {
		_, err := actor.Create(ctx, "Initial Item", "not_started")
		if err != nil {
			t.Fatalf("Create failed: %v", err)
		}
	}

	const numGoroutines = 30
	var wg sync.WaitGroup
	errChan := make(chan error, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			var err error

			switch index % 4 {
			case 0: // Create
				_, err = actor.Create(ctx, "New Item", "not_started")
			case 1: // Read
				_, err = actor.ListAll(ctx)
			case 2: // Update
				_, err = actor.Update(ctx, 1, "Updated", "in_progress")
			case 3: // Read single
				_, err = actor.List(ctx, 1)
			}

			if err != nil {
				errChan <- err
			}
		}(i)
	}

	wg.Wait()
	close(errChan)

	for err := range errChan {
		t.Errorf("Concurrent mixed operation failed: %v", err)
	}
}

// TestActor_Concurrency_SequentialOperations tests a sequence of operations.
func TestActor_Concurrency_SequentialOperations(t *testing.T) {
	_, cleanup := setupTestStorage(t)
	defer cleanup()

	ctx := context.Background()
	actor := NewActor(ctx)

	// Create
	created, err := actor.Create(ctx, "Sequential Test", "not_started")
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	// Read
	retrieved, err := actor.List(ctx, created.ID)
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if retrieved.Description != "Sequential Test" {
		t.Errorf("Expected 'Sequential Test', got '%s'", retrieved.Description)
	}

	// Update
	updated, err := actor.Update(ctx, created.ID, "Updated Sequential", "in_progress")
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}
	if updated.Description != "Updated Sequential" {
		t.Errorf("Expected 'Updated Sequential', got '%s'", updated.Description)
	}

	// List all
	items, err := actor.ListAll(ctx)
	if err != nil {
		t.Fatalf("ListAll failed: %v", err)
	}
	if len(items) == 0 {
		t.Error("Expected at least one item")
	}

	// Delete
	err = actor.Delete(ctx, created.ID)
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	// Verify deletion
	_, err = actor.List(ctx, created.ID)
	if err == nil {
		t.Error("Expected error after deletion, got nil")
	}
}
