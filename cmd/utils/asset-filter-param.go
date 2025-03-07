package utils

import (
	"github.com/gin-gonic/gin"
	"github.com/unicornultrafoundation/go-u2u/common"
)

// GetAssetFilterParam get asset filter parameters
func GetAssetFilterParam(ctx *gin.Context) (bool, []string, string) {
	// Get the query parameters
	tokenIds := ctx.QueryArray("token_id")
	owner := ctx.Query("owner")

	hasQuery := false
	if len(tokenIds) != 0 || owner != "" {
		hasQuery = true
	}

	return hasQuery, tokenIds, owner
}

// GetAssetCollectionFilterParam get asset collection filter parameters
func GetAssetCollectionFilterParam(ctx *gin.Context) (bool, string) {
	// Get the query parameters
	collectionAddress := ctx.Query("collection_address")
	collectionAddress = common.HexToAddress(collectionAddress).Hex()
	hasQuery := false
	if collectionAddress != "" && collectionAddress != "0x0000000000000000000000000000000000000000" {
		hasQuery = true
	}

	return hasQuery, collectionAddress
}
