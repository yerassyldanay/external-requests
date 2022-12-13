package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"go.uber.org/zap"

	"github.com/yerassyldanay/requestmaker/connections/kafkaconn"
	"github.com/yerassyldanay/requestmaker/connections/postgresconn"
	"github.com/yerassyldanay/requestmaker/connections/redisconn"
	"github.com/yerassyldanay/requestmaker/pkg/configx"
	"github.com/yerassyldanay/requestmaker/pkg/errorx"
	"github.com/yerassyldanay/requestmaker/provider/msgbroker"
	"github.com/yerassyldanay/requestmaker/provider/ratelimiter"
	"github.com/yerassyldanay/requestmaker/provider/taskprovider"
	"github.com/yerassyldanay/requestmaker/server/backgroundworker"
	"github.com/yerassyldanay/requestmaker/server/rest/v1/handler"
	"github.com/yerassyldanay/requestmaker/service/taskservice"
)

func main() {
	logger, err := zap.NewDevelopment()
	errorx.PanicIfError(err)

	// priting the version of the application
	logger.Info("[SERVICE] version v1.0.0...")

	// parsing the conf parameters
	conf, err := configx.NewConfiguration()
	errorx.PanicIfError(err)
	logger.Debug("parsed configuration", zap.Any("conf", conf))

	// DB connection is established here (PostgreSQL in our case)
	db, err := postgresconn.NewDB(*conf)
	errorx.PanicIfError(err)
	defer func() {
		logger.Info("closing the connection with datastore", zap.Any("err", db.Close()))
	}()
	errorx.PanicIfError(db.Ping())
	logger.Info("successfully pinged database...")

	dir, err := os.Getwd()
	errorx.PanicIfError(err)

	filePath := filepath.Join(dir, "/connections/postgresconn/migration")

	// migrate
	driver, err := postgres.WithInstance(db, &postgres.Config{})
	errorx.PanicIfError(err)

	migrateInstance, err := migrate.NewWithDatabaseInstance(fmt.Sprintf("file://%s", filePath), "postgres", driver)
	errorx.PanicIfError(err)

	_ = migrateInstance.Up()

	// kafka
	kafkaWriter := kafkaconn.NewMBWriter(conf.GetMBHosts(), conf.MBRequestsTopic)
	kafkaReader := kafkaconn.NewMBReader(conf.GetMBHosts(), conf.MBRequestsTopic, conf.MBRequestsPartition)
	mbConnection, err := msgbroker.NewMBConnection(
		kafkaWriter,
		kafkaReader,
		logger.With(zap.String("layer", "provider"), zap.String("type", "mbconnection")),
	)
	errorx.PanicIfError(err)
	defer mbConnection.Close()

	// redis connection
	redisConn, err := redisconn.NewRedisConnection(conf.RedisHost, conf.RedisPort, conf.RedisDatabase)
	errorx.PanicIfError(err)

	statusCmd := redisConn.Ping(context.Background())
	errorx.PanicIfError(statusCmd.Err())

	// rate limiter
	rateLimiter := ratelimiter.NewExternalReqLimit(logger, redisConn)

	// establishing connection to message broker
	backgroundWorker := backgroundworker.NewWorker(
		db,
		mbConnection,
		rateLimiter,
		logger.With(zap.String("layer", "server"), zap.String("type", "worker")),
	)

	ctxBackgroundWorker, cancalBackgroundWorker := context.WithCancel(context.Background())
	defer cancalBackgroundWorker()

	// this function is to handle requests tasks in the background
	go backgroundWorker.Start(ctxBackgroundWorker)

	// services
	taskService := taskservice.NewTaskHandle(
		taskprovider.New(db),
		mbConnection,
		logger.With(zap.String("layer", "service"), zap.String("type", "taskhandle")),
	)

	// server
	server := handler.NewTaskServer(taskService)
	go func(server *handler.TaskServer) {
		server.Router.Run(conf.ListenAddr)
	}(server)

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	<-c
}
