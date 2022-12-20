package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/yerassyldanay/requestmaker/provider/metricsprovider"
	"github.com/yerassyldanay/requestmaker/server/rest/v1/middleware"
	"github.com/yerassyldanay/requestmaker/service/taskservice"

	"go.uber.org/zap"
)

type TaskServer struct {
	Router      *gin.Engine
	TaskService taskservice.TaskHandler
	logger      *zap.Logger
}

type TaskServerOption func(taskServer *TaskServer)

func NewTaskServer(logger *zap.Logger, taskService taskservice.TaskHandler, opts ...TaskServerOption) *TaskServer {
	taskServer := &TaskServer{
		TaskService: taskService,
		logger:      logger,
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

	registrer := prometheus.NewRegistry()
	httpMetrics := metricsprovider.GetHttpMetrics(registrer)

	router.GET("/metrics", middleware.PrometheusHandler(registrer))

	v1 := router.Group("/api/v1/")
	v1.Use(middleware.HttpRequestStats(httpMetrics), middleware.ValidateHeader())
	{
		v1.GET("/task/:task_id", s.CheckTask)
		v1.POST("/task", s.AcceptTask)
	}

	s.Router = router
}
