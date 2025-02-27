// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.28.0
// source: api_key.sql

package database

import (
	"context"
)

const createAPIKey = `-- name: CreateAPIKey :one
insert into api_keys (
    user_id,
    credential,
    signature
) values (
    $1, $2, $3
) returning id, user_id, credential, signature, created_at, updated_at
`

type CreateAPIKeyParams struct {
	UserID     int32  `json:"user_id"`
	Credential []byte `json:"credential"`
	Signature  []byte `json:"signature"`
}

func (q *Queries) CreateAPIKey(ctx context.Context, arg CreateAPIKeyParams) (ApiKey, error) {
	row := q.db.QueryRow(ctx, createAPIKey, arg.UserID, arg.Credential, arg.Signature)
	var i ApiKey
	err := row.Scan(
		&i.ID,
		&i.UserID,
		&i.Credential,
		&i.Signature,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}

const deleteAPIKey = `-- name: DeleteAPIKey :exec
delete from api_keys
where credential = $1
`

func (q *Queries) DeleteAPIKey(ctx context.Context, credential []byte) error {
	_, err := q.db.Exec(ctx, deleteAPIKey, credential)
	return err
}

const getAPIKey = `-- name: GetAPIKey :one
select user_id, signature from api_keys
where credential = $1
`

type GetAPIKeyRow struct {
	UserID    int32  `json:"user_id"`
	Signature []byte `json:"signature"`
}

func (q *Queries) GetAPIKey(ctx context.Context, credential []byte) (GetAPIKeyRow, error) {
	row := q.db.QueryRow(ctx, getAPIKey, credential)
	var i GetAPIKeyRow
	err := row.Scan(&i.UserID, &i.Signature)
	return i, err
}
