package handler

import (
	"net/http"
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

}

func getByIDHandler(w http.ResponseWriter, r *http.Request) {

}

func createItemHandler(w http.ResponseWriter, r *http.Request) {
}

func updateItemHandler(w http.ResponseWriter, r *http.Request) {

}

func deleteItemHandler(w http.ResponseWriter, r *http.Request) {
}

func aboutPageHandler(w http.ResponseWriter, r *http.Request) {
}

func dynamicListHandler(w http.ResponseWriter, r *http.Request) {
}
