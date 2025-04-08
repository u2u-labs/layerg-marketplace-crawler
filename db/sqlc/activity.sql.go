// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.27.0
// source: activity.sql

package db

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
)

const upsertActivity = `-- name: UpsertActivity :exec
INSERT INTO "Activity"
    ("id", "from", "to", "collectionId", "nftId", "userAddress", "type", "qty",
     "price", "createdAt", "logId", "blockNumber", "txHash")
VALUES ($1, $2, $3, $4, $5, $6, $7, $8,
        $9, $10, $11, $12, $13)
ON CONFLICT ("logId") DO UPDATE
SET "blockNumber" = EXCLUDED."blockNumber",
    "txHash"      = EXCLUDED."txHash"
`

type UpsertActivityParams struct {
	ID           string         `json:"id"`
	From         string         `json:"from"`
	To           string         `json:"to"`
	CollectionId uuid.UUID      `json:"collectionId"`
	NftId        string         `json:"nftId"`
	UserAddress  string         `json:"userAddress"`
	Type         string         `json:"type"`
	Qty          int32          `json:"qty"`
	Price        sql.NullString `json:"price"`
	CreatedAt    time.Time      `json:"createdAt"`
	LogId        sql.NullString `json:"logId"`
	BlockNumber  sql.NullString `json:"blockNumber"`
	TxHash       sql.NullString `json:"txHash"`
}

func (q *Queries) UpsertActivity(ctx context.Context, arg UpsertActivityParams) error {
	_, err := q.db.ExecContext(ctx, upsertActivity,
		arg.ID,
		arg.From,
		arg.To,
		arg.CollectionId,
		arg.NftId,
		arg.UserAddress,
		arg.Type,
		arg.Qty,
		arg.Price,
		arg.CreatedAt,
		arg.LogId,
		arg.BlockNumber,
		arg.TxHash,
	)
	return err
}
