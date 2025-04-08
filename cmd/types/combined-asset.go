package types

import (
	"github.com/u2u-labs/layerg-crawler/cmd/utils"
	db "github.com/u2u-labs/layerg-crawler/db/sqlc"
)

type CombinedAsset struct {
	TokenType  string  `json:"tokenType"`
	ChainID    int     `json:"chainId"`
	AssetID    string  `json:"assetId"`
	TokenID    string  `json:"tokenId"`
	Owner      *string `json:"owner,omitempty"`
	Attributes string  `json:"attributes"`
	CreatedAt  string  `json:"createdAt"`
}

type Erc721CollectionAssetExtended struct {
	*db.Erc721CollectionAsset
	TxHash            string `json:"txHash"`
	CollectionAddress string `json:"collectionAddress"`
	From              string `json:"from"`
	To                string `json:"to"`
	LogIndex          uint   `json:"logIndex"`
	BlockNumber       string `json:"blockNumber"`
}

type FulfillOrderEvent struct {
	GameId       string  `json:"gameId"`
	CollectionId string  `json:"collectionId"`
	Amount       int32   `json:"amount"`
	NftId        string  `json:"nftId"`
	QuoteToken   string  `json:"quoteToken"`
	ChainId      int64   `json:"chainId"`
	PriceWei     string  `json:"priceWei"`
	Price        float64 `json:"price"`
	Timestamp    int64   `json:"timestamp"`
	OrderIndex   int32   `json:"orderIndex"`
	OrderSig     string  `json:"orderSig"`
	FilledQty    int32   `json:"filledQty"`
}

type Erc1155TransferSingleEventExtended struct {
	*utils.Erc1155TransferSingleEvent
	TxHash      string `json:"txHash"`
	AssetId     string `json:"assetId"`
	LogIndex    uint   `json:"logIndex"`
	BlockNumber string `json:"blockNumber"`
}

type OrderEventExtended struct {
	*db.OrderAsset
	BlockNumber string `json:"blockNumber"`
}
