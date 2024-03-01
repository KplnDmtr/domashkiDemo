package main

import (
	"fmt"
	"sync"
)

type NotesMemory struct {
	Notes map[int]Note
	mu    *sync.Mutex
	curID int
}

func NewNotesMemory() *NotesMemory {
	return &NotesMemory{
		Notes: make(map[int]Note),
		mu:    &sync.Mutex{},
		curID: 0,
	}
}

func (n *NotesMemory) Add(note Note) (int, error) {
	n.mu.Lock()
	defer n.mu.Unlock()
	n.curID++
	note.ID = n.curID
	if _, ok := n.Notes[n.curID]; ok {
		n.curID--
		return -1, fmt.Errorf("already exist")
	}
	n.Notes[n.curID] = note
	return n.curID, nil
}

func (n *NotesMemory) Delete(id int) error {
	n.mu.Lock()
	defer n.mu.Unlock()
	if _, ok := n.Notes[id]; !ok {
		return fmt.Errorf("not exist")
	}
	delete(n.Notes, id)
	return nil
}

func (n *NotesMemory) GetOne(id int) (*Note, error) {
	n.mu.Lock()
	defer n.mu.Unlock()
	if _, ok := n.Notes[id]; !ok {
		return nil, fmt.Errorf("not exist")
	}
	note := &Note{}
	*note = n.Notes[id]
	return note, nil
}

func (n *NotesMemory) Get() ([]Note, error) {
	n.mu.Lock()
	notes := make([]Note, 0)
	for _, val := range n.Notes {
		notes = append(notes, val)
	}
	n.mu.Unlock()
	return notes, nil
}

func (n *NotesMemory) Update(id int, note Note) error {
	n.mu.Lock()
	defer n.mu.Unlock()
	if _, ok := n.Notes[id]; !ok {
		return fmt.Errorf("not exist")
	}
	n.Notes[id] = note
	return nil
}
