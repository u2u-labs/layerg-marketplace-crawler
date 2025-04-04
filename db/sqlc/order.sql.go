// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.27.0
// source: order.sql

package db

import (
	"context"

	"github.com/lib/pq"
)

const getOrderBySignature = `-- name: GetOrderBySignature :one
SELECT "Order".index, "Order".sig, "Order"."makerId", "Order"."makeAssetType", "Order"."makeAssetAddress", "Order"."makeAssetValue", "Order"."makeAssetId", "Order"."takerId", "Order"."takeAssetType", "Order"."takeAssetAddress", "Order"."takeAssetValue", "Order"."takeAssetId", "Order".salt, "Order".start, "Order"."end", "Order"."orderStatus", "Order"."orderType", "Order".root, "Order".proof, "Order"."tokenId", "Order"."collectionId", "Order".quantity, "Order".price, "Order"."priceNum", "Order"."netPrice", "Order"."netPriceNum", "Order"."createdAt", "Order"."updatedAt", "Order"."quoteToken", "Order"."filledQty"
FROM "Order"
    INNER JOIN "User" maker ON "Order"."makerId" = maker."id"
    LEFT JOIN "User" taker ON "Order"."takerId" = taker."id"
WHERE "sig" = $1 AND "index" = $2
`

type GetOrderBySignatureParams struct {
	Sig   string `json:"sig"`
	Index int32  `json:"index"`
}

func (q *Queries) GetOrderBySignature(ctx context.Context, arg GetOrderBySignatureParams) (Order, error) {
	row := q.db.QueryRowContext(ctx, getOrderBySignature, arg.Sig, arg.Index)
	var i Order
	err := row.Scan(
		&i.Index,
		&i.Sig,
		&i.MakerId,
		&i.MakeAssetType,
		&i.MakeAssetAddress,
		&i.MakeAssetValue,
		&i.MakeAssetId,
		&i.TakerId,
		&i.TakeAssetType,
		&i.TakeAssetAddress,
		&i.TakeAssetValue,
		&i.TakeAssetId,
		&i.Salt,
		&i.Start,
		&i.End,
		&i.OrderStatus,
		&i.OrderType,
		&i.Root,
		pq.Array(&i.Proof),
		&i.TokenId,
		&i.CollectionId,
		&i.Quantity,
		&i.Price,
		&i.PriceNum,
		&i.NetPrice,
		&i.NetPriceNum,
		&i.CreatedAt,
		&i.UpdatedAt,
		&i.QuoteToken,
		&i.FilledQty,
	)
	return i, err
}
