-- name: GetAppById :one
SELECT * FROM apps WHERE id = $1;

-- name: GetAllApp :many
SELECT * FROM apps;


-- name: CreateApp :exec
INSERT INTO apps (
    name, secret_key
    )
VALUES (
    $1, $2
) RETURNING *;

