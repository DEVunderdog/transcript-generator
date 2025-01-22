-- name: CreateEncryptionKeys :one
insert into encryption_keys (
    public_key,
    private_key,
    is_active,
    purpose
) values (
    $1, $2, $3, $4
) returning *;

-- name: GetActiveKeyBasedOnPurpose :one
select id, public_key, private_key, created_at from encryption_keys
where is_active = 'true' and purpose = sqlc.arg('purpose');

-- name: UpdateKeyStatus :exec
update encryption_keys
set
    is_active = coalesce(sqlc.narg('is_active'), is_active),
    updated_at = current_timestamp
where
    id = sqlc.arg('id')
returning *;

-- name: CountEncryptionKeys :one
select count(*) from encryption_keys;