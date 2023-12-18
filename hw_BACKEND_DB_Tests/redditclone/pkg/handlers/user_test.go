package handlers

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	handlerstestsutils "redditclone/pkg/handlers/testing"
	"redditclone/pkg/key"
	"redditclone/pkg/logger"
	"redditclone/pkg/posts"
	"redditclone/pkg/session"
	"redditclone/pkg/user"

	"github.com/golang/mock/gomock"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func funcSwitcherUser(funcName string, service interface{}, req *http.Request, w *httptest.ResponseRecorder) {
	serviceReal := service.(*UserHandler)
	switch funcName {
	case "GetUserPosts":
		serviceReal.GetUserPosts(w, req)
	case "LogIn":
		serviceReal.LogIn(w, req)
	case "SignIn":
		serviceReal.SignIn(w, req)
	default:
		return
	}
}

func InitiateHandlerUser(rep *posts.MockPostsRepository, users *user.MockUserRepo, sess *session.MockSessionManager) *UserHandler {
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
	service := &UserHandler{
		Logger:     logger,
		UserRepo:   users,
		Session:    sess,
		PostsRepo:  rep,
		ContextKey: key,
	}
	return service
}

func TestGetUserPosts(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	postsRepo := posts.NewMockPostsRepository(ctrl)
	users := user.NewMockUserRepo(ctrl)
	session := session.NewMockSessionManager(ctrl)
	service := InitiateHandlerUser(postsRepo, users, session)

	// user login empty
	test := handlerstestsutils.Testing{}
	test.Req = httptest.NewRequest("GET", "/api/user/{USER_LOGIN}", nil)
	test.W = httptest.NewRecorder()
	test.FuncName = "GetUserPosts"
	test.ExpectedStatus = 400
	test.Service = service
	test.T = t
	handlerstestsutils.StatusTesting(test, funcSwitcherUser)

	// GetByUserLogin error
	test.Req = httptest.NewRequest("GET", "/api/user/{USER_LOGIN}", nil)
	test.Req = mux.SetURLVars(test.Req, map[string]string{"USER_LOGIN": "abc"})
	test.W = httptest.NewRecorder()
	test.ExpectedStatus = 500
	postsRepo.EXPECT().GetByUserLogin("abc").Return(nil, fmt.Errorf("some error"))
	handlerstestsutils.StatusTesting(test, funcSwitcherUser)

	// OK
	postsResult := []posts.Post{posts.Post{ID: "1"}, posts.Post{ID: "2"}, posts.Post{ID: "3"}}
	test.Req = httptest.NewRequest("GET", "/api/user/{USER_LOGIN}", nil)
	test.Req = mux.SetURLVars(test.Req, map[string]string{"USER_LOGIN": "abc"})
	test.W = httptest.NewRecorder()
	test.ExpectedStatus = 200
	postsRepo.EXPECT().GetByUserLogin("abc").Return(postsResult, nil).MaxTimes(2)
	test.Expected = handlerstestsutils.ConvertToJSON(t, postsResult)
	handlerstestsutils.BodyTesting(test, funcSwitcherUser)
	handlerstestsutils.StatusTesting(test, funcSwitcherUser)
}

func TestSignIn(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	postsRepo := posts.NewMockPostsRepository(ctrl)
	users := user.NewMockUserRepo(ctrl)
	session := session.NewMockSessionManager(ctrl)
	service := InitiateHandlerUser(postsRepo, users, session)

	// read All error
	test := handlerstestsutils.Testing{}
	test.Req = httptest.NewRequest("GET", "/api/user/register", &handlerstestsutils.MockReader{A: 5})
	test.W = httptest.NewRecorder()
	test.FuncName = "SignIn"
	test.ExpectedStatus = 500
	test.Service = service
	test.T = t
	handlerstestsutils.StatusTesting(test, funcSwitcherUser)

	// unmarshal error
	invalidJSON := `{"sad: "sdasd"}`
	requestBody := bytes.NewBuffer([]byte(invalidJSON))
	test.Req = httptest.NewRequest("GET", "/api/user/register", requestBody)
	test.W = httptest.NewRecorder()
	handlerstestsutils.StatusTesting(test, funcSwitcherUser)

	// addnew user error
	user := user.User{
		Username: "abc",
		Password: "sad",
	}
	validJSON := handlerstestsutils.ConvertToJSON(t, user)
	requestBody = bytes.NewBuffer(validJSON)
	test.Req = httptest.NewRequest("GET", "/api/user/register", requestBody)
	test.W = httptest.NewRecorder()
	users.EXPECT().AddNewUser(user).Return("", fmt.Errorf("db error"))
	handlerstestsutils.StatusTesting(test, funcSwitcherUser)

	// AddNewSess error
	requestBody = bytes.NewBuffer(validJSON)
	test.Req = httptest.NewRequest("GET", "/api/user/register", requestBody)
	test.W = httptest.NewRecorder()
	users.EXPECT().AddNewUser(user).Return("1", nil)
	session.EXPECT().AddNewSess("1", gomock.Any(), gomock.Any()).Return(fmt.Errorf("something gone wrong"))
	handlerstestsutils.StatusTesting(test, funcSwitcherUser)

	// signed string error
	requestBody = bytes.NewBuffer(validJSON)
	test.Req = httptest.NewRequest("GET", "/api/user/register", requestBody)
	test.W = httptest.NewRecorder()
	users.EXPECT().AddNewUser(user).Return("1", nil)
	session.EXPECT().GetExp("1", gomock.Any()).Return(time.Now().Add(120 * time.Hour).Unix())
	session.EXPECT().AddNewSess("1", gomock.Any(), gomock.Any()).Return(nil)
	session.EXPECT().GetKey().Return(1)
	handlerstestsutils.StatusTesting(test, funcSwitcherUser)

	// OK
	requestBody = bytes.NewBuffer(validJSON)
	test.Req = httptest.NewRequest("GET", "/api/user/register", requestBody)
	test.W = httptest.NewRecorder()
	users.EXPECT().AddNewUser(user).Return("1", nil)
	session.EXPECT().GetKey().Return([]byte("babuka"))
	session.EXPECT().AddNewSess("1", gomock.Any(), gomock.Any()).Return(nil)
	session.EXPECT().GetExp("1", gomock.Any()).Return(time.Now().Add(120 * time.Hour).Unix())
	test.ExpectedStatus = 201
	handlerstestsutils.StatusTesting(test, funcSwitcherUser)

}

func TestLogIn(t *testing.T) {
	ctrl := gomock.NewController(t)

	// Finish сравнит последовательсноть вызовов и выведет ошибку если последовательность другая
	defer ctrl.Finish()

	postsRepo := posts.NewMockPostsRepository(ctrl)
	users := user.NewMockUserRepo(ctrl)
	session := session.NewMockSessionManager(ctrl)
	service := InitiateHandlerUser(postsRepo, users, session)

	// read All error
	test := handlerstestsutils.Testing{}
	test.Req = httptest.NewRequest("GET", "/api/user/login", &handlerstestsutils.MockReader{A: 5})
	test.W = httptest.NewRecorder()
	test.FuncName = "LogIn"
	test.ExpectedStatus = 500
	test.Service = service
	test.T = t
	handlerstestsutils.StatusTesting(test, funcSwitcherUser)

	// unmarshal error
	invalidJSON := `{"sad: "sdasd"}`
	requestBody := bytes.NewBuffer([]byte(invalidJSON))
	test.Req = httptest.NewRequest("GET", "/api/user/login", requestBody)
	test.W = httptest.NewRecorder()
	handlerstestsutils.StatusTesting(test, funcSwitcherUser)

	// authenticate user error
	user := user.User{
		Username: "abc",
		Password: "sad",
	}
	userData := []byte(`{"username":"abc", "password":"sad"}`)
	requestBody = bytes.NewBuffer(userData)
	test.Req = httptest.NewRequest("GET", "/api/user/login", requestBody)
	test.W = httptest.NewRecorder()
	users.EXPECT().Authenticate(user).Return("\"", fmt.Errorf("db error"))
	handlerstestsutils.StatusTesting(test, funcSwitcherUser)

	// AddNewSess error
	requestBody = bytes.NewBuffer(userData)
	test.Req = httptest.NewRequest("GET", "/api/user/login", requestBody)
	test.W = httptest.NewRecorder()
	users.EXPECT().Authenticate(user).Return("1", nil)
	session.EXPECT().AddNewSess("1", gomock.Any(), gomock.Any()).Return(fmt.Errorf("something gone wrong"))
	handlerstestsutils.StatusTesting(test, funcSwitcherUser)

	// signed string error
	requestBody = bytes.NewBuffer(userData)
	test.Req = httptest.NewRequest("GET", "/api/user/login", requestBody)
	test.W = httptest.NewRecorder()
	users.EXPECT().Authenticate(user).Return("1", nil)
	session.EXPECT().AddNewSess("1", gomock.Any(), gomock.Any()).Return(nil)
	session.EXPECT().GetExp("1", gomock.Any()).Return(time.Now().Add(120 * time.Hour).Unix())
	session.EXPECT().GetKey().Return(1)
	handlerstestsutils.StatusTesting(test, funcSwitcherUser)

	// OK
	requestBody = bytes.NewBuffer(userData)
	test.Req = httptest.NewRequest("GET", "/api/user/login", requestBody)
	test.W = httptest.NewRecorder()
	users.EXPECT().Authenticate(user).Return("1", nil)
	session.EXPECT().AddNewSess("1", gomock.Any(), gomock.Any()).Return(nil)
	session.EXPECT().GetExp("1", gomock.Any()).Return(time.Now().Add(120 * time.Hour).Unix())
	session.EXPECT().GetKey().Return([]byte("babuka"))
	test.ExpectedStatus = 201
	handlerstestsutils.StatusTesting(test, funcSwitcherUser)

}
