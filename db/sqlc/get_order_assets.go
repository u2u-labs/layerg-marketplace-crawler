package db

import (
	"fmt"
	"strings"

	"github.com/gin-gonic/gin"
)

func placeholder(i int) string {
	return fmt.Sprintf("$%d", i)

}

func OptionalString(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

// GetOrderAssetsQueryScript returns get order query, get total order query and args
func GetOrderAssetsQueryScript(c *gin.Context, limit, offset int) (string, string, []any, []any) {
	where, args := []string{"1 = 1"}, []any{}
	maker := OptionalString(c.Query("maker"))
	taker := OptionalString(c.Query("taker"))
	sig := OptionalString(c.Query("sig"))
	status := OptionalString(c.Query("status"))
	txHash := OptionalString(c.Query("tx_hash"))
	chainId := OptionalString(c.Query("chain_id"))
	assetId := OptionalString(c.Query("asset_id"))
	if v := maker; v != nil {
		where, args = append(where, fmt.Sprintf("maker ILIKE %s", placeholder(len(args)+1))), append(args, *v)
	}
	if v := taker; v != nil {
		where, args = append(where, fmt.Sprintf("taker ILIKE %s", placeholder(len(args)+1))), append(args, *v)
	}
	if v := sig; v != nil {
		where, args = append(where, fmt.Sprintf("sig ILIKE %s", placeholder(len(args)+1))), append(args, *v)
	}
	if v := status; v != nil {
		where, args = append(where, fmt.Sprintf("status ILIKE %s", placeholder(len(args)+1))), append(args, *v)
	}
	if v := txHash; v != nil {
		where, args = append(where, fmt.Sprintf("tx_hash ILIKE %s", placeholder(len(args)+1))), append(args, *v)
	}
	if v := chainId; v != nil {
		where, args = append(where, fmt.Sprintf("chain_id ILIKE %s", placeholder(len(args)+1))), append(args, *v)
	}
	if v := assetId; v != nil {
		where, args = append(where, fmt.Sprintf("asset_id ILIKE %s", placeholder(len(args)+1))), append(args, *v)
	}

	query := fmt.Sprintf(`
	SELECT id, maker, taker, sig, index, status, take_qty, filled_qty, nonce, timestamp, remaining, asset_id, chain_id, tx_hash, created_at, updated_at
	FROM order_asset
	WHERE %s 
	LIMIT %s OFFSET %s`, strings.Join(where, " AND "), placeholder(len(args)+1), placeholder(len(args)+2))
	queryArgs := append(args, limit, offset)

	totalQuery := fmt.Sprintf(`
	SELECT COUNT(*)
	FROM order_asset
	WHERE %s`, strings.Join(where, " AND "))

	return query, totalQuery, queryArgs, args
}
