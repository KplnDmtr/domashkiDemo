package main

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"sort"
	"strconv"
	"time"

	"github.com/gorilla/mux"
)

type Handler struct {
	Repo NoteRepository
	Log  *log.Logger
}

func (h *Handler) GetNoteByID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	noteID, err := strconv.Atoi(vars["id"])
	if err != nil {
		w.WriteHeader(500)
		return
	}
	note, err := h.Repo.GetOne(noteID)
	if err != nil {
		h.Log.Println(err.Error())
		w.WriteHeader(500)
		return
	}
	ServerResponseWriter(w, 200, note)
}

func (h *Handler) NewNote(w http.ResponseWriter, r *http.Request) {
	note := Note{}
	js, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(500)
		return
	}
	err = json.Unmarshal(js, &note)
	if err != nil {
		w.WriteHeader(500)
		return
	}
	note.CreatedAt = time.Now()
	note.UpdatedAt = time.Now()
	note.ID, err = h.Repo.Add(note)
	if err != nil {
		h.Log.Println(err.Error())
		w.WriteHeader(500)
		return
	}
	ServerResponseWriter(w, 200, note)
}

func (h *Handler) UpdateNote(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	noteID, err := strconv.Atoi(vars["id"])
	if err != nil {
		w.WriteHeader(500)
		return
	}
	note := Note{}
	js, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(500)
		return
	}
	err = json.Unmarshal(js, &note)
	if err != nil {
		w.WriteHeader(500)
		return
	}
	note.ID = noteID
	note.UpdatedAt = time.Now()
	err = h.Repo.Update(noteID, note)
	if err != nil {
		h.Log.Println(err.Error())
		w.WriteHeader(500)
		return
	}
	w.WriteHeader(200)
}

func (h *Handler) DeleteNote(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	noteID, err := strconv.Atoi(vars["id"])
	if err != nil {
		w.WriteHeader(500)
		return
	}
	err = h.Repo.Delete(noteID)
	if err != nil {
		h.Log.Println(err.Error())
		w.WriteHeader(500)
		return
	}
	w.WriteHeader(200)
}

func (h *Handler) sorting(notes []Note, orderBy string) {
	switch orderBy {
	case "id":
		sort.Slice(notes, func(i, j int) bool {
			return notes[i].ID < notes[j].ID
		})
		return
	case "text":
		sort.Slice(notes, func(i, j int) bool {
			return notes[i].Text < notes[j].Text
		})
		return
	case "created_at":
		sort.Slice(notes, func(i, j int) bool {
			ans := notes[i].CreatedAt.Compare(notes[j].CreatedAt)
			return ans < 0
		})
		return
	case "updated_at":
		sort.Slice(notes, func(i, j int) bool {
			ans := notes[i].UpdatedAt.Compare(notes[j].UpdatedAt)
			return ans > 0
		})
		return
	default:
		return
	}
}

func (h *Handler) GetNotes(w http.ResponseWriter, r *http.Request) {
	notes, err := h.Repo.Get()
	if err != nil {
		h.Log.Println(err.Error())
		w.WriteHeader(500)
		return
	}
	queryParams := r.URL.Query()
	sortingFeature := queryParams.Get("order_by")
	h.sorting(notes, sortingFeature)
	ServerResponseWriter(w, 200, notes)
}
