package cmd

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"math/big"
	"net/http"
	"strings"
	"time"

	u2u "github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	utypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/hibiken/asynq"
	"github.com/redis/go-redis/v9"
	"github.com/u2u-labs/layerg-crawler/cmd/helpers"
	"github.com/u2u-labs/layerg-crawler/cmd/types"
	"github.com/u2u-labs/layerg-crawler/cmd/utils"
	"github.com/u2u-labs/layerg-crawler/config"
	rdb "github.com/u2u-labs/layerg-crawler/db"
	db "github.com/u2u-labs/layerg-crawler/db/sqlc"
	"go.uber.org/zap"
)

type EvmRPCClient struct {
	EthClient *ethclient.Client
	RpcClient *rpc.Client
}

func StartChainCrawler(ctx context.Context, sugar *zap.SugaredLogger, client *EvmRPCClient, dbConn *db.DBManager, chain *db.Chain, rdb *redis.Client) {
	sugar.Infow("Start chain crawler", "chain", chain)
	timer := time.NewTimer(time.Duration(chain.BlockTime) * time.Millisecond)
	defer timer.Stop()
	for {
		select {
		case <-ctx.Done():
			sugar.Infow("Stopping chain crawler", "chain", chain)
			return
		case <-timer.C:
			// Process new blocks
			err := ProcessLatestBlocks(ctx, sugar, client, dbConn, chain, rdb)
			if err != nil {
				sugar.Errorw("Error processing latest blocks", "err", err)
				return
			}
			timer.Reset(time.Duration(chain.BlockTime) * time.Millisecond)
		}
	}
}

func AddBackfillCrawlerTask(ctx context.Context, sugar *zap.SugaredLogger, client *ethclient.Client, q *db.Queries, chain *db.Chain, bf *db.GetCrawlingBackfillCrawlerRow, queueClient *asynq.Client) {
	blockRangeScan := int64(config.BackfillBlockRangeScan) * 100
	if bf.CurrentBlock%blockRangeScan == 0 {
		sugar.Infow("Backfill crawler", "chain", chain, "block", bf.CurrentBlock)
	}

	timer := time.NewTimer(time.Duration(chain.BlockTime) * time.Millisecond)
	defer timer.Stop()

	for {
		select {
		case <-ctx.Done():
			sugar.Infow("Stopping backfill crawler", "chain", chain, "block", bf.CurrentBlock)
			return
		case <-timer.C:
			newBf, err := q.GetCrawlingBackfillCrawlerById(ctx, db.GetCrawlingBackfillCrawlerByIdParams{
				ChainID:           bf.ChainID,
				CollectionAddress: bf.CollectionAddress,
			})
			if err != nil {
				log.Fatalf("Error getting backfill crawler: %v", err)
			}
			if bf.CurrentBlock < newBf.CurrentBlock {
				bf.CurrentBlock = newBf.CurrentBlock
			}

			task, err := NewBackfillCollectionTask(bf)
			if err != nil {
				log.Fatalf("could not create task: %v", err)
			}
			_, err = queueClient.Enqueue(task)
			if err != nil {
				log.Fatalf("could not enqueue task: %v", err)
			}
			timer.Reset(time.Duration(chain.BlockTime) * time.Millisecond)
		}
	}
}

func ProcessLatestBlocks(ctx context.Context, sugar *zap.SugaredLogger, client *EvmRPCClient, dbConn *db.DBManager, chain *db.Chain, rdb *redis.Client) error {
	// Get latest block number with retries
	latest, err := helpers.RetryWithBackoff(ctx, sugar, config.MaxRetriesAttempt, "fetch latest block number", func() (uint64, error) {
		return client.EthClient.BlockNumber(ctx)
	})
	if err != nil {
		sugar.Errorw("Failed to fetch latest blocks", "err", err, "chain", chain)
		return err
	}

	c, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	tx, err := dbConn.CrawlerDB.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
	if err != nil {
		sugar.Errorw("Failed to start transaction", "err", err)
		return err
	}
	defer tx.Rollback()

	qtx := dbConn.CrQueries.WithTx(tx)
	latestUpdatedBlock := chain.LatestBlock
	// Process each block between
	for i := chain.LatestBlock + 1; i <= int64(latest) && i <= chain.LatestBlock+config.ProcessBatchBlocksLimit; i++ {
		if i%50 == 0 {
			sugar.Infow("Importing block receipts", "chain", chain.Chain+" "+chain.Name, "block", i, "latest", latest)
		}

		//block, err := client.EthClient.BlockByNumber(c, big.NewInt(int64(i)))
		//if err != nil {
		//	sugar.Errorw("Failed to fetch latest block", "err", err, "height", i, "chain", chain)
		//	break
		//}
		//receipts, err := client.EthClient.BlockReceipts(ctx, rpc.BlockNumberOrHashWithNumber(rpc.BlockNumber(i)))
		//if err != nil {
		//	sugar.Errorw("Failed to fetch latest block receipts", "err", err, "height", i, "chain", chain)
		//	break
		//}

		block, receipts, err := helpers.FetchBlockAndReceipts(client.RpcClient, sugar, c, big.NewInt(i))
		if err != nil {
			if !strings.Contains(err.Error(), "context deadline exceeded") {
				sugar.Errorw("Failed to fetch latest block receipts", "err", err, "height", i, "chain", chain)
			}
			break
		}

		if err = FilterEvents(ctx, sugar, qtx, client.EthClient, chain, rdb, receipts, block.Time()); err != nil {
			sugar.Errorw("Failed to filter events", "err", err, "height", i, "chain", chain)
			break
		}
		latestUpdatedBlock = i
	}
	// Update latest block processed
	if latestUpdatedBlock > chain.LatestBlock {
		if err = qtx.UpdateChainLatestBlock(ctx, db.UpdateChainLatestBlockParams{
			ID:          chain.ID,
			LatestBlock: latestUpdatedBlock,
		}); err != nil {
			sugar.Errorw("Failed to update chain latest blocks in DB", "err", err, "chain", chain)
			return err
		}
		chain.LatestBlock = latestUpdatedBlock
	}

	if err = tx.Commit(); err != nil {
		sugar.Errorw("Failed to commit transaction", "err", err)
		return err
	}

	return nil
}

func FilterEvents(ctx context.Context, sugar *zap.SugaredLogger, q *db.Queries, client *ethclient.Client,
	chain *db.Chain, rdb *redis.Client, receipts utypes.Receipts, ts uint64) error {

	for _, r := range receipts {
		for _, l := range r.Logs { //sugar.Debugw("FilterEvents", "txHash", r.TxHash.Hex(), "l.Address.Hex", l.Address.Hex(), "info", contractType[chain.ID][l.Address.Hex()])

			switch contractType[chain.ID][l.Address.Hex()].Type {
			case db.AssetTypeERC20:
				if err := handleErc20Transfer(ctx, sugar, q, client, chain, rdb, l); err != nil {
					sugar.Errorw("handleErc20Transfer", "err", err)
					return err
				}
			case db.AssetTypeERC721:
				if err := handleErc721Transfer(ctx, sugar, q, client, chain, rdb, l); err != nil {
					sugar.Errorw("handleErc721Transfer", "err", err)
					return err
				}
			case db.AssetTypeERC1155:
				if l.Topics[0].Hex() == utils.TransferSingleSig {
					if err := handleErc1155TransferSingle(ctx, sugar, q, client, chain, rdb, l); err != nil {
						sugar.Errorw("handleErc1155TransferSingle", "err", err)
						return err
					}
				}
				if l.Topics[0].Hex() == utils.TransferBatchSig {
					if err := handleErc1155TransferBatch(ctx, sugar, q, client, chain, rdb, l); err != nil {
						sugar.Errorw("handleErc1155TransferBatch", "err", err)
						return err
					}
				}
			case db.AssetTypeORDER:
				if l.Topics[0].Hex() == utils.FillOrderSig {
					if err := handleFillOrder(ctx, sugar, q, client, chain, rdb, l, ts); err != nil {
						sugar.Errorw("handleFillOrder", "err", err)
						return err
					}
				} else if l.Topics[0].Hex() == utils.CancelOrderSig {
					if err := handleCancelOrder(ctx, sugar, q, client, chain, rdb, l, ts); err != nil {
						sugar.Errorw("handleCancelOrder", "err", err)
						return err
					}
				}
			default:
				continue
			}
		}
	}
	return nil
}

func handleErc20Transfer(ctx context.Context, sugar *zap.SugaredLogger, q *db.Queries, client *ethclient.Client,
	chain *db.Chain, rc *redis.Client, l *utypes.Log) error {

	if l.Topics[0].Hex() != utils.TransferEventSig {
		return nil
	}
	// Unpack the log data
	var event utils.Erc20TransferEvent

	err := utils.ERC20ABI.UnpackIntoInterface(&event, "Transfer", l.Data)
	if err != nil {
		sugar.Fatalf("Failed to unpack log: %v", err)
		return err
	}

	// Decode the indexed fields manually
	event.From = common.BytesToAddress(l.Topics[1].Bytes())
	event.To = common.BytesToAddress(l.Topics[2].Bytes())
	amount := event.Value.String()

	_, err = q.AddOnchainTransaction(ctx, db.AddOnchainTransactionParams{
		From:      event.From.Hex(),
		To:        event.To.Hex(),
		AssetID:   contractType[chain.ID][l.Address.Hex()].ID,
		TokenID:   "0",
		Amount:    amount,
		TxHash:    l.TxHash.Hex(),
		Timestamp: time.Now(),
	})

	if err != nil {
		return err
	}

	// Update sender's balances
	balance, err := getErc20BalanceOf(ctx, sugar, client, &l.Address, &event.From)

	if err != nil {
		sugar.Errorw("Failed to get ERC20 balance", "err", err)
	}
	if err = q.Add20Asset(ctx, db.Add20AssetParams{
		AssetID: contractType[chain.ID][l.Address.Hex()].ID,
		ChainID: chain.ID,
		Owner:   event.From.Hex(),
		Balance: balance.String(),
	}); err != nil {
		return err
	}

	// Update receiver's balances
	balance, err = getErc20BalanceOf(ctx, sugar, client, &l.Address, &event.To)

	if err != nil {
		sugar.Errorw("Failed to get ERC20 balance", "err", err)
	}
	if err = q.Add20Asset(ctx, db.Add20AssetParams{
		AssetID: contractType[chain.ID][l.Address.Hex()].ID,
		ChainID: chain.ID,
		Owner:   event.To.Hex(),
		Balance: balance.String(),
	}); err != nil {
		return err
	}

	return nil
}

func handleErc20BackFill(ctx context.Context, sugar *zap.SugaredLogger, q *db.Queries, client *ethclient.Client,
	chain *db.Chain, logs []utypes.Log) error {

	// Initialize the AddressSet
	addressSet := helpers.NewAddressSet()

	if len(logs) == 0 {
		return nil
	}

	var contractAddress *common.Address
	for _, l := range logs {
		contractAddress = &l.Address
		var event utils.Erc20TransferEvent

		err := utils.ERC20ABI.UnpackIntoInterface(&event, "Transfer", l.Data)
		if err != nil {
			sugar.Fatalf("Failed to unpack log: %v", err)
			return err
		}

		if l.Topics[0].Hex() != utils.TransferEventSig {
			return nil
		}

		event.From = common.BytesToAddress(l.Topics[1].Bytes())
		event.To = common.BytesToAddress(l.Topics[2].Bytes())
		amount := event.Value.String()

		_, err = q.AddOnchainTransaction(ctx, db.AddOnchainTransactionParams{
			From:      event.From.Hex(),
			To:        event.To.Hex(),
			AssetID:   contractType[chain.ID][l.Address.Hex()].ID,
			TokenID:   "0",
			Amount:    amount,
			TxHash:    l.TxHash.Hex(),
			Timestamp: time.Now(),
		})

		// adding sender and receiver to the address set
		addressSet.AddAddress(event.From)
		addressSet.AddAddress(event.To)
	}

	rpcClient, _ := helpers.InitNewRPCClient(chain.RpcUrl)

	addressList := addressSet.GetAddresses()

	results := make([]string, len(addressList))
	calls := make([]rpc.BatchElem, len(addressList))

	for i, addr := range addressList {
		// Pack the data for the balanceOf function
		data, err := utils.ERC20ABI.Pack("balanceOf", addr)
		if err != nil {
			sugar.Errorf("Failed to pack data for balanceOf: %v", err)
			return err
		}

		encodedData := "0x" + common.Bytes2Hex(data)

		// Append the BatchElem for the eth_call
		calls[i] = rpc.BatchElem{
			Method: "eth_call",
			Args: []interface{}{
				map[string]interface{}{
					"to":   contractAddress,
					"data": encodedData,
				},
				"latest",
			},
			Result: &results[i],
		}
	}

	// Execute batch call
	if err := rpcClient.BatchCallContext(ctx, calls); err != nil {
		log.Fatalf("Failed to execute batch call: %v", err)
	}

	// Iterate over the results and update the balances
	for i, result := range results {
		var balance *big.Int

		utils.ERC20ABI.UnpackIntoInterface(&balance, "balanceOf", common.FromHex(result))

		if err := q.Add20Asset(ctx, db.Add20AssetParams{
			AssetID: contractType[chain.ID][contractAddress.Hex()].ID,
			ChainID: chain.ID,
			Owner:   addressList[i].Hex(),
			Balance: balance.String(),
		}); err != nil {
			return err
		}
	}

	addressSet.Reset()
	return nil
}

func handleErc721BackFill(ctx context.Context, sugar *zap.SugaredLogger, q *db.Queries, rdb *redis.Client, client *ethclient.Client,
	chain *db.Chain, logs []utypes.Log) (uint64, error) {
	scannedBlock := uint64(0)
	// Initialize the NewTokenIdSet
	tokenIdSet := helpers.NewTokenIdSet()

	if len(logs) == 0 {
		return 0, nil
	}

	var contractAddress *common.Address
	for i, l := range logs {
		contractAddress = &l.Address

		// Decode the indexed fields manually
		event := utils.Erc721TransferEvent{
			From:    common.BytesToAddress(l.Topics[1].Bytes()),
			To:      common.BytesToAddress(l.Topics[2].Bytes()),
			TokenID: l.Topics[3].Big(),
		}
		_, err := q.AddOnchainTransaction(ctx, db.AddOnchainTransactionParams{
			From:      event.From.Hex(),
			To:        event.To.Hex(),
			AssetID:   contractType[chain.ID][l.Address.Hex()].ID,
			TokenID:   event.TokenID.String(),
			Amount:    "0",
			TxHash:    l.TxHash.Hex(),
			Timestamp: time.Now(),
		})
		if err != nil {
			return 0, err
		}

		// adding token Id
		tokenIdSet.AddTokenId(event.TokenID)
		if i > config.MaxScanLogsLimit {
			sugar.Infow("Batch size is too big, split the batch", "total", len(logs), "batch", i)
			scannedBlock = l.BlockNumber
			break
		}
	}

	rpcClient, _ := helpers.InitNewRPCClient(chain.RpcUrl)

	tokenIdList := tokenIdSet.GetTokenIds()

	_, err := fetchTokenUriAndOwner(ctx, sugar, q, rdb, tokenIdList, contractAddress, rpcClient, chain)
	if err != nil {
		return 0, err
	}

	tokenIdSet.Reset()
	return scannedBlock, nil
}

func fetchTokenUriAndOwner(ctx context.Context, sugar *zap.SugaredLogger, q *db.Queries, rdb *redis.Client, tokenIdList []*big.Int, contractAddress *common.Address, rpcClient *rpc.Client, chain *db.Chain) ([]string, error) {
	results := make([]string, len(tokenIdList)*2)
	calls := make([]rpc.BatchElem, len(tokenIdList)*2)

	for i, tokenId := range tokenIdList {
		// Pack the data for the tokenURI function
		data, err := utils.ERC721ABI.Pack("tokenURI", tokenId)
		if err != nil {
			sugar.Errorf("Failed to pack data for tokenURI: %v", err)
			return nil, err
		}

		encodedUriData := "0x" + common.Bytes2Hex(data)

		// Append the BatchElem for the eth_call
		calls[2*i] = rpc.BatchElem{
			Method: "eth_call",
			Args: []interface{}{
				map[string]interface{}{
					"to":   contractAddress,
					"data": encodedUriData,
				},
				"latest",
			},
			Result: &results[2*i],
		}

		// Pack the data for the ownerOf function
		ownerData, err := utils.ERC721ABI.Pack("ownerOf", tokenId)
		if err != nil {
			sugar.Errorf("Failed to pack data for ownerOf: %v", err)
			return nil, err
		}

		encodedOwnerData := "0x" + common.Bytes2Hex(ownerData)

		// Append the BatchElem for the eth_call
		calls[2*i+1] = rpc.BatchElem{
			Method: "eth_call",
			Args: []interface{}{
				map[string]interface{}{
					"to":   contractAddress,
					"data": encodedOwnerData,
				},
				"latest",
			},
			Result: &results[2*i+1],
		}

	}

	// Execute batch call
	err := helpers.BatchCallRpcClient(ctx, sugar, calls, rpcClient)
	if err != nil {
		return nil, err
	}

	// Iterate over the results and update the balances
	for i := 0; i < len(results); i += 2 {
		var uri string
		var owner common.Address
		utils.ERC721ABI.UnpackIntoInterface(&uri, "tokenURI", common.FromHex(results[i]))
		utils.ERC721ABI.UnpackIntoInterface(&owner, "ownerOf", common.FromHex(results[i+1]))
		attrs := make(map[string]string)
		attrs["image"] = uri
		rawAttrs, err := json.Marshal(attrs)
		if err != nil {
			sugar.Errorw("Failed to marshal attributes", "err", err)
			return nil, err
		}

		asset, err := q.Add721Asset(ctx, db.Add721AssetParams{
			AssetID: contractType[chain.ID][contractAddress.Hex()].ID,
			ChainID: chain.ID,
			TokenID: tokenIdList[i/2].String(),
			Owner:   owner.Hex(),
			Attributes: sql.NullString{
				String: string(rawAttrs),
				Valid:  true,
			},
		})
		if err != nil {
			return nil, err
		}

		// publish to redis
		go func() {
			data := &types.Erc721CollectionAssetExtended{
				Erc721CollectionAsset: &asset,
				TxHash:                "",
				CollectionAddress:     contractAddress.Hex(),
			}
			dataBytes, err := json.Marshal(data)
			if err != nil {
				sugar.Errorw("Failed to marshal data", "err", err)
				return
			}
			err = rdb.Publish(ctx, erc721TransferEvent, dataBytes).Err()
			if err != nil {
				sugar.Errorw("Failed to publish to redis", "err", err)
				return
			}
		}()
	}
	return results, nil
}

func handleErc1155Backfill(ctx context.Context, sugar *zap.SugaredLogger, q *db.Queries, client *ethclient.Client,
	chain *db.Chain, logs []utypes.Log) (uint64, error) {
	scannedBlock := uint64(0)
	// Initialize the NewTokenIdSet
	tokenIdContractAddressSet := helpers.NewTokenIdContractAddressSet()

	if len(logs) == 0 {
		return 0, nil
	}

	var contractAddress *common.Address
	for i, l := range logs {
		contractAddress = &l.Address
		if l.Topics[0].Hex() == utils.TransferSingleSig {
			// handleTransferSingle

			// Decode TransferSingle log
			var event utils.Erc1155TransferSingleEvent
			err := utils.ERC1155ABI.UnpackIntoInterface(&event, "TransferSingle", l.Data)
			if err != nil {
				sugar.Errorw("Failed to unpack TransferSingle log:", "err", err)
			}

			// Decode the indexed fields for TransferSingle
			event.Operator = common.BytesToAddress(l.Topics[1].Bytes())
			event.From = common.BytesToAddress(l.Topics[2].Bytes())
			event.To = common.BytesToAddress(l.Topics[3].Bytes())

			amount := event.Value.String()
			_, err = q.AddOnchainTransaction(ctx, db.AddOnchainTransactionParams{
				From:      event.From.Hex(),
				To:        event.To.Hex(),
				AssetID:   contractType[chain.ID][l.Address.Hex()].ID,
				TokenID:   event.Id.String(),
				Amount:    amount,
				TxHash:    l.TxHash.Hex(),
				Timestamp: time.Now(),
			})
			if err != nil {
				return 0, err
			}

			// adding data to set
			tokenIdContractAddressSet.AddTokenIdContractAddress(event.Id, event.From.Hex())
			tokenIdContractAddressSet.AddTokenIdContractAddress(event.Id, event.To.Hex())
		}

		if l.Topics[0].Hex() == utils.TransferBatchSig {
			var event utils.Erc1155TransferBatchEvent
			err := utils.ERC1155ABI.UnpackIntoInterface(&event, "TransferBatch", l.Data)
			if err != nil {
				sugar.Errorw("Failed to unpack TransferBatch log:", "err", err)
			}

			// Decode the indexed fields for TransferBatch
			event.Operator = common.BytesToAddress(l.Topics[1].Bytes())
			event.From = common.BytesToAddress(l.Topics[2].Bytes())
			event.To = common.BytesToAddress(l.Topics[3].Bytes())

			for i := range event.Ids {
				amount := event.Values[i].String()
				_, err := q.AddOnchainTransaction(ctx, db.AddOnchainTransactionParams{
					From:      event.From.Hex(),
					To:        event.To.Hex(),
					AssetID:   contractType[chain.ID][l.Address.Hex()].ID,
					TokenID:   event.Ids[i].String(),
					Amount:    amount,
					TxHash:    l.TxHash.Hex(),
					Timestamp: time.Now(),
				})
				if err != nil {
					return 0, err
				}

				// adding data to set
				tokenIdContractAddressSet.AddTokenIdContractAddress(event.Ids[i], event.From.Hex())
				tokenIdContractAddressSet.AddTokenIdContractAddress(event.Ids[i], event.To.Hex())
			}
		}

		if i > config.MaxScanLogsLimit {
			sugar.Errorw("Reached max scan logs limit", "total_logs", len(logs))
			scannedBlock = l.BlockNumber
			break
		}
	}

	rpcClient, _ := helpers.InitNewRPCClient(chain.RpcUrl)

	tokenIdList := tokenIdContractAddressSet.GetTokenIdContractAddressses()

	results := make([]string, len(tokenIdList)*2)
	calls := make([]rpc.BatchElem, len(tokenIdList)*2)

	for i, pairData := range tokenIdList {
		tokenId := pairData.TokenId
		ownerAddress := common.HexToAddress(pairData.ContractAddress)

		// Pack the data for the tokenURI function
		data, err := utils.ERC1155ABI.Pack("uri", tokenId)
		if err != nil {
			sugar.Errorf("Failed to pack data for tokenURI: %v", err)
			return 0, err
		}

		encodedUriData := "0x" + common.Bytes2Hex(data)

		// Append the BatchElem for the eth_call
		calls[2*i] = rpc.BatchElem{
			Method: "eth_call",
			Args: []interface{}{
				map[string]interface{}{
					"to":   contractAddress,
					"data": encodedUriData,
				},
				"latest",
			},
			Result: &results[2*i],
		}

		// 	// Pack the data for the ownerOf function
		ownerData, err := utils.ERC1155ABI.Pack("balanceOf", ownerAddress, tokenId)
		if err != nil {
			sugar.Errorf("Failed to pack data for balanceOf: %v", err)
			return 0, err
		}

		encodedBalanceData := "0x" + common.Bytes2Hex(ownerData)

		// 	// Append the BatchElem for the eth_call
		calls[2*i+1] = rpc.BatchElem{
			Method: "eth_call",
			Args: []interface{}{
				map[string]interface{}{
					"to":   contractAddress,
					"data": encodedBalanceData,
				},
				"latest",
			},
			Result: &results[2*i+1],
		}
	}

	// // Execute batch call
	if err := rpcClient.BatchCallContext(ctx, calls); err != nil {
		log.Fatalf("Failed to execute batch call: %v", err)
	}

	// // Iterate over the results and update the balances
	for i := 0; i < len(results); i += 2 {
		var uri string
		var balance *big.Int
		utils.ERC1155ABI.UnpackIntoInterface(&uri, "uri", common.FromHex(results[i]))
		utils.ERC1155ABI.UnpackIntoInterface(&balance, "balanceOf", common.FromHex(results[i+1]))

		if err := q.Add1155Asset(ctx, db.Add1155AssetParams{
			AssetID: contractType[chain.ID][contractAddress.Hex()].ID,
			ChainID: chain.ID,
			TokenID: tokenIdList[i/2].TokenId.String(),
			Owner:   tokenIdList[i/2].ContractAddress,
			Attributes: sql.NullString{
				String: uri,
				Valid:  true,
			},
		}); err != nil {
			return 0, err
		}
	}

	tokenIdContractAddressSet.Reset()
	return scannedBlock, nil
}

func handleErc721Transfer(ctx context.Context, sugar *zap.SugaredLogger, q *db.Queries, client *ethclient.Client,
	chain *db.Chain, rc *redis.Client, l *utypes.Log) error {

	if l.Topics[0].Hex() != utils.TransferEventSig {
		return nil
	}
	// Decode the indexed fields manually
	event := utils.Erc721TransferEvent{
		From:    common.BytesToAddress(l.Topics[1].Bytes()),
		To:      common.BytesToAddress(l.Topics[2].Bytes()),
		TokenID: l.Topics[3].Big(),
	}
	_, err := q.AddOnchainTransaction(ctx, db.AddOnchainTransactionParams{
		From:      event.From.Hex(),
		To:        event.To.Hex(),
		AssetID:   contractType[chain.ID][l.Address.Hex()].ID,
		TokenID:   event.TokenID.String(),
		Amount:    "0",
		TxHash:    l.TxHash.Hex(),
		Timestamp: time.Now(),
	})
	if err != nil {
		return err
	}

	tokenIdSet := helpers.NewTokenIdSet()
	tokenIdSet.AddTokenId(event.TokenID)
	rpcClient, err := helpers.InitNewRPCClient(chain.RpcUrl)
	if err != nil {
		return err
	}
	results, err := fetchTokenUriAndOwner(ctx, sugar, q, rc, tokenIdSet.GetTokenIds(), &l.Address, rpcClient, chain)
	if err != nil {
		return err
	}
	var uri string
	var owner common.Address
	sugar.Infow("results", "results", results, "tokenIdSet", tokenIdSet.GetTokenIds(), "event", event)
	utils.ERC721ABI.UnpackIntoInterface(&uri, "tokenURI", common.FromHex(results[0]))
	utils.ERC721ABI.UnpackIntoInterface(&owner, "ownerOf", common.FromHex(results[1]))
	attrs := make(map[string]string)
	attrs["image"] = uri
	rawAttrs, err := json.Marshal(attrs)
	if err != nil {
		sugar.Errorw("Failed to marshal attributes", "err", err)
		return err
	}

	asset, err := q.Add721Asset(ctx, db.Add721AssetParams{
		AssetID: contractType[chain.ID][l.Address.Hex()].ID,
		ChainID: chain.ID,
		TokenID: event.TokenID.String(),
		Owner:   owner.Hex(),
		Attributes: sql.NullString{
			String: string(rawAttrs),
			Valid:  true,
		},
	})
	if err != nil {
		return err
	}

	// publish to redis
	go func() {
		data := &types.Erc721CollectionAssetExtended{
			Erc721CollectionAsset: &asset,
			TxHash:                l.TxHash.Hex(),
			CollectionAddress:     l.Address.Hex(),
		}
		dataBytes, err := json.Marshal(data)
		if err != nil {
			sugar.Errorw("Failed to marshal data", "err", err)
			return
		}
		err = rc.Publish(ctx, erc721TransferEvent, dataBytes).Err()
		if err != nil {
			sugar.Errorw("Failed to publish to redis", "err", err)
			return
		}
	}()

	return nil
}

func addingErc721Elem(ctx context.Context, sugar *zap.SugaredLogger, q *db.Queries, client *ethclient.Client,
	chain *db.Chain, rc *redis.Client, l *utypes.Log, backfill bool) error {

	if l.Topics[0].Hex() != utils.TransferEventSig {
		return nil
	}
	// Decode the indexed fields manually
	event := utils.Erc721TransferEvent{
		From:    common.BytesToAddress(l.Topics[1].Bytes()),
		To:      common.BytesToAddress(l.Topics[2].Bytes()),
		TokenID: l.Topics[3].Big(),
	}
	history, err := q.AddOnchainTransaction(ctx, db.AddOnchainTransactionParams{
		From:      event.From.Hex(),
		To:        event.To.Hex(),
		AssetID:   contractType[chain.ID][l.Address.Hex()].ID,
		TokenID:   event.TokenID.String(),
		Amount:    "0",
		TxHash:    l.TxHash.Hex(),
		Timestamp: time.Now(),
	})
	if err != nil {
		return err
	}

	// Cache the new onchain transaction
	if !backfill {
		if err = rdb.SetHistoryCache(ctx, rc, history); err != nil {
			return err
		}
	}

	// Update NFT holder
	uri, err := getErc721TokenURI(ctx, sugar, client, rc, chain.Chain, &l.Address, event.TokenID)
	if err != nil {
		sugar.Errorw("Failed to get ERC721 uri", "err", err, "tokenID", event.TokenID, "contract", l.Address.Hex())
	}

	// Get owner of the token
	owner, err := getErc721OwnerOf(ctx, sugar, client, &l.Address, event.TokenID)
	if err != nil {
		sugar.Errorw("Failed to get ERC721 owner", "err", err, "tokenID", event.TokenID, "contract", l.Address.Hex())
	}

	_, err = q.Add721Asset(ctx, db.Add721AssetParams{
		AssetID: contractType[chain.ID][l.Address.Hex()].ID,
		ChainID: chain.ID,
		TokenID: event.TokenID.String(),
		Owner:   owner.Hex(),
		Attributes: sql.NullString{
			String: uri,
			Valid:  true,
		},
	})
	if err != nil {
		return err
	}

	return nil
}

func handleErc1155TransferBatch(ctx context.Context, sugar *zap.SugaredLogger, q *db.Queries, client *ethclient.Client,
	chain *db.Chain, rc *redis.Client, l *utypes.Log) error {
	// Decode TransferBatch log
	var event utils.Erc1155TransferBatchEvent
	err := utils.ERC1155ABI.UnpackIntoInterface(&event, "TransferBatch", l.Data)
	if err != nil {
		sugar.Errorw("Failed to unpack TransferBatch log:", "err", err)
	}

	// Decode the indexed fields for TransferBatch
	event.Operator = common.BytesToAddress(l.Topics[1].Bytes())
	event.From = common.BytesToAddress(l.Topics[2].Bytes())
	event.To = common.BytesToAddress(l.Topics[3].Bytes())

	for i := range event.Ids {
		amount := event.Values[i].String()
		_, err := q.AddOnchainTransaction(ctx, db.AddOnchainTransactionParams{
			From:      event.From.Hex(),
			To:        event.To.Hex(),
			AssetID:   contractType[chain.ID][l.Address.Hex()].ID,
			TokenID:   event.Ids[i].String(),
			Amount:    amount,
			TxHash:    l.TxHash.Hex(),
			Timestamp: time.Now(),
		})
		if err != nil {
			return err
		}

		uri, err := getErc1155TokenURI(ctx, sugar, client, rc, chain.Chain, &l.Address, event.Ids[i])
		if err != nil {
			sugar.Errorw("Failed to get ERC1155 token URI", "err", err, "tokenID", event.Ids[i])
			return err
		}

		// Update sender's balance
		balance, err := getErc1155BalanceOf(ctx, sugar, client, &l.Address, &event.From, event.Ids[i])
		if err != nil {
			sugar.Errorw("Failed to get ERC1155 balance", "err", err, "tokenID", event.Ids[i])
			return err
		}

		if err = q.Add1155Asset(ctx, db.Add1155AssetParams{
			AssetID: contractType[chain.ID][l.Address.Hex()].ID,
			ChainID: chain.ID,
			TokenID: event.Ids[i].String(),
			Owner:   event.From.Hex(),
			Balance: balance.String(),
			Attributes: sql.NullString{
				String: uri,
				Valid:  true,
			},
		}); err != nil {
			return err
		}

		// Update receiver's balance
		balance, err = getErc1155BalanceOf(ctx, sugar, client, &l.Address, &event.To, event.Ids[i])
		if err != nil {
			sugar.Errorw("Failed to get ERC1155 balance", "err", err, "tokenID", event.Ids[i])
			return err
		}

		if err = q.Add1155Asset(ctx, db.Add1155AssetParams{
			AssetID: contractType[chain.ID][l.Address.Hex()].ID,
			ChainID: chain.ID,
			TokenID: event.Ids[i].String(),
			Owner:   event.To.Hex(),
			Balance: balance.String(),
			Attributes: sql.NullString{
				String: uri,
				Valid:  true,
			},
		}); err != nil {
			return err
		}

	}

	return nil
}

func handleErc1155TransferSingle(ctx context.Context, sugar *zap.SugaredLogger, q *db.Queries, client *ethclient.Client,
	chain *db.Chain, rc *redis.Client, l *utypes.Log) error {

	// Decode TransferSingle log
	var event utils.Erc1155TransferSingleEvent
	err := utils.ERC1155ABI.UnpackIntoInterface(&event, "TransferSingle", l.Data)
	if err != nil {
		sugar.Errorw("Failed to unpack TransferSingle log:", "err", err)
	}

	// Decode the indexed fields for TransferSingle
	event.Operator = common.BytesToAddress(l.Topics[1].Bytes())
	event.From = common.BytesToAddress(l.Topics[2].Bytes())
	event.To = common.BytesToAddress(l.Topics[3].Bytes())

	amount := event.Value.String()
	_, err = q.AddOnchainTransaction(ctx, db.AddOnchainTransactionParams{
		From:      event.From.Hex(),
		To:        event.To.Hex(),
		AssetID:   contractType[chain.ID][l.Address.Hex()].ID,
		TokenID:   event.Id.String(),
		Amount:    amount,
		TxHash:    l.TxHash.Hex(),
		Timestamp: time.Now(),
	})
	if err != nil {
		return err
	}

	uri, err := getErc1155TokenURI(ctx, sugar, client, rc, chain.Chain, &l.Address, event.Id)
	if err != nil {
		sugar.Errorw("Failed to get ERC1155 token URI", "err", err, "tokenID", event.Id)
		return err
	}

	// Update Sender's balance
	balance, err := getErc1155BalanceOf(ctx, sugar, client, &l.Address, &event.From, event.Id)
	if err != nil {
		sugar.Errorw("Failed to get ERC1155 balance", "err", err, "tokenID", event.Id)
		return err
	}

	if err = q.Add1155Asset(ctx, db.Add1155AssetParams{
		AssetID: contractType[chain.ID][l.Address.Hex()].ID,
		ChainID: chain.ID,
		TokenID: event.Id.String(),
		Owner:   event.From.Hex(),
		Balance: balance.String(),
		Attributes: sql.NullString{
			String: uri,
			Valid:  true,
		},
	}); err != nil {
		return err
	}

	// Update Sender's balance
	balance, err = getErc1155BalanceOf(ctx, sugar, client, &l.Address, &event.To, event.Id)
	if err != nil {
		sugar.Errorw("Failed to get ERC1155 balance", "err", err, "tokenID", event.Id)
		return err
	}

	if err = q.Add1155Asset(ctx, db.Add1155AssetParams{
		AssetID: contractType[chain.ID][l.Address.Hex()].ID,
		ChainID: chain.ID,
		TokenID: event.Id.String(),
		Owner:   event.To.Hex(),
		Balance: balance.String(),
		Attributes: sql.NullString{
			String: uri,
			Valid:  true,
		},
	}); err != nil {
		return err
	}

	return nil
}

func getErc721TokenURI(ctx context.Context, sugar *zap.SugaredLogger, client *ethclient.Client, rc *redis.Client,
	chainName string, contractAddress *common.Address, tokenId *big.Int) (string, error) {
	hashKey := fmt.Sprintf("erc721_token_uri:%s", chainName)
	field := contractAddress.Hex()

	// Try to get from cache first
	rs, err := rc.HGet(ctx, hashKey, field).Result()
	if err == nil {
		return rs, nil
	} else if !errors.Is(err, redis.Nil) {
		sugar.Errorw("Failed to get ERC721 token URI from cache", "err", err)
	}

	// Prepare the function call data
	data, err := utils.ERC721ABI.Pack("tokenURI", tokenId)
	if err != nil {
		sugar.Errorf("Failed to pack data for tokenURI: %v", err)
		return "", err
	}

	// Call the contract
	msg := u2u.CallMsg{
		To:   contractAddress,
		Data: data,
	}

	// Execute the call
	result, err := client.CallContract(context.Background(), msg, nil)
	if err != nil {
		sugar.Errorf("Failed to call contract: %v", err)
		return "", err
	}

	// Unpack the result to get the token URI
	var tokenURI string
	err = utils.ERC721ABI.UnpackIntoInterface(&tokenURI, "tokenURI", result)
	if err != nil {
		sugar.Errorf("Failed to unpack tokenURI: %v", err)
		return "", err
	}

	// Cache the token URI
	err = rc.HSet(ctx, hashKey, field, tokenURI).Err()
	if err != nil {
		sugar.Warnw("Failed to cache token URI", "err", err)
		// Continue even if caching fails
	}

	// Set an expiration on the entire hash if this is the first entry
	if rc.Exists(ctx, hashKey).Val() == 1 {
		rc.Expire(ctx, hashKey, 24*time.Hour)
	}

	return tokenURI, nil
}

func getErc1155TokenURI(ctx context.Context, sugar *zap.SugaredLogger, client *ethclient.Client, rc *redis.Client,
	chainName string, contractAddress *common.Address, tokenId *big.Int) (string, error) {
	hashKey := fmt.Sprintf("erc1155_token_uri:%s", chainName)
	field := contractAddress.Hex()

	// Try to get from cache first
	rs, err := rc.HGet(ctx, hashKey, field).Result()
	if err == nil {
		return rs, nil
	} else if !errors.Is(err, redis.Nil) {
		sugar.Errorw("Failed to get ERC1155 token URI from cache", "err", err)
	}

	// Prepare the function call data
	data, err := utils.ERC1155ABI.Pack("uri", tokenId)

	if err != nil {
		sugar.Errorf("Failed to pack data for tokenURI: %v", err)
		return "", err
	}

	// Call the contract
	msg := u2u.CallMsg{
		To:   contractAddress,
		Data: data,
	}

	// Execute the call
	result, err := client.CallContract(context.Background(), msg, nil)
	if err != nil {
		sugar.Errorf("Failed to call contract: %v", err)
		return "", err
	}

	// Unpack the result to get the token URI
	var tokenURI string
	err = utils.ERC1155ABI.UnpackIntoInterface(&tokenURI, "uri", result)
	if err != nil {
		sugar.Errorf("Failed to unpack tokenURI: %v", err)
		return "", err
	}
	// Replace {id} in the URI template with the actual token ID in hexadecimal form
	tokenIDHex := fmt.Sprintf("%x", tokenId)
	tokenURI = replaceTokenIDPlaceholder(tokenURI, tokenIDHex)

	// Cache the token URI
	err = rc.HSet(ctx, hashKey, field, tokenURI).Err()
	if err != nil {
		sugar.Warnw("Failed to cache token URI", "err", err)
		// Continue even if caching fails
	}

	// Set an expiration on the entire hash if this is the first entry
	if rc.Exists(ctx, hashKey).Val() == 1 {
		rc.Expire(ctx, hashKey, 24*time.Hour)
	}
	return tokenURI, nil
}

// replaceTokenIDPlaceholder replaces the "{id}" placeholder with the actual token ID in hexadecimal
func replaceTokenIDPlaceholder(uriTemplate, tokenIDHex string) string {
	return strings.ReplaceAll(uriTemplate, "{id}", tokenIDHex)
}

func retrieveNftMetadata(tokenURI string) ([]byte, error) {
	res, err := http.Get(tokenURI)
	if err != nil {
		return nil, err
	}
	return io.ReadAll(res.Body)
}

func getErc20BalanceOf(ctx context.Context, sugar *zap.SugaredLogger, client *ethclient.Client,
	contractAddress *common.Address, ownerAddress *common.Address) (*big.Int, error) {

	// Prepare the function call data
	data, err := utils.ERC20ABI.Pack("balanceOf", ownerAddress)
	if err != nil {
		sugar.Errorf("Failed to pack data for balanceOf: %v", err)
		return nil, err
	}

	// Call the contract
	msg := u2u.CallMsg{
		To:   contractAddress,
		Data: data,
	}

	// Execute the call
	result, err := client.CallContract(context.Background(), msg, nil)
	if err != nil {
		sugar.Errorf("Failed to call contract: %v", err)
		return nil, err
	}

	// Unpack the result to get the balance
	var balance *big.Int
	err = utils.ERC20ABI.UnpackIntoInterface(&balance, "balanceOf", result)
	if err != nil {
		sugar.Errorf("Failed to unpack balanceOf: %v", err)
		return nil, err
	}

	return balance, nil
}

func getErc721OwnerOf(ctx context.Context, sugar *zap.SugaredLogger, client *ethclient.Client,
	contractAddress *common.Address, tokenId *big.Int) (common.Address, error) {

	// Prepare the function call data
	data, err := utils.ERC721ABI.Pack("ownerOf", tokenId)
	if err != nil {
		sugar.Errorf("Failed to pack data for balanceOf: %v", err)
		return common.Address{}, err
	}

	// Call the contract
	msg := u2u.CallMsg{
		To:   contractAddress,
		Data: data,
	}

	// Execute the call
	result, err := client.CallContract(context.Background(), msg, nil)
	if err != nil {
		sugar.Errorf("Failed to call contract: %v", err)
		return common.Address{}, err
	}

	// Unpack the result to get the balance
	var owner common.Address
	err = utils.ERC721ABI.UnpackIntoInterface(&owner, "ownerOf", result)

	if err != nil {
		sugar.Errorf("Failed to unpack ownerOf: %v", err)
		return common.Address{}, err
	}

	return owner, nil
}

func getErc1155BalanceOf(ctx context.Context, sugar *zap.SugaredLogger, client *ethclient.Client,
	contractAddress *common.Address, ownerAddress *common.Address, tokenId *big.Int) (*big.Int, error) {
	// Prepare the function call data
	data, err := utils.ERC1155ABI.Pack("balanceOf", ownerAddress, tokenId)
	if err != nil {
		sugar.Errorf("Failed to pack data for balanceOf: %v", err)
		return nil, err
	}

	// Call the contract
	msg := u2u.CallMsg{
		To:   contractAddress,
		Data: data,
	}

	// Execute the call
	result, err := client.CallContract(context.Background(), msg, nil)
	if err != nil {
		sugar.Errorf("Failed to call contract: %v", err)
		return nil, err
	}

	// Unpack the result to get the balance
	var balance *big.Int
	err = utils.ERC1155ABI.UnpackIntoInterface(&balance, "balanceOf", result)
	if err != nil {
		sugar.Errorf("Failed to unpack balanceOf: %v", err)
		return nil, err
	}

	return balance, nil
}

func handleFillOrder(ctx context.Context, sugar *zap.SugaredLogger, q *db.Queries, client *ethclient.Client,
	chain *db.Chain, rc *redis.Client, l *utypes.Log, ts uint64) error {

	// Decode FillOrder log
	// FillOrder(uint8 orderType, address maker, address taker, bytes sig, uint16 index, uint256 takeQty, uint256 takeAssetId, uint256 currentFilledValue, uint256 randomValue)
	var event utils.FillOrderEvent
	err := utils.EXCHANGEABI.UnpackIntoInterface(&event, "FillOrder", l.Data)
	if err != nil {
		sugar.Errorw("Failed to unpack FillOrder log:", "err", err)
		return err
	}

	// insert maker, taker accounts
	_, err = q.GetOrCreateEvmAccount(ctx, event.Maker.String())
	if err != nil {
		sugar.Errorw("Failed to get or create maker account", "err", err)
		return err
	}

	if event.Taker.String() != "0x0000000000000000000000000000000000000000" {
		_, err = q.GetOrCreateEvmAccount(ctx, event.Taker.String())
		if err != nil {
			sugar.Errorw("Failed to get or create taker account", "err", err)
			return err
		}
	}

	// insert order
	order, err := q.CreateOrderAsset(ctx, db.CreateOrderAssetParams{
		Maker:     event.Maker.String(),
		Taker:     sql.NullString{String: event.Taker.String(), Valid: true},
		Sig:       fmt.Sprintf("%x", event.Sig),
		Index:     int32(event.Index),
		Status:    db.OrderStatusFILLED,
		TakeQty:   event.TakeQty.String(),
		FilledQty: event.FilledQty.String(),
		Nonce:     event.RandomValue.String(),
		Timestamp: time.Unix(int64(ts), 0),
		Remaining: "0",
		AssetID:   contractType[chain.ID][l.Address.Hex()].ID,
		ChainID:   contractType[chain.ID][l.Address.Hex()].ChainID,
		TxHash:    l.TxHash.Hex(),
	})
	if err != nil {
		sugar.Errorw("Failed to create order", "err", err)
		return err
	}

	// publish to redis queue
	if rc != nil {
		go func() {
			orderBytes, _ := json.Marshal(order)
			_ = rc.Publish(ctx, config.FillOrderChannel, orderBytes).Err()
		}()
	}

	return nil
}

func handleCancelOrder(ctx context.Context, sugar *zap.SugaredLogger, q *db.Queries, client *ethclient.Client,
	chain *db.Chain, rc *redis.Client, l *utypes.Log, ts uint64) error {

	// Decode CancelOrder log
	var event utils.CancelOrderEvent
	err := utils.EXCHANGEABI.UnpackIntoInterface(&event, "CancelOrder", l.Data)
	if err != nil {
		sugar.Errorw("Failed to unpack CancelOrder log:", "err", err)
		return err
	}

	// insert maker, taker accounts
	_, err = q.GetOrCreateEvmAccount(ctx, event.Maker.String())
	if err != nil {
		sugar.Errorw("Failed to get or create maker account", "err", err)
		return err
	}

	// insert order
	order, err := q.CreateOrderAsset(ctx, db.CreateOrderAssetParams{
		Maker:     event.Maker.String(),
		Taker:     sql.NullString{String: "", Valid: false},
		Sig:       fmt.Sprintf("0x%x", event.Sig),
		Index:     int32(event.Index),
		Status:    db.OrderStatusCANCELED,
		TakeQty:   "0",
		FilledQty: "0",
		Nonce:     "0",
		Timestamp: time.Unix(int64(ts), 0),
		Remaining: "0",
		AssetID:   contractType[chain.ID][l.Address.Hex()].ID,
		ChainID:   contractType[chain.ID][l.Address.Hex()].ChainID,
		TxHash:    l.TxHash.Hex(),
	})
	if err != nil {
		sugar.Errorw("Failed to create order", "err", err)
		return err
	}

	// publish to redis queue
	if rc != nil {
		go func() {
			orderBytes, _ := json.Marshal(order)
			_ = rc.Publish(ctx, config.CancelOrderChannel, orderBytes).Err()
		}()
	}

	return nil
}

func handleExchangeBackfill(ctx context.Context, sugar *zap.SugaredLogger, q *db.Queries, client *ethclient.Client,
	chain *db.Chain, logs []utypes.Log, rc *redis.Client) error {
	if len(logs) == 0 {
		return nil
	}

	for _, l := range logs {
		blk, err := helpers.RetryWithBackoff(ctx, sugar, config.MaxRetriesAttempt, "fetch latest block number", func() (*utypes.Block, error) {
			return client.BlockByNumber(ctx, big.NewInt(int64(l.BlockNumber)))
		})
		//blk, err := client.BlockByNumber(ctx, big.NewInt(int64(l.BlockNumber)))
		if err != nil {
			sugar.Errorw("Failed to get block by number", "err", err)
			return err
		}
		ts := blk.Time()

		switch l.Topics[0].Hex() {
		case utils.FillOrderSig:
			_ = handleFillOrder(ctx, sugar, q, client, chain, rc, &l, ts)
		case utils.CancelOrderSig:
			_ = handleCancelOrder(ctx, sugar, q, client, chain, rc, &l, ts)
		}
	}

	return nil
}
