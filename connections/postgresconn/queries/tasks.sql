-- name: CreateWithStatus :one
insert into requests.tasks (task_status) values ($1) returning task_id, task_status;

-- name: UpdateStatus :exec
update requests.tasks set task_status = $1 where task_id = $2;

-- name: AddHttpResponseData :exec
update requests.tasks set task_status = $1, content_length = $2, status_code = $3, headers = $4 where task_id = $5;

-- name: GetOne :one
select * from requests.tasks where task_id = $1;
