package services

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/u2u-labs/layerg-crawler/cmd/helpers"
	"github.com/u2u-labs/layerg-crawler/cmd/response"
	"github.com/u2u-labs/layerg-crawler/cmd/types"
	"github.com/u2u-labs/layerg-crawler/cmd/utils"
	"github.com/u2u-labs/layerg-crawler/config"
	rdb "github.com/u2u-labs/layerg-crawler/db"
	db "github.com/u2u-labs/layerg-crawler/db/sqlc"
)

type AssetService struct {
	db    *db.Queries
	rawDb *sql.DB
	ctx   context.Context
	rdb   *redis.Client
}

func NewAssetService(db *db.Queries, rawDb *sql.DB, ctx context.Context, rdb *redis.Client) *AssetService {
	return &AssetService{db, rawDb, ctx, rdb}
}

func (as *AssetService) AddNewAsset(ctx *gin.Context) {

	var params *utils.AddNewAssetParamsUtil
	chainIdStr := ctx.Param("chain_id")
	chainId, err := strconv.Atoi(chainIdStr)

	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"status": "failed", "message": "Invalid chainId"})
		return
	}

	// Read the raw body
	rawBodyData, err := ctx.GetRawData()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read body"})
		return
	}

	// Unmarshal JSON into the struct
	if err := json.Unmarshal(rawBodyData, &params); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// read lastest block number from chain
	chain, err := as.db.GetChainById(ctx, int32(chainId))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	latestBlockNumber, err := helpers.GetLastestBlockFromChainUrl(chain.RpcUrl)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	params.ChainID = int32(chainId)
	params.ID = strconv.Itoa(int(chainId)) + ":" + params.CollectionAddress
	params.InitialBlock = utils.JsonNullInt64{
		NullInt64: sql.NullInt64{
			Int64: int64(latestBlockNumber),
			Valid: true,
		},
	}

	assetParam := utils.ConvertCustomTypeToSqlParams(params)

	// add to db
	if err := as.db.AddNewAsset(ctx, assetParam); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Output the result
	jsonResponse, err := utils.MarshalAssetParams(assetParam)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Cache the new added asset
	a, err := as.db.GetAssetById(ctx, assetParam.ID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if err = rdb.SetPendingAssetToCache(as.ctx, as.rdb, a); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	go func() {
		jsonBytes, err := json.Marshal(assetParam)
		if err != nil {
			fmt.Println("Error marshaling asset:", err)
			return
		}
		_ = as.rdb.Publish(as.ctx, config.NewAssetChannel, jsonBytes).Err()
		fmt.Println("New asset added to db", jsonResponse)
	}()

	response.SuccessReponseData(ctx, http.StatusCreated, jsonResponse)
}

func (as *AssetService) GetAssetByChainId(ctx *gin.Context) {
	chainIdStr := ctx.Param("chain_id")

	chainId, err := strconv.Atoi(chainIdStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"status": "failed", "message": "Invalid chainId"})
		return
	}

	page, limit, offset := db.GetLimitAndOffset(ctx)

	assets, err := as.db.GetPaginatedAssetsByChainId(ctx, db.GetPaginatedAssetsByChainIdParams{
		ChainID: int32(chainId),
		Limit:   int32(limit),
		Offset:  int32(offset),
	})

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Query total items count
	totalAssets, err := as.db.CountAssetByChainId(ctx, int32(chainId))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Create pagination response
	paginationResponse := db.Pagination[utils.AssetResponse]{
		Page:       page,
		Limit:      limit,
		TotalItems: totalAssets,
		TotalPages: (totalAssets + int64(limit) - 1) / int64(limit), // Calculate total pages
		Data:       utils.ConvertToAssetResponses(assets),
	}

	response.SuccessReponseData(ctx, http.StatusOK, paginationResponse)

}

func (as *AssetService) GetAssetCollectionByChainIdAndContractAddress(ctx *gin.Context, collectionAddress string) {
	chainIdStr := ctx.Param("chain_id")
	assetId := chainIdStr + ":" + collectionAddress
	assetCollection, err := as.db.GetAssetById(ctx, assetId)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			response.ErrorResponseData(ctx, http.StatusNotFound, "Failed to retrieve asset collection with this contract address in the chain")
			return
		}
		response.ErrorResponseData(ctx, http.StatusBadGateway, err.Error())
		return
	}

	response.SuccessReponseData(ctx, http.StatusOK, utils.ConvertAssetToAssetResponse(assetCollection))
}

func (as *AssetService) DeleteAssetCollection(ctx *gin.Context) {
	chainIdStr := ctx.Param("chain_id")
	var body struct {
		CollectionAddress string `json:"collectionAddress"`
	}
	if err := ctx.BindJSON(&body); err != nil {
		response.ErrorResponseData(ctx, http.StatusBadRequest, err.Error())
		return
	}
	tx, err := as.rawDb.BeginTx(ctx, nil)
	if err != nil {
		response.ErrorResponseData(ctx, http.StatusInternalServerError, err.Error())
		return
	}
	defer tx.Rollback()
	q := as.db.WithTx(tx)
	assetId := chainIdStr + ":" + body.CollectionAddress

	assetCollection, err := q.GetAssetById(ctx, assetId)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			response.ErrorResponseData(ctx, http.StatusNotFound, "Failed to retrieve asset collection with this contract address in the chain")
			return
		}
		response.ErrorResponseData(ctx, http.StatusBadGateway, err.Error())
		return
	}

	err = q.DeleteOnchainHistoriesByAssetId(ctx, assetId)
	if err != nil {
		response.ErrorResponseData(ctx, http.StatusInternalServerError, err.Error())
		return
	}
	err = q.Delete20AssetByAssetId(ctx, assetId)
	if err != nil {
		response.ErrorResponseData(ctx, http.StatusInternalServerError, err.Error())
		return
	}
	err = q.Delete721AssetByAssetId(ctx, assetId)
	if err != nil {
		response.ErrorResponseData(ctx, http.StatusInternalServerError, err.Error())
		return
	}
	err = q.Delete1155AssetByAssetId(ctx, assetId)
	if err != nil {
		response.ErrorResponseData(ctx, http.StatusInternalServerError, err.Error())
		return
	}
	err = q.DeleteOrderAssetByAssetId(ctx, assetId)
	if err != nil {
		response.ErrorResponseData(ctx, http.StatusInternalServerError, err.Error())
		return
	}

	err = q.DeleteAsset(ctx, assetId)
	if err != nil {
		response.ErrorResponseData(ctx, http.StatusInternalServerError, err.Error())
		return
	}
	if err = rdb.DeletePendingAssetInCache(as.ctx, as.rdb, assetCollection); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	err = tx.Commit()
	if err != nil {
		response.ErrorResponseData(ctx, http.StatusInternalServerError, err.Error())
		return
	}
	response.SuccessReponseData(ctx, http.StatusOK, utils.ConvertAssetToAssetResponse(assetCollection))

	go func() {
		jsonBytes, err := json.Marshal(assetCollection)
		if err != nil {
			fmt.Println("Error marshaling asset:", err)
			return
		}
		_ = as.rdb.Publish(as.ctx, config.NewAssetChannel, jsonBytes).Err()
	}()
}

func (as *AssetService) GetAssetsFromAssetCollectionId(ctx *gin.Context, assetId string) {
	assetCollection, err := as.db.GetAssetById(ctx, assetId)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			ctx.JSON(http.StatusNotFound, gin.H{"status": "failed", "message": "Failed to retrieve asset collection with this contract address in the chain"})
			return
		}
		ctx.JSON(http.StatusBadGateway, gin.H{"status": "Failed retrieving Asset", "error": err.Error()})
	}
	page, limit, offset := db.GetLimitAndOffset(ctx)

	switch assetType := assetCollection.Type; assetType {
	case db.AssetTypeERC721:

		holdersCount, err := as.db.Count721AssetHolderByAssetId(ctx, assetCollection.ID)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		totalAssets, err := as.db.Count721AssetByAssetId(ctx, assetCollection.ID)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		assets, _ := as.db.GetPaginated721AssetByAssetId(ctx, db.GetPaginated721AssetByAssetIdParams{
			AssetID: assetCollection.ID,
			Limit:   int32(limit),
			Offset:  int32(offset),
		})

		paginationResponse := db.Pagination[utils.Erc721CollectionAssetResponse]{
			Page:       page,
			Limit:      limit,
			TotalItems: totalAssets,
			TotalPages: (totalAssets + int64(limit) - 1) / int64(limit),
			Holders:    &holdersCount,
			Data:       utils.ConvertToErc721CollectionAssetResponses(assets),
		}

		ctx.JSON(http.StatusOK, gin.H{"status": "Successfully retrived id", "type": "ERC721", "asset": paginationResponse})
	case db.AssetTypeERC1155:
		holdersCount, err := as.db.Count1155AssetHolderByAssetId(ctx, assetCollection.ID)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		totalAssets, err := as.db.Count1155AssetByAssetId(ctx, assetCollection.ID)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		assets, _ := as.db.GetPaginated1155AssetByAssetId(ctx, db.GetPaginated1155AssetByAssetIdParams{
			AssetID: assetCollection.ID,
			Limit:   int32(limit),
			Offset:  int32(offset),
		})

		paginationResponse := db.Pagination[utils.Erc1155CollectionAssetResponse]{
			Page:       page,
			Limit:      limit,
			TotalItems: totalAssets,
			TotalPages: (totalAssets + int64(limit) - 1) / int64(limit),
			Holders:    &holdersCount,
			Data:       utils.ConvertToErc1155CollectionAssetResponses(assets),
		}

		ctx.JSON(http.StatusOK, gin.H{"status": "Successfully retrived id", "type": "ERC1155", "asset": paginationResponse})

	case db.AssetTypeERC20:
		holdersCount, err := as.db.Count20AssetHoldersByAssetId(ctx, assetCollection.ID)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		totalAssets, err := as.db.Count20AssetByAssetId(ctx, assetCollection.ID)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		assets, _ := as.db.GetPaginated20AssetByAssetId(ctx, db.GetPaginated20AssetByAssetIdParams{
			AssetID: assetCollection.ID,
			Limit:   int32(limit),
			Offset:  int32(offset),
		})

		paginationResponse := db.Pagination[db.Erc20CollectionAsset]{
			Page:       page,
			Limit:      limit,
			TotalItems: totalAssets,
			TotalPages: (totalAssets + int64(limit) - 1) / int64(limit),
			Holders:    &holdersCount,
			Data:       assets,
		}

		ctx.JSON(http.StatusOK, gin.H{"status": "Successfully retrived id", "type": "ERC20", "asset": paginationResponse})
	}
}

func (as *AssetService) GetAssetsFromCollectionWithFilter(ctx *gin.Context, assetId string, tokenIds []string, owner string) {
	assetCollection, err := as.db.GetAssetById(ctx, assetId)
	filterConditions := make(map[string][]string)

	if len(tokenIds) > 0 {
		filterConditions["token_id"] = tokenIds
	}

	if owner != "" {
		filterConditions["owner"] = []string{owner}
	}

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			ctx.JSON(http.StatusNotFound, gin.H{"status": "failed", "message": "Failed to retrieve asset collection with this contract address in the chain"})
			return
		}
		ctx.JSON(http.StatusBadGateway, gin.H{"status": "Failed retrieving Asset", "error": err.Error()})
	}
	page, limit, offset := db.GetLimitAndOffset(ctx)

	switch assetType := assetCollection.Type; assetType {
	case db.AssetTypeERC721:
		totalAssets, holdersCount, err := db.CountItemsWithFilter(as.rawDb, "erc_721_collection_assets", filterConditions)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		assets, _ := db.QueryWithDynamicFilter[db.Erc721CollectionAsset](as.rawDb, "erc_721_collection_assets", limit, offset, filterConditions)

		paginationResponse := db.Pagination[utils.Erc721CollectionAssetResponse]{
			Page:       page,
			Limit:      limit,
			TotalItems: int64(totalAssets),
			TotalPages: int64(totalAssets+(limit)-1) / int64(limit),
			Holders:    &holdersCount,
			Data:       utils.ConvertToErc721CollectionAssetResponses(assets),
		}

		ctx.JSON(http.StatusOK, gin.H{"status": "Successfully retrived id", "type": "ERC721", "asset": paginationResponse})

	case db.AssetTypeERC1155:
		totalAssets, holdersCount, err := db.CountItemsWithFilter(as.rawDb, "erc_1155_collection_assets", filterConditions)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		assets, _ := db.QueryWithDynamicFilter[db.Erc1155CollectionAsset](as.rawDb, "erc_1155_collection_assets", limit, offset, filterConditions)

		paginationResponse := db.Pagination[utils.Erc1155CollectionAssetResponse]{
			Page:       page,
			Limit:      limit,
			TotalItems: int64(totalAssets),
			TotalPages: int64(totalAssets+(limit)-1) / int64(limit),
			Holders:    &holdersCount,
			Data:       utils.ConvertToErc1155CollectionAssetResponses(assets),
		}

		ctx.JSON(http.StatusOK, gin.H{"status": "Successfully retrived id", "type": "ERC721", "asset": paginationResponse})

	case db.AssetTypeERC20:
		totalAssets, holdersCount, err := db.CountItemsWithFilter(as.rawDb, "erc_20_collection_assets", filterConditions)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		assets, _ := db.QueryWithDynamicFilter[db.Erc20CollectionAsset](as.rawDb, "erc_20_collection_assets", limit, offset, filterConditions)

		paginationResponse := db.Pagination[db.Erc20CollectionAsset]{
			Page:       page,
			Limit:      limit,
			TotalItems: int64(totalAssets),
			TotalPages: int64(totalAssets+(limit)-1) / int64(limit),
			Holders:    &holdersCount,
			Data:       assets,
		}

		ctx.JSON(http.StatusOK, gin.H{"status": "Successfully retrived id", "type": "ERC721", "asset": paginationResponse})

	}
}

func (as *AssetService) GetAssetByChainIdAndContractAddressDetail(ctx *gin.Context, assetId string, tokenId string, owner string) {
	assetCollection, err := as.db.GetAssetById(ctx, assetId)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			ctx.JSON(http.StatusNotFound, gin.H{"status": "failed", "message": "Failed to retrieve asset collection with this contract address in the chain"})
			return
		}
		ctx.JSON(http.StatusBadGateway, gin.H{"status": "Failed retrieving Asset", "error": err.Error()})
	}

	switch assetType := assetCollection.Type; assetType {
	case db.AssetTypeERC721:
		assetDetail, err := as.db.Get721AssetByAssetIdAndTokenId(ctx, db.Get721AssetByAssetIdAndTokenIdParams{
			AssetID: assetId,
			TokenID: tokenId,
		})

		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}

		ctx.JSON(http.StatusOK, gin.H{"status": "Successfully retrived id", "type": "ERC721", "asset": utils.ConvertToDetailErc721CollectionAssetResponses(assetDetail)})

	case db.AssetTypeERC1155:
		page, limit, offset := db.GetLimitAndOffset(ctx)
		countQuery, countArgs :=
			db.Count1155AssetHolderByAssetIdAndTokenIdQuery(assetId, tokenId, owner)

		countRow := as.rawDb.QueryRowContext(ctx, countQuery, countArgs...)
		var count int64
		err := countRow.Scan(&count)

		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}

		detailErc1155Params := db.GetDetailERC1155AssetsParams{
			AssetID: assetId,
			TokenID: tokenId,
		}

		query, args := db.GetPaginated1155AssetOwnersByAssetIdAndTokenIdQuery(ctx, assetId,
			tokenId,
			owner,
			limit,
			offset,
		)

		rows, err := as.rawDb.QueryContext(ctx, query, args...)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		defer rows.Close()

		var erc1155Owners []utils.ERC1155AssetOwnerResponse
		for rows.Next() {
			var owner utils.ERC1155AssetOwnerResponse
			if err := rows.Scan(&owner.Id, &owner.Owner, &owner.Balance, &owner.CreatedAt, &owner.UpdatedAt); err != nil {
				ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			}
			erc1155Owners = append(erc1155Owners, owner)
		}

		assetDetail, err := as.db.GetDetailERC1155Assets(ctx, detailErc1155Params)

		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		ownerPagination := db.Pagination[utils.ERC1155AssetOwnerResponse]{
			Page:       page,
			Limit:      limit,
			TotalItems: count,
			TotalPages: (count + int64(limit) - 1) / int64(limit),
			Data:       erc1155Owners,
		}

		ctx.JSON(http.StatusOK, gin.H{"status": "Successfully retrived id", "type": "ERC1155", "asset": utils.ConvertToDetailERC1155AssetResponse(assetDetail, ownerPagination)})

	}

}

func (as *AssetService) GetNFTCombinedAsset(ctx *gin.Context) {
	page, limit, offset := db.GetLimitAndOffset(ctx)

	// get combined asset
	query, args := db.GetCombinedNFTAssetQueryScript(ctx, limit, offset)

	rows, err := as.rawDb.QueryContext(ctx, query, args...)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}
	defer rows.Close()

	var assets []types.CombinedAsset
	for rows.Next() {
		var asset types.CombinedAsset

		// Ensure the order of Scan matches the SELECT statement
		if err := rows.Scan(&asset.TokenType, &asset.ChainID, &asset.AssetID, &asset.TokenID, &asset.Owner, &asset.Attributes, &asset.CreatedAt); err != nil {
			response.ErrorResponseData(ctx, http.StatusInternalServerError, err.Error())
		}
		assets = append(assets, asset)
	}

	// get count of combined asset
	query, args = db.GeCountCombinedNFTAssetQueryScript(ctx)

	countRow := as.rawDb.QueryRowContext(ctx, query, args...)
	var totalAssets int64
	err = countRow.Scan(&totalAssets)

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Check for errors from iterating over rows
	if err = rows.Err(); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}

	paginationResponse := db.Pagination[types.CombinedAsset]{
		Page:       page,
		Limit:      limit,
		TotalItems: totalAssets,
		TotalPages: (totalAssets + int64(limit) - 1) / int64(limit),
		Data:       assets,
	}

	response.SuccessReponseData(ctx, http.StatusOK, paginationResponse)
}

func (as *AssetService) GetOrderAssets(ctx *gin.Context) {
	page, limit, offset := db.GetLimitAndOffset(ctx)

	// get order asset
	query, countQuery, queryArgs, countArgs := db.GetOrderAssetsQueryScript(ctx, limit, offset)

	rows, err := as.rawDb.QueryContext(ctx, query, queryArgs...)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}
	defer rows.Close()

	var assets []db.OrderAsset
	for rows.Next() {
		var asset db.OrderAsset

		// Ensure the order of Scan matches the SELECT statement
		if err := rows.Scan(
			&asset.ID,
			&asset.Maker,
			&asset.Taker,
			&asset.Sig,
			&asset.Index,
			&asset.Status,
			&asset.TakeQty,
			&asset.FilledQty,
			&asset.Nonce,
			&asset.Timestamp,
			&asset.Remaining,
			&asset.AssetID,
			&asset.TxHash,
			&asset.CreatedAt,
			&asset.UpdatedAt,
		); err != nil {
			response.ErrorResponseData(ctx, http.StatusInternalServerError, err.Error())
		}
		assets = append(assets, asset)
	}

	// get count of order asset
	countRow := as.rawDb.QueryRowContext(ctx, countQuery, countArgs...)
	var totalAssets int64
	err = countRow.Scan(&totalAssets)

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Check for errors from iterating over rows
	if err = rows.Err(); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}

	paginationResponse := db.Pagination[db.OrderAsset]{
		Page:       page,
		Limit:      limit,
		TotalItems: totalAssets,
		TotalPages: (totalAssets + int64(limit) - 1) / int64(limit),
		Data:       assets,
	}

	response.SuccessReponseData(ctx, http.StatusOK, paginationResponse)
}
