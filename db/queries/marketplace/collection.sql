-- name: UpdateCollectionVolumeFloor :one
UPDATE "Collection"
SET vol = $1, "volumeWei" = $2,
    floor = $3, "floorWei" = $4, "floorPrice" = $5
WHERE "id" = $6
RETURNING *;

-- name: GetCollectionByAddressAndChainId :one
SELECT *
FROM "Collection"
WHERE "address" ILIKE $1
  AND "chainId" = $2;

-- name: GetCollectionById :one
SELECT *
FROM "Collection"
WHERE "id" = $1;
