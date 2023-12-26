package posts

import (
	"context"
	"testing"

	"redditclone/pkg/comments"
	"redditclone/pkg/vote"

	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/integration/mtest"
)

var CursorError string = "cursor error"
var SomeError string = "some error"
var CALLError string = "c.All error"

func ConvertToBSON(t *testing.T, data interface{}) []byte {
	bytes, err := bson.Marshal(data)
	if err != nil {
		t.Errorf("unexpected err: %s", err.Error())
		return nil
	}
	return bytes
}

func ConvertToPrimtive(t *testing.T, data []byte) *primitive.D {
	var primitive primitive.D
	err := bson.Unmarshal(data, &primitive)
	if err != nil {
		t.Errorf("unexpected error %s", err.Error())
		return nil
	}
	return &primitive
}

type Testing struct {
	t                *testing.T
	mt               *mtest.T
	funcName         string
	testName         string
	expected         interface{}
	args             []interface{}
	mockResponses    []primitive.D
	noEqualAssertion bool
}

func funcSwitcher(repo *PostsMongoRepo, funcName string, args ...interface{}) (interface{}, error) {
	var ans interface{}
	var err error
	switch funcName {
	case "GetAllPosts":
		ans, err = repo.GetAllPosts(context.Background())
		return ans, err
	case "UpdatePost":
		err = repo.UpdatePost(context.Background(), args[0].(Post))
		return ans, err
	case "DeletePost":
		err = repo.DeletePost(context.Background(), args[0].(string))
		return ans, err
	case "AddPost":
		ans, err = repo.AddPost(context.Background(), args[0].(Post))
		return ans, err
	case "GetCategory":
		ans, err = repo.GetCategory(context.Background(), args[0].(string))
		return ans, err
	case "GetPostByID":
		ans, err = repo.GetPostByID(context.Background(), args[0].(string))
		return ans, err
	case "AddComment":
		ans, err = repo.AddComment(context.Background(), args[0].(string), args[1].(comments.Comment))
		return ans, err
	case "DeleteComment":
		ans, err = repo.DeleteComment(context.Background(), args[0].(string), args[1].(string))
		return ans, err
	case "GetByUserLogin":
		ans, err = repo.GetByUserLogin(context.Background(), args[0].(string))
		return ans, err
	case "Vote":
		ans, err = repo.Vote(context.Background(), args[0].(string), args[1].(vote.Vote))
		return ans, err
	case "UnVote":
		ans, err = repo.UnVote(context.Background(), args[0].(string), args[1].(string))
		return ans, err
	}

	return ans, err
}

func EqualityTesting(test Testing) {
	mt := test.mt
	mt.Run(test.testName, func(mt *mtest.T) {
		postsCollection := mt.Coll
		postsRepo := NewPostsMongoRepo(postsCollection)
		for _, resp := range test.mockResponses {
			mt.AddMockResponses(resp)
		}
		ans, err := funcSwitcher(postsRepo, test.funcName, test.args...)
		assert.Nil(test.t, err)
		if !test.noEqualAssertion {
			assert.Equal(test.t, test.expected, ans)
		}
	})
}

func ErrorTesting(test Testing) {
	mt := test.mt
	mt.Run(test.testName, func(mt *mtest.T) {
		postsCollection := mt.Coll
		postsRepo := NewPostsMongoRepo(postsCollection)
		for _, resp := range test.mockResponses {
			mt.AddMockResponses(resp)
		}
		_, err := funcSwitcher(postsRepo, test.funcName, test.args...)
		if err == nil {
			test.t.Errorf("expected error,got nil")
		}
	})
}

func TestGetAllPosts(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))
	// defer mt.Close()

	Post1 := Post{
		ID: "1",
	}
	Post2 := Post{
		ID: "2",
	}

	primitivePost1 := ConvertToPrimtive(t, ConvertToBSON(t, Post1))
	primitivePost2 := ConvertToPrimtive(t, ConvertToBSON(t, Post2))
	first := mtest.CreateCursorResponse(1, "foo.bar", mtest.FirstBatch, *primitivePost1)
	second := mtest.CreateCursorResponse(1, "foo.bar", mtest.NextBatch, *primitivePost2)
	killCursors := mtest.CreateCursorResponse(0, "foo.bar", mtest.NextBatch)

	test := Testing{
		t:             t,
		mt:            mt,
		funcName:      "GetAllPosts",
		testName:      "cursorTestOK",
		expected:      []Post{Post1, Post2},
		mockResponses: []primitive.D{first, second, killCursors},
	}
	EqualityTesting(test)

	test.mockResponses = []primitive.D{bson.D{{Key: "ok", Value: 0}}}
	test.testName = CursorError
	ErrorTesting(test)

	test.mockResponses = []primitive.D{mtest.CreateCursorResponse(1, "foo.bar", mtest.FirstBatch, bson.D{
		{Key: "author", Value: 5},
	})}
	test.testName = CALLError
	ErrorTesting(test)
}

func TestUpdatePost(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))
	// defer mt.Close()

	test := Testing{
		t:        t,
		mt:       mt,
		funcName: "UpdatePost",
		testName: "OK",
		mockResponses: []primitive.D{
			bson.D{
				{Key: "ok", Value: 1},
				{Key: "nModified", Value: 1},
			},
		},
		args: []interface{}{Post{}},
	}
	EqualityTesting(test)

	test.mockResponses = []primitive.D{
		bson.D{
			{Key: "ok", Value: 1},
			{Key: "nModified", Value: 0},
		},
	}
	test.testName = "0 modified"
	ErrorTesting(test)

	test.mockResponses = []primitive.D{
		bson.D{
			{Key: "ok", Value: 0},
		},
	}
	test.testName = SomeError
	ErrorTesting(test)
}

func TestDeletePost(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))
	// defer mt.Close()

	test := Testing{
		t:        t,
		mt:       mt,
		funcName: "DeletePost",
		testName: "OK",
		mockResponses: []primitive.D{
			bson.D{
				{Key: "ok", Value: 1},
				{Key: "acknowledged", Value: true},
				{Key: "n", Value: 1},
			},
		},
		args: []interface{}{"1"},
	}
	EqualityTesting(test)

	test.mockResponses = []primitive.D{bson.D{
		{Key: "ok", Value: 1},
		{Key: "acknowledged", Value: true},
		{Key: "n", Value: 0},
	}}
	test.testName = "0 Deleted"
	ErrorTesting(test)

	test.mockResponses = []primitive.D{bson.D{
		{Key: "ok", Value: 0},
	}}
	test.testName = "error deleting"
	ErrorTesting(test)

}

func TestAddPost(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))
	// defer mt.Close()

	test := Testing{
		t:                t,
		mt:               mt,
		funcName:         "AddPost",
		testName:         "OK",
		expected:         Post{ID: "1"},
		args:             []interface{}{Post{ID: "1"}},
		mockResponses:    []primitive.D{mtest.CreateSuccessResponse()},
		noEqualAssertion: true,
	}
	EqualityTesting(test)

	test.mockResponses = []primitive.D{bson.D{{Key: "ok", Value: 0}}}
	test.testName = SomeError
	ErrorTesting(test)
}

func TestGetPostByID(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))
	// defer mt.Close()

	post := Post{ID: "1"}
	primitivePost := ConvertToPrimtive(t, ConvertToBSON(t, post))
	test := Testing{
		t:             t,
		mt:            mt,
		funcName:      "GetPostByID",
		testName:      "OK",
		expected:      Post{ID: "1"},
		args:          []interface{}{"1"},
		mockResponses: []primitive.D{mtest.CreateCursorResponse(1, "foo.bar", mtest.FirstBatch, *primitivePost)},
	}
	EqualityTesting(test)

	test.mockResponses = []primitive.D{bson.D{{Key: "ok", Value: 0}}}
	test.testName = SomeError
	ErrorTesting(test)

	test.mockResponses = []primitive.D{mtest.CreateCursorResponse(1, "foo.bar", mtest.FirstBatch, bson.D{
		{Key: "author", Value: 5},
	})}
	test.testName = CALLError
	ErrorTesting(test)
}

func TestGetByCategory(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))
	// defer mt.Close()

	Post1 := Post{
		ID: "1",
	}
	Post2 := Post{
		ID: "2",
	}

	primitivePost1 := ConvertToPrimtive(t, ConvertToBSON(t, Post1))
	primitivePost2 := ConvertToPrimtive(t, ConvertToBSON(t, Post2))
	first := mtest.CreateCursorResponse(1, "foo.bar", mtest.FirstBatch, *primitivePost1)
	second := mtest.CreateCursorResponse(1, "foo.bar", mtest.NextBatch, *primitivePost2)
	killCursors := mtest.CreateCursorResponse(0, "foo.bar", mtest.NextBatch)

	test := Testing{
		t:             t,
		mt:            mt,
		funcName:      "GetCategory",
		testName:      "cursorTestOK",
		expected:      []Post{Post1, Post2},
		mockResponses: []primitive.D{first, second, killCursors},
		args:          []interface{}{"news"},
	}
	EqualityTesting(test)

	test.mockResponses = []primitive.D{bson.D{{Key: "ok", Value: 0}}}
	test.testName = CursorError
	ErrorTesting(test)

	test.mockResponses = []primitive.D{mtest.CreateCursorResponse(1, "foo.bar", mtest.FirstBatch, bson.D{
		{Key: "author", Value: 5},
	})}
	test.testName = CALLError
	ErrorTesting(test)
}

func TestAddComment(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))
	// defer mt.Close()

	test := Testing{
		t:        t,
		mt:       mt,
		funcName: "AddComment",
		testName: "OK",
		expected: Post{ID: "1"},
		args:     []interface{}{"1", comments.Comment{}},
		mockResponses: []primitive.D{bson.D{
			{Key: "ok", Value: 1},
			{Key: "value", Value: bson.D{
				{Key: "_id", Value: "1"},
			},
			},
		}},
	}
	EqualityTesting(test)

	test.mockResponses = []primitive.D{bson.D{{Key: "ok", Value: 0}, {Key: "value", Value: bson.D{
		{Key: "_id", Value: "1"},
	}}}}
	test.testName = SomeError
	ErrorTesting(test)

	test.mockResponses = []primitive.D{bson.D{
		{Key: "ok", Value: 1},
		{Key: "value", Value: bson.D{
			{Key: "author", Value: "1"},
		},
		},
	}}
	test.testName = "decode error"
	ErrorTesting(test)

}

func TestDeleteComment(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))
	// defer mt.Close()

	test := Testing{
		t:        t,
		mt:       mt,
		funcName: "DeleteComment",
		testName: "OK",
		expected: Post{ID: "1"},
		args:     []interface{}{"1", "2"},
		mockResponses: []primitive.D{bson.D{
			{Key: "ok", Value: 1},
			{Key: "value", Value: bson.D{
				{Key: "_id", Value: "1"},
			},
			},
		}},
	}
	EqualityTesting(test)

	test.mockResponses = []primitive.D{bson.D{{Key: "ok", Value: 0}, {Key: "value", Value: bson.D{
		{Key: "_id", Value: "1"},
	}}}}
	test.testName = "error"
	ErrorTesting(test)

	test.mockResponses = []primitive.D{bson.D{
		{Key: "ok", Value: 1},
		{Key: "value", Value: bson.D{
			{Key: "author", Value: "1"},
		},
		},
	}}
	test.testName = "decode error"
	ErrorTesting(test)

}

func TestGetByUserLogin(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))
	// defer mt.Close()

	Post1 := Post{
		ID: "1",
	}
	Post2 := Post{
		ID: "2",
	}

	primitivePost1 := ConvertToPrimtive(t, ConvertToBSON(t, Post1))
	primitivePost2 := ConvertToPrimtive(t, ConvertToBSON(t, Post2))
	first := mtest.CreateCursorResponse(1, "foo.bar", mtest.FirstBatch, *primitivePost1)
	second := mtest.CreateCursorResponse(1, "foo.bar", mtest.NextBatch, *primitivePost2)
	killCursors := mtest.CreateCursorResponse(0, "foo.bar", mtest.NextBatch)

	test := Testing{
		t:             t,
		mt:            mt,
		funcName:      "GetByUserLogin",
		testName:      "cursorTestOK",
		expected:      []Post{Post1, Post2},
		mockResponses: []primitive.D{first, second, killCursors},
		args:          []interface{}{"asd"},
	}
	EqualityTesting(test)

	test.mockResponses = []primitive.D{bson.D{{Key: "ok", Value: 0}}}
	test.testName = CursorError
	ErrorTesting(test)

	test.mockResponses = []primitive.D{mtest.CreateCursorResponse(1, "foo.bar", mtest.FirstBatch, bson.D{
		{Key: "author", Value: 5},
	})}
	test.testName = CALLError
	ErrorTesting(test)
}

func TestVote(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))
	// defer mt.Close()

	test := Testing{
		t:        t,
		mt:       mt,
		funcName: "Vote",
		testName: "get post error",
		mockResponses: []primitive.D{bson.D{
			{Key: "ok", Value: 0},
		}},
		args: []interface{}{"1", vote.Vote{Vote: 1}},
	}
	ErrorTesting(test)

	primitivePost := ConvertToPrimtive(t, ConvertToBSON(t, Post{ID: "1"}))
	test.mockResponses = []primitive.D{mtest.CreateCursorResponse(1, "foo.bar", mtest.FirstBatch, *primitivePost)}
	test.testName = "update error"
	ErrorTesting(test)

	voteTmp := vote.Vote{
		Vote: 1,
		User: "1",
	}
	Post1 := Post{
		ID:      "1",
		Votes:   []vote.Vote{voteTmp},
		Upvotes: 1,
		Score:   1,
	}
	primitivePost1 := ConvertToPrimtive(t, ConvertToBSON(t, Post1))
	test.mockResponses = []primitive.D{
		mtest.CreateCursorResponse(1, "foo.bar", mtest.FirstBatch, *primitivePost1),
		mtest.CreateSuccessResponse(),
		bson.D{
			{Key: "ok", Value: 1},
			{Key: "nModified", Value: 1},
		},
	}

	voteTmp.Vote = -1
	Post1.UpvotePercentage = 0
	Post1.Upvotes--
	Post1.Votes[0].Vote = -1
	Post1.Score = -1

	test.args[1] = voteTmp
	test.expected = Post1
	test.testName = "Ok"
	EqualityTesting(test)

}

func TestUnVote(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))
	// defer mt.Close()

	test := Testing{
		t:        t,
		mt:       mt,
		funcName: "UnVote",
		testName: "get post error",
		mockResponses: []primitive.D{bson.D{
			{Key: "ok", Value: 0},
		}},
		args: []interface{}{"1", "1"},
	}
	ErrorTesting(test)

	Post1 := Post{
		ID: "1",
	}
	primitivePost := ConvertToPrimtive(t, ConvertToBSON(t, Post1))
	test.testName = "not exist"
	test.mockResponses = []primitive.D{mtest.CreateCursorResponse(1, "foo.bar", mtest.FirstBatch, *primitivePost)}
	ErrorTesting(test)

	voteTmp := vote.Vote{
		Vote: 1,
		User: "1",
	}
	Post1 = Post{
		ID:      "1",
		Votes:   []vote.Vote{voteTmp},
		Upvotes: 1,
		Score:   1,
	}
	primitivePost = ConvertToPrimtive(t, ConvertToBSON(t, Post1))

	test.mockResponses = []primitive.D{
		mtest.CreateCursorResponse(1, "foo.bar", mtest.FirstBatch, *primitivePost),
		mtest.CreateSuccessResponse(),
		bson.D{
			{Key: "ok", Value: 1},
			{Key: "nModified", Value: 1},
		},
	}

	Post1.UpvotePercentage = 0
	Post1.Upvotes--
	Post1.Votes = make([]vote.Vote, 0)
	Post1.Score = 0

	test.expected = Post1
	test.testName = "ok"
	EqualityTesting(test)

	test.testName = "update error"
	test.mockResponses = []primitive.D{
		mtest.CreateCursorResponse(1, "foo.bar", mtest.FirstBatch, *primitivePost),
		mtest.CreateSuccessResponse(),
		bson.D{
			{Key: "ok", Value: 0},
		},
	}
	ErrorTesting(test)
}
