package main

import (
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
)

func main() {
	repo := NewNotesMemory()
	logger := log.New(os.Stdout, "note app: ", log.Ldate|log.Ltime)
	handler := Handler{
		Repo: repo,
		Log:  logger,
	}
	// test comment2
	r := mux.NewRouter()
	r.HandleFunc("/note/{id}", handler.GetNoteByID).Methods("GET")
	r.HandleFunc("/note", handler.NewNote).Methods("POST")
	r.HandleFunc("/note/{id}", handler.UpdateNote).Methods("PUT")
	r.HandleFunc("/note/{id}", handler.DeleteNote).Methods("DELETE")
	r.HandleFunc("/note", handler.GetNotes).Methods("GET")

	err := http.ListenAndServe(":8080", r)
	if err != nil {
		panic("server didn't start")
	}
}
