package main

import "time"

type Note struct {
	ID        int
	Text      string
	CreatedAt time.Time
	UpdatedAt time.Time
}

type NoteRepository interface {
	GetOne(id int) (*Note, error)
	Get() ([]Note, error)
	Update(id int, note Note) error
	Delete(id int) error
	Add(note Note) (int, error)
}
