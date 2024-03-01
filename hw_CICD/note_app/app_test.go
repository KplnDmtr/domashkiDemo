package main

import (
	"encoding/json"
	"io"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
)

func TestGetNoteByID(t *testing.T) {
	repo := NewNotesMemory()
	note := Note{
		ID:   1,
		Text: "asdasd",
	}
	_, err := repo.Add(note)
	if err != nil {
		t.Errorf("unexpected error %s", err.Error())
	}
	handler := Handler{
		Repo: repo,
	}

	req := httptest.NewRequest("GET", "/note/1", nil)
	w := httptest.NewRecorder()
	handler.GetNoteByID(w, req)
	resp := w.Result()
	defer resp.Body.Close()
	if resp.StatusCode != 500 {
		t.Errorf("Expected 500 got %d", resp.StatusCode)
	}

	req = httptest.NewRequest("GET", "/note/1", nil)
	req = mux.SetURLVars(req, map[string]string{"id": "1"})
	w = httptest.NewRecorder()
	handler.GetNoteByID(w, req)

	resp = w.Result()
	body, err := io.ReadAll(resp.Body)
	defer resp.Body.Close()
	if err != nil {
		t.Errorf("unexpected error %s", err.Error())
	}

	noteFromTest := Note{}
	err = json.Unmarshal(body, &noteFromTest)
	if err != nil {
		t.Errorf("unexpected error %s", err.Error())
	}
	assert.Equal(t, note, noteFromTest)

}
