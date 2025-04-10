-- name: CreateOrderAsset :one
INSERT INTO order_asset (
    maker,
    taker,
    sig,
    "index",
    status,
    take_qty,
    filled_qty,
    nonce,
    timestamp,
    remaining,
    asset_id,
    chain_id,
    tx_hash
) VALUES (
    $1,   -- maker
    $2,   -- taker (nullable)
    $3,   -- sig
    $4,   -- index
    $5,   -- status
    $6,   -- take_qty
    $7,   -- filled_qty
    $8,   -- nonce
    $9,   -- timestamp
    $10,  -- remaining
    $11,  -- asset_id
    $12,  -- chain_id
    $13   -- tx_hash
         )
ON CONFLICT (asset_id, tx_hash)
    DO UPDATE SET asset_id = order_asset.asset_id -- No actual change, just to trigger RETURNING
RETURNING *;

-- name: GetOrderAssets :many
SELECT * FROM order_asset LIMIT $1 OFFSET $2;

-- name: CountOrderAssets :one
SELECT COUNT(*) FROM order_asset;

-- name: GetOrderAssetByTxHashAndChainId :one
SELECT * FROM order_asset WHERE tx_hash = $1 AND chain_id = $2;

-- name: GetOrderAssetByAssetId :one
SELECT * FROM order_asset WHERE asset_id = $1;
