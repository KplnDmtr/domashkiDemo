package middleware

import (
	"fmt"
	"net/http"

	"redditclone/pkg/logger"
)

func Panic(log logger.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("panicMiddleware", r.URL.Path)
		defer func() {
			if err := recover(); err != nil {
				fieldsMap := map[string]interface{}{
					"panic": err,
				}
				log.LogW("Panic", "recovered: ", fieldsMap)
				http.Error(w, "Internal server error", 500)
			}
		}()
		next.ServeHTTP(w, r)
	})
}
