-- name: GetOrderBySignature :one
SELECT * FROM "Order" WHERE "sig" = $1 AND "index" = $2;
