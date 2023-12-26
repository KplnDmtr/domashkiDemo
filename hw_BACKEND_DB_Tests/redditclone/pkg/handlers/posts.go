package handlers

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"time"

	"redditclone/pkg/author"
	"redditclone/pkg/comments"
	"redditclone/pkg/key"
	"redditclone/pkg/logger"
	"redditclone/pkg/posts"
	"redditclone/pkg/response"
	"redditclone/pkg/session"
	"redditclone/pkg/vote"

	"github.com/gorilla/mux"
)

type PostsHandler struct {
	Session    session.SessionManager
	PostsRepo  posts.PostsRepository
	Logger     logger.Logger
	ContextKey key.Key
}

var dbError string = "DB error"

func (p *PostsHandler) All(w http.ResponseWriter, r *http.Request) {
	posts, err := p.PostsRepo.GetAllPosts(r.Context())
	if err != nil {
		p.Logger.Log("Error", err.Error())
		response.ServerResponseWriter(w, 500, map[string]interface{}{"message": dbError})
	}
	response.ServerResponseWriter(w, 200, posts)
}

func (p *PostsHandler) NewPost(w http.ResponseWriter, r *http.Request) {
	post := posts.Post{
		Score:            1,
		Created:          time.Now(),
		UpvotePercentage: 100,
		Comments:         make([]comments.Comment, 0),
	}

	js, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(500)
		return
	}

	err = json.Unmarshal(js, &post)
	if err != nil {
		w.WriteHeader(500)
		return
	}

	author, ok := r.Context().Value(p.ContextKey).(*author.Author)
	if !ok {
		w.WriteHeader(500)
		return
	}
	post.Author.ID = author.ID
	post.Author.Username = author.Username
	post.Votes = []vote.Vote{{
		User: post.Author.ID,
		Vote: 1,
	}}

	post, err = p.PostsRepo.AddPost(r.Context(), post)
	if err != nil {
		p.Logger.Log("Error", err.Error())
		response.ServerResponseWriter(w, 500, map[string]interface{}{"message": dbError})
		return
	}

	response.ServerResponseWriter(w, 201, post)
}

func (p *PostsHandler) GetPost(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["POST_ID"]
	if id == "" {
		w.WriteHeader(400)
		return
	}

	post, err := p.PostsRepo.GetPostByID(r.Context(), id)
	if err != nil {
		p.Logger.Log("Error", err.Error())
		response.ServerResponseWriter(w, 500, map[string]interface{}{"message": dbError})
		return
	}

	post.Views += 1

	err = p.PostsRepo.UpdatePost(r.Context(), post)
	if err != nil {
		p.Logger.Log("Error", err.Error())
		response.ServerResponseWriter(w, 500, map[string]interface{}{"message": dbError})
		return
	}

	response.ServerResponseWriter(w, 201, post)
}

func (p *PostsHandler) GetPostsByCategory(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	category := vars["CATEGORY_NAME"]
	if category == "" {
		w.WriteHeader(400)
		return
	}

	posts, err := p.PostsRepo.GetCategory(r.Context(), category)
	if err != nil {
		p.Logger.Log("Error", err.Error())
		response.ServerResponseWriter(w, 500, map[string]interface{}{"message": dbError})
		return
	}
	response.ServerResponseWriter(w, 200, posts)
}

func (p *PostsHandler) AddComment(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["POST_ID"]
	if id == "" {
		w.WriteHeader(400)
		return
	}

	author, ok := r.Context().Value(p.ContextKey).(*author.Author)
	if !ok {
		w.WriteHeader(500)
		return
	}

	js, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(500)
		return
	}

	simple := comments.SimpleComment{}

	err = json.Unmarshal(js, &simple)
	if err != nil {
		w.WriteHeader(500)
		return
	}

	newComm := comments.Comment{
		Author:  *author,
		Created: time.Now(),
		Body:    simple.Comment,
	}

	post, err := p.PostsRepo.AddComment(r.Context(), id, newComm)
	if err != nil {
		p.Logger.Log("Error", err.Error())
		response.ServerResponseWriter(w, 500, map[string]interface{}{"message": dbError})
		return
	}

	response.ServerResponseWriter(w, 201, post)
}

func (p *PostsHandler) DeleteComment(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	postID := vars["POST_ID"]
	commID := vars["COMMENT_ID"]
	if postID == "" || commID == "" {
		w.WriteHeader(400)
		return
	}
	post, err := p.PostsRepo.DeleteComment(r.Context(), postID, commID)
	if err != nil {
		p.Logger.Log("Error", err.Error())
		response.ServerResponseWriter(w, 500, map[string]interface{}{"message": dbError})
		return
	}
	response.ServerResponseWriter(w, 200, post)
}

func (p *PostsHandler) DeletePost(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	postID := vars["POST_ID"]
	if postID == "" {
		w.WriteHeader(400)
		return
	}
	err := p.PostsRepo.DeletePost(r.Context(), postID)
	if err != nil {
		p.Logger.Log("Error", err.Error())
		response.ServerResponseWriter(w, 500, map[string]interface{}{"message": dbError})
		return
	}
	response.ServerResponseWriter(w, 200, map[string]interface{}{"message": "success"})
}

func (p *PostsHandler) Vote(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)
	id := vars["POST_ID"]
	if id == "" {
		w.WriteHeader(400)
		return
	}

	author, ok := r.Context().Value(p.ContextKey).(*author.Author)
	if !ok {
		w.WriteHeader(500)
		return
	}

	newVote := vote.Vote{
		User: author.ID,
		Vote: 1,
	}

	paths := strings.Split(r.URL.Path, "/")
	voteType := paths[len(paths)-1]
	if voteType == "downvote" {
		newVote.Vote = -1
	}

	post, err := p.PostsRepo.Vote(r.Context(), id, newVote)
	if err != nil {
		p.Logger.Log("Error", err.Error())
		response.ServerResponseWriter(w, 500, map[string]interface{}{"message": dbError})
		return
	}

	response.ServerResponseWriter(w, 200, post)
}

func (p *PostsHandler) UnVote(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["POST_ID"]
	if id == "" {
		w.WriteHeader(400)
		return
	}

	author, ok := r.Context().Value(p.ContextKey).(*author.Author)
	if !ok {
		w.WriteHeader(500)
		return
	}

	post, err := p.PostsRepo.UnVote(r.Context(), author.ID, id)
	if err != nil {
		p.Logger.Log("Error", err.Error())
		response.ServerResponseWriter(w, 500, map[string]interface{}{"message": dbError})
		return
	}

	response.ServerResponseWriter(w, 200, post)
}
