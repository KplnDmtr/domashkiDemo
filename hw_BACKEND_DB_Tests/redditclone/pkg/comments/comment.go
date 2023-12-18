package comments

import (
	"time"

	"redditclone/pkg/author"
)

type Comment struct {
	Created time.Time     `json:"created"`
	Author  author.Author `json:"author"`
	Body    string        `json:"body"`
	ID      string        `json:"id"`
}

type SimpleComment struct {
	Comment string
}
