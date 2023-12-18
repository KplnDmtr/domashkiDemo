package posts

import (
	"time"

	"redditclone/pkg/author"
	"redditclone/pkg/comments"
	"redditclone/pkg/vote"
)

type Post struct {
	Score            int                `json:"score" bson:"score"`
	Views            int                `json:"views" bson:"views"`
	Type             string             `json:"type" bson:"type"`
	Title            string             `json:"title" bson:"title"`
	Author           author.Author      `json:"author" bson:"author"`
	Category         string             `json:"category" bson:"category"`
	URL              string             `json:"url,omitempty" bson:"url,omitempty"`
	Text             string             `json:"text,omitempty" bson:"text,omitempty"`
	Votes            []vote.Vote        `json:"votes" bson:"votes"`
	Comments         []comments.Comment `json:"comments" bson:"comments"`
	Created          time.Time          `json:"created" bson:"created"`
	UpvotePercentage int                `json:"upvotePercentage" bson:"upvotePercentage"`
	ID               string             `json:"id" bson:"_id"`
	Upvotes          int                `json:"-" bson:"upvotes"`
}

//go:generate mockgen -source posts.go -destination posts_mock.go -package posts PostsRepository
type PostsRepository interface {
	GetAllPosts() ([]Post, error)
	AddPost(Post) (Post, error)
	UpdatePost(Post) error
	GetPostByID(string) (Post, error)
	GetCategory(string) ([]Post, error)
	AddComment(string, comments.Comment) (Post, error)
	DeleteComment(string, string) (Post, error)
	DeletePost(string) error
	GetByUserLogin(string) ([]Post, error)
	Vote(string, vote.Vote) (Post, error)
	UnVote(string, string) (Post, error)
}
