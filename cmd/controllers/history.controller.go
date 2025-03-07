package controllers

import (
	"context"
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/u2u-labs/layerg-crawler/cmd/response"
	rdb "github.com/u2u-labs/layerg-crawler/db"
	db "github.com/u2u-labs/layerg-crawler/db/sqlc"
)

type HistoryController struct {
	db    *db.Queries
	rawDb *sql.DB
	ctx   context.Context
	rdb   *redis.Client
}

func NewHistoryController(db *db.Queries, rawDb *sql.DB, ctx context.Context, rdb *redis.Client) *HistoryController {
	return &HistoryController{db, rawDb, ctx, rdb}
}

// Get onchain history godoc
// @Summary      Get History of a transaction
// @Description  Get History of a transaction
// @Tags         history
// @Accept       json
// @Produce      json
// @Param tx_hash query string true "Tx Hash"
// @Success      200 {object} response.ResponseData
// @Security     ApiKeyAuth
// @Router       /history [get]
func (hc *HistoryController) GetHistory(ctx *gin.Context) {
	txHash := ctx.Query("tx_hash")

	// get onchain history in cache or db

	histories, err := rdb.GetHistoryCache(ctx, hc.rdb, txHash)

	if err != nil {
		response.ErrorResponseData(ctx, http.StatusInternalServerError, err.Error())
		return
	}

	if len(histories) != 0 {
		response.SuccessReponseData(ctx, http.StatusOK, histories)
		return
	}

	tx, err := hc.db.GetOnchainHistoriesByTxHash(ctx, txHash)

	if err != nil {
		response.ErrorResponseData(ctx, http.StatusInternalServerError, err.Error())
		return
	}
	// Cache new added chain
	if err = rdb.SetHistoriesCache(hc.ctx, hc.rdb, tx); err != nil {
		response.ErrorResponseData(ctx, http.StatusInternalServerError, err.Error())
		return
	}
	response.SuccessReponseData(ctx, http.StatusOK, tx)
}
