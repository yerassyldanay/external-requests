package taskservice_test

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/yerassyldanay/requestmaker/pkg/convx"
	"github.com/yerassyldanay/requestmaker/pkg/mockx"
	mock_msgbroker "github.com/yerassyldanay/requestmaker/provider/msgbroker/mock"
	"github.com/yerassyldanay/requestmaker/provider/taskprovider"
	mock_taskprovider "github.com/yerassyldanay/requestmaker/provider/taskprovider/mock"
	"github.com/yerassyldanay/requestmaker/service/taskservice"
)

type MockTaskHandle struct {
	TaskHandle    *taskservice.TaskHandle
	Querier       *mock_taskprovider.MockQuerier
	MessageBroker *mock_msgbroker.MockMessageBroker
}

func getMockTaskHandle(t *testing.T) MockTaskHandle {
	controller := gomock.NewController(t)
	logger, err := zap.NewDevelopment()
	require.NoErrorf(t, err, "failed to establish zap logger")
	messageBroker := mock_msgbroker.NewMockMessageBroker(controller)
	querier := mock_taskprovider.NewMockQuerier(controller)
	return MockTaskHandle{
		TaskHandle:    taskservice.NewTaskHandle(querier, messageBroker, logger),
		Querier:       querier,
		MessageBroker: messageBroker,
	}
}

func TestGetByID(t *testing.T) {
	t.Parallel()

	mockTaskHandle := getMockTaskHandle(t)
	_ = mockTaskHandle

	taskId := mockx.GetUUID(t)

	headers := json.RawMessage([]byte(`{"one":"one","three":"three","two":"two"}`))
	response := taskservice.ResponseGetByID{
		TaskID:        taskId,
		TaskStatus:    "status",
		StatusCode:    convx.Pointer(200),
		Headers:       &headers,
		ContentLength: convx.Pointer[int64](150),
	}

	var testCases = []struct {
		name     string
		params   taskservice.ParamGetByID
		response taskservice.ResponseGetByID
		err      error
		hasError bool
		prepare  func(m MockTaskHandle)
	}{
		{
			name: "ok",
			params: taskservice.ParamGetByID{
				TaskID: taskId,
			},
			response: response,
			hasError: false,
			err:      nil,
			prepare: func(m MockTaskHandle) {
				m.Querier.EXPECT().GetOne(gomock.Any(), taskId).Return(taskprovider.RequestsTask{
					TaskID:        response.TaskID,
					TaskStatus:    response.TaskStatus,
					StatusCode:    response.StatusCode,
					Headers:       response.Headers,
					ContentLength: response.ContentLength,
				}, nil)
			},
		},
		{
			name: "not found",
			params: taskservice.ParamGetByID{
				TaskID: taskId,
			},
			response: taskservice.ResponseGetByID{},
			hasError: true,
			err:      sql.ErrNoRows,
			prepare: func(m MockTaskHandle) {
				m.Querier.EXPECT().GetOne(gomock.Any(), taskId).Return(taskprovider.RequestsTask{}, sql.ErrNoRows)
			},
		},
		{
			name: "other error",
			params: taskservice.ParamGetByID{
				TaskID: taskId,
			},
			response: taskservice.ResponseGetByID{},
			hasError: true,
			err:      fmt.Errorf("failed to get task in TaskService. err: %v", errors.New("")),
			prepare: func(m MockTaskHandle) {
				m.Querier.EXPECT().GetOne(gomock.Any(), taskId).Return(taskprovider.RequestsTask{}, errors.New(""))
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.prepare(mockTaskHandle)
			resp, err := mockTaskHandle.TaskHandle.GetByID(context.Background(), testCase.params)
			require.Equalf(t, testCase.hasError, err != nil, "hasErr: %v & err: %v", testCase.hasError, err)
			if testCase.hasError {
				require.Equal(t, testCase.err, err)
			}
			require.Equal(t, testCase.response, resp)
		})
	}
}

func TestHandle(t *testing.T) {
	mockTaskHandle := getMockTaskHandle(t)

	taskId := mockx.GetUUID(t)
	urlExample, err := url.Parse("http://example.com/")
	require.NoErrorf(t, err, "failed to parse url")

	var testCases = []struct {
		name     string
		params   taskservice.ParamsHandle
		response taskservice.ResponseHandle
		err      error
		hasError bool
		prepare  func()
	}{
		{
			name: "ok",
			params: taskservice.ParamsHandle{
				Method: "get",
				Url:    *urlExample,
				Headers: map[string][]string{
					"First":  {"First"},
					"Second": {"Second"},
					"Third":  {"Third"},
				},
			},
			response: taskservice.ResponseHandle{
				TaskID:     taskId,
				TaskStatus: "new",
			},
			err:      nil,
			hasError: false,
			prepare: func() {
				callToDB := mockTaskHandle.Querier.EXPECT().CreateWithStatus(gomock.Any(), "new").
					Return(
						taskprovider.CreateWithStatusRow{
							TaskID:     taskId,
							TaskStatus: "new",
						}, nil,
					)

				b, err := json.Marshal(taskservice.TaskParams{
					TaskID: taskId,
					ParamsHandle: taskservice.ParamsHandle{
						Method: "GET",
						Url:    *urlExample,
						Headers: map[string][]string{
							"First":  {"First"},
							"Second": {"Second"},
							"Third":  {"Third"},
						},
					},
				})
				require.NoErrorf(t, err, "failed to marshal")
				mockTaskHandle.MessageBroker.EXPECT().Publish(gomock.Any(), b).Return(nil).After(callToDB)
			},
		},
	}

	for _, testCase := range testCases {
		testCase.prepare()
		resp, err := mockTaskHandle.TaskHandle.Handle(context.Background(), testCase.params)
		require.Equalf(t, testCase.hasError, err != nil, "hasErr: %v & err: %v", testCase.hasError, err)
		if testCase.hasError {
			require.Equal(t, testCase.err, err)
		}
		require.Equal(t, testCase.response, resp)
	}
}
