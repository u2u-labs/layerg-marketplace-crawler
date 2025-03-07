-- name: GetPaginated20AssetByAssetId :many
SELECT * FROM erc_20_collection_assets 
WHERE asset_id = $1
LIMIT $2 OFFSET $3;
;

-- name: Count20AssetByAssetId :one
SELECT COUNT(*) FROM erc_20_collection_assets 
WHERE asset_id = $1;

-- name: Count20AssetHoldersByAssetId :one
SELECT COUNT(DISTINCT(owner)) FROM erc_20_collection_assets 
WHERE asset_id = $1;


-- name: Get20AssetByAssetIdAndTokenId :one
SELECT * FROM erc_20_collection_assets
WHERE
    asset_id = $1
    AND owner = $2;

-- name: GetPaginated20AssetByOwnerAddress :many
SELECT * FROM erc_20_collection_assets
WHERE
    owner = $1
LIMIT $2 OFFSET $3;

-- name: Count20AssetByOwner :one
SELECT COUNT(*) FROM erc_20_collection_assets 
WHERE owner = $1;


-- name: Add20Asset :exec
INSERT INTO
    erc_20_collection_assets (asset_id, chain_id, owner, balance)
VALUES (
    $1, $2, $3, $4
) ON CONFLICT (owner) DO UPDATE SET
    balance = $4
RETURNING *;

-- name: Update20Asset :exec
UPDATE erc_20_collection_assets
SET
    owner = $2 
WHERE 
    id = $1;

-- name: Delete20Asset :exec
DELETE 
FROM erc_20_collection_assets
WHERE
    id = $1;