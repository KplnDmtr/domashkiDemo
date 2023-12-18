package handlers

import (
	"encoding/json"
	"io"
	"net/http"
	"time"

	"redditclone/pkg/key"
	"redditclone/pkg/logger"
	"redditclone/pkg/posts"
	"redditclone/pkg/response"
	"redditclone/pkg/session"
	"redditclone/pkg/user"

	"github.com/dgrijalva/jwt-go"
	"github.com/gorilla/mux"
)

type UserHandler struct {
	Logger     logger.Logger
	UserRepo   user.UserRepo
	Session    session.SessionManager
	PostsRepo  posts.PostsRepository
	ContextKey key.Key
}

func (u *UserHandler) makeToken(user user.User, iat int64) (string, error) {
	mp := make(map[string]string, 2)
	mp["username"] = user.Username
	mp["id"] = user.ID
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user": mp,
		"iat":  iat,
		"exp":  u.Session.GetExp(user.ID, iat),
	})
	tokenString, err := token.SignedString(u.Session.GetKey())
	if err != nil {
		return "", err
	}
	return tokenString, nil
}

func (u *UserHandler) SignIn(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(500)
		return
	}

	var user user.User
	if err = json.Unmarshal(body, &user); err != nil {
		w.WriteHeader(500)
		return
	}

	user.ID, err = u.UserRepo.AddNewUser(user)
	if err != nil {
		u.Logger.Log("Error", err.Error())
		response.ServerResponseWriter(w, 500, map[string]interface{}{"message": "user already exists"})
		return
	}

	iat := time.Now().Unix()
	err = u.Session.AddNewSess(user.ID, time.Now().Add(120*time.Hour).Unix(), iat)
	if err != nil {
		u.Logger.Log("Error", err.Error())
		w.WriteHeader(500)
		return
	}

	tokenString, err := u.makeToken(user, iat)
	if err != nil {
		w.WriteHeader(500)
		return
	}

	response.ServerResponseWriter(w, 201, map[string]interface{}{"token": tokenString})
}

func (u *UserHandler) LogIn(w http.ResponseWriter, r *http.Request) {

	body, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(500)
		return
	}

	var user user.User
	if err = json.Unmarshal(body, &user); err != nil {
		w.WriteHeader(500)
		return
	}

	user.ID, err = u.UserRepo.Authenticate(user)
	if err != nil {
		u.Logger.Log("Error", err.Error())
		response.ServerResponseWriter(w, 500, map[string]interface{}{"message": "password or login not right"})
		return
	}

	iat := time.Now().Unix()
	err = u.Session.AddNewSess(user.ID, time.Now().Add(120*time.Hour).Unix(), iat)
	if err != nil {
		w.WriteHeader(500)
		return
	}

	tokenString, err := u.makeToken(user, iat)
	if err != nil {
		w.WriteHeader(500)
		return
	}

	response.ServerResponseWriter(w, 201, map[string]interface{}{"token": tokenString})
}

func (u *UserHandler) GetUserPosts(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userLogin := vars["USER_LOGIN"]
	if userLogin == "" {
		w.WriteHeader(400)
		return
	}
	posts, err := u.PostsRepo.GetByUserLogin(userLogin)
	if err != nil {
		u.Logger.Log("Info", err.Error())
		response.ServerResponseWriter(w, 500, map[string]interface{}{"message": dbError})
		return
	}
	response.ServerResponseWriter(w, 200, posts)
}
