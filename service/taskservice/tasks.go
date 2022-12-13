package taskservice

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strings"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/yerassyldanay/requestmaker/pkg/convx"
	"github.com/yerassyldanay/requestmaker/provider/msgbroker"
	"github.com/yerassyldanay/requestmaker/provider/taskprovider"
)

type StringList []string

func (h StringList) Contains(str string) bool {
	for _, eachStr := range h {
		if eachStr == str {
			return true
		}
	}

	return false
}

var httpMethods StringList = StringList{"GET", "HEAD", "POST", "PUT", "PATCH", "DELETE", "CONNECT", "OPTIONS", "TRACE"}

// TaskHandle is responsible for methods connected with handling tasks,
// and knowing its status
type TaskHandle struct {
	Querier       taskprovider.Querier
	MessageBroker msgbroker.MessageBroker
	logger        *zap.Logger
}

// NewTaskHandle creates a new instance of TaskHandle
func NewTaskHandle(taskQuerier taskprovider.Querier, messageBroker msgbroker.MessageBroker, logger *zap.Logger) *TaskHandle {
	return &TaskHandle{
		Querier:       taskQuerier,
		MessageBroker: messageBroker,
		logger:        logger,
	}
}

var _ TaskHandler = (*TaskHandle)(nil)

type ParamGetByID struct {
	TaskID uuid.UUID
}

type ResponseGetByID struct {
	TaskID        uuid.UUID        `json:"task_id"`
	TaskStatus    string           `json:"task_status"`
	StatusCode    *int             `json:"status_code"`
	Headers       *json.RawMessage `json:"headers"`
	ContentLength *int64           `json:"content_length"`
}

// GetByID returns the info on the status of the task by id
func (service TaskHandle) GetByID(ctx context.Context, args ParamGetByID) (ResponseGetByID, error) {
	// there is no need to validate the arguments
	taskInfo, err := service.Querier.GetOne(ctx, args.TaskID)
	switch {
	case errors.Is(err, sql.ErrNoRows):
		return ResponseGetByID{}, sql.ErrNoRows
	case err != nil:
		return ResponseGetByID{}, fmt.Errorf("failed to get task in TaskService. err: %v", err)
	default:
	}

	var response ResponseGetByID
	if err := convx.Copy(taskInfo, &response); err != nil {
		return ResponseGetByID{}, fmt.Errorf("failed to copy task info. err: %v", err)
	}

	return response, nil
}

type ParamsHandle struct {
	Method  string              `json:"method"`
	Url     url.URL             `json:"url"`
	Headers map[string][]string `json:"headers"`
}

type ResponseHandle struct {
	TaskID     uuid.UUID `json:"task_id"`
	TaskStatus string    `json:"task_status"`
}

type TaskParams struct {
	TaskID uuid.UUID `json:"task_id"`
	ParamsHandle
}

// Handle accepts the tasks and handles it in the background
func (service TaskHandle) Handle(ctx context.Context, args ParamsHandle) (ResponseHandle, error) {
	// validation at the business logic
	args.Method = strings.ToUpper(args.Method)
	if ok := httpMethods.Contains(args.Method); !ok {
		return ResponseHandle{}, fmt.Errorf("received %q, but it must be one of %v", args.Method, httpMethods)
	}

	// business logic
	task, err := service.Querier.CreateWithStatus(context.Background(), "new")
	if err != nil {
		return ResponseHandle{}, fmt.Errorf("failed to create a task with default values. err: %v", err)
	}

	b, err := json.Marshal(TaskParams{
		TaskID:       task.TaskID,
		ParamsHandle: args,
	})
	if err != nil {
		service.logger.Error("failed to marshal the task", zap.Error(err))
		return ResponseHandle{}, fmt.Errorf("failed to marshal the task. err: %v", err)
	}

	if err := service.MessageBroker.Publish(ctx, b); err != nil {
		service.logger.Error("failed to publish the message", zap.Error(err))
		return ResponseHandle{}, fmt.Errorf("failed to publish the message. err: %v", err)
	}

	service.logger.Debug("published a task & updated its status", zap.String("task_id", task.TaskID.String()))

	return ResponseHandle{
		TaskID:     task.TaskID,
		TaskStatus: task.TaskStatus,
	}, nil
}
