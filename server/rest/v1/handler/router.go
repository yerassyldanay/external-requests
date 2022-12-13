package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/yerassyldanay/requestmaker/service/taskservice"
)

type TaskServer struct {
	Router      *gin.Engine
	TaskService taskservice.TaskHandler
}

type TaskServerOption func(taskServer *TaskServer)

func NewTaskServer(taskService taskservice.TaskHandler, opts ...TaskServerOption) *TaskServer {
	taskServer := &TaskServer{
		TaskService: taskService,
	}
	taskServer.setRouter()

	for _, opt := range opts {
		opt(taskServer)
	}

	return taskServer
}

// @title           Task Handler Service
// @version         1.0.0
// @description     service accepts and handler requests to 3rd party services

// @contact.name   Yerassyl Danay

// @BasePath  /api/v1/
func (s *TaskServer) setRouter() {
	router := gin.Default()
	v1 := router.Group("/api/v1/")
	{
		v1.GET("/task/:task_id", s.CheckTask)
		v1.POST("/task", s.AcceptTask)
	}

	s.Router = router
}
