package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"

	"redditclone/pkg/handlers"
	"redditclone/pkg/key"
	"redditclone/pkg/logger"
	"redditclone/pkg/middleware"
	"redditclone/pkg/posts"
	"redditclone/pkg/session"
	"redditclone/pkg/user"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// 05_web_app\99_hw\redditclone\cmd\redditclone\main.go
func main() {
	zapConfig := zap.Config{
		Level:            zap.NewAtomicLevelAt(zapcore.InfoLevel),
		Development:      true,
		Encoding:         "console",
		EncoderConfig:    zap.NewProductionEncoderConfig(),
		OutputPaths:      []string{"stdout"},
		ErrorOutputPaths: []string{"stderr"},
	}
	logger, err := logger.NewCustomLogger(zapConfig)
	if err != nil {
		panic(err.Error())
	}

	dsn := "root:1234@tcp(localhost:3306)/redditclone?"
	dsn += "charset=utf8"
	dsn += "&interpolateParams=true"

	db, err := sql.Open("mysql", dsn)

	db.SetMaxOpenConns(10)

	err = db.Ping()
	if err != nil {
		panic(err)
	}
	fmt.Println("Connected to SQL!")

	// Mongo connection
	connectionString := "mongodb://localhost:27017"
	connection, err := mongo.Connect(context.TODO(), options.Client().ApplyURI(connectionString))
	if err != nil {
		panic(err)
	}
	err = connection.Ping(context.TODO(), nil)
	if err != nil {
		panic(err)
	}

	fmt.Println("Connected to MongoDB!")
	collection := connection.Database("redditclone").Collection("posts")

	defer func() {
		if err = connection.Disconnect(context.TODO()); err != nil {
			log.Fatal(err)
		}
		fmt.Println("Connection closed.")
	}()

	////
	repo := user.NewUserSQLRepo(db)
	sess := session.NewSessionSQLRepo(db)
	postsRepo := posts.NewPostsMongoRepo(collection)
	if err = sess.DownloadKey(); err != nil {
		panic(err.Error())
	}
	var key key.Key = "author"
	userHandler := &handlers.UserHandler{
		Logger:     logger,
		UserRepo:   repo,
		Session:    sess,
		PostsRepo:  postsRepo,
		ContextKey: key,
	}
	postsHandler := &handlers.PostsHandler{
		Logger:     logger,
		PostsRepo:  postsRepo,
		Session:    sess,
		ContextKey: key,
	}
	r := mux.NewRouter()
	r.Handle("/", http.FileServer(http.Dir("../../static/html")))
	r.HandleFunc("/api/login", userHandler.LogIn)
	r.HandleFunc("/api/register", userHandler.SignIn)
	r.HandleFunc("/api/posts/", postsHandler.All).Methods("GET")
	r.HandleFunc("/api/user/{USER_LOGIN}", userHandler.GetUserPosts).Methods("GET")

	newPostHandler := middleware.JWT(key, logger, sess, middleware.Authenticate(key, logger, repo, http.HandlerFunc(postsHandler.NewPost)))
	r.Handle("/api/posts", newPostHandler).Methods("POST")

	r.HandleFunc("/api/post/{POST_ID}", postsHandler.GetPost).Methods("GET")
	r.HandleFunc("/api/posts/{CATEGORY_NAME}", postsHandler.GetPostsByCategory).Methods("GET")

	deletePostHandler := middleware.JWT(key, logger, sess, middleware.Authenticate(key, logger, repo, middleware.Authorize(key, logger, postsRepo, http.HandlerFunc(postsHandler.DeletePost))))
	r.Handle("/api/post/{POST_ID}", deletePostHandler).Methods("DELETE")

	newCommentHandler := middleware.JWT(key, logger, sess, middleware.Authenticate(key, logger, repo, http.HandlerFunc(postsHandler.AddComment)))
	r.Handle("/api/post/{POST_ID}", newCommentHandler).Methods("POST")

	deleteCommentHandler := middleware.JWT(key, logger, sess, middleware.Authenticate(key, logger, repo, middleware.Authorize(key, logger, postsRepo, http.HandlerFunc(postsHandler.DeleteComment))))
	r.Handle("/api/post/{POST_ID}/{COMMENT_ID}", deleteCommentHandler).Methods("DELETE")

	upvoteHandler := middleware.JWT(key, logger, sess, middleware.Authenticate(key, logger, repo, http.HandlerFunc(postsHandler.Vote)))
	r.Handle("/api/post/{POST_ID}/upvote", upvoteHandler).Methods("GET")

	downvoteHandler := middleware.JWT(key, logger, sess, middleware.Authenticate(key, logger, repo, http.HandlerFunc(postsHandler.Vote)))
	r.Handle("/api/post/{POST_ID}/downvote", downvoteHandler).Methods("GET")

	unvoteHandler := middleware.JWT(key, logger, sess, middleware.Authenticate(key, logger, repo, http.HandlerFunc(postsHandler.UnVote)))
	r.Handle("/api/post/{POST_ID}/unvote", unvoteHandler).Methods("GET")

	r.PathPrefix("/static/css/").Handler(http.StripPrefix("/static/css/", http.FileServer(http.Dir("../../static/css"))))
	r.PathPrefix("/static/js/").Handler(http.StripPrefix("/static/js/", http.FileServer(http.Dir("../../static/js"))))

	nw := middleware.Logging(logger, r)
	nw = middleware.Panic(logger, nw)
	err = http.ListenAndServe(":8080", nw)
	if err != nil {
		panic("server didn't start")
	}
}
