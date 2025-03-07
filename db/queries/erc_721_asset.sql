-- name: GetPaginated721AssetByAssetId :many
SELECT * FROM erc_721_collection_assets 
WHERE asset_id = $1
LIMIT $2 OFFSET $3;

-- name: Count721AssetByAssetId :one
SELECT COUNT(*) FROM erc_721_collection_assets 
WHERE asset_id = $1;

-- name: Count721AssetHolderByAssetId :one
SELECT COUNT(DISTINCT(owner)) FROM erc_721_collection_assets 
WHERE asset_id = $1;


-- name: Get721AssetByAssetIdAndTokenId :one
SELECT * FROM erc_721_collection_assets
WHERE
    asset_id = $1
    AND token_id = $2;


-- name: GetPaginated721AssetByOwnerAddress :many
SELECT * FROM erc_721_collection_assets
WHERE
    owner = $1
LIMIT $2 OFFSET $3;

-- name: Count721AssetByOwnerAddress :one
SELECT COUNT(*) FROM erc_721_collection_assets 
WHERE owner = $1;

-- name: Add721Asset :one
INSERT INTO
    erc_721_collection_assets (asset_id, chain_id, token_id, owner, attributes)
VALUES (
    $1, $2, $3, $4, $5
) ON CONFLICT ON CONSTRAINT UC_ERC721 DO UPDATE SET
    owner = $4,
    attributes = $5
RETURNING *;

-- name: Update721Asset :exec
UPDATE erc_721_collection_assets
SET
    owner = $2 
WHERE 
    id = $1;

-- name: Delete721Asset :exec
DELETE 
FROM erc_721_collection_assets
WHERE
    id = $1;