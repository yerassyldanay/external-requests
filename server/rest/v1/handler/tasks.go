package handler

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/yerassyldanay/requestmaker/service/taskservice"
	"go.uber.org/zap"
)

type ErrMsg struct {
	Err string `json:"err"`
}

type ParamsAcceptTask struct {
	Method  string              `json:"method" binding:"required,oneof=GET POST PUT PATCH DELETE HEAD CONNECT OPTIONS TRACE"`
	Url     string              `json:"url" binding:"required,url"`
	Headers map[string][]string `json:"headers"`
}

type ResponseAcceptTask struct {
	TaskID     uuid.UUID `json:"task_id"`
	TaskStatus string    `json:"task_status"`
}

// AcceptTask
// @Tags task
// @Summary send task to the server
// @Description this endpoint accepts task from the client
// and handles it in the background
// @Accept  json
// @Produce  json
// @Param args body ParamsAcceptTask true "task info"
// @Success 200 {object} ResponseAcceptTask
// @Failure 400 {object} ErrMsg
// @Router /api/v1/task [POST]
func (server *TaskServer) AcceptTask(c *gin.Context) {
	var args = ParamsAcceptTask{
		Headers: make(map[string][]string, 10),
	}
	if err := c.ShouldBindJSON(&args); err != nil {
		c.JSON(http.StatusBadRequest, ErrMsg{
			Err: fmt.Sprintf("failed to parse params. err: %v", err),
		})
		return
	}

	parseUrl, err := url.Parse(args.Url)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrMsg{
			Err: fmt.Sprintf("failed to parse url provided. err: %v", err),
		})
		return
	}

	server.logger.Debug("parsed url", zap.Any("url", parseUrl))

	task, err := server.TaskService.Handle(context.Background(), taskservice.ParamsHandle{
		Method:  args.Method,
		Url:     *parseUrl,
		Headers: args.Headers,
	})
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrMsg{
			Err: fmt.Sprintf("failed to create a task with default values:  %v", err),
		})
		return
	}

	c.JSON(http.StatusOK, ResponseAcceptTask{
		TaskID:     task.TaskID,
		TaskStatus: task.TaskStatus,
	})
}

type ParamsCheckTask struct {
	TaskID string `uri:"task_id" binding:"required,uuid"`
}

type ResponseCheckTask struct {
	TaskID        string           `json:"task_id"`
	TaskStatus    string           `json:"task_status"`
	StatusCode    *int             `json:"status_code"`
	Headers       *json.RawMessage `json:"headers"`
	ContentLength *int64           `json:"content_length"`
}

// CheckTask
// @Tags task
// @Summary fetch info on the task
// @Description this endpoint provides info on the task (status, http response info, etc.)
// @Accept  json
// @Produce  json
// @Param        task_id    path     string  false  "id of the task"
// @Success 200 {object} ResponseCheckTask
// @Failure 400 {object} ErrMsg
// @Router /api/v1/task/{task_id} [GET]
func (server *TaskServer) CheckTask(c *gin.Context) {
	var args ParamsCheckTask

	if err := c.ShouldBindUri(&args); err != nil {
		c.JSON(http.StatusBadRequest, ErrMsg{
			Err: fmt.Sprintf("failed to parse params. err: %v", err),
		})
		return
	}

	uuidTaskId, err := uuid.Parse(args.TaskID)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrMsg{
			Err: fmt.Sprintf("failed to parse task id. err: %v", err),
		})
		return
	}

	taskInfo, err := server.TaskService.GetByID(context.Background(), taskservice.ParamGetByID{
		TaskID: uuidTaskId,
	})
	switch {
	case errors.Is(err, sql.ErrNoRows):
		c.JSON(http.StatusNotFound, ErrMsg{
			Err: err.Error(),
		})
		return
	case err != nil:
		c.JSON(http.StatusBadRequest, ErrMsg{
			Err: fmt.Sprintf("failed to fetch task info. err: %v", err),
		})
		return
	default:
	}

	c.JSON(http.StatusOK, ResponseCheckTask{
		TaskID:        taskInfo.TaskID.String(),
		TaskStatus:    taskInfo.TaskStatus,
		StatusCode:    taskInfo.StatusCode,
		Headers:       taskInfo.Headers,
		ContentLength: taskInfo.ContentLength,
	})
}
