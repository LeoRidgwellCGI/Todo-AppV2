package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"text/template"
	"todo-app/actor"
	"todo-app/storage"
)

// ActorInterface defines the methods required by handlers
type ActorInterface interface {
	Create(ctx context.Context, description string, status string) (storage.Item, error)
	Update(ctx context.Context, id int, description string, status string) (storage.Item, error)
	Delete(ctx context.Context, id int) error
	ListAll(ctx context.Context) (storage.Items, error)
	List(ctx context.Context, id int) (storage.Item, error)
}

var actorInstance ActorInterface

// InitActor initializes the actor instance.
func InitActor(ctx context.Context) {
	actorInstance = actor.NewActor(ctx)
}

// AddRoutes adds HTTP routes to the provided ServeMux.
func AddRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/create", createItemHandler)
	mux.HandleFunc("/update", updateItemHandler)
	mux.HandleFunc("/delete", deleteItemHandler)
	mux.HandleFunc("/get/{itemid}", getByIDHandler)
	mux.HandleFunc("/get", getListHandler)
	mux.HandleFunc("/list", dynamicListHandler)

	mux.Handle("/about/", http.StripPrefix("/about/", http.FileServer(http.Dir("static/about"))))
	mux.HandleFunc("/about", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/about/", http.StatusMovedPermanently)
	})
}

// getListHandler handles requests to retrieve all todo items.
func getListHandler(w http.ResponseWriter, r *http.Request) {
	if actorInstance == nil {
		http.Error(w, "Actor not initialized", http.StatusInternalServerError)
		return
	}
	items, err := actorInstance.ListAll(context.Background())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	todos := make([]storage.Item, 0, len(items))
	for _, v := range items {
		todos = append(todos, storage.Item{
			ID:          v.ID,
			Description: v.Description,
			Status:      v.Status,
			Created:     v.Created,
		})
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(todos)
}

// getByIDHandler handles requests to retrieve a todo item by ID.
func getByIDHandler(w http.ResponseWriter, r *http.Request) {
	if actorInstance == nil {
		http.Error(w, "Actor not initialized", http.StatusInternalServerError)
		return
	}
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 3 {
		http.Error(w, "Missing item ID", http.StatusBadRequest)
		return
	}
	id, err := strconv.Atoi(parts[2])
	if err != nil {
		http.Error(w, "Invalid item ID", http.StatusBadRequest)
		return
	}
	item, err := actorInstance.List(context.Background(), id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(storage.Item{
		ID:          item.ID,
		Description: item.Description,
		Status:      item.Status,
		Created:     item.Created,
	})
}

// createItemHandler handles requests to create a new todo item.
func createItemHandler(w http.ResponseWriter, r *http.Request) {
	if actorInstance == nil {
		http.Error(w, "Actor not initialized", http.StatusInternalServerError)
		return
	}
	var todo storage.Item
	err := json.NewDecoder(r.Body).Decode(&todo)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	item, err := actorInstance.Create(context.Background(), todo.Description, todo.Status)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(storage.Item{
		ID:          item.ID,
		Description: item.Description,
		Status:      item.Status,
		Created:     item.Created,
	})
}

// updateItemHandler handles requests to update an existing todo item.
func updateItemHandler(w http.ResponseWriter, r *http.Request) {
	if actorInstance == nil {
		http.Error(w, "Actor not initialized", http.StatusInternalServerError)
		return
	}
	var todo storage.Item
	err := json.NewDecoder(r.Body).Decode(&todo)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	item, err := actorInstance.Update(context.Background(), todo.ID, todo.Description, todo.Status)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(storage.Item{
		ID:          item.ID,
		Description: item.Description,
		Status:      item.Status,
		Created:     item.Created,
	})
}

// deleteItemHandler handles requests to delete a todo item by ID.
func deleteItemHandler(w http.ResponseWriter, r *http.Request) {
	if actorInstance == nil {
		http.Error(w, "Actor not initialized", http.StatusInternalServerError)
		return
	}
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 3 {
		http.Error(w, "Missing item ID", http.StatusBadRequest)
		return
	}
	id, err := strconv.Atoi(parts[2])
	if err != nil {
		http.Error(w, "Invalid item ID", http.StatusBadRequest)
		return
	}
	err = actorInstance.Delete(context.Background(), id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"deleted": id})
}

// dynamicListHandler handles requests to retrieve all todo items.
func dynamicListHandler(w http.ResponseWriter, r *http.Request) {
	const listTemplate = "<!doctype html><html><head><meta charset=\"utf-8\"><title>Todos</title><style>body{font-family:Arial,sans-serif;margin:2em;background:#f9f9f9;}h1{color: #007acc;}p{max-width:600px;}ul{display:table;border-collapse:collapse;width:100%;padding:0;margin:0;}ul li{display:table-row;}ul li span{display:table-cell;border:1px solid #007acc;padding:8px;text-align:left;}ul li.header span{font-weight:bold;background-color: #007acc;color: #ffffff;}</style></head><body><h1>Todos</h1><ul><li class='header'><span>ID</span><span>Description</span><span>Status</span></li>{{range .Items}}<li><span>{{.ID}}</span><span>{{.Description}}</span><span>{{.Status}}</span></li>{{else}}<li><span colspan=\"3\">none</span></li>{{end}}</ul></body></html>"
	list, err := actorInstance.ListAll(context.Background())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	tpl := template.Must(template.New("list").Parse(listTemplate))
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_ = tpl.Execute(w, struct{ Items storage.Items }{Items: list})
}
