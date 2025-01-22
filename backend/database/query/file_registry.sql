-- name: CreateEmptyFile :one
insert into file_registry (
    user_id,
    file_name,
    lock_status,
    upload_status
) values (
    $1, $2, $3, $4
) returning *;

-- name: GetFileByID :one
select upload_status, lock_status, updated_at
from file_registry
where id = sqlc.arg(id)
for update;

-- name: GetFileByName :one
select * from file_registry
where
    file_name = sqlc.arg(file_name) 
    and user_id = sqlc.arg(user_id)
for update;

-- name: UpdateFileMetadata :one
update file_registry
set
    object_key = sqlc.arg(object_key),
    upload_status = sqlc.arg(upload_status),
    lock_status = sqlc.arg(lock_status),
    updated_at = current_timestamp
where id = sqlc.arg(id)
returning *;

-- name: UpdateFileName :one
update file_registry
set
    file_name = sqlc.arg(new_file_name),
    updated_at = current_timestamp
where id = sqlc.arg(id)
returning *;

-- name: LockFile :one
update file_registry
set
    lock_status = sqlc.arg(lock_status),
    updated_at = current_timestamp
where id = sqlc.arg(id)
returning *;

-- name: DeleteFile :exec
delete from file_registry
where id = sqlc.arg(id);