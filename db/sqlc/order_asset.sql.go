// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.29.0
// source: order_asset.sql

package db

import (
	"context"
	"database/sql"
	"time"
)

const countOrderAssets = `-- name: CountOrderAssets :one
SELECT COUNT(*) FROM order_asset
`

func (q *Queries) CountOrderAssets(ctx context.Context) (int64, error) {
	row := q.db.QueryRowContext(ctx, countOrderAssets)
	var count int64
	err := row.Scan(&count)
	return count, err
}

const createOrderAsset = `-- name: CreateOrderAsset :one
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
RETURNING id, maker, taker, sig, index, status, take_qty, filled_qty, nonce, timestamp, remaining, asset_id, chain_id, tx_hash, created_at, updated_at
`

type CreateOrderAssetParams struct {
	Maker     string         `json:"maker"`
	Taker     sql.NullString `json:"taker"`
	Sig       string         `json:"sig"`
	Index     int32          `json:"index"`
	Status    OrderStatus    `json:"status"`
	TakeQty   string         `json:"takeQty"`
	FilledQty string         `json:"filledQty"`
	Nonce     string         `json:"nonce"`
	Timestamp time.Time      `json:"timestamp"`
	Remaining string         `json:"remaining"`
	AssetID   string         `json:"assetId"`
	ChainID   int32          `json:"chainId"`
	TxHash    string         `json:"txHash"`
}

func (q *Queries) CreateOrderAsset(ctx context.Context, arg CreateOrderAssetParams) (OrderAsset, error) {
	row := q.db.QueryRowContext(ctx, createOrderAsset,
		arg.Maker,
		arg.Taker,
		arg.Sig,
		arg.Index,
		arg.Status,
		arg.TakeQty,
		arg.FilledQty,
		arg.Nonce,
		arg.Timestamp,
		arg.Remaining,
		arg.AssetID,
		arg.ChainID,
		arg.TxHash,
	)
	var i OrderAsset
	err := row.Scan(
		&i.ID,
		&i.Maker,
		&i.Taker,
		&i.Sig,
		&i.Index,
		&i.Status,
		&i.TakeQty,
		&i.FilledQty,
		&i.Nonce,
		&i.Timestamp,
		&i.Remaining,
		&i.AssetID,
		&i.ChainID,
		&i.TxHash,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}

const deleteOrderAssetByAssetId = `-- name: DeleteOrderAssetByAssetId :exec
DELETE FROM order_asset WHERE asset_id = $1
`

func (q *Queries) DeleteOrderAssetByAssetId(ctx context.Context, assetID string) error {
	_, err := q.db.ExecContext(ctx, deleteOrderAssetByAssetId, assetID)
	return err
}

const getOrderAssetByAssetId = `-- name: GetOrderAssetByAssetId :one
SELECT id, maker, taker, sig, index, status, take_qty, filled_qty, nonce, timestamp, remaining, asset_id, chain_id, tx_hash, created_at, updated_at FROM order_asset WHERE asset_id = $1
`

func (q *Queries) GetOrderAssetByAssetId(ctx context.Context, assetID string) (OrderAsset, error) {
	row := q.db.QueryRowContext(ctx, getOrderAssetByAssetId, assetID)
	var i OrderAsset
	err := row.Scan(
		&i.ID,
		&i.Maker,
		&i.Taker,
		&i.Sig,
		&i.Index,
		&i.Status,
		&i.TakeQty,
		&i.FilledQty,
		&i.Nonce,
		&i.Timestamp,
		&i.Remaining,
		&i.AssetID,
		&i.ChainID,
		&i.TxHash,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}

const getOrderAssetByTxHashAndChainId = `-- name: GetOrderAssetByTxHashAndChainId :one
SELECT id, maker, taker, sig, index, status, take_qty, filled_qty, nonce, timestamp, remaining, asset_id, chain_id, tx_hash, created_at, updated_at FROM order_asset WHERE tx_hash = $1 AND chain_id = $2
`

type GetOrderAssetByTxHashAndChainIdParams struct {
	TxHash  string `json:"txHash"`
	ChainID int32  `json:"chainId"`
}

func (q *Queries) GetOrderAssetByTxHashAndChainId(ctx context.Context, arg GetOrderAssetByTxHashAndChainIdParams) (OrderAsset, error) {
	row := q.db.QueryRowContext(ctx, getOrderAssetByTxHashAndChainId, arg.TxHash, arg.ChainID)
	var i OrderAsset
	err := row.Scan(
		&i.ID,
		&i.Maker,
		&i.Taker,
		&i.Sig,
		&i.Index,
		&i.Status,
		&i.TakeQty,
		&i.FilledQty,
		&i.Nonce,
		&i.Timestamp,
		&i.Remaining,
		&i.AssetID,
		&i.ChainID,
		&i.TxHash,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}

const getOrderAssets = `-- name: GetOrderAssets :many
SELECT id, maker, taker, sig, index, status, take_qty, filled_qty, nonce, timestamp, remaining, asset_id, chain_id, tx_hash, created_at, updated_at FROM order_asset LIMIT $1 OFFSET $2
`

type GetOrderAssetsParams struct {
	Limit  int32 `json:"limit"`
	Offset int32 `json:"offset"`
}

func (q *Queries) GetOrderAssets(ctx context.Context, arg GetOrderAssetsParams) ([]OrderAsset, error) {
	rows, err := q.db.QueryContext(ctx, getOrderAssets, arg.Limit, arg.Offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []OrderAsset
	for rows.Next() {
		var i OrderAsset
		if err := rows.Scan(
			&i.ID,
			&i.Maker,
			&i.Taker,
			&i.Sig,
			&i.Index,
			&i.Status,
			&i.TakeQty,
			&i.FilledQty,
			&i.Nonce,
			&i.Timestamp,
			&i.Remaining,
			&i.AssetID,
			&i.ChainID,
			&i.TxHash,
			&i.CreatedAt,
			&i.UpdatedAt,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}
