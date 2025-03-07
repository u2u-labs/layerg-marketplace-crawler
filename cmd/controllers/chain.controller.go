package controllers

import (
	"context"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"

	"github.com/u2u-labs/layerg-crawler/cmd/response"
	"github.com/u2u-labs/layerg-crawler/cmd/services"
)

type ChainController struct {
	service *services.ChainService
	ctx     context.Context
	rdb     *redis.Client
}

func NewChainController(service *services.ChainService, ctx context.Context, rdb *redis.Client) *ChainController {
	return &ChainController{service, ctx, rdb}
}

// AddNewChain godoc
// @Summary      Add a new chain
// @Description  Add a new chain
// @Tags         chains
// @Accept       json
// @Produce      json
// @Param body body db.AddChainParams true "Add a new chain"
// @Security     ApiKeyAuth
// @Success      200 {object} response.ResponseData
// @Failure      500 {object} response.ErrorResponse
// @Example      { "id": 1, "chain": "U2U", "name": "Nebulas Testnet", "RpcUrl": "sre", "ChainId": 2484, "Explorer": "str", "BlockTime": 500 }
// @Router       /chain [post]
func (cc *ChainController) AddNewChain(ctx *gin.Context) {
	cc.service.AddNewChain(ctx)
}

// GetAllSupportedChain godoc
// @Summary      Get all supported chains
// @Description  Get all supported chains
// @Tags         chains
// @Accept       json
// @Produce      json
// @Security     ApiKeyAuth
// @Param chain_id query string false "Chain Id"
// @Router       /chain [get]
func (cc *ChainController) GetAllChains(ctx *gin.Context) {

	chainIdStr := ctx.Query("chain_id")

	if chainIdStr != "" {
		// get chain by id
		chainId, err := strconv.Atoi(chainIdStr)
		if err != nil {
			response.ErrorResponseData(ctx, http.StatusInternalServerError, err.Error())
			return
		}

		cc.service.GetChainById(chainId, ctx)
		return
	}

	// get all chains
	cc.service.GetAllSupportedChains(ctx)

}
