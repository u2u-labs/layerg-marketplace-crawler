// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.28.0

package db

import (
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type AssetType string

const (
	AssetTypeERC721  AssetType = "ERC721"
	AssetTypeERC1155 AssetType = "ERC1155"
	AssetTypeERC20   AssetType = "ERC20"
	AssetTypeORDER   AssetType = "ORDER"
)

func (e *AssetType) Scan(src interface{}) error {
	switch s := src.(type) {
	case []byte:
		*e = AssetType(s)
	case string:
		*e = AssetType(s)
	default:
		return fmt.Errorf("unsupported scan type for AssetType: %T", src)
	}
	return nil
}

type NullAssetType struct {
	AssetType AssetType `json:"assetType"`
	Valid     bool      `json:"valid"` // Valid is true if AssetType is not NULL
}

// Scan implements the Scanner interface.
func (ns *NullAssetType) Scan(value interface{}) error {
	if value == nil {
		ns.AssetType, ns.Valid = "", false
		return nil
	}
	ns.Valid = true
	return ns.AssetType.Scan(value)
}

// Value implements the driver Valuer interface.
func (ns NullAssetType) Value() (driver.Value, error) {
	if !ns.Valid {
		return nil, nil
	}
	return string(ns.AssetType), nil
}

type CrawlerStatus string

const (
	CrawlerStatusCRAWLING CrawlerStatus = "CRAWLING"
	CrawlerStatusCRAWLED  CrawlerStatus = "CRAWLED"
)

func (e *CrawlerStatus) Scan(src interface{}) error {
	switch s := src.(type) {
	case []byte:
		*e = CrawlerStatus(s)
	case string:
		*e = CrawlerStatus(s)
	default:
		return fmt.Errorf("unsupported scan type for CrawlerStatus: %T", src)
	}
	return nil
}

type NullCrawlerStatus struct {
	CrawlerStatus CrawlerStatus `json:"crawlerStatus"`
	Valid         bool          `json:"valid"` // Valid is true if CrawlerStatus is not NULL
}

// Scan implements the Scanner interface.
func (ns *NullCrawlerStatus) Scan(value interface{}) error {
	if value == nil {
		ns.CrawlerStatus, ns.Valid = "", false
		return nil
	}
	ns.Valid = true
	return ns.CrawlerStatus.Scan(value)
}

// Value implements the driver Valuer interface.
func (ns NullCrawlerStatus) Value() (driver.Value, error) {
	if !ns.Valid {
		return nil, nil
	}
	return string(ns.CrawlerStatus), nil
}

type OrderStatus string

const (
	OrderStatusCANCELED       OrderStatus = "CANCELED"
	OrderStatusFILLED         OrderStatus = "FILLED"
	OrderStatusTRANSFER       OrderStatus = "TRANSFER"
	OrderStatusTRANSFERFILLED OrderStatus = "TRANSFERFILLED"
)

func (e *OrderStatus) Scan(src interface{}) error {
	switch s := src.(type) {
	case []byte:
		*e = OrderStatus(s)
	case string:
		*e = OrderStatus(s)
	default:
		return fmt.Errorf("unsupported scan type for OrderStatus: %T", src)
	}
	return nil
}

type NullOrderStatus struct {
	OrderStatus OrderStatus `json:"orderStatus"`
	Valid       bool        `json:"valid"` // Valid is true if OrderStatus is not NULL
}

// Scan implements the Scanner interface.
func (ns *NullOrderStatus) Scan(value interface{}) error {
	if value == nil {
		ns.OrderStatus, ns.Valid = "", false
		return nil
	}
	ns.Valid = true
	return ns.OrderStatus.Scan(value)
}

// Value implements the driver Valuer interface.
func (ns NullOrderStatus) Value() (driver.Value, error) {
	if !ns.Valid {
		return nil, nil
	}
	return string(ns.OrderStatus), nil
}

type App struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	SecretKey string    `json:"secretKey"`
}

type Asset struct {
	ID                string        `json:"id"`
	ChainID           int32         `json:"chainId"`
	CollectionAddress string        `json:"collectionAddress"`
	Type              AssetType     `json:"type"`
	CreatedAt         time.Time     `json:"createdAt"`
	UpdatedAt         time.Time     `json:"updatedAt"`
	DecimalData       sql.NullInt16 `json:"decimalData"`
	InitialBlock      sql.NullInt64 `json:"initialBlock"`
	LastUpdated       sql.NullTime  `json:"lastUpdated"`
}

type BackfillCrawler struct {
	ChainID           int32         `json:"chainId"`
	CollectionAddress string        `json:"collectionAddress"`
	CurrentBlock      int64         `json:"currentBlock"`
	Status            CrawlerStatus `json:"status"`
	CreatedAt         time.Time     `json:"createdAt"`
}

type Chain struct {
	ID          int32  `json:"id"`
	Chain       string `json:"chain"`
	Name        string `json:"name"`
	RpcUrl      string `json:"rpcUrl"`
	ChainID     int64  `json:"chainId"`
	Explorer    string `json:"explorer"`
	LatestBlock int64  `json:"latestBlock"`
	BlockTime   int32  `json:"blockTime"`
}

type Erc1155CollectionAsset struct {
	ID         uuid.UUID      `json:"id"`
	ChainID    int32          `json:"chainId"`
	AssetID    string         `json:"assetId"`
	TokenID    string         `json:"tokenId"`
	Owner      string         `json:"owner"`
	Balance    string         `json:"balance"`
	Attributes sql.NullString `json:"attributes"`
	CreatedAt  time.Time      `json:"createdAt"`
	UpdatedAt  time.Time      `json:"updatedAt"`
}

type Erc1155TotalSupply struct {
	AssetID     string         `json:"assetId"`
	TokenID     string         `json:"tokenId"`
	Attributes  sql.NullString `json:"attributes"`
	TotalSupply int64          `json:"totalSupply"`
}

type Erc20CollectionAsset struct {
	ID        uuid.UUID `json:"id"`
	ChainID   int32     `json:"chainId"`
	AssetID   string    `json:"assetId"`
	Owner     string    `json:"owner"`
	Balance   string    `json:"balance"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type Erc721CollectionAsset struct {
	ID         uuid.UUID      `json:"id"`
	ChainID    int32          `json:"chainId"`
	AssetID    string         `json:"assetId"`
	TokenID    string         `json:"tokenId"`
	Owner      string         `json:"owner"`
	Attributes sql.NullString `json:"attributes"`
	CreatedAt  time.Time      `json:"createdAt"`
	UpdatedAt  time.Time      `json:"updatedAt"`
}

type EvmAccount struct {
	ID           uuid.UUID `json:"id"`
	Address      string    `json:"address"`
	OnSaleCount  int64     `json:"onSaleCount"`
	HoldingCount int64     `json:"holdingCount"`
	CreatedAt    time.Time `json:"createdAt"`
	UpdatedAt    time.Time `json:"updatedAt"`
}

type Nft struct {
	Uid            uuid.UUID       `json:"uid"`
	TokenId        string          `json:"tokenId"`
	Name           string          `json:"name"`
	CreatedAt      time.Time       `json:"createdAt"`
	UpdatedAt      time.Time       `json:"updatedAt"`
	Status         string          `json:"status"`
	TokenUri       string          `json:"tokenUri"`
	TxCreationHash string          `json:"txCreationHash"`
	CreatorId      sql.NullString  `json:"creatorId"`
	CollectionId   string          `json:"collectionId"`
	ChainId        int64           `json:"chainId"`
	Image          sql.NullString  `json:"image"`
	Description    sql.NullString  `json:"description"`
	AnimationUrl   sql.NullString  `json:"animationUrl"`
	NameSlug       sql.NullString  `json:"nameSlug"`
	MetricPoint    int64           `json:"metricPoint"`
	MetricDetail   json.RawMessage `json:"metricDetail"`
	Source         sql.NullString  `json:"source"`
}

type OnchainHistory struct {
	ID        uuid.UUID `json:"id"`
	From      string    `json:"from"`
	To        string    `json:"to"`
	AssetID   string    `json:"assetId"`
	TokenID   string    `json:"tokenId"`
	Amount    string    `json:"amount"`
	TxHash    string    `json:"txHash"`
	Timestamp time.Time `json:"timestamp"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type OrderAsset struct {
	ID        uuid.UUID      `json:"id"`
	Maker     string         `json:"maker"`
	Taker     sql.NullString `json:"taker"`
	Sig       string         `json:"sig"`
	Index     int32          `json:"index"`
	Status    OrderStatus    `json:"status"`
	TakeQty   string         `json:"takeQty"`
	FilledQty string         `json:"filledQty"`
	Nonce     string         `json:"nonce"`
	Timestamp time.Time      `json:"timestamp"`
	Remaining string         `json:"remaining"`
	AssetID   string         `json:"assetId"`
	ChainID   int32          `json:"chainId"`
	TxHash    string         `json:"txHash"`
	CreatedAt time.Time      `json:"createdAt"`
	UpdatedAt time.Time      `json:"updatedAt"`
}
