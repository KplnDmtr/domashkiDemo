package middleware

import (
	"fmt"
	"net/http"

	"redditclone/pkg/author"
	"redditclone/pkg/key"
	"redditclone/pkg/logger"
	"redditclone/pkg/response"
	"redditclone/pkg/user"
)

func Authenticate(contextKey key.Key, logger logger.Logger, uRepo user.UserRepo, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("AUTHENTICATE MIDDLEWARE", r.URL.Path)
		author, ok := r.Context().Value(contextKey).(*author.Author)
		if !ok {
			logger.Log("Info", "Not in context")
			return
		}

		isUser, err := uRepo.IsUser(r.Context(), author.Username, author.ID)
		if err != nil {
			response.ServerResponseWriter(w, 401, map[string]interface{}{"message": "db error"})
			return
		}

		if isUser {
			next.ServeHTTP(w, r)
			return
		}

		response.ServerResponseWriter(w, 401, map[string]interface{}{"message": "user not exists"})
	})
}
