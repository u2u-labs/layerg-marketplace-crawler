-- name: GetAssetById :one
SELECT * FROM assets 
WHERE id = $1;

-- name: GetPaginatedAssetsByChainId :many
SELECT * FROM assets 
WHERE chain_id = $1
LIMIT $2 OFFSET $3;

-- name: CountAssetByChainId :one
SELECT COUNT(*) FROM assets 
WHERE chain_id = $1;

-- name: AddNewAsset :exec
INSERT INTO assets (
    id, chain_id, collection_address, type, decimal_data, initial_block, last_updated
)
VALUES (
    $1, $2, $3, $4, $5, $6, $7
) RETURNING *;
