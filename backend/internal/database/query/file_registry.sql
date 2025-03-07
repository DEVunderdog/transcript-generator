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
where id = sqlc.arg(id) and user_id = sqlc.arg(user_id)
for update;

-- name: GetFileByName :one
select * from file_registry
where file_name = sqlc.arg(file_name) and user_id = sqlc.arg(user_id);

-- name: GetFileByNameByLocking :one
select * from file_registry
where
    file_name = sqlc.arg(file_name)
    and user_id = sqlc.arg(user_id)
for update;

-- name: ListAllFiles :many
select id, file_name from file_registry
where
    user_id = sqlc.arg(user_id)
    and
    upload_status = sqlc.arg(upload_status)
    and
    lock_status = sqlc.arg(lock_status);

-- name: ListConflictingFiles :many
select id, object_key from file_registry
where ((lock_status = sqlc.arg(first_lock_condition) AND upload_status = sqlc.arg(first_upload_status)) OR
    (lock_status = sqlc.arg(second_lock_condition) AND upload_status = sqlc.arg(second_upload_status)) OR
    (lock_status = sqlc.arg(third_lock_condition) AND upload_status = sqlc.arg(third_upload_status)) OR
    (lock_status = sqlc.arg(fourth_lock_condition) AND upload_status = sqlc.arg(fourth_upload_status)) OR
    (lock_status = sqlc.arg(fifth_lock_condition) AND upload_status = sqlc.arg(fifth_upload_status)))
    AND
    user_id = sqlc.arg(user_id);

-- name: UpdateFileMetadata :one
update file_registry
set
    object_key = sqlc.arg(object_key),
    upload_status = sqlc.arg(upload_status),
    lock_status = sqlc.arg(lock_status),
    updated_at = current_timestamp
where id = sqlc.arg(id) and user_id = sqlc.arg(user_id)
returning *;

-- name: UpdateFileName :one
update file_registry
set
    file_name = sqlc.arg(new_file_name),
    updated_at = current_timestamp
where id = sqlc.arg(id) and user_id = sqlc.arg(user_id)
returning *;

-- name: UnlockAndLockFile :one
update file_registry
set
    upload_status = sqlc.arg(status),
    lock_status = sqlc.arg(lock_status),
    updated_at = current_timestamp
where id = sqlc.arg(id) and user_id = sqlc.arg(user_id)
returning *;

-- name: DeleteFiles :exec
delete from file_registry
where
    user_id = sqlc.arg(user_id)
    and
    id = sqlc.arg(id);