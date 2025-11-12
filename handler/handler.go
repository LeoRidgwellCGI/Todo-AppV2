package handler

import (
	"fmt"
	"net/http"
	"todo-app/actor"
)

func addRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/create", createItemHandler)
	mux.HandleFunc("/update", updateItemHandler)
	mux.HandleFunc("/delete", deleteItemHandler)
	mux.HandleFunc("/get/{itemid}", getByIDHandler)
	mux.HandleFunc("/get", getListHandler)
	mux.HandleFunc("/about", aboutPageHandler)

	mux.HandleFunc("/list", dynamicListHandler)
}

func getListHandler(w http.ResponseWriter, r *http.Request) {
	items := actor.ListAll()
	fmt.Println(items)
}

func getByIDHandler(w http.ResponseWriter, r *http.Request) {
	id := 1
	item := actor.List(id)
	fmt.Println(item)
}

func createItemHandler(w http.ResponseWriter, r *http.Request) {
	description := "Sample Task"
	status := "not_started"
	actor.Create(description, status)
}

func updateItemHandler(w http.ResponseWriter, r *http.Request) {
	id := 1
	description := "Updated Task"
	status := "has_started"
	actor.Update(id, description, status)
}

func deleteItemHandler(w http.ResponseWriter, r *http.Request) {
	id := 1
	actor.Delete(id)
}

func aboutPageHandler(w http.ResponseWriter, r *http.Request) {

}

func dynamicListHandler(w http.ResponseWriter, r *http.Request) {
}
