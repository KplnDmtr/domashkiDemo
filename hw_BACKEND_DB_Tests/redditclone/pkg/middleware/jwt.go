package middleware

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"redditclone/pkg/author"
	"redditclone/pkg/key"
	"redditclone/pkg/logger"
	"redditclone/pkg/response"
	"redditclone/pkg/session"

	"github.com/dgrijalva/jwt-go"
)

func JWT(contextKey key.Key, logger logger.Logger, sess session.SessionManager, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("JWTMiddleware", r.URL.Path)
		token := r.Header.Get("Authorization")
		if token == "" {
			response.ServerResponseWriter(w, 400, map[string]interface{}{"message": "token not found"})
			return
		}

		token = strings.TrimPrefix(token, "Bearer ")
		hashSecretGetter := func(token *jwt.Token) (interface{}, error) {
			method, ok := token.Method.(*jwt.SigningMethodHMAC)
			if !ok || method.Alg() != "HS256" {
				return nil, fmt.Errorf("bad sign method")
			}
			return sess.GetKey(), nil
		}

		tkn, err := jwt.Parse(token, hashSecretGetter)
		if err != nil || !tkn.Valid {
			response.ServerResponseWriter(w, 400, map[string]interface{}{"message": "bad token"})
			return
		}

		payload, ok := tkn.Claims.(jwt.MapClaims)
		if !ok {
			response.ServerResponseWriter(w, 400, map[string]interface{}{"message": "bad token"})
			return
		}

		iat := int64(payload["iat"].(float64))
		exp := int64(payload["exp"].(float64))
		curTime := time.Now().Unix()
		if curTime <= iat || curTime >= exp {
			response.ServerResponseWriter(w, 400, map[string]interface{}{"message": "token expired or has incorrect time"})
			return
		}

		var author author.Author
		authorData := payload["user"].(map[string]interface{})
		author.ID = authorData["id"].(string)
		author.Username = authorData["username"].(string)

		timeExp := sess.GetExp(r.Context(), author.ID, iat)
		if curTime >= timeExp {
			err = sess.DeleteSess(r.Context(), author.ID, iat)
			if err != nil {
				logger.Log("Error", err.Error())
				response.ServerResponseWriter(w, 400, map[string]interface{}{"message": "db error"})
			}
			response.ServerResponseWriter(w, 400, map[string]interface{}{"message": "token expired or has incorrect time"})
			return
		}

		ctx := r.Context()
		ctx = context.WithValue(ctx,
			contextKey,
			&author,
		)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
