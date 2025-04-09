-- name: CreateOrderHistory :exec
INSERT INTO "OrderHistory" (id, index, sig, nonce, "fromId", "toId", "qtyMatch", price, "priceNum", timestamp)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10);

-- name: CheckOrderHistoryExists :one
SELECT EXISTS (
    SELECT 1
    FROM "OrderHistory"
    WHERE sig = $1 AND index = $2
) AS exists;
