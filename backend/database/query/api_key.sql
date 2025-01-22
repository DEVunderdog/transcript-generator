-- name: CreateAPIKey :one
insert into api_keys (
    user_id,
    credential,
    signature
) values (
    $1, $2, $3
) returning *;

-- name: GetAPIKey :one
select user_id, signature from api_keys
where credential = sqlc.arg('credential');

-- name: DeleteAPIKey :exec
delete from api_keys
where credential = sqlc.arg('credential');