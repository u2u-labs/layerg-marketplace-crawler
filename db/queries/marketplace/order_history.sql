-- name: UpsertOrderHistory :exec
INSERT INTO "OrderHistory" (id, index, sig, nonce, "fromId", "toId", "qtyMatch", price, "priceNum", timestamp)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
ON CONFLICT ("sig", "index") DO NOTHING;
