package db

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// GetCombinedAssetQueryScript
func GetCombinedNFTAssetQueryScript(ctx *gin.Context, limit int, offset int) (string, []interface{}) {
	whereClause, args := GetCombinedAssetWhereClause(ctx)
	filterClause, combinedWhereClause, filterArgs := GetCombinedAssetFilterClause(args, ctx)

	// Get the query parameters
	query := fmt.Sprintf(`
	SELECT * FROM (
		SELECT 
			'ERC721' as token_type,
			MIN(chain_id) as chain_id,
			asset_id,
			token_id,
			%s,
			attributes,
			MIN(created_at) as created_at
			
		FROM erc_721_collection_assets
		%s
		GROUP BY asset_id, token_id, owner, attributes
		UNION ALL
		SELECT 
			'ERC1155' as token_type,
			MIN(chain_id) as chain_id,
			asset_id,
			token_id,
			%s,
			attributes,
			MIN(created_at) as created_at
			
		FROM erc_1155_collection_assets
		%s
		GROUP BY asset_id, token_id, owner, attributes
	) combined
	%s
	ORDER BY combined.created_at ASC
	LIMIT %d OFFSET %d
	
	`, filterClause, whereClause, filterClause, whereClause, combinedWhereClause, limit, offset)

	return query, filterArgs
}

// GetCombinedAssetQueryScript
func GeCountCombinedNFTAssetQueryScript(ctx *gin.Context) (string, []interface{}) {
	whereClause, args := GetCombinedAssetWhereClause(ctx)
	filterClause, combinedWhereClause, filterArgs := GetCombinedAssetFilterClause(args, ctx)

	// Get the query parameters
	query := fmt.Sprintf(`
	SELECT COUNT(*) FROM (
		SELECT 
			'ERC721' as token_type,
			MIN(chain_id) as chain_id,
			asset_id,
			token_id,
			%s,
			attributes,
			MIN(created_at) as created_at
			
		FROM erc_721_collection_assets
		%s
		GROUP BY asset_id, token_id, owner, attributes
		UNION ALL
		SELECT 
			'ERC1155' as token_type,
			MIN(chain_id) as chain_id,
			asset_id,
			token_id,
			%s,
			attributes,
			MIN(created_at) as created_at
			
		FROM erc_1155_collection_assets
		%s
		GROUP BY asset_id, token_id, owner, attributes
	) 
	combined
	%s	
	`, filterClause, whereClause, filterClause, whereClause, combinedWhereClause)

	return query, filterArgs
}

// GetCombinedAssetWhereClause
func GetCombinedAssetWhereClause(ctx *gin.Context) (string, []interface{}) {
	chainIdStr := ctx.Param("chain_id")

	args := []interface{}{}
	whereClause := "WHERE chain_id = $1"

	if chainIdStr != "" {
		chainId, err := strconv.Atoi(chainIdStr)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return "", nil
		}
		args = append(args, chainId)
	}

	created_at_from := ctx.Query("created_at_from")
	created_at_to := ctx.Query("created_at_to")
	if created_at_from != "" {
		whereClause += " AND created_at >= $2"
		args = append(args, created_at_from)
	}

	if created_at_to != "" {
		whereClause += " AND created_at <= $3"
		args = append(args, created_at_to)
	}
	return whereClause, args
}

// GetCombinedAssetFilterClause
func GetCombinedAssetFilterClause(whereArgs []interface{}, ctx *gin.Context) (string, string, []interface{}) {
	owner := ctx.Query("owner")
	tokenIds := ctx.QueryArray("token_id")

	args := whereArgs
	ownerSelect := "owner"
	whereClause := fmt.Sprintf("WHERE combined.owner = $%d", len(args)+1)

	if owner == "" {
		ownerSelect = "NULL as owner" // Exclude owner if chainId is not provided
		whereClause = "WHERE combined.owner IS NULL"
	} else {
		args = append(args, owner)
	}

	if len(tokenIds) > 0 {
		whereClause += " AND combined.token_id IN ("
		startCount := len(args) + 1
		for i, tokenId := range tokenIds {
			whereClause += fmt.Sprintf("$%d", startCount+i)
			if i < len(tokenIds)-1 {
				whereClause += ","
			}

			args = append(args, tokenId)
		}
		whereClause += ")"
	}

	return ownerSelect, whereClause, args
}

func Count1155AssetHolderByAssetIdAndTokenIdQuery(assetId string, tokenId string, owner string) (string, []interface{}) {
	query := `
	SELECT COUNT(*) FROM erc_1155_collection_assets
	WHERE asset_id = $1 AND token_id = $2
	`
	args := []interface{}{assetId, tokenId}

	if owner != "" {
		query += " AND owner = $3"
		args = append(args, owner)
	}

	return query, args
}

func GetPaginated1155AssetOwnersByAssetIdAndTokenIdQuery(ctx *gin.Context, assetId string, tokenId string, owner string, limit int, offset int) (string, []interface{}) {
	query := `
	SELECT 
	id,
    owner, 
    balance, 
    created_at, 
    updated_at
	
	FROM erc_1155_collection_assets
	WHERE asset_id = $1 AND token_id = $2
	`

	args := []interface{}{assetId, tokenId, limit, offset}

	if owner != "" {
		query += " AND owner = $5"
		args = append(args, owner)
	}

	query += "LIMIT $3 OFFSET $4"

	return query, args
}
