-- name: UpsertActivity :exec
INSERT INTO "Activity"
    ("id", "from", "to", "collectionId", "nftId", "userAddress", "type", "qty", "price", "createdAt")
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
ON CONFLICT ("id") DO NOTHING;
