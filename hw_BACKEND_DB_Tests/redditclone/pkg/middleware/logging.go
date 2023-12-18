package middleware

import (
	"fmt"
	"net/http"
	"time"

	"redditclone/pkg/logger"
)

func Logging(logger logger.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("access log middleware")
		start := time.Now()
		next.ServeHTTP(w, r)
		fieldsMap := map[string]interface{}{
			"method":      r.Method,
			"remote_addr": r.RemoteAddr,
			"url":         r.URL.Path,
			"time":        time.Since(start),
		}
		logger.LogW("Info", "new request with params: ", fieldsMap)
	})
}
