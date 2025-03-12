-- name: UpdateCollectionVolume :one
UPDATE "Collection"
SET vol = $1, "volumeWei" = $2
WHERE "id" = $3
RETURNING *;

-- name: GetCollectionByAddressAndChainId :one
SELECT *
FROM "Collection"
WHERE "address" ILIKE $1
  AND "chainId" = $2;
