// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.15.0
// source: tasks.sql

package taskprovider

import (
	"context"

	"encoding/json"
	"github.com/google/uuid"
)

const addHttpResponseData = `-- name: AddHttpResponseData :exec
update requests.tasks set task_status = $1, content_length = $2, status_code = $3, headers = $4 where task_id = $5
`

type AddHttpResponseDataParams struct {
	TaskStatus    string           `json:"task_status"`
	ContentLength *int64           `json:"content_length"`
	StatusCode    *int             `json:"status_code"`
	Headers       *json.RawMessage `json:"headers"`
	TaskID        uuid.UUID        `json:"task_id"`
}

func (q *Queries) AddHttpResponseData(ctx context.Context, arg AddHttpResponseDataParams) error {
	_, err := q.exec(ctx, q.addHttpResponseDataStmt, addHttpResponseData,
		arg.TaskStatus,
		arg.ContentLength,
		arg.StatusCode,
		arg.Headers,
		arg.TaskID,
	)
	return err
}

const createWithStatus = `-- name: CreateWithStatus :one
insert into requests.tasks (task_status) values ($1) returning task_id, task_status
`

type CreateWithStatusRow struct {
	TaskID     uuid.UUID `json:"task_id"`
	TaskStatus string    `json:"task_status"`
}

func (q *Queries) CreateWithStatus(ctx context.Context, taskStatus string) (CreateWithStatusRow, error) {
	row := q.queryRow(ctx, q.createWithStatusStmt, createWithStatus, taskStatus)
	var i CreateWithStatusRow
	err := row.Scan(&i.TaskID, &i.TaskStatus)
	return i, err
}

const getOne = `-- name: GetOne :one
select task_id, task_status, status_code, headers, content_length from requests.tasks where task_id = $1
`

func (q *Queries) GetOne(ctx context.Context, taskID uuid.UUID) (RequestsTask, error) {
	row := q.queryRow(ctx, q.getOneStmt, getOne, taskID)
	var i RequestsTask
	err := row.Scan(
		&i.TaskID,
		&i.TaskStatus,
		&i.StatusCode,
		&i.Headers,
		&i.ContentLength,
	)
	return i, err
}

const updateStatus = `-- name: UpdateStatus :exec
update requests.tasks set task_status = $1 where task_id = $2
`

type UpdateStatusParams struct {
	TaskStatus string    `json:"task_status"`
	TaskID     uuid.UUID `json:"task_id"`
}

func (q *Queries) UpdateStatus(ctx context.Context, arg UpdateStatusParams) error {
	_, err := q.exec(ctx, q.updateStatusStmt, updateStatus, arg.TaskStatus, arg.TaskID)
	return err
}
