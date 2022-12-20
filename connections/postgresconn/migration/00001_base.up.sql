create schema if not exists requests;
create table requests.tasks (
    task_id uuid not null primary key DEFAULT gen_random_uuid(),
    task_status varchar not null default 'new',
    status_code int,
    headers json,
    content_length bigint
);
create index task_id_idx on requests.tasks using hash (task_id);