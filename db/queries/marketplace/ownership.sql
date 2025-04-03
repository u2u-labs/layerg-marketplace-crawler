-- name: GetOwnershipByUserAddressAndCollectionId :one
SELECT * FROM "Ownership"
WHERE "userAddress" ILIKE $1 AND "collectionId" = $2 AND "nftId" = $3;

-- name: UpsertOwnership :one
INSERT INTO "Ownership"
    ("id", "userAddress", "nftId", "collectionId", "quantity", "createdAt", "updatedAt", "chainId")
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
ON CONFLICT ("userAddress", "nftId", "collectionId")
    DO UPDATE
    SET "quantity"  = $5,
        "updatedAt" = COALESCE($7, now())
RETURNING *;

-- name: DeleteOwnership :exec
DELETE FROM "Ownership"
WHERE "userAddress" = $1 AND "nftId" = $2 AND "collectionId" = $3;
