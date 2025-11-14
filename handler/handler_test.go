package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"todo-app/storage"
)

// mockActor implements actor interface for testing.
type mockActor struct {
	items map[int]storage.Item
}

// ListAll returns all items.
func (m *mockActor) ListAll(ctx context.Context) (storage.Items, error) {
	result := make(storage.Items)
	for k, v := range m.items {
		result[k] = v
	}
	return result, nil
}

// List returns item by ID.
func (m *mockActor) List(ctx context.Context, id int) (storage.Item, error) {
	item, ok := m.items[id]
	if !ok {
		return storage.Item{}, errors.New("not found")
	}
	return item, nil
}

// Create creates a new item.
func (m *mockActor) Create(ctx context.Context, desc, status string) (storage.Item, error) {
	id := len(m.items) + 1
	item := storage.Item{ID: id, Description: desc, Status: status}
	m.items[id] = item
	return item, nil
}

// Update updates an existing item.
func (m *mockActor) Update(ctx context.Context, id int, desc, status string) (storage.Item, error) {
	item, ok := m.items[id]
	if !ok {
		return storage.Item{}, errors.New("not found")
	}
	item.Description = desc
	item.Status = status
	m.items[id] = item
	return item, nil
}

// Delete deletes an item by ID.
func (m *mockActor) Delete(ctx context.Context, id int) error {
	if _, ok := m.items[id]; !ok {
		return errors.New("not found")
	}
	delete(m.items, id)
	return nil
}

// setupMockActor initializes the mock actor for testing.
func setupMockActor() {
	mock := &mockActor{items: map[int]storage.Item{
		1: {ID: 1, Description: "Test", Status: "open"},
	}}
	actorInstance = mock
}

// TestHandler_GetListHandler tests the getListHandler function.
func TestHandler_GetListHandler(t *testing.T) {
	setupMockActor()
	req := httptest.NewRequest("GET", "/get", nil)
	w := httptest.NewRecorder()
	getListHandler(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var items []storage.Item
	if err := json.NewDecoder(w.Body).Decode(&items); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if len(items) != 1 || items[0].ID != 1 {
		t.Errorf("unexpected items: %+v", items)
	}
}

// TestHandler_GetByIDHandler tests the getByIDHandler function.
func TestHandler_GetByIDHandler(t *testing.T) {
	setupMockActor()
	req := httptest.NewRequest("GET", "/get/1", nil)
	w := httptest.NewRecorder()
	getByIDHandler(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var item storage.Item
	if err := json.NewDecoder(w.Body).Decode(&item); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if item.ID != 1 {
		t.Errorf("unexpected item: %+v", item)
	}
}

// TestHandler_GetByIDHandler_NotFound tests getByIDHandler for non-existent ID.
func TestHandler_GetByIDHandler_NotFound(t *testing.T) {
	setupMockActor()
	req := httptest.NewRequest("GET", "/get/999", nil)
	w := httptest.NewRecorder()
	getByIDHandler(w, req)
	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", w.Code)
	}
}

// TestHandler_CreateItemHandler tests the createItemHandler function.
func TestHandler_CreateItemHandler(t *testing.T) {
	setupMockActor()
	body := `{"Description":"New","Status":"open"}`
	req := httptest.NewRequest("POST", "/create", strings.NewReader(body))
	w := httptest.NewRecorder()
	createItemHandler(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var item storage.Item
	if err := json.NewDecoder(w.Body).Decode(&item); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if item.Description != "New" {
		t.Errorf("unexpected item: %+v", item)
	}
}

// TestHandler_UpdateItemHandler tests the updateItemHandler function.
func TestHandler_UpdateItemHandler(t *testing.T) {
	setupMockActor()
	body := `{"ID":1,"Description":"Updated","Status":"done"}`
	req := httptest.NewRequest("PUT", "/update", strings.NewReader(body))
	w := httptest.NewRecorder()
	updateItemHandler(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var item storage.Item
	if err := json.NewDecoder(w.Body).Decode(&item); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if item.Description != "Updated" || item.Status != "done" {
		t.Errorf("unexpected item: %+v", item)
	}
}

// TestHandler_DeleteItemHandler tests the deleteItemHandler function.
func TestHandler_DeleteItemHandler(t *testing.T) {
	setupMockActor()
	req := httptest.NewRequest("DELETE", "/delete/1", nil)
	w := httptest.NewRecorder()
	deleteItemHandler(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var resp map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if resp["deleted"] != float64(1) {
		t.Errorf("unexpected response: %+v", resp)
	}
}

// TestHandler_DeleteItemHandler_NotFound tests deleteItemHandler for non-existent ID.
func TestHandler_DeleteItemHandler_NotFound(t *testing.T) {
	setupMockActor()
	req := httptest.NewRequest("DELETE", "/delete/999", nil)
	w := httptest.NewRecorder()
	deleteItemHandler(w, req)
	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", w.Code)
	}
}

// TestHandler_DynamicListHandler tests the dynamicListHandler function.
func TestHandler_DynamicListHandler(t *testing.T) {
	setupMockActor()
	req := httptest.NewRequest("GET", "/list", nil)
	w := httptest.NewRecorder()
	dynamicListHandler(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "<title>Todos</title>") {
		t.Errorf("unexpected html: %s", w.Body.String())
	}
}

// TestHandler_ActorNotInitialized tests handler behavior when actor is not initialized.
func TestHandler_ActorNotInitialized(t *testing.T) {
	actorInstance = nil
	req := httptest.NewRequest("GET", "/get", nil)
	w := httptest.NewRecorder()
	getListHandler(w, req)
	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", w.Code)
	}
}

// TestHandler_Concurrency_MultipleReads tests concurrent read operations.
func TestHandler_Concurrency_MultipleReads(t *testing.T) {
	setupMockActor()

	const numGoroutines = 50
	done := make(chan bool, numGoroutines)
	errChan := make(chan error, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func() {
			req := httptest.NewRequest("GET", "/get", nil)
			w := httptest.NewRecorder()
			getListHandler(w, req)

			if w.Code != http.StatusOK {
				errChan <- nil
			} else {
				var items []storage.Item
				if err := json.NewDecoder(w.Body).Decode(&items); err != nil {
					errChan <- err
				} else {
					errChan <- nil
				}
			}
			done <- true
		}()
	}

	// Wait for all goroutines to complete
	for i := 0; i < numGoroutines; i++ {
		<-done
		if err := <-errChan; err != nil {
			t.Errorf("concurrent read failed: %v", err)
		}
	}
}

// TestHandler_Concurrency_MultipleWrites tests concurrent write operations.
func TestHandler_Concurrency_MultipleWrites(t *testing.T) {
	setupMockActor()

	const numGoroutines = 20
	done := make(chan bool, numGoroutines)
	errChan := make(chan error, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func(index int) {
			body := `{"Description":"Concurrent Task","Status":"open"}`
			req := httptest.NewRequest("POST", "/create", strings.NewReader(body))
			w := httptest.NewRecorder()
			createItemHandler(w, req)

			if w.Code != http.StatusOK {
				errChan <- errors.New("create failed")
			} else {
				var item storage.Item
				if err := json.NewDecoder(w.Body).Decode(&item); err != nil {
					errChan <- err
				} else {
					errChan <- nil
				}
			}
			done <- true
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < numGoroutines; i++ {
		<-done
		if err := <-errChan; err != nil {
			t.Errorf("concurrent write failed: %v", err)
		}
	}
}

// TestHandler_Concurrency_MixedOperations tests concurrent mixed read/write operations.
func TestHandler_Concurrency_MixedOperations(t *testing.T) {
	setupMockActor()

	const numGoroutines = 30
	done := make(chan bool, numGoroutines)
	errChan := make(chan error, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func(index int) {
			var req *http.Request
			var w *httptest.ResponseRecorder

			switch index % 3 {
			case 0: // Read operation
				req = httptest.NewRequest("GET", "/get/1", nil)
				w = httptest.NewRecorder()
				getByIDHandler(w, req)
				if w.Code != http.StatusOK {
					errChan <- errors.New("get failed")
				} else {
					errChan <- nil
				}
			case 1: // Create operation
				body := `{"Description":"Mixed Task","Status":"open"}`
				req = httptest.NewRequest("POST", "/create", strings.NewReader(body))
				w = httptest.NewRecorder()
				createItemHandler(w, req)
				if w.Code != http.StatusOK {
					errChan <- errors.New("create failed")
				} else {
					errChan <- nil
				}
			case 2: // List operation
				req = httptest.NewRequest("GET", "/get", nil)
				w = httptest.NewRecorder()
				getListHandler(w, req)
				if w.Code != http.StatusOK {
					errChan <- errors.New("list failed")
				} else {
					errChan <- nil
				}
			}
			done <- true
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < numGoroutines; i++ {
		<-done
		if err := <-errChan; err != nil {
			t.Errorf("concurrent mixed operation failed: %v", err)
		}
	}
}

// TestHandler_Concurrency_UpdateAndRead tests concurrent update and read operations on same item.
func TestHandler_Concurrency_UpdateAndRead(t *testing.T) {
	setupMockActor()

	const numGoroutines = 40
	done := make(chan bool, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func(index int) {
			if index%2 == 0 {
				// Update operation
				body := `{"ID":1,"Description":"Updated Concurrent","Status":"done"}`
				req := httptest.NewRequest("PUT", "/update", strings.NewReader(body))
				w := httptest.NewRecorder()
				updateItemHandler(w, req)
			} else {
				// Read operation
				req := httptest.NewRequest("GET", "/get/1", nil)
				w := httptest.NewRecorder()
				getByIDHandler(w, req)
			}
			done <- true
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < numGoroutines; i++ {
		<-done
	}

	// Verify the item still exists and is accessible
	req := httptest.NewRequest("GET", "/get/1", nil)
	w := httptest.NewRecorder()
	getByIDHandler(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected item to exist after concurrent updates, got status %d", w.Code)
	}
}

// TestHandler_Concurrency_DeleteAndRead tests concurrent delete and read operations.
func TestHandler_Concurrency_DeleteAndRead(t *testing.T) {
	setupMockActor()

	const numReads = 10
	done := make(chan bool, numReads+1)

	// Start read operations
	for i := 0; i < numReads; i++ {
		go func() {
			req := httptest.NewRequest("GET", "/get/1", nil)
			w := httptest.NewRecorder()
			getByIDHandler(w, req)
			// Don't fail on 404 as delete might have succeeded
			done <- true
		}()
	}

	// Start one delete operation
	go func() {
		req := httptest.NewRequest("DELETE", "/delete/1", nil)
		w := httptest.NewRecorder()
		deleteItemHandler(w, req)
		done <- true
	}()

	// Wait for all goroutines to complete
	for i := 0; i < numReads+1; i++ {
		<-done
	}
}
