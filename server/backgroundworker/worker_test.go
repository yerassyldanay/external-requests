package backgroundworker

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"github.com/yerassyldanay/requestmaker/pkg/convx"
	"github.com/yerassyldanay/requestmaker/pkg/mockx"
	mock_msgbroker "github.com/yerassyldanay/requestmaker/provider/msgbroker/mock"
	"github.com/yerassyldanay/requestmaker/provider/taskprovider"
	mock_taskprovider "github.com/yerassyldanay/requestmaker/provider/taskprovider/mock"
	mock_httpclient "github.com/yerassyldanay/requestmaker/service/httpclient/mock"
)

type mockWorker struct {
	worker            Worker
	mockMessageBroker *mock_msgbroker.MockMessageBroker
	mockTaskWriter    *mock_taskprovider.MockQuerier
	mockHttpClient    *mock_httpclient.MockRequestMaker
}

func NewMockWorker(t *testing.T) *mockWorker {
	controller := gomock.NewController(t)
	mockMessageBroker := mock_msgbroker.NewMockMessageBroker(controller)
	mockTaskWriter := mock_taskprovider.NewMockQuerier(controller)
	mockHttpClient := mock_httpclient.NewMockRequestMaker(controller)
	return &mockWorker{
		worker: Worker{
			MessageBroker: mockMessageBroker,
			TaskWriter:    mockTaskWriter,
			HttpClient:    mockHttpClient,
		},
		mockMessageBroker: mockMessageBroker,
		mockTaskWriter:    mockTaskWriter,
		mockHttpClient:    mockHttpClient,
	}
}

func TestGetHttpResponseParsed(t *testing.T) {
	taskId := mockx.GetUUID(t)

	var headerStr = `{"Content-Language":["language1","language2"],"Content-Type":["application/json"]}`

	type param struct {
		taskId       uuid.UUID
		httpResponse *http.Response
		err          error
	}

	testCases := []struct {
		name     string
		params   param
		response taskprovider.AddHttpResponseDataParams
	}{
		{
			name: "ok",
			params: param{
				taskId: taskId,
				httpResponse: &http.Response{
					Status:        "OK",
					StatusCode:    http.StatusOK,
					ContentLength: 1,
					Header: map[string][]string{
						"Content-Type": {
							"application/json",
						},
						"Content-Language": {
							"language1",
							"language2",
						},
					},
				},
				err: nil,
			},
			response: taskprovider.AddHttpResponseDataParams{
				TaskStatus:    "done",
				ContentLength: convx.Pointer[int64](1),
				StatusCode:    convx.Pointer(http.StatusOK),
				Headers:       convx.Pointer[json.RawMessage]([]byte(headerStr)),
				TaskID:        taskId,
			},
		},
		{
			name: "ok",
			params: param{
				taskId: taskId,
				httpResponse: &http.Response{
					Status:        "OK",
					StatusCode:    http.StatusOK,
					ContentLength: 1,
					Header: map[string][]string{
						"Content-Type": {
							"application/json",
						},
						"Content-Language": {
							"language1",
							"language2",
						},
					},
				},
				err: http.ErrHandlerTimeout,
			},
			response: taskprovider.AddHttpResponseDataParams{
				TaskStatus:    "error",
				ContentLength: nil,
				StatusCode:    nil,
				Headers:       nil,
				TaskID:        taskId,
			},
		},
		// TODO more test cases must be added
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			var mockBackgroundWorker = NewMockWorker(t)
			parsedHttpResponse := mockBackgroundWorker.worker.getHttpResponseParsed(
				testCase.params.taskId,
				testCase.params.httpResponse,
				testCase.params.err,
			)
			require.Equalf(t, testCase.response, parsedHttpResponse, "got diff parsed parameters")
		})
	}
}
