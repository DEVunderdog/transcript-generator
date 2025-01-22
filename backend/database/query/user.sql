-- name: CreateUsers :one
insert into users (
    email
) values (
    $1
) returning *;

-- name: GetUsersID :one
select id from users
where email = sqlc.arg('email');

