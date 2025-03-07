package types

import db "github.com/u2u-labs/layerg-crawler/db/sqlc"

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
}
