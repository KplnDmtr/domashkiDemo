package posts

import (
	"context"
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
	GetAllPosts(ctx context.Context) ([]Post, error)
	AddPost(ctx context.Context, post Post) (Post, error)
	UpdatePost(ctx context.Context, post Post) error
	GetPostByID(ctx context.Context, id string) (Post, error)
	GetCategory(ctx context.Context, category string) ([]Post, error)
	AddComment(ctx context.Context, postID string, comment comments.Comment) (Post, error)
	DeleteComment(ctx context.Context, postID string, commentID string) (Post, error)
	DeletePost(ctx context.Context, postID string) error
	GetByUserLogin(ctx context.Context, login string) ([]Post, error)
	Vote(ctx context.Context, postID string, vote vote.Vote) (Post, error)
	UnVote(ctx context.Context, username string, postID string) (Post, error)
}
