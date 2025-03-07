package controllers

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/u2u-labs/layerg-crawler/cmd/response"
	"github.com/u2u-labs/layerg-crawler/config"
	db "github.com/u2u-labs/layerg-crawler/db/sqlc"
	"github.com/unicornultrafoundation/go-u2u/common"
)

type BackFillController struct {
	db    *db.Queries
	rawDb *sql.DB
	ctx   context.Context
	rdb   *redis.Client
}

func NewBackFillController(db *db.Queries, rawDb *sql.DB, ctx context.Context, rdb *redis.Client) *BackFillController {
	return &BackFillController{db, rawDb, ctx, rdb}
}

// AddBackFillTracker godoc
// @Summary      Add a new asset collection to the chain
// @Description  Add a new asset collection to the chain
// @Tags         asset, backfill
// @Accept       json
// @Produce      json
// @Security     ApiKeyAuth
// @Param body body db.AddBackfillCrawlerParams true "Asset collection information"
// @Example      { "chainId": 1, "collectionAddress": "0x1234567890123456789012345678901234567890", "blockNumber": 1234567890 }
// @Router       /backfill [post]
func (bfc *BackFillController) AddBackFillTracker(ctx *gin.Context) {
	var params *db.AddBackfillCrawlerParams

	// Read the request body
	if err := ctx.ShouldBindJSON(&params); err != nil {
		response.ErrorResponseData(ctx, http.StatusInternalServerError, err.Error())
		return
	}

	// convert params.contractAddress to checksum
	params.CollectionAddress = common.HexToAddress(params.CollectionAddress).Hex()

	// add to db
	if err := bfc.db.AddBackfillCrawler(ctx, *params); err != nil {
		response.ErrorResponseData(ctx, http.StatusInternalServerError, err.Error())
		return
	}

	_ = bfc.rdb.Publish(bfc.ctx, config.NewAssetChannel, "ok").Err()
	fmt.Println("New asset added to db", "ok")
}
