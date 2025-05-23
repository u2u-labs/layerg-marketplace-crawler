// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.29.0
// source: erc_1155_asset.sql

package db

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
)

const add1155Asset = `-- name: Add1155Asset :one
INSERT INTO
    erc_1155_collection_assets (asset_id, chain_id, token_id, owner, balance, attributes)
VALUES (
    $1, $2, $3, $4, $5, $6
) ON CONFLICT ON CONSTRAINT UC_ERC1155_OWNER DO UPDATE SET
    balance = $5,
    attributes = $6
    
RETURNING id, chain_id, asset_id, token_id, owner, balance, attributes, created_at, updated_at
`

type Add1155AssetParams struct {
	AssetID    string         `json:"assetId"`
	ChainID    int32          `json:"chainId"`
	TokenID    string         `json:"tokenId"`
	Owner      string         `json:"owner"`
	Balance    string         `json:"balance"`
	Attributes sql.NullString `json:"attributes"`
}

func (q *Queries) Add1155Asset(ctx context.Context, arg Add1155AssetParams) (Erc1155CollectionAsset, error) {
	row := q.db.QueryRowContext(ctx, add1155Asset,
		arg.AssetID,
		arg.ChainID,
		arg.TokenID,
		arg.Owner,
		arg.Balance,
		arg.Attributes,
	)
	var i Erc1155CollectionAsset
	err := row.Scan(
		&i.ID,
		&i.ChainID,
		&i.AssetID,
		&i.TokenID,
		&i.Owner,
		&i.Balance,
		&i.Attributes,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}

const count1155AssetByAssetId = `-- name: Count1155AssetByAssetId :one
SELECT COUNT(*) FROM erc_1155_collection_assets 
WHERE asset_id = $1
`

func (q *Queries) Count1155AssetByAssetId(ctx context.Context, assetID string) (int64, error) {
	row := q.db.QueryRowContext(ctx, count1155AssetByAssetId, assetID)
	var count int64
	err := row.Scan(&count)
	return count, err
}

const count1155AssetByOwner = `-- name: Count1155AssetByOwner :one
SELECT COUNT(*) FROM erc_1155_collection_assets 
WHERE owner = $1
`

func (q *Queries) Count1155AssetByOwner(ctx context.Context, owner string) (int64, error) {
	row := q.db.QueryRowContext(ctx, count1155AssetByOwner, owner)
	var count int64
	err := row.Scan(&count)
	return count, err
}

const count1155AssetHolderByAssetId = `-- name: Count1155AssetHolderByAssetId :one
SELECT COUNT(DISTINCT(owner)) FROM erc_1155_collection_assets 
WHERE asset_id = $1
`

func (q *Queries) Count1155AssetHolderByAssetId(ctx context.Context, assetID string) (int64, error) {
	row := q.db.QueryRowContext(ctx, count1155AssetHolderByAssetId, assetID)
	var count int64
	err := row.Scan(&count)
	return count, err
}

const count1155AssetHolderByAssetIdAndTokenId = `-- name: Count1155AssetHolderByAssetIdAndTokenId :one
SELECT COUNT(DISTINCT(owner)) FROM erc_1155_collection_assets 
WHERE asset_id = $1
AND token_id = $2
AND owner = COALESCE($3, owner)
`

type Count1155AssetHolderByAssetIdAndTokenIdParams struct {
	AssetID string `json:"assetId"`
	TokenID string `json:"tokenId"`
	Owner   string `json:"owner"`
}

func (q *Queries) Count1155AssetHolderByAssetIdAndTokenId(ctx context.Context, arg Count1155AssetHolderByAssetIdAndTokenIdParams) (int64, error) {
	row := q.db.QueryRowContext(ctx, count1155AssetHolderByAssetIdAndTokenId, arg.AssetID, arg.TokenID, arg.Owner)
	var count int64
	err := row.Scan(&count)
	return count, err
}

const delete1155Asset = `-- name: Delete1155Asset :exec
DELETE 
FROM erc_1155_collection_assets
WHERE
    id = $1
`

func (q *Queries) Delete1155Asset(ctx context.Context, id uuid.UUID) error {
	_, err := q.db.ExecContext(ctx, delete1155Asset, id)
	return err
}

const delete1155AssetByAssetId = `-- name: Delete1155AssetByAssetId :exec
DELETE
FROM erc_1155_collection_assets
WHERE
    asset_id = $1
`

func (q *Queries) Delete1155AssetByAssetId(ctx context.Context, assetID string) error {
	_, err := q.db.ExecContext(ctx, delete1155AssetByAssetId, assetID)
	return err
}

const get1155AssetByAssetIdAndTokenId = `-- name: Get1155AssetByAssetIdAndTokenId :one
SELECT id, chain_id, asset_id, token_id, owner, balance, attributes, created_at, updated_at FROM erc_1155_collection_assets
WHERE
    asset_id = $1
    AND token_id = $2
`

type Get1155AssetByAssetIdAndTokenIdParams struct {
	AssetID string `json:"assetId"`
	TokenID string `json:"tokenId"`
}

func (q *Queries) Get1155AssetByAssetIdAndTokenId(ctx context.Context, arg Get1155AssetByAssetIdAndTokenIdParams) (Erc1155CollectionAsset, error) {
	row := q.db.QueryRowContext(ctx, get1155AssetByAssetIdAndTokenId, arg.AssetID, arg.TokenID)
	var i Erc1155CollectionAsset
	err := row.Scan(
		&i.ID,
		&i.ChainID,
		&i.AssetID,
		&i.TokenID,
		&i.Owner,
		&i.Balance,
		&i.Attributes,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}

const get1155AssetChain = `-- name: Get1155AssetChain :one
SELECT erc_1155_collection_assets.id, erc_1155_collection_assets.chain_id, erc_1155_collection_assets.asset_id, erc_1155_collection_assets.token_id, erc_1155_collection_assets.owner, erc_1155_collection_assets.balance, erc_1155_collection_assets.attributes, erc_1155_collection_assets.created_at, erc_1155_collection_assets.updated_at, assets.collection_address, chains.chain_id
FROM erc_1155_collection_assets
INNER JOIN assets ON assets.id = erc_1155_collection_assets.asset_id
INNER JOIN chains ON chains.id = assets.chain_id
WHERE
    asset_id = $1
  AND token_id = $2
`

type Get1155AssetChainParams struct {
	AssetID string `json:"assetId"`
	TokenID string `json:"tokenId"`
}

type Get1155AssetChainRow struct {
	ID                uuid.UUID      `json:"id"`
	ChainID           int32          `json:"chainId"`
	AssetID           string         `json:"assetId"`
	TokenID           string         `json:"tokenId"`
	Owner             string         `json:"owner"`
	Balance           string         `json:"balance"`
	Attributes        sql.NullString `json:"attributes"`
	CreatedAt         time.Time      `json:"createdAt"`
	UpdatedAt         time.Time      `json:"updatedAt"`
	CollectionAddress string         `json:"collectionAddress"`
	ChainID_2         int64          `json:"chainId2"`
}

func (q *Queries) Get1155AssetChain(ctx context.Context, arg Get1155AssetChainParams) (Get1155AssetChainRow, error) {
	row := q.db.QueryRowContext(ctx, get1155AssetChain, arg.AssetID, arg.TokenID)
	var i Get1155AssetChainRow
	err := row.Scan(
		&i.ID,
		&i.ChainID,
		&i.AssetID,
		&i.TokenID,
		&i.Owner,
		&i.Balance,
		&i.Attributes,
		&i.CreatedAt,
		&i.UpdatedAt,
		&i.CollectionAddress,
		&i.ChainID_2,
	)
	return i, err
}

const getDetailERC1155Assets = `-- name: GetDetailERC1155Assets :one
SELECT 
    asset_id,
    token_id,
    attributes,
    total_supply
FROM 
    erc_1155_total_supply
WHERE 
    asset_id = $1
AND 
    token_id = $2
`

type GetDetailERC1155AssetsParams struct {
	AssetID string `json:"assetId"`
	TokenID string `json:"tokenId"`
}

func (q *Queries) GetDetailERC1155Assets(ctx context.Context, arg GetDetailERC1155AssetsParams) (Erc1155TotalSupply, error) {
	row := q.db.QueryRowContext(ctx, getDetailERC1155Assets, arg.AssetID, arg.TokenID)
	var i Erc1155TotalSupply
	err := row.Scan(
		&i.AssetID,
		&i.TokenID,
		&i.Attributes,
		&i.TotalSupply,
	)
	return i, err
}

const getPaginated1155AssetByAssetId = `-- name: GetPaginated1155AssetByAssetId :many
SELECT id, chain_id, asset_id, token_id, owner, balance, attributes, created_at, updated_at FROM erc_1155_collection_assets 
WHERE asset_id = $1
LIMIT $2 OFFSET $3
`

type GetPaginated1155AssetByAssetIdParams struct {
	AssetID string `json:"assetId"`
	Limit   int32  `json:"limit"`
	Offset  int32  `json:"offset"`
}

func (q *Queries) GetPaginated1155AssetByAssetId(ctx context.Context, arg GetPaginated1155AssetByAssetIdParams) ([]Erc1155CollectionAsset, error) {
	rows, err := q.db.QueryContext(ctx, getPaginated1155AssetByAssetId, arg.AssetID, arg.Limit, arg.Offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []Erc1155CollectionAsset
	for rows.Next() {
		var i Erc1155CollectionAsset
		if err := rows.Scan(
			&i.ID,
			&i.ChainID,
			&i.AssetID,
			&i.TokenID,
			&i.Owner,
			&i.Balance,
			&i.Attributes,
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

const getPaginated1155AssetByOwnerAddress = `-- name: GetPaginated1155AssetByOwnerAddress :many
SELECT id, chain_id, asset_id, token_id, owner, balance, attributes, created_at, updated_at FROM erc_1155_collection_assets
WHERE
    owner = $1
LIMIT $2 OFFSET $3
`

type GetPaginated1155AssetByOwnerAddressParams struct {
	Owner  string `json:"owner"`
	Limit  int32  `json:"limit"`
	Offset int32  `json:"offset"`
}

func (q *Queries) GetPaginated1155AssetByOwnerAddress(ctx context.Context, arg GetPaginated1155AssetByOwnerAddressParams) ([]Erc1155CollectionAsset, error) {
	rows, err := q.db.QueryContext(ctx, getPaginated1155AssetByOwnerAddress, arg.Owner, arg.Limit, arg.Offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []Erc1155CollectionAsset
	for rows.Next() {
		var i Erc1155CollectionAsset
		if err := rows.Scan(
			&i.ID,
			&i.ChainID,
			&i.AssetID,
			&i.TokenID,
			&i.Owner,
			&i.Balance,
			&i.Attributes,
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

const update1155Asset = `-- name: Update1155Asset :exec
UPDATE erc_1155_collection_assets
SET
    owner = $2 
WHERE 
    id = $1
`

type Update1155AssetParams struct {
	ID    uuid.UUID `json:"id"`
	Owner string    `json:"owner"`
}

func (q *Queries) Update1155Asset(ctx context.Context, arg Update1155AssetParams) error {
	_, err := q.db.ExecContext(ctx, update1155Asset, arg.ID, arg.Owner)
	return err
}
