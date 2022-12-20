package taskprovider

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/google/uuid"

	"github.com/stretchr/testify/require"
	"github.com/yerassyldanay/requestmaker/connections/postgresconn"
	"github.com/yerassyldanay/requestmaker/pkg/configx"
	"github.com/yerassyldanay/requestmaker/pkg/convx"
	"github.com/yerassyldanay/requestmaker/pkg/errorx"
	"github.com/yerassyldanay/requestmaker/provider/taskprovider"
)

func dbDropDatabase(t *testing.T, db *sql.DB) {
	_, err := db.Exec(`SELECT pg_terminate_backend(pid) FROM pg_stat_activity 
	WHERE pid <> pg_backend_pid() AND datname = 'test_db';`)
	require.NoError(t, err, "failed to revoke database from public")
	_, err = db.Exec("drop database if exists test_db")
	require.NoError(t, err, "failed to empty test playground")
}

func dbCreateDatabase(t *testing.T, db *sql.DB) {
	_, err := db.Exec("create database test_db")
	require.NoError(t, err, "failed to empty test playground")
	_, err = db.Exec("GRANT CONNECT ON DATABASE test_db TO public")
	require.NoError(t, err, "failed to grant database to public")
}

func getDbConnection(t *testing.T) *sql.DB {
	conf, err := configx.NewConfiguration()
	require.NoErrorf(t, err, "failed to parse configurations")

	{
		// connect to datastore & prepare a playground for tests
		db, err := sql.Open("postgres", fmt.Sprintf("host=%s port=%d user=%s password=%s sslmode=%s",
			conf.PostgresHostname, conf.PostgresPort, conf.PostgresUsername, conf.PostgresPassword, conf.PostgresSSLMode))
		require.NoErrorf(t, err, "failed to establish connection with database")
		require.NoErrorf(t, db.Ping(), "failed to ping datastore")

		dbDropDatabase(t, db)
		dbCreateDatabase(t, db)
	}

	conf.PostgresDBName = "test_db"
	db, err := postgresconn.NewDB(*conf)
	require.NoErrorf(t, err, "failed to establish connection with database")

	dir, err := os.Getwd()
	require.NoErrorf(t, err, "failed to Getwd")

	filePath := filepath.Join(dir, "/../../../connections/postgresconn/migration")

	// migrate
	driver, err := postgres.WithInstance(db, &postgres.Config{})
	errorx.PanicIfError(err)

	migrateInstance, err := migrate.NewWithDatabaseInstance(fmt.Sprintf("file://%s", filePath), "postgres", driver)
	errorx.PanicIfError(err)

	_ = migrateInstance.Up()

	return db
}

func TestConnection(t *testing.T) {
	db := getDbConnection(t)
	defer db.Close()
}

func TestTaskQueries(t *testing.T) {
	db := getDbConnection(t)
	defer db.Close()

	taskQuerier := taskprovider.New(db)

	type TaskLimited struct {
		TaskId     uuid.UUID
		TaskStatus string
	}

	var numberOfTasks = 10
	var tasks = make([]TaskLimited, 0, numberOfTasks)

	// creating tasks with provided status
	{
		for i := 0; i < numberOfTasks; i++ {
			task, err := taskQuerier.CreateWithStatus(context.Background(), "new")
			require.NoErrorf(t, err, "failed to create tasks with default values and provided status")
			tasks = append(tasks, TaskLimited{
				TaskId:     task.TaskID,
				TaskStatus: task.TaskStatus,
			})
		}
		_ = tasks
	}

	// updating status of tasks by their ids & check update by fetching tasks one by one
	{
		for i := 0; i < numberOfTasks; i++ {
			err := taskQuerier.UpdateStatus(context.Background(), taskprovider.UpdateStatusParams{
				TaskStatus: "progress",
				TaskID:     tasks[i].TaskId,
			})
			require.NoErrorf(t, err, "failed to update the status of the task")
		}

		for i := 0; i < numberOfTasks; i++ {
			task := tasks[i]
			taskFetched, err := taskQuerier.GetOne(context.Background(), task.TaskId)
			require.NoErrorf(t, err, "failed get task from database")
			taskExpected := taskprovider.RequestsTask{
				TaskID:     task.TaskId,
				TaskStatus: "progress",
			}
			require.Equalf(t, taskExpected, taskFetched, "got diff tasks")
		}
	}

	// prepare http response info & check updates
	var payload = json.RawMessage(`{"Content-Type": "application/json"}`)
	{
		for i := 0; i < numberOfTasks; i++ {
			err := taskQuerier.AddHttpResponseData(context.Background(), taskprovider.AddHttpResponseDataParams{
				TaskStatus:    "updated",
				ContentLength: convx.Pointer[int64](1),
				StatusCode:    convx.Pointer(200),
				Headers:       &payload,
				TaskID:        tasks[i].TaskId,
			})
			require.NoErrorf(t, err, "failed get task from database")
		}

		for i := 0; i < numberOfTasks; i++ {
			taskFetched, err := taskQuerier.GetOne(context.Background(), tasks[i].TaskId)
			require.NoErrorf(t, err, "failed get task from database")
			taskExpected := taskprovider.RequestsTask{
				TaskID:        tasks[i].TaskId,
				TaskStatus:    "updated",
				ContentLength: convx.Pointer[int64](1),
				StatusCode:    convx.Pointer(200),
				Headers:       &payload,
			}
			require.Equalf(t, taskExpected, taskFetched, "got diff tasks")
		}
	}
}
