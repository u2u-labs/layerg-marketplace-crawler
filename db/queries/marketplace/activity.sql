-- name: UpsertActivity :exec
INSERT INTO "Activity"
    ("id", "from", "to", "collectionId", "nftId", "userAddress", "type", "qty", "price", "createdAt", "logId")
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
ON CONFLICT ("logId") DO NOTHING;
