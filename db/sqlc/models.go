// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.27.0

package db

import (
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/sqlc-dev/pqtype"
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

type Activity struct {
	ID           string         `json:"id"`
	From         string         `json:"from"`
	To           string         `json:"to"`
	CollectionId uuid.UUID      `json:"collectionId"`
	NftId        string         `json:"nftId"`
	UserAddress  string         `json:"userAddress"`
	Type         string         `json:"type"`
	Qty          int32          `json:"qty"`
	Price        sql.NullString `json:"price"`
	CreatedAt    time.Time      `json:"createdAt"`
	LogId        sql.NullString `json:"logId"`
	BlockNumber  sql.NullString `json:"blockNumber"`
	TxHash       sql.NullString `json:"txHash"`
}

type AnalysisCollection struct {
	ID           string    `json:"id"`
	CollectionId uuid.UUID `json:"collectionId"`
	KeyTime      string    `json:"keyTime"`
	Address      string    `json:"address"`
	Type         string    `json:"type"`
	Volume       string    `json:"volume"`
	Vol          float64   `json:"vol"`
	VolumeWei    string    `json:"volumeWei"`
	FloorPrice   int64     `json:"floorPrice"`
	Floor        float64   `json:"floor"`
	FloorWei     string    `json:"floorWei"`
	Items        int64     `json:"items"`
	Owner        int64     `json:"owner"`
	CreatedAt    time.Time `json:"createdAt"`
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
	BlockScanInterval sql.NullInt64 `json:"blockScanInterval"`
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

type Collection struct {
	ID             uuid.UUID             `json:"id"`
	TxCreationHash sql.NullString        `json:"txCreationHash"`
	Name           string                `json:"name"`
	NameSlug       sql.NullString        `json:"nameSlug"`
	Symbol         string                `json:"symbol"`
	Description    sql.NullString        `json:"description"`
	Address        sql.NullString        `json:"address"`
	ShortUrl       sql.NullString        `json:"shortUrl"`
	Metadata       sql.NullString        `json:"metadata"`
	IsU2U          bool                  `json:"isU2U"`
	Status         string                `json:"status"`
	Type           string                `json:"type"`
	CreatedAt      time.Time             `json:"createdAt"`
	UpdatedAt      time.Time             `json:"updatedAt"`
	CoverImage     sql.NullString        `json:"coverImage"`
	Avatar         sql.NullString        `json:"avatar"`
	ProjectId      uuid.NullUUID         `json:"projectId"`
	IsVerified     bool                  `json:"isVerified"`
	FloorPrice     int64                 `json:"floorPrice"`
	Floor          float64               `json:"floor"`
	FloorWei       string                `json:"floorWei"`
	IsActive       bool                  `json:"isActive"`
	FlagExtend     sql.NullBool          `json:"flagExtend"`
	IsSync         sql.NullBool          `json:"isSync"`
	SubgraphUrl    sql.NullString        `json:"subgraphUrl"`
	LastTimeSync   sql.NullInt32         `json:"lastTimeSync"`
	MetricPoint    sql.NullInt64         `json:"metricPoint"`
	MetricDetail   pqtype.NullRawMessage `json:"metricDetail"`
	MetadataJson   pqtype.NullRawMessage `json:"metadataJson"`
	GameLayergId   sql.NullString        `json:"gameLayergId"`
	Source         sql.NullString        `json:"source"`
	Vol            float64               `json:"vol"`
	VolumeWei      string                `json:"volumeWei"`
	ChainId        int64                 `json:"chainId"`
	TotalAssets    int32                 `json:"totalAssets"`
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

type NFT struct {
	ID             string          `json:"id"`
	Name           string          `json:"name"`
	CreatedAt      time.Time       `json:"createdAt"`
	UpdatedAt      time.Time       `json:"updatedAt"`
	Status         string          `json:"status"`
	TokenUri       string          `json:"tokenUri"`
	TxCreationHash string          `json:"txCreationHash"`
	CreatorId      uuid.NullUUID   `json:"creatorId"`
	CollectionId   uuid.UUID       `json:"collectionId"`
	Image          sql.NullString  `json:"image"`
	IsActive       bool            `json:"isActive"`
	Description    sql.NullString  `json:"description"`
	AnimationUrl   sql.NullString  `json:"animationUrl"`
	NameSlug       sql.NullString  `json:"nameSlug"`
	MetricPoint    int64           `json:"metricPoint"`
	MetricDetail   json.RawMessage `json:"metricDetail"`
	Source         sql.NullString  `json:"source"`
	OwnerId        string          `json:"ownerId"`
	Slug           sql.NullString  `json:"slug"`
}

type OnchainHistory struct {
	ID        uuid.UUID `json:"id"`
	From      string    `json:"from"`
	To        string    `json:"to"`
	AssetID   string    `json:"assetId"`
	TokenID   string    `json:"tokenId"`
	TxHash    string    `json:"txHash"`
	Timestamp time.Time `json:"timestamp"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
	Amount    string    `json:"amount"`
}

type Order struct {
	Index            int32         `json:"index"`
	Sig              string        `json:"sig"`
	MakerId          uuid.UUID     `json:"makerId"`
	MakeAssetType    int32         `json:"makeAssetType"`
	MakeAssetAddress string        `json:"makeAssetAddress"`
	MakeAssetValue   string        `json:"makeAssetValue"`
	MakeAssetId      string        `json:"makeAssetId"`
	TakerId          uuid.NullUUID `json:"takerId"`
	TakeAssetType    int32         `json:"takeAssetType"`
	TakeAssetAddress string        `json:"takeAssetAddress"`
	TakeAssetValue   string        `json:"takeAssetValue"`
	TakeAssetId      string        `json:"takeAssetId"`
	Salt             string        `json:"salt"`
	Start            int32         `json:"start"`
	End              int32         `json:"end"`
	OrderStatus      string        `json:"orderStatus"`
	OrderType        string        `json:"orderType"`
	Root             string        `json:"root"`
	Proof            []string      `json:"proof"`
	TokenId          string        `json:"tokenId"`
	CollectionId     uuid.UUID     `json:"collectionId"`
	Quantity         int32         `json:"quantity"`
	Price            string        `json:"price"`
	PriceNum         float64       `json:"priceNum"`
	NetPrice         string        `json:"netPrice"`
	NetPriceNum      float64       `json:"netPriceNum"`
	CreatedAt        time.Time     `json:"createdAt"`
	UpdatedAt        sql.NullTime  `json:"updatedAt"`
	QuoteToken       string        `json:"quoteToken"`
	FilledQty        int32         `json:"filledQty"`
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

type Ownership struct {
	ID           string         `json:"id"`
	UserAddress  string         `json:"userAddress"`
	NftId        sql.NullString `json:"nftId"`
	CollectionId uuid.NullUUID  `json:"collectionId"`
	Quantity     int32          `json:"quantity"`
	CreatedAt    time.Time      `json:"createdAt"`
	UpdatedAt    sql.NullTime   `json:"updatedAt"`
	ChainId      int32          `json:"chainId"`
}

type User struct {
	ID            uuid.UUID             `json:"id"`
	UaId          sql.NullString        `json:"uaId"`
	Mode          sql.NullString        `json:"mode"`
	Email         sql.NullString        `json:"email"`
	Avatar        sql.NullString        `json:"avatar"`
	Username      sql.NullString        `json:"username"`
	Signature     sql.NullString        `json:"signature"`
	SignedMessage sql.NullString        `json:"signedMessage"`
	Signer        string                `json:"signer"`
	PublicKey     sql.NullString        `json:"publicKey"`
	SignDate      sql.NullTime          `json:"signDate"`
	AcceptedTerms bool                  `json:"acceptedTerms"`
	CreatedAt     time.Time             `json:"createdAt"`
	UpdatedAt     time.Time             `json:"updatedAt"`
	Bio           sql.NullString        `json:"bio"`
	FacebookLink  sql.NullString        `json:"facebookLink"`
	TwitterLink   sql.NullString        `json:"twitterLink"`
	TelegramLink  sql.NullString        `json:"telegramLink"`
	ShortLink     sql.NullString        `json:"shortLink"`
	DiscordLink   sql.NullString        `json:"discordLink"`
	WebURL        sql.NullString        `json:"webURL"`
	CoverImage    sql.NullString        `json:"coverImage"`
	Followers     sql.NullInt32         `json:"followers"`
	Following     sql.NullInt32         `json:"following"`
	AccountStatus bool                  `json:"accountStatus"`
	VerifyEmail   bool                  `json:"verifyEmail"`
	IsActive      bool                  `json:"isActive"`
	MetricPoint   sql.NullInt64         `json:"metricPoint"`
	MetricDetail  pqtype.NullRawMessage `json:"metricDetail"`
	Type          sql.NullString        `json:"type"`
}
