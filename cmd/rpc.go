package cmd

import (
	"github.com/ethereum/go-ethereum/ethclient"
	db "github.com/u2u-labs/layerg-crawler/db/sqlc"
)

func initChainClient(chain *db.Chain) (*ethclient.Client, error) {
	return ethclient.Dial(chain.RpcUrl)
}
