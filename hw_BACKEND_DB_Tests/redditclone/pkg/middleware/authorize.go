package middleware

import (
	"fmt"
	"net/http"

	"redditclone/pkg/author"
	"redditclone/pkg/key"
	"redditclone/pkg/logger"
	"redditclone/pkg/posts"
	"redditclone/pkg/response"

	"github.com/gorilla/mux"
)

func Authorize(contextKey key.Key, logger logger.Logger, pRepo posts.PostsRepository, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("AuthorizeMiddleware", r.URL.Path)
		author, ok := r.Context().Value(contextKey).(*author.Author)
		if !ok {
			logger.Log("Info", "not in context")
			return
		}
		vars := mux.Vars(r)
		postID := vars["POST_ID"]
		commentID := vars["COMMENT_ID"]
		if commentID != "" && postID != "" {
			err := CheckComment(postID, commentID, *author, pRepo)
			if err != nil {
				logger.Log("Error", err.Error())
				response.ServerResponseWriter(w, 400, err.Error())
				return
			}
			next.ServeHTTP(w, r)
			return
		}
		if postID != "" {
			err := CheckPost(postID, *author, pRepo)
			if err != nil {
				logger.Log("Error", err.Error())
				response.ServerResponseWriter(w, 400, err.Error())
				return
			}
			next.ServeHTTP(w, r)
			return
		}
		response.ServerResponseWriter(w, 400, map[string]interface{}{"message": "bad auth"})
	})
}

func CheckComment(postID string, commentID string, author author.Author, pRepo posts.PostsRepository) error {
	post, err := pRepo.GetPostByID(postID)
	if err != nil {
		return err
	}
	for _, comment := range post.Comments {
		if comment.ID == commentID && comment.Author.ID == author.ID && comment.Author.Username == author.Username {
			return nil
		}
	}
	return fmt.Errorf("comment not found")
}

func CheckPost(postID string, author author.Author, pRepo posts.PostsRepository) error {
	post, err := pRepo.GetPostByID(postID)
	if err != nil {
		return err
	}
	if post.Author.ID == author.ID && post.Author.Username == author.Username {
		return nil
	}
	return fmt.Errorf("post not found")
}
