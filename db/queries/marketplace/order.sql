-- name: GetOrderBySignature :one
SELECT "Order".*
FROM "Order"
    INNER JOIN "User" maker ON "Order"."makerId" = maker."id"
    LEFT JOIN "User" taker ON "Order"."takerId" = taker."id"
WHERE "sig" = $1 AND "index" = $2;

-- name: UpdateOrderBySignature :exec
UPDATE "Order"
SET "orderStatus" = $1,
    "updatedAt" = coalesce($2, now())
WHERE "sig" = $3 AND "index" = $4;
