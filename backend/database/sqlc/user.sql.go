// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.27.0
// source: user.sql

package database

import (
	"context"
)

const createUsers = `-- name: CreateUsers :one
insert into users (
    email
) values (
    $1
) returning id, email, created_at, updated_at
`

func (q *Queries) CreateUsers(ctx context.Context, email string) (User, error) {
	row := q.db.QueryRow(ctx, createUsers, email)
	var i User
	err := row.Scan(
		&i.ID,
		&i.Email,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}

const getUsersID = `-- name: GetUsersID :one
select id from users
where email = $1
`

func (q *Queries) GetUsersID(ctx context.Context, email string) (int32, error) {
	row := q.db.QueryRow(ctx, getUsersID, email)
	var id int32
	err := row.Scan(&id)
	return id, err
}
