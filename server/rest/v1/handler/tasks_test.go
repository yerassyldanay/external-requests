package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/yerassyldanay/requestmaker/pkg/mockx"
	"github.com/yerassyldanay/requestmaker/service/taskservice"
	mock_taskservice "github.com/yerassyldanay/requestmaker/service/taskservice/mock"
)

type mockTaskServer struct {
	taskServer      *TaskServer
	mockTaskHandler *mock_taskservice.MockTaskHandler
}

func newMockTaskServer(t *testing.T) *mockTaskServer {
	logger, err := zap.NewDevelopment()
	require.Errorf(t, err, "failed to create a logger")

	gin.SetMode(gin.ReleaseMode)
	ctr := gomock.NewController(t)
	mockService := mock_taskservice.NewMockTaskHandler(ctr)
	return &mockTaskServer{
		taskServer:      NewTaskServer(logger, mockService),
		mockTaskHandler: mockService,
	}
}

func TestAcceptTask(t *testing.T) {
	taskId := mockx.GetUUID(t)
	type Task struct {
		Method  string              `json:"method"`
		Url     string              `json:"url"`
		Headers map[string][]string `json:"headers"`
	}

	task := Task{
		Method: http.MethodGet,
		Url:    "https://google.com",
		Headers: map[string][]string{
			"Content-Type": {
				"application/json",
			},
		},
	}
	urlString, err := url.Parse(task.Url)
	require.NoErrorf(t, err, "failed to parse url")

	testCases := []struct {
		name        string
		status      int
		prepareMock func(m *mockTaskServer)
		prepareHttp func() (*httptest.ResponseRecorder, *http.Request)
		check       func(r *httptest.ResponseRecorder)
	}{
		{
			name:   "ok",
			status: http.StatusOK,
			prepareHttp: func() (*httptest.ResponseRecorder, *http.Request) {
				req := httptest.NewRequest(http.MethodPost, "/api/v1/task", mockx.GetBuffer(t, task))
				req.Header.Set("Content-Type", "application/json")

				rec := httptest.NewRecorder()

				return rec, req
			},
			check: func(rec *httptest.ResponseRecorder) {
				var result ResponseAcceptTask
				b := rec.Body.Bytes()
				require.NoErrorf(t, json.Unmarshal(b, &result), "failed to parse http response body")

				expected := ResponseAcceptTask{
					TaskID:     taskId,
					TaskStatus: "new",
				}
				require.Equalf(t, expected, result, "diff response body")
			},
			prepareMock: func(m *mockTaskServer) {
				m.mockTaskHandler.EXPECT().Handle(gomock.Any(), taskservice.ParamsHandle{
					Method:  task.Method,
					Url:     *urlString,
					Headers: task.Headers,
				}).Return(taskservice.ResponseHandle{
					TaskID:     taskId,
					TaskStatus: "new",
				}, nil)
			},
		},
		// TODO add more test cases
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			mockServer := newMockTaskServer(t)

			testCase.prepareMock(mockServer)

			recordWriter, req := testCase.prepareHttp()
			mockServer.taskServer.Router.ServeHTTP(recordWriter, req)

			assert.Equal(t, testCase.status, recordWriter.Code)
			testCase.check(recordWriter)
		})
	}
}

// TODO add test for CheckTask
