package handlerstestsutils

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

type Testing struct {
	Service        interface{}
	Req            *http.Request
	W              *httptest.ResponseRecorder
	FuncName       string
	Expected       []byte
	ExpectedStatus int
	T              *testing.T
}

func BodyTesting(test Testing, funcSwitcher func(funcName string, service interface{}, req *http.Request, w *httptest.ResponseRecorder)) {
	funcSwitcher(test.FuncName, test.Service, test.Req, test.W)
	resp := test.W.Result()
	body, err := io.ReadAll(resp.Body)
	defer resp.Body.Close()
	if err != nil {
		test.T.Errorf("unexpected error %s", err.Error())
	}
	assert.Equal(test.T, test.Expected, body)

}

func StatusTesting(test Testing, funcSwitcher func(funcName string, service interface{}, req *http.Request, w *httptest.ResponseRecorder)) {
	funcSwitcher(test.FuncName, test.Service, test.Req, test.W)
	resp := test.W.Result()
	if resp.StatusCode != test.ExpectedStatus {
		test.T.Errorf("Expected status code %d got %d", test.ExpectedStatus, resp.StatusCode)
	}
}
