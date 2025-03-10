package cmd

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"runtime/debug"
	"strings"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/google/uuid"
	"github.com/hibiken/asynq"
	"github.com/u2u-labs/layerg-crawler/cmd/helpers"
	"github.com/u2u-labs/layerg-crawler/cmd/types"
	"github.com/u2u-labs/layerg-crawler/config"

	_ "github.com/lib/pq"
	"github.com/redis/go-redis/v9"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.uber.org/zap"

	"github.com/u2u-labs/layerg-crawler/cmd/utils"
	"github.com/u2u-labs/layerg-crawler/db"
	dbCon "github.com/u2u-labs/layerg-crawler/db/sqlc"
)

var contractType = make(map[int32]map[string]dbCon.Asset)

func startCrawler(cmd *cobra.Command, args []string) {
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

	crawlerConn, err := sql.Open(
		viper.GetString("COCKROACH_DB_DRIVER"),
		viper.GetString("COCKROACH_DB_URL"),
	)
	if err != nil {
		sugar.Errorw("Could not connect to database", "err", err)
	}
	mkpConn, err := sql.Open(
		viper.GetString("COCKROACH_DB_DRIVER"), // same postgres driver
		viper.GetString("POSTGRES_DB_URL"),
	)
	if err != nil {
		sugar.Errorw("Could not connect to database", "err", err)
	}
	dbStore := dbCon.NewDBManager(crawlerConn, mkpConn)

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

	err = crawlSupportedChains(ctx, sugar, dbStore.CrawlerDB, dbStore.CrQueries, rdb)
	if err != nil {
		sugar.Errorw("Error init supported chains", "err", err)
		return
	}

	// start consumer handle transfer events
	go erc721TransferEventHandler(ctx, dbStore, sugar, rdb)

	timer := time.NewTimer(config.RetriveAddedChainsAndAssetsInterval)
	defer timer.Stop()
	for {
		select {
		case <-ctx.Done():
			sugar.Infow("Crawler stopped")
			return
		case <-timer.C:
			var wg sync.WaitGroup
			iterCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
			// redis subscribe to new assets channel to restart the crawler
			go subscribeToNewAsset(iterCtx, sugar, cancel, &wg, rdb)

			// Process new chains
			ProcessNewChains(iterCtx, sugar, dbStore.CrawlerDB, rdb, dbStore.CrQueries, &wg)
			// Process new assets
			ProcessNewChainAssets(iterCtx, sugar, rdb)
			// Process backfill collection
			ProcessCrawlingBackfillCollection(iterCtx, sugar, dbStore.CrQueries, rdb, queueClient, &wg)

			wg.Wait()

			cancel()
			timer.Reset(config.RetriveAddedChainsAndAssetsInterval)
		}
	}
}

func ProcessNewChains(ctx context.Context, sugar *zap.SugaredLogger, dbConn *sql.DB, rdb *redis.Client, q *dbCon.Queries, wg *sync.WaitGroup) {
	chains, err := db.GetCachedPendingChain(ctx, rdb)
	if err != nil {
		sugar.Errorw("ProcessNewChains failed to get cached pending chains", "err", err)
		return
	}
	if err = db.DeletePendingChainsInCache(ctx, rdb); err != nil {
		sugar.Errorw("ProcessNewChains failed to delete cached pending chains", "err", err)
		return
	}
	for _, c := range chains {
		client, err := initChainClient(&c)

		if err != nil {
			sugar.Errorw("ProcessNewChains failed to init chain client", "err", err, "chain", c)
			return
		}
		rpcClient, err := helpers.InitNewRPCClient(c.RpcUrl)
		if err != nil {
			sugar.Errorw("ProcessNewChains failed to init rpc client", "err", err, "chain", c)
			return
		}

		wg.Add(1)
		go func(chain dbCon.Chain) {
			defer wg.Done()
			StartChainCrawler(ctx, sugar, &EvmRPCClient{client, rpcClient}, dbConn, q, &chain, rdb)
		}(c)
		sugar.Infow("Initiated new chain, start crawling", "chain", c)
	}
}

func ProcessNewChainAssets(ctx context.Context, sugar *zap.SugaredLogger, rdb *redis.Client) {
	assets, err := db.GetCachedPendingAsset(ctx, rdb)
	if err != nil {
		sugar.Errorw("ProcessNewChainAssets failed to get cached pending assets", "err", err)
		return
	}
	if err = db.DeletePendingAssetsInCache(ctx, rdb); err != nil {
		sugar.Errorw("ProcessNewChainAssets failed to delete cached pending assets", "err", err)
		return
	}
	for _, a := range assets {
		if contractType[a.ChainID] == nil {
			contractType[a.ChainID] = make(map[string]dbCon.Asset)
		}
		contractType[a.ChainID][a.CollectionAddress] = a
		sugar.Infow("Initiated new assets, start crawling",
			"chain", a.ChainID,
			"address", a.CollectionAddress,
			"type", a.Type,
		)
	}
}

func crawlSupportedChains(ctx context.Context, sugar *zap.SugaredLogger, dbConn *sql.DB, q *dbCon.Queries, rdb *redis.Client) error {
	// Query, flush cache and connect all supported chains
	chains, err := q.GetAllChain(ctx)
	if err != nil {
		return err
	}
	for _, c := range chains {
		contractType[c.ID] = make(map[string]dbCon.Asset)
		if err = db.DeleteChainInCache(ctx, rdb, c.ID); err != nil {
			return err
		}
		if err = db.DeleteChainAssetsInCache(ctx, rdb, c.ID); err != nil {
			return err
		}

		// Query all assets of one chain
		var (
			assets []dbCon.Asset
			limit  int32 = 10
			offset int32 = 0
		)
		for {
			a, err := q.GetPaginatedAssetsByChainId(ctx, dbCon.GetPaginatedAssetsByChainIdParams{
				ChainID: c.ID,
				Limit:   limit,
				Offset:  offset,
			})
			if err != nil {
				return err
			}
			assets = append(assets, a...)
			offset = offset + limit
			if len(a) < int(limit) {
				break
			}
		}

		if err = db.SetChainToCache(ctx, rdb, &c); err != nil {
			return err
		}
		if err = db.SetAssetsToCache(ctx, rdb, assets); err != nil {
			return err
		}
		for _, a := range assets {
			contractType[a.ChainID][a.CollectionAddress] = a
		}
		client, err := initChainClient(&c)
		if err != nil {
			return err
		}
		rpcClient, err := helpers.InitNewRPCClient(c.RpcUrl)
		if err != nil {
			return err
		}
		go StartChainCrawler(ctx, sugar, &EvmRPCClient{client, rpcClient}, dbConn, q, &c, rdb)
	}
	return nil
}

func ProcessCrawlingBackfillCollection(ctx context.Context, sugar *zap.SugaredLogger, q *dbCon.Queries, rdb *redis.Client, queueClient *asynq.Client, wg *sync.WaitGroup) error {
	// Get all Backfill Collection with status CRAWLING
	crawlingBackfill, err := q.GetCrawlingBackfillCrawler(ctx)

	if err != nil {
		return err
	}

	for _, c := range crawlingBackfill {
		chain, err := q.GetChainById(ctx, c.ChainID)
		if err != nil {
			return err
		}

		client, err := initChainClient(&chain)
		if err != nil {
			return err
		}
		wg.Add(1)
		go func(bf dbCon.GetCrawlingBackfillCrawlerRow) {
			defer wg.Done()
			AddBackfillCrawlerTask(ctx, sugar, client, q, &chain, &bf, queueClient)
		}(c)
	}
	return nil
}

const erc721TransferEvent = "erc721_transfer"

func erc721TransferEventHandler(ctx context.Context, dbStore *dbCon.DBManager, logger *zap.SugaredLogger, rdb *redis.Client) {
	logger.Info("Starting erc721 transfer event handler")

	defer func() {
		if r := recover(); r != nil {
			logger.Errorw("Recovered from panic in erc721TransferEventHandler", "error", r, "stacktrace", string(debug.Stack()))
		}
	}()

	ps := rdb.Subscribe(ctx, erc721TransferEvent)
	defer ps.Close()

	ch := ps.Channel()

	for {
		select {
		case <-ctx.Done():
			// Context canceled, time to shut down
			logger.Info("Shutting down erc721 transfer event handler due to context cancellation")
			return

		case msg, ok := <-ch:
			// Channel closed or error occurred
			if !ok {
				logger.Warn("Redis subscription channel closed, exiting handler")
				return
			}

			// Process the message
			if err := processErc721Transfer(ctx, dbStore, logger, msg); err != nil {
				logger.Errorw("Error processing ERC721 transfer", "error", err)
			}
		}
	}
}

func processErc721Transfer(ctx context.Context, dbStore *dbCon.DBManager, logger *zap.SugaredLogger, msg *redis.Message) error {
	var payload types.Erc721CollectionAssetExtended
	if err := json.Unmarshal([]byte(msg.Payload), &payload); err != nil {
		return fmt.Errorf("failed to unmarshal: %w", err)
	}

	var (
		name         = payload.TokenID
		ipfsUrl      = ""
		description  = ""
		image        = ""
		animationUrl = ""
	)

	if payload.Attributes.Valid && len(payload.Attributes.String) > 0 {
		var attrs map[string]string
		// try to unmarshal attributes to a map
		if err := json.Unmarshal([]byte(payload.Attributes.String), &attrs); err == nil {
			name = attrs["name"]
			ipfsUrl = attrs["image"]
			description = attrs["description"]
			image = attrs["image"]
			animationUrl = attrs["animationUrl"]
		}
	}

	// get collection
	col, err := dbStore.MpQueries.GetCollectionByAddressAndChainId(ctx, dbCon.GetCollectionByAddressAndChainIdParams{
		Address: sql.NullString{String: strings.ToLower(payload.CollectionAddress), Valid: true},
		ChainId: int64(payload.ChainID),
	})
	if errors.Is(err, sql.ErrNoRows) {
		return nil
	}
	if err != nil {
		return err
	}

	// upsert nft to layerg marketplace db
	upsertedNft, err := dbStore.MpQueries.UpsertNFT(ctx, dbCon.UpsertNFTParams{
		ID:             payload.TokenID,
		Name:           name,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
		Status:         "SUCCESS",
		TokenUri:       ipfsUrl,
		TxCreationHash: payload.TxHash,
		CreatorId:      uuid.NullUUID{},
		CollectionId:   col.ID,
		Image:          sql.NullString{Valid: true, String: image},
		Description:    sql.NullString{Valid: true, String: description},
		AnimationUrl:   sql.NullString{Valid: true, String: animationUrl},
		NameSlug:       sql.NullString{Valid: true, String: name},
		MetricPoint:    0,
		MetricDetail:   []byte("{}"),
		Source:         sql.NullString{Valid: true, String: "crawler"},
		OwnerId:        payload.Owner,
	})

	if err != nil {
		return fmt.Errorf("failed to upsert NFT: %w", err)
	}

	logger.Infow("Upserted NFT successfully", "tokenID", payload.TokenID, "collection", upsertedNft.CollectionId,
		"chainId", payload.ChainID)
	return nil
}

func subscribeToNewAsset(ctx context.Context, sugar *zap.SugaredLogger, cancel context.CancelFunc, wg *sync.WaitGroup, rdb *redis.Client) {
	ps := rdb.Subscribe(ctx, config.NewAssetChannel)
	defer ps.Close()
	ch := ps.Channel()

	for {
		select {
		case <-ctx.Done():
			return
		case msg, ok := <-ch:
			if !ok {
				sugar.Error("Redis subscription channel closed, exiting handler")
				return
			}
			sugar.Infow("Received new asset", "message", msg.Payload)
			cancel()
			return
		}
	}
}
