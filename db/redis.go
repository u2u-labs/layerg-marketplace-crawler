package db

import (
	"context"
	"encoding/json"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"

	db "github.com/u2u-labs/layerg-crawler/db/sqlc"
)

type RedisConfig struct {
	Url      string
	Db       int
	Password string
}

func NewRedisClient(cfg *RedisConfig) (*redis.Client, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     cfg.Url,
		Password: cfg.Password,
		DB:       cfg.Db,
	})
	return rdb, nil
}

func ChainCacheKey(chainId int32) string {
	return "chain:" + strconv.Itoa(int(chainId))
}

func AssetCacheKey(chainId int32) string {
	return "assets:" + strconv.Itoa(int(chainId))
}

func PendingChainKey() string {
	return "pendingChains"
}

func PendingAssetKey() string {
	return "pendingAssets"
}

func OnchainHistoryKey(txHash string) string {
	return "txHash:" + txHash
}

func GetCachedChain(ctx context.Context, rdb *redis.Client, chainId int32) (*db.Chain, error) {
	res := rdb.Get(ctx, ChainCacheKey(chainId))
	if res.Err() != nil {
		return nil, res.Err()
	}
	var chain *db.Chain
	err := json.Unmarshal([]byte(res.Val()), &chain)
	return chain, err
}

func SetChainToCache(ctx context.Context, rdb *redis.Client, chain *db.Chain) error {
	jsonChain, err := json.Marshal(chain)
	if err != nil {
		return err
	}
	return rdb.Set(ctx, ChainCacheKey(chain.ID), string(jsonChain), 0).Err()
}

func DeleteChainInCache(ctx context.Context, rdb *redis.Client, chainId int32) error {
	return rdb.Del(ctx, ChainCacheKey(chainId)).Err()
}

func GetCachedAssets(ctx context.Context, rdb *redis.Client, chainId int32) ([]db.Asset, error) {
	assetsStr, err := rdb.LRange(ctx, AssetCacheKey(chainId), 0, -1).Result()
	if err != nil {
		return nil, err
	}
	var assets []db.Asset
	for _, a := range assetsStr {
		var asset db.Asset
		err = json.Unmarshal([]byte(a), &asset)
		if err != nil {
			continue
		}
		assets = append(assets, asset)
	}
	return assets, err
}

func SetAssetsToCache(ctx context.Context, rdb *redis.Client, assets []db.Asset) error {
	for _, a := range assets {
		jsonAsset, err := json.Marshal(a)
		if err != nil {
			return err
		}
		err = rdb.LPush(ctx, AssetCacheKey(a.ChainID), string(jsonAsset), 0).Err()
		if err != nil {
			return err
		}
	}
	return nil
}

func DeleteChainAssetsInCache(ctx context.Context, rdb *redis.Client, chainId int32) error {
	return rdb.Del(ctx, AssetCacheKey(chainId)).Err()
}

func GetCachedPendingAsset(ctx context.Context, rdb *redis.Client) ([]db.Asset, error) {
	assetsStr, err := rdb.LRange(ctx, PendingAssetKey(), 0, -1).Result()
	if err != nil {
		return nil, err
	}
	var assets []db.Asset
	for _, a := range assetsStr {
		var asset db.Asset
		err = json.Unmarshal([]byte(a), &asset)
		if err != nil {
			continue
		}
		assets = append(assets, asset)
	}
	return assets, err
}

func SetPendingAssetToCache(ctx context.Context, rdb *redis.Client, asset db.Asset) error {
	jsonAsset, err := json.Marshal(asset)
	if err != nil {
		return err
	}
	err = rdb.LPush(ctx, PendingAssetKey(), string(jsonAsset), 0).Err()
	if err != nil {
		return err
	}
	return nil
}

func DeletePendingAssetsInCache(ctx context.Context, rdb *redis.Client) error {
	return rdb.Del(ctx, PendingAssetKey()).Err()
}

func GetCachedPendingChain(ctx context.Context, rdb *redis.Client) ([]db.Chain, error) {
	chainStr, err := rdb.LRange(ctx, PendingChainKey(), 0, -1).Result()
	if err != nil {
		return nil, err
	}
	var chains []db.Chain
	for _, c := range chainStr {
		var chain db.Chain
		err = json.Unmarshal([]byte(c), &chain)
		if err != nil {
			continue
		}
		chains = append(chains, chain)
	}
	return chains, err
}

func SetPendingChainToCache(ctx context.Context, rdb *redis.Client, chain db.Chain) error {
	jsonChain, err := json.Marshal(chain)
	if err != nil {
		return err
	}
	err = rdb.LPush(ctx, PendingChainKey(), string(jsonChain), 0).Err()
	if err != nil {
		return err
	}
	return nil
}

func DeletePendingChainsInCache(ctx context.Context, rdb *redis.Client) error {
	return rdb.Del(ctx, PendingChainKey()).Err()
}

func SetHistoriesCache(ctx context.Context, rdb *redis.Client, onChainHistories []db.OnchainHistory) error {
	for _, history := range onChainHistories {
		SetHistoryCache(ctx, rdb, history)
	}
	return nil
}

// SetHistoryCache stores a transaction hash in Redis with an expiration time of 15 minutes.
func SetHistoryCache(ctx context.Context, rdb *redis.Client, onChainHistory db.OnchainHistory) error {
	jsonHistory, err := json.Marshal(onChainHistory)
	if err != nil {
		return err
	}

	// Store the JSON string in Redis with an expiration of 15 minutes
	onchainHistoryKey := OnchainHistoryKey(onChainHistory.TxHash)
	err = rdb.LPush(ctx, onchainHistoryKey, string(jsonHistory)).Err()
	if err != nil {
		return err
	}

	err = rdb.Expire(ctx, onchainHistoryKey, 15*time.Minute).Err()
	if err != nil {
		return err
	}
	return nil
}

// GetHistoryCache retrieves a transaction hash from Redis.
func GetHistoryCache(ctx context.Context, rdb *redis.Client, txHash string) ([]db.OnchainHistory, error) {
	histories, err := rdb.LRange(ctx, OnchainHistoryKey(txHash), 0, -1).Result()

	if err != nil {
		return nil, err
	}

	onchainHistories := make([]db.OnchainHistory, 0, len(histories))
	for _, history := range histories {

		var onchainHistory db.OnchainHistory
		err = json.Unmarshal([]byte(history), &onchainHistory)
		if err != nil {
			return nil, err
		}
		onchainHistories = append(onchainHistories, onchainHistory)
	}
	return onchainHistories, nil
}
