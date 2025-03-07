package config

import "time"

const (
	RetriveAddedChainsAndAssetsInterval = 2 * time.Second
	BackfillBlockRangeScan              = 100
	WorkerConcurrency                   = 10
	ProcessBatchBlocksLimit             = 100
	NewAssetChannel                     = "new_asset"
	MaxRetriesAttempt                   = 4
	MaxScanLogsLimit                    = 1000
)
