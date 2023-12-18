package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"redditclone/pkg/author"
	"redditclone/pkg/comments"
	handlersTestsUtils "redditclone/pkg/handlers/testing"
	"redditclone/pkg/key"
	"redditclone/pkg/logger"
	"redditclone/pkg/posts"
	"redditclone/pkg/vote"

	"github.com/golang/mock/gomock"
	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func funcSwitcher(funcName string, service interface{}, req *http.Request, w *httptest.ResponseRecorder) {
	servicePosts := service.(*PostsHandler)
	switch funcName {
	case "All":
		servicePosts.All(w, req)
	case "AddComment":
		servicePosts.AddComment(w, req)
	case "DeleteComment":
		servicePosts.DeleteComment(w, req)
	case "DeletePost":
		servicePosts.DeletePost(w, req)
	case "GetPost":
		servicePosts.GetPost(w, req)
	case "GetPostsByCategory":
		servicePosts.GetPostsByCategory(w, req)
	case "NewPost":
		servicePosts.NewPost(w, req)
	case "UnVote":
		servicePosts.UnVote(w, req)
	case "Vote":
		servicePosts.Vote(w, req)
	default:
		return
	}
}

func InitiateHandler(rep *posts.MockPostsRepository) *PostsHandler {
	zapConfig := zap.Config{
		Level:            zap.NewAtomicLevelAt(zapcore.InfoLevel),
		Development:      true,
		Encoding:         "console",
		EncoderConfig:    zap.NewProductionEncoderConfig(),
		OutputPaths:      []string{"Nop"},
		ErrorOutputPaths: []string{"Nop"},
	}
	logger, err := logger.NewCustomLogger(zapConfig)
	if err != nil {
		panic(err.Error())
	}
	var key key.Key = "author"
	service := &PostsHandler{
		PostsRepo:  rep,
		ContextKey: key,
		Logger:     logger,
	}
	return service
}

func TestGetAllPosts(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	var test handlersTestsUtils.Testing
	st := posts.NewMockPostsRepository(ctrl)
	service := InitiateHandler(st)
	// Ok
	test.Req = httptest.NewRequest("GET", "/api/posts/", nil)
	test.W = httptest.NewRecorder()
	test.ExpectedStatus = 200
	test.FuncName = "All"
	test.T = t
	test.Service = service
	resultPosts := []posts.Post{
		posts.Post{ID: "1"},
		posts.Post{ID: "2"},
	}
	test.Expected = handlersTestsUtils.ConvertToJSON(t, resultPosts)
	st.EXPECT().GetAllPosts().Return(resultPosts, nil)
	handlersTestsUtils.BodyTesting(test, funcSwitcher)

	// Error
	test.Req = httptest.NewRequest("GET", "/api/posts/", nil)
	test.W = httptest.NewRecorder()
	test.ExpectedStatus = 500
	st.EXPECT().GetAllPosts().Return(nil, fmt.Errorf("Something"))
	handlersTestsUtils.StatusTesting(test, funcSwitcher)
}

func TestAddPost(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	funcSwitch := funcSwitcher
	st := posts.NewMockPostsRepository(ctrl)
	service := InitiateHandler(st)

	// io.ReadALl error
	req := httptest.NewRequest("POST", "/api/posts", &handlersTestsUtils.MockReader{A: 5})
	w := httptest.NewRecorder()
	test := handlersTestsUtils.Testing{
		Req:            req,
		W:              w,
		ExpectedStatus: 500,
		Service:        service,
		FuncName:       "NewPost",
		T:              t,
	}
	handlersTestsUtils.StatusTesting(test, funcSwitch)

	// unmarshal error
	invalidJSON := `{"sdsfs:sadas",}`
	requestBody := bytes.NewBuffer([]byte(invalidJSON))
	test.Req = httptest.NewRequest("POST", "/api/posts", requestBody)
	test.W = httptest.NewRecorder()
	handlersTestsUtils.StatusTesting(test, funcSwitch)

	// context error
	validJSON := `{"sdsfs":"sadas"}`
	requestBody = bytes.NewBuffer([]byte(validJSON))
	test.Req = httptest.NewRequest("POST", "/api/posts", requestBody)
	test.W = httptest.NewRecorder()
	handlersTestsUtils.StatusTesting(test, funcSwitch)

	// AddPost error
	st.EXPECT().AddPost(gomock.Any()).Return(posts.Post{}, fmt.Errorf("Someerror"))
	requestBody = bytes.NewBuffer([]byte(validJSON))
	test.Req = httptest.NewRequest("POST", "/api/posts", requestBody)
	author := author.Author{
		Username: "abc",
		ID:       "1",
	}
	ctx := test.Req.Context()
	ctx = context.WithValue(ctx,
		service.ContextKey,
		&author,
	)
	test.W = httptest.NewRecorder()
	test.Req = test.Req.WithContext(ctx)
	handlersTestsUtils.StatusTesting(test, funcSwitch)

	// OK
	testPost := posts.Post{
		Comments: []comments.Comment{
			comments.Comment{
				Author: author,
				Body:   "kasd;lasd",
				ID:     "2",
			},
		},
		Votes:  []vote.Vote{vote.Vote{User: "1", Vote: 1}},
		Author: author,
	}
	testPostJSON, err := json.Marshal(testPost)
	if err != nil {
		t.Error("unexpected error\n")
	}
	returnPost := testPost
	returnPost.ID = "3"
	returnPostJSON, err := json.Marshal(returnPost)
	if err != nil {
		t.Error("unexpected error\n")
	}
	st.EXPECT().AddPost(testPost).Return(returnPost, nil)
	requestBody = bytes.NewBuffer(testPostJSON)
	test.Req = httptest.NewRequest("POST", "/api/posts", requestBody)
	ctx = test.Req.Context()
	ctx = context.WithValue(ctx,
		service.ContextKey,
		&author,
	)
	test.W = httptest.NewRecorder()
	test.Req = test.Req.WithContext(ctx)
	test.Expected = returnPostJSON
	handlersTestsUtils.BodyTesting(test, funcSwitch)
}

func TestGetPost(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	funcSwitch := funcSwitcher
	st := posts.NewMockPostsRepository(ctrl)
	service := InitiateHandler(st)
	var test handlersTestsUtils.Testing

	// empty post
	test.Req = httptest.NewRequest("GET", "/api/post/", nil)
	test.W = httptest.NewRecorder()
	test.FuncName = "GetPost"
	test.ExpectedStatus = 400
	test.Service = service
	test.T = t
	handlersTestsUtils.StatusTesting(test, funcSwitch)

	// GetPostError
	test.Req = httptest.NewRequest("GET", "/api/post/", nil)
	test.Req = mux.SetURLVars(test.Req, map[string]string{"POST_ID": "1"})
	test.W = httptest.NewRecorder()
	test.ExpectedStatus = 500
	st.EXPECT().GetPostByID("1").Return(posts.Post{}, fmt.Errorf(mongo.ErrNoDocuments.Error()))
	handlersTestsUtils.StatusTesting(test, funcSwitch)

	// UpdatePost error
	outPost := posts.Post{
		Views: 0,
		ID:    "1",
	}
	test.Req = httptest.NewRequest("GET", "/api/post/{POST_ID}", nil)
	test.Req = mux.SetURLVars(test.Req, map[string]string{"POST_ID": "1"})
	test.W = httptest.NewRecorder()
	st.EXPECT().GetPostByID("1").Return(outPost, nil)
	outPost.Views += 1
	st.EXPECT().UpdatePost(outPost).Return(fmt.Errorf("skddlsf"))
	handlersTestsUtils.StatusTesting(test, funcSwitch)

	// OK
	req := httptest.NewRequest("GET", "/api/post/{POST_ID}", nil)
	test.Req = mux.SetURLVars(req, map[string]string{"POST_ID": "1"})
	test.W = httptest.NewRecorder()
	test.ExpectedStatus = 201
	outPost.Views -= 1
	st.EXPECT().GetPostByID("1").Return(outPost, nil).MaxTimes(2)
	outPost.Views += 1
	st.EXPECT().UpdatePost(outPost).Return(nil).MaxTimes(2)
	test.Expected = handlersTestsUtils.ConvertToJSON(t, outPost)
	handlersTestsUtils.BodyTesting(test, funcSwitch)
	handlersTestsUtils.StatusTesting(test, funcSwitch)
}

func TestGetPostByCategory(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	funcSwitch := funcSwitcher
	st := posts.NewMockPostsRepository(ctrl)
	service := InitiateHandler(st)
	var test handlersTestsUtils.Testing

	// empty category error
	test.Req = httptest.NewRequest("GET", "/api/posts/", nil)
	test.W = httptest.NewRecorder()
	test.FuncName = "GetPostsByCategory"
	test.ExpectedStatus = 400
	test.Service = service
	test.T = t
	handlersTestsUtils.StatusTesting(test, funcSwitch)

	// get category error
	test.Req = httptest.NewRequest("GET", "/api/posts/", nil)
	test.Req = mux.SetURLVars(test.Req, map[string]string{"CATEGORY_NAME": "funnyyyy"})
	test.W = httptest.NewRecorder()
	test.ExpectedStatus = 500
	st.EXPECT().GetCategory("funnyyyy").Return(nil, fmt.Errorf("no such category"))
	handlersTestsUtils.StatusTesting(test, funcSwitch)

	// OK
	posts := make([]posts.Post, 5)
	for i := range posts {
		posts[i].Category = "funny"
		posts[i].ID = strconv.Itoa(i)
	}
	test.Req = httptest.NewRequest("GET", "/api/posts/", nil)
	test.Req = mux.SetURLVars(test.Req, map[string]string{"CATEGORY_NAME": "funny"})
	test.W = httptest.NewRecorder()
	test.ExpectedStatus = 200
	st.EXPECT().GetCategory("funny").Return(posts, nil).MaxTimes(2)
	test.Expected = handlersTestsUtils.ConvertToJSON(t, posts)
	handlersTestsUtils.StatusTesting(test, funcSwitch)
	handlersTestsUtils.BodyTesting(test, funcSwitch)
}

func TestAddComment(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	funcSwitch := funcSwitcher
	st := posts.NewMockPostsRepository(ctrl)
	service := InitiateHandler(st)
	var test handlersTestsUtils.Testing

	// empty POST ID
	test.Req = httptest.NewRequest("POST", "/api/post/", nil)
	test.W = httptest.NewRecorder()
	test.FuncName = "AddComment"
	test.ExpectedStatus = 400
	test.Service = service
	test.T = t
	handlersTestsUtils.StatusTesting(test, funcSwitch)

	// context error
	test.Req = httptest.NewRequest("POST", "/api/post/", nil)
	test.Req = mux.SetURLVars(test.Req, map[string]string{"POST_ID": "1"})
	test.W = httptest.NewRecorder()
	test.ExpectedStatus = 500
	handlersTestsUtils.StatusTesting(test, funcSwitch)

	// broken body
	test.Req = httptest.NewRequest("POST", "/api/post/", &handlersTestsUtils.MockReader{A: 5})
	test.Req = mux.SetURLVars(test.Req, map[string]string{"POST_ID": "1"})
	ctx := test.Req.Context()
	author := author.Author{
		Username: "abc",
		ID:       "12",
	}
	ctx = context.WithValue(ctx,
		service.ContextKey,
		&author,
	)
	test.Req = test.Req.WithContext(ctx)
	test.W = httptest.NewRecorder()
	test.ExpectedStatus = 500
	handlersTestsUtils.StatusTesting(test, funcSwitch)

	// unmarshal error
	invalidJSON := `{"sdas:"sda}`
	requestBody := bytes.NewBuffer([]byte(invalidJSON))
	test.Req = httptest.NewRequest("POST", "/api/post/", requestBody)
	test.Req = mux.SetURLVars(test.Req, map[string]string{"POST_ID": "1"})
	ctx = test.Req.Context()
	ctx = context.WithValue(ctx, service.ContextKey, &author)
	test.Req = test.Req.WithContext(ctx)
	test.W = httptest.NewRecorder()
	handlersTestsUtils.StatusTesting(test, funcSwitch)

	// AddComment error
	validJSON := `{"comment":"sdasf"}`
	requestBody = bytes.NewBuffer([]byte(validJSON))
	test.Req = httptest.NewRequest("POST", "/api/post/", requestBody)
	test.Req = mux.SetURLVars(test.Req, map[string]string{"POST_ID": "1"})
	ctx = test.Req.Context()
	ctx = context.WithValue(ctx, service.ContextKey, &author)
	test.Req = test.Req.WithContext(ctx)
	test.W = httptest.NewRecorder()
	st.EXPECT().AddComment("1", gomock.Any()).Return(posts.Post{}, fmt.Errorf("no such post"))
	handlersTestsUtils.StatusTesting(test, funcSwitch)

	// OK
	requestBody = bytes.NewBuffer([]byte(validJSON))
	test.Req = httptest.NewRequest("POST", "/api/post/", requestBody)
	test.Req = mux.SetURLVars(test.Req, map[string]string{"POST_ID": "1"})
	ctx = test.Req.Context()
	ctx = context.WithValue(ctx, service.ContextKey, &author)
	test.Req = test.Req.WithContext(ctx)
	test.W = httptest.NewRecorder()
	returnPost := posts.Post{
		ID: "1",
		Comments: []comments.Comment{
			comments.Comment{
				Author: author,
				Body:   "sdasf",
				ID:     "2",
			},
		},
	}
	st.EXPECT().AddComment("1", gomock.Any()).Return(returnPost, nil).MaxTimes(2)
	test.Expected = handlersTestsUtils.ConvertToJSON(t, returnPost)
	test.ExpectedStatus = 201
	handlersTestsUtils.StatusTesting(test, funcSwitch)
	handlersTestsUtils.BodyTesting(test, funcSwitch)
}

func TestDeleteComment(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	funcSwitch := funcSwitcher
	st := posts.NewMockPostsRepository(ctrl)
	service := InitiateHandler(st)
	var test handlersTestsUtils.Testing

	// empty URL
	test.Req = httptest.NewRequest("DELETE", "/api/post/{POST_ID}/{COMMENT_ID}", nil)
	test.W = httptest.NewRecorder()
	test.FuncName = "DeleteComment"
	test.ExpectedStatus = 400
	test.Service = service
	test.T = t
	handlersTestsUtils.StatusTesting(test, funcSwitch)

	// DeleteComment error
	test.Req = httptest.NewRequest("DELETE", "/api/post/{POST_ID}/{COMMENT_ID}", nil)
	test.Req = mux.SetURLVars(test.Req, map[string]string{"POST_ID": "1", "COMMENT_ID": "2"})
	test.W = httptest.NewRecorder()
	test.ExpectedStatus = 500
	st.EXPECT().DeleteComment("1", "2").Return(posts.Post{}, fmt.Errorf("no such comment/post"))
	handlersTestsUtils.StatusTesting(test, funcSwitch)

	// Ok
	test.Req = httptest.NewRequest("DELETE", "/api/post/{POST_ID}/{COMMENT_ID}", nil)
	test.Req = mux.SetURLVars(test.Req, map[string]string{"POST_ID": "1", "COMMENT_ID": "2"})
	test.W = httptest.NewRecorder()
	test.ExpectedStatus = 200
	test.Expected = handlersTestsUtils.ConvertToJSON(t, posts.Post{ID: "1"})
	st.EXPECT().DeleteComment("1", "2").Return(posts.Post{ID: "1"}, nil).MaxTimes(2)
	handlersTestsUtils.StatusTesting(test, funcSwitch)
	handlersTestsUtils.BodyTesting(test, funcSwitch)
}

func TestDeletePost(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	funcSwitch := funcSwitcher
	st := posts.NewMockPostsRepository(ctrl)
	service := InitiateHandler(st)
	var test handlersTestsUtils.Testing

	// empty URL
	test.Req = httptest.NewRequest("DELETE", "/api/post/{POST_ID}", nil)
	test.W = httptest.NewRecorder()
	test.FuncName = "DeletePost"
	test.ExpectedStatus = 400
	test.Service = service
	test.T = t
	handlersTestsUtils.StatusTesting(test, funcSwitch)

	// DeletePost error
	test.Req = httptest.NewRequest("DELETE", "/api/post/{POST_ID}/{COMMENT_ID}", nil)
	test.Req = mux.SetURLVars(test.Req, map[string]string{"POST_ID": "1"})
	test.W = httptest.NewRecorder()
	test.ExpectedStatus = 500
	st.EXPECT().DeletePost("1").Return(fmt.Errorf("no such comment/post"))
	handlersTestsUtils.StatusTesting(test, funcSwitch)

	// Ok
	test.Req = httptest.NewRequest("DELETE", "/api/post/{POST_ID}/{COMMENT_ID}", nil)
	test.Req = mux.SetURLVars(test.Req, map[string]string{"POST_ID": "1"})
	test.W = httptest.NewRecorder()
	test.ExpectedStatus = 200
	test.Expected = handlersTestsUtils.ConvertToJSON(t, posts.Post{ID: "1"})
	st.EXPECT().DeletePost("1").Return(nil).MaxTimes(2)
	handlersTestsUtils.StatusTesting(test, funcSwitch)
}

func TestVote(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	funcSwitch := funcSwitcher
	st := posts.NewMockPostsRepository(ctrl)
	service := InitiateHandler(st)
	var test handlersTestsUtils.Testing

	// empty URL
	test.Req = httptest.NewRequest("GET", "/api/post/{POST_ID}/downvote", nil)
	test.W = httptest.NewRecorder()
	test.FuncName = "Vote"
	test.ExpectedStatus = 400
	test.Service = service
	test.T = t
	handlersTestsUtils.StatusTesting(test, funcSwitch)

	// empty context
	test.Req = httptest.NewRequest("GET", "/api/post/{POST_ID}/downvote", nil)
	test.Req = mux.SetURLVars(test.Req, map[string]string{"POST_ID": "1"})
	test.W = httptest.NewRecorder()
	test.ExpectedStatus = 500
	handlersTestsUtils.StatusTesting(test, funcSwitch)

	// Vote error
	test.Req = httptest.NewRequest("GET", "/api/post/{POST_ID}/downvote", nil)
	test.Req = mux.SetURLVars(test.Req, map[string]string{"POST_ID": "1"})
	ctx := test.Req.Context()
	author := author.Author{
		ID:       "12",
		Username: "abc",
	}
	ctx = context.WithValue(ctx, service.ContextKey, &author)
	test.Req = test.Req.WithContext(ctx)
	test.W = httptest.NewRecorder()
	test.ExpectedStatus = 500
	testVote := vote.Vote{
		User: "12",
		Vote: -1,
	}
	st.EXPECT().Vote("1", testVote).Return(posts.Post{}, fmt.Errorf("no such post"))
	handlersTestsUtils.StatusTesting(test, funcSwitch)

	// OK
	test.Req = httptest.NewRequest("GET", "/api/post/{POST_ID}/downvote", nil)
	test.Req = mux.SetURLVars(test.Req, map[string]string{"POST_ID": "1"})
	ctx = test.Req.Context()
	ctx = context.WithValue(ctx, service.ContextKey, &author)
	test.Req = test.Req.WithContext(ctx)
	test.W = httptest.NewRecorder()
	test.ExpectedStatus = 200
	st.EXPECT().Vote("1", testVote).Return(posts.Post{ID: "1", Votes: []vote.Vote{testVote}}, nil).MaxTimes(2)
	test.Expected = handlersTestsUtils.ConvertToJSON(t, posts.Post{ID: "1", Votes: []vote.Vote{testVote}})
	handlersTestsUtils.StatusTesting(test, funcSwitch)
	handlersTestsUtils.BodyTesting(test, funcSwitch)
}

func TestUnVote(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	funcSwitch := funcSwitcher
	st := posts.NewMockPostsRepository(ctrl)
	service := InitiateHandler(st)
	var test handlersTestsUtils.Testing

	// empty URL
	test.Req = httptest.NewRequest("GET", "/api/post/{POST_ID}/unvote", nil)
	test.W = httptest.NewRecorder()
	test.FuncName = "UnVote"
	test.ExpectedStatus = 400
	test.Service = service
	test.T = t
	handlersTestsUtils.StatusTesting(test, funcSwitch)

	// empty context
	test.Req = httptest.NewRequest("GET", "/api/post/{POST_ID}/unvote", nil)
	test.Req = mux.SetURLVars(test.Req, map[string]string{"POST_ID": "1"})
	test.W = httptest.NewRecorder()
	test.ExpectedStatus = 500
	handlersTestsUtils.StatusTesting(test, funcSwitch)

	// UnVote error
	test.Req = httptest.NewRequest("GET", "/api/post/{POST_ID}/uvote", nil)
	test.Req = mux.SetURLVars(test.Req, map[string]string{"POST_ID": "1"})
	ctx := test.Req.Context()
	author := author.Author{
		ID:       "12",
		Username: "abc",
	}
	ctx = context.WithValue(ctx, service.ContextKey, &author)
	test.Req = test.Req.WithContext(ctx)
	test.W = httptest.NewRecorder()
	test.ExpectedStatus = 500
	testVote := vote.Vote{
		User: "12",
		Vote: -1,
	}
	st.EXPECT().UnVote(testVote.User, "1").Return(posts.Post{}, fmt.Errorf("no such post"))
	handlersTestsUtils.StatusTesting(test, funcSwitch)

	// OK
	test.Req = httptest.NewRequest("GET", "/api/post/{POST_ID}/unvote", nil)
	test.Req = mux.SetURLVars(test.Req, map[string]string{"POST_ID": "1"})
	ctx = test.Req.Context()
	ctx = context.WithValue(ctx, service.ContextKey, &author)
	test.Req = test.Req.WithContext(ctx)
	test.W = httptest.NewRecorder()
	test.ExpectedStatus = 200
	st.EXPECT().UnVote(testVote.User, "1").Return(posts.Post{ID: "1"}, nil).MaxTimes(2)
	test.Expected = handlersTestsUtils.ConvertToJSON(t, posts.Post{ID: "1"})
	handlersTestsUtils.StatusTesting(test, funcSwitch)
	handlersTestsUtils.BodyTesting(test, funcSwitch)
}
