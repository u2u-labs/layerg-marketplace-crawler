package config

import "time"

const (
	RetriveAddedChainsAndAssetsInterval = 2 * time.Second
	BackfillBlockRangeScan              = 100
	WorkerConcurrency                   = 10
	ProcessBatchBlocksLimit             = 100
	NewAssetChannel                     = "new_asset"
	MaxRetriesAttempt                   = 4
	MaxScanLogsLimit                    = 800
	FillOrderChannel                    = "fill_order"
	CancelOrderChannel                  = "cancel_order"
	PusherChannelOrder                  = "marketplace_order"
	BackfillTimeInterval                = 2 * time.Second
	Erc721TransferEvent                 = "erc721_transfer"
	Erc1155TransferEvent                = "erc1155_transfer"

	ActivityEventTransfer    = "Transfer"
	ActivityEventFilledOrder = "FilledOrder"
	ActivityEventCancelOrder = "CancelOrder"
)
