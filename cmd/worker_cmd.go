package cmd

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"math/big"
	"strconv"
	"strings"

	u2u "github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/hibiken/asynq"
	"github.com/redis/go-redis/v9"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/u2u-labs/layerg-crawler/cmd/helpers"
	"github.com/u2u-labs/layerg-crawler/cmd/utils"
	"github.com/u2u-labs/layerg-crawler/config"
	"github.com/u2u-labs/layerg-crawler/db"
	dbCon "github.com/u2u-labs/layerg-crawler/db/sqlc"
	"go.uber.org/zap"
)

func startWorker(cmd *cobra.Command, args []string) {
	var (
		ctx    = context.Background()
		logger = &zap.Logger{}
	)
	if viper.GetString("LOG_LEVEL") == "PROD" {
		logger, _ = zap.NewProduction(zap.AddStacktrace(zap.FatalLevel))
	} else {
		logger, _ = zap.NewDevelopment(zap.AddStacktrace(zap.PanicLevel))
	}
	defer logger.Sync() // flushes buffer, if any
	sugar := logger.Sugar()

	conn, err := sql.Open(
		viper.GetString("COCKROACH_DB_DRIVER"),
		viper.GetString("COCKROACH_DB_URL"),
	)
	if err != nil {
		sugar.Errorw("Failed to connect to database", "err", err)
	}

	sqlDb := dbCon.New(conn)

	rdb, err := db.NewRedisClient(&db.RedisConfig{
		Url:      viper.GetString("REDIS_DB_URL"),
		Db:       viper.GetInt("REDIS_DB"),
		Password: viper.GetString("REDIS_DB_PASSWORD"),
	})
	if err != nil {
		sugar.Errorw("Failed to connect to redis", "err", err)
	}

	queueClient := asynq.NewClient(asynq.RedisClientOpt{Addr: viper.GetString("REDIS_DB_URL")})
	defer queueClient.Close()

	if utils.ERC20ABI, err = abi.JSON(strings.NewReader(utils.ERC20ABIStr)); err != nil {
		sugar.Errorw("Failed to parse ERC20 ABI", "err", err)
	}
	if utils.ERC721ABI, err = abi.JSON(strings.NewReader(utils.ERC721ABIStr)); err != nil {
		sugar.Errorw("Failed to parse ERC721 ABI", "err", err)
	}
	if utils.ERC1155ABI, err = abi.JSON(strings.NewReader(utils.ERC1155ABIStr)); err != nil {
		sugar.Errorw("Failed to parse ERC1155 ABI", "err", err)
	}
	if utils.EXCHANGEABI, err = abi.JSON(strings.NewReader(utils.EXCHANGEABIStr)); err != nil {
		sugar.Errorw("Failed to parse EXCHANGE ABI", "err", err)
	}

	InitBackfillProcessor(ctx, sugar, sqlDb, rdb, queueClient)
}

func InitBackfillProcessor(ctx context.Context, sugar *zap.SugaredLogger, q *dbCon.Queries, rdb *redis.Client, queueClient *asynq.Client) error {
	// Get all chains
	chains, err := q.GetAllChain(ctx)
	if err != nil {
		return err
	}
	for _, chain := range chains {
		client, err := initChainClient(&chain)
		if err != nil {
			return err
		}

		// handle queue
		srv := asynq.NewServer(
			asynq.RedisClientOpt{Addr: viper.GetString("REDIS_DB_URL")},
			asynq.Config{
				Concurrency: config.WorkerConcurrency,
			},
		)

		// mux maps a type to a handler
		mux := asynq.NewServeMux()
		taskName := BackfillCollection + ":" + strconv.Itoa(int(chain.ID))
		mux.Handle(taskName, NewBackfillProcessor(sugar, client, q, &chain, rdb))

		if err := srv.Run(mux); err != nil {
			log.Fatalf("could not run server: %v", err)
		}
	}

	return nil
}

// ----------------------------------------------
// Task
// ----------------------------------------------
const (
	BackfillCollection = "backfill_collection"
)

//----------------------------------------------
// Write a function NewXXXTask to create a task.
// A task consists of a type and a payload.
//----------------------------------------------

func NewBackfillCollectionTask(bf *dbCon.GetCrawlingBackfillCrawlerRow) (*asynq.Task, error) {

	payload, err := json.Marshal(bf)
	if err != nil {
		return nil, err
	}

	taskName := BackfillCollection + ":" + strconv.Itoa(int(bf.ChainID))

	return asynq.NewTask(taskName, payload), nil
}

func (p *BackfillProcessor) ProcessTask(ctx context.Context, t *asynq.Task) error {
	var bf dbCon.GetCrawlingBackfillCrawlerRow

	if err := json.Unmarshal(t.Payload(), &bf); err != nil {
		return fmt.Errorf("json.Unmarshal failed: %v: %w", err, asynq.SkipRetry)
	}

	blockRangeScan := int64(config.BackfillBlockRangeScan)

	toScanBlock := bf.CurrentBlock + config.BackfillBlockRangeScan
	// get the nearest upper block that multiple of blockRangeScan
	if (bf.CurrentBlock % blockRangeScan) != 0 {
		toScanBlock = ((bf.CurrentBlock / blockRangeScan) + 1) * blockRangeScan
	}
	if bf.InitialBlock.Valid && toScanBlock >= bf.InitialBlock.Int64 {
		toScanBlock = bf.InitialBlock.Int64
		bf.Status = dbCon.CrawlerStatusCRAWLED
	}

	var transferEventSig []string

	switch bf.Type {
	case dbCon.AssetTypeERC20, dbCon.AssetTypeERC721:
		transferEventSig = []string{utils.TransferEventSig}
	case dbCon.AssetTypeERC1155:
		transferEventSig = []string{utils.TransferSingleSig, utils.TransferBatchSig}
	case dbCon.AssetTypeORDER:
		transferEventSig = []string{utils.FillOrderSig, utils.CancelOrderSig}
	}

	// Initialize topics slice
	var topics [][]common.Hash

	// Populate the topics slice
	innerSlice := make([]common.Hash, len(transferEventSig))
	for i, sig := range transferEventSig {
		innerSlice[i] = common.HexToHash(sig) // Convert each signature to common.Hash
	}
	topics = append(topics, innerSlice) // Add the inner slice to topics

	logs, err := helpers.RetryWithBackoff(ctx, p.sugar, config.MaxRetriesAttempt, "fetch latest block number", func() ([]types.Log, error) {
		return p.ethClient.FilterLogs(ctx, u2u.FilterQuery{
			Topics:    topics,
			BlockHash: nil,
			FromBlock: big.NewInt(bf.CurrentBlock),
			ToBlock:   big.NewInt(toScanBlock),
			Addresses: []common.Address{common.HexToAddress(bf.CollectionAddress)},
		})
	})

	if err != nil {
		p.sugar.Warnf("Failed to get filter logs %s %s", "err", err)
	}
	if bf.CurrentBlock%1000 == 0 {
		p.sugar.Infof("Get filter logs from block %d to block %d for assetType %s, contractAddress %s", bf.CurrentBlock, toScanBlock, bf.Type, bf.CollectionAddress)
	}

	switch bf.Type {
	case dbCon.AssetTypeERC20:
		handleErc20BackFill(ctx, p.sugar, p.q, p.ethClient, p.chain, logs)
	case dbCon.AssetTypeERC721:
		scannedBlock, err := handleErc721BackFill(ctx, p.sugar, p.q, p.rdb, p.ethClient, p.chain, logs)
		if scannedBlock > 0 {
			toScanBlock = int64(scannedBlock)
		}
		if err != nil {
			toScanBlock = bf.CurrentBlock
		}
	case dbCon.AssetTypeERC1155:
		scannedBlock, _ := handleErc1155Backfill(ctx, p.sugar, p.q, p.ethClient, p.chain, logs)
		if scannedBlock > 0 {
			toScanBlock = int64(scannedBlock)
		}
	case dbCon.AssetTypeORDER:
		handleExchangeBackfill(ctx, p.sugar, p.q, p.ethClient, p.chain, logs)
	}

	bf.CurrentBlock = toScanBlock

	p.q.UpdateCrawlingBackfill(ctx, dbCon.UpdateCrawlingBackfillParams{
		ChainID:           bf.ChainID,
		CollectionAddress: bf.CollectionAddress,
		Status:            bf.Status,
		CurrentBlock:      bf.CurrentBlock,
	})

	if bf.Status == dbCon.CrawlerStatusCRAWLED {
		return nil
	}
	return nil
}

// BackfillProcessor implements asynq.Handler interface.
type BackfillProcessor struct {
	sugar     *zap.SugaredLogger
	ethClient *ethclient.Client
	q         *dbCon.Queries
	chain     *dbCon.Chain
	rdb       *redis.Client
}

func NewBackfillProcessor(sugar *zap.SugaredLogger, ethClient *ethclient.Client, q *dbCon.Queries, chain *dbCon.Chain, rdb *redis.Client) *BackfillProcessor {
	sugar.Infow("Initiated new chain backfill, start crawling", "chain", chain.Chain)
	return &BackfillProcessor{sugar, ethClient, q, chain, rdb}
}
