// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.28.0
// source: marketplace_nft.sql

package db

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

const getCollectionByAddressAndChainId = `-- name: GetCollectionByAddressAndChainId :one
SELECT id, "txCreationHash", name, "nameSlug", symbol, description, address, "shortUrl", metadata, "isU2U", status, type, "categoryId", "createdAt", "updatedAt", "coverImage", avatar, "projectId", "isVerified", "floorPrice", floor, "floorWei", "isActive", "flagExtend", "isSync", "subgraphUrl", "lastTimeSync", "metricPoint", "metricDetail", "metadataJson", "gameId", source, "categoryG", vol, "volumeWei", "chainId"
FROM "Collection"
WHERE "address" ILIKE $1
  AND "chainId" = $2
`

type GetCollectionByAddressAndChainIdParams struct {
	Address sql.NullString `json:"address"`
	ChainId int64          `json:"chainId"`
}

func (q *Queries) GetCollectionByAddressAndChainId(ctx context.Context, arg GetCollectionByAddressAndChainIdParams) (Collection, error) {
	row := q.db.QueryRowContext(ctx, getCollectionByAddressAndChainId, arg.Address, arg.ChainId)
	var i Collection
	err := row.Scan(
		&i.ID,
		&i.TxCreationHash,
		&i.Name,
		&i.NameSlug,
		&i.Symbol,
		&i.Description,
		&i.Address,
		&i.ShortUrl,
		&i.Metadata,
		&i.IsU2U,
		&i.Status,
		&i.Type,
		&i.CategoryId,
		&i.CreatedAt,
		&i.UpdatedAt,
		&i.CoverImage,
		&i.Avatar,
		&i.ProjectId,
		&i.IsVerified,
		&i.FloorPrice,
		&i.Floor,
		&i.FloorWei,
		&i.IsActive,
		&i.FlagExtend,
		&i.IsSync,
		&i.SubgraphUrl,
		&i.LastTimeSync,
		&i.MetricPoint,
		&i.MetricDetail,
		&i.MetadataJson,
		&i.GameId,
		&i.Source,
		&i.CategoryG,
		&i.Vol,
		&i.VolumeWei,
		&i.ChainId,
	)
	return i, err
}

const upsertNFT = `-- name: UpsertNFT :one
INSERT INTO "NFT" ("id", name, "createdAt", "updatedAt", status, "tokenUri", "txCreationHash",
                   "creatorId", "collectionId", image, description, "animationUrl",
                   "nameSlug", "metricPoint", "metricDetail", source, "ownerId")
VALUES ($1, $2, $3, $4, $5, $6, $7,
        $8, $9, $10, $11, $12, $13,
        $14, $15, $16, $17)
ON CONFLICT ("id", "collectionId")
    DO UPDATE SET name             = EXCLUDED.name,
                  "updatedAt"      = CURRENT_TIMESTAMP,
                  status           = EXCLUDED.status,
                  "tokenUri"       = EXCLUDED."tokenUri",
                  "txCreationHash" = EXCLUDED."txCreationHash",
                  "creatorId"      = EXCLUDED."creatorId",
                  image            = EXCLUDED.image,
                  description      = EXCLUDED.description,
                  "animationUrl"   = EXCLUDED."animationUrl",
                  "nameSlug"       = EXCLUDED."nameSlug",
                  "metricPoint"    = EXCLUDED."metricPoint",
                  source           = EXCLUDED.source,
                  "ownerId"        = EXCLUDED."ownerId"
RETURNING id, name, "createdAt", "updatedAt", status, "tokenUri", "txCreationHash", "creatorId", "collectionId", image, "isActive", description, "animationUrl", "nameSlug", "metricPoint", "metricDetail", source, "ownerId"
`

type UpsertNFTParams struct {
	ID             string          `json:"id"`
	Name           string          `json:"name"`
	CreatedAt      time.Time       `json:"createdAt"`
	UpdatedAt      time.Time       `json:"updatedAt"`
	Status         string          `json:"status"`
	TokenUri       string          `json:"tokenUri"`
	TxCreationHash string          `json:"txCreationHash"`
	CreatorId      uuid.NullUUID   `json:"creatorId"`
	CollectionId   uuid.UUID       `json:"collectionId"`
	Image          sql.NullString  `json:"image"`
	Description    sql.NullString  `json:"description"`
	AnimationUrl   sql.NullString  `json:"animationUrl"`
	NameSlug       sql.NullString  `json:"nameSlug"`
	MetricPoint    int64           `json:"metricPoint"`
	MetricDetail   json.RawMessage `json:"metricDetail"`
	Source         sql.NullString  `json:"source"`
	OwnerId        string          `json:"ownerId"`
}

func (q *Queries) UpsertNFT(ctx context.Context, arg UpsertNFTParams) (NFT, error) {
	row := q.db.QueryRowContext(ctx, upsertNFT,
		arg.ID,
		arg.Name,
		arg.CreatedAt,
		arg.UpdatedAt,
		arg.Status,
		arg.TokenUri,
		arg.TxCreationHash,
		arg.CreatorId,
		arg.CollectionId,
		arg.Image,
		arg.Description,
		arg.AnimationUrl,
		arg.NameSlug,
		arg.MetricPoint,
		arg.MetricDetail,
		arg.Source,
		arg.OwnerId,
	)
	var i NFT
	err := row.Scan(
		&i.ID,
		&i.Name,
		&i.CreatedAt,
		&i.UpdatedAt,
		&i.Status,
		&i.TokenUri,
		&i.TxCreationHash,
		&i.CreatorId,
		&i.CollectionId,
		&i.Image,
		&i.IsActive,
		&i.Description,
		&i.AnimationUrl,
		&i.NameSlug,
		&i.MetricPoint,
		&i.MetricDetail,
		&i.Source,
		&i.OwnerId,
	)
	return i, err
}
