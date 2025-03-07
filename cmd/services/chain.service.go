package services

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

type ChainService struct {
	db    *db.Queries
	rawDb *sql.DB
	ctx   context.Context
	rdb   *redis.Client
}

func NewChainService(db *db.Queries, rawDb *sql.DB, ctx context.Context, rdb *redis.Client) *ChainService {
	return &ChainService{db, rawDb, ctx, rdb}
}

func (cs *ChainService) AddNewChain(ctx *gin.Context) {
	var params *db.AddChainParams

	// Read the request body
	if err := ctx.ShouldBindJSON(&params); err != nil {
		response.ErrorResponseData(ctx, http.StatusInternalServerError, err.Error())
		return
	}

	// add to db
	if err := cs.db.AddChain(ctx, *params); err != nil {
		response.ErrorResponseData(ctx, http.StatusInternalServerError, err.Error())
		return
	}

	// response
	c, err := cs.db.GetChainById(ctx, params.ID)
	if err != nil {
		response.ErrorResponseData(ctx, http.StatusInternalServerError, err.Error())
		return
	}

	// Cache new added chain
	if err = rdb.SetPendingChainToCache(cs.ctx, cs.rdb, c); err != nil {
		response.ErrorResponseData(ctx, http.StatusInternalServerError, err.Error())
		return
	}

	response.SuccessReponseData(ctx, http.StatusOK, c)

}

func (cs *ChainService) GetAllSupportedChains(ctx *gin.Context) {
	// get all chains
	chains, err := cs.db.GetAllChain(ctx)
	if err != nil {
		response.ErrorResponseData(ctx, http.StatusInternalServerError, err.Error())
		return
	}
	response.SuccessReponseData(ctx, http.StatusOK, chains)
}

func (cs *ChainService) GetChainById(chainId int, ctx *gin.Context) {

	chain, err := cs.db.GetChainById(ctx, int32(chainId))
	if err != nil {
		response.ErrorResponseData(ctx, http.StatusInternalServerError, err.Error())
		return
	}
	response.SuccessReponseData(ctx, http.StatusOK, chain)
}
