package backgroundworker

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/url"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/yerassyldanay/requestmaker/provider/msgbroker"
	"github.com/yerassyldanay/requestmaker/provider/ratelimiter"
	"github.com/yerassyldanay/requestmaker/provider/taskprovider"
	"github.com/yerassyldanay/requestmaker/service/httpclient"
)

const MaxConcurrentExternalRequests = 1e3

// Worker runs at the background
// -> consumes messages from message broker
// -> makes requests to 3rd party resources
// -> writes http response info to database
type Worker struct {
	MessageBroker           msgbroker.MessageBroker
	controlExternalRequests chan struct{}
	TaskWriter              taskprovider.Querier
	HttpClient              httpclient.RequestMaker
	logger                  *zap.Logger
	RateLimiter             ratelimiter.ExternalReqLimiter
}

func NewWorker(
	dbConn *sql.DB,
	messageBroker msgbroker.MessageBroker,
	rateLimiter ratelimiter.ExternalReqLimiter,
	logger *zap.Logger,
) *Worker {
	return &Worker{
		MessageBroker:           messageBroker,
		controlExternalRequests: make(chan struct{}, MaxConcurrentExternalRequests),
		TaskWriter:              taskprovider.New(dbConn),
		HttpClient:              http.DefaultClient,
		logger:                  logger,
		RateLimiter:             rateLimiter,
	}
}

type Task struct {
	TaskId  uuid.UUID           `json:"task_id"`
	Method  string              `json:"method"`
	Url     url.URL             `json:"url"`
	Headers map[string][]string `json:"headers"`
}

func (worker *Worker) Start(ctx context.Context) {

	worker.logger.Info("started consumer at the background...")

	taskChan, err := worker.MessageBroker.Consume(ctx)
	if err != nil {
		worker.logger.Panic("failed to add http response info", zap.Error(err))
	}

Loop:
	for {
		select {
		case <-ctx.Done():
			break Loop
		case taskByte := <-taskChan:
			go worker.handleMessage(ctx, taskByte)
		}
	}
}

func (worker *Worker) handleMessage(ctx context.Context, taskByte []byte) {
	worker.lock()
	defer worker.unlock()

	var task Task
	if err := json.Unmarshal(taskByte, &task); err != nil {
		worker.logger.Error("failed to unmarshal MB msg", zap.Error(err))
		return
	}

	// TODO check rate limit
	// thus we can avoid problems with IPs being blocked / load etc.

	// handle task here
	newRequest, err := http.NewRequest(task.Method, task.Url.String(), nil)
	if err != nil {
		worker.logger.Error("failed to create a new request", zap.Error(err))
		return
	}

	for key, values := range task.Headers {
		for _, value := range values {
			newRequest.Header.Add(key, value)
		}
	}

	response, err := http.DefaultClient.Do(newRequest)

	httpResponseParsed := worker.getHttpResponseParsed(task.TaskId, response, err)
	err = worker.TaskWriter.AddHttpResponseData(ctx, httpResponseParsed)
	if err != nil {
		worker.logger.Error("failed to add http response info to DB", zap.Error(err))
		return
	}
}

func (mb *Worker) getHttpResponseParsed(taskId uuid.UUID, response *http.Response, err error) taskprovider.AddHttpResponseDataParams {
	var taskStatus = "done"
	var statusCode *int = nil
	var contentLength *int64 = nil
	var rawMessage *json.RawMessage = nil

	switch {
	case err != nil:
		taskStatus = "error"
	default:
		if response.StatusCode%100 >= 3 {
			taskStatus = "error"
		}
		statusCode = &response.StatusCode
		contentLength = &response.ContentLength

		b, err := json.Marshal(response.Header)
		if err == nil {
			var tempRawMessage = json.RawMessage(b)
			rawMessage = &tempRawMessage
		}
	}

	return taskprovider.AddHttpResponseDataParams{
		TaskStatus:    taskStatus,
		ContentLength: contentLength,
		StatusCode:    statusCode,
		Headers:       rawMessage,
		TaskID:        taskId,
	}
}

func (mb *Worker) Close() {
	// TODO wait for goroutines to finish its work
}

func (worker *Worker) lock() {
	worker.controlExternalRequests <- struct{}{}
}

func (worker *Worker) unlock() {
	<-worker.controlExternalRequests
}
