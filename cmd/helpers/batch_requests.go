package helpers

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"math/big"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/u2u-labs/layerg-crawler/config"
	"go.uber.org/zap"
	"golang.org/x/time/rate"
)

func FetchBlockAndReceipts(client *rpc.Client, sugar *zap.SugaredLogger, ctx context.Context, number *big.Int) (*types.Block, []*types.Receipt, error) {
	var (
		rawBlock       json.RawMessage
		receiptsResult []*types.Receipt
	)

	batch := []rpc.BatchElem{
		{
			Method: "eth_getBlockByNumber",
			Args:   []any{hexutil.EncodeBig(number), true},
			Result: &rawBlock,
		},
		{
			Method: "eth_getBlockReceipts",
			Args:   []any{hexutil.EncodeBig(number)},
			Result: &receiptsResult,
		},
	}

	err := BatchCallRpcClient(ctx, sugar, batch, client)
	if err != nil {
		return nil, nil, err
	}

	if batch[0].Error != nil {
		return nil, nil, batch[0].Error
	}
	if batch[1].Error != nil {
		return nil, nil, batch[1].Error
	}

	// Use the proper decoder
	var header *types.Header
	var body *types.Body

	// Decode the header and body separately
	if err := json.Unmarshal(rawBlock, &header); err != nil {
		return nil, nil, fmt.Errorf("failed to decode block header: %w", err)
	}

	if err := json.Unmarshal(rawBlock, &body); err != nil {
		return nil, nil, fmt.Errorf("failed to decode block body: %w", err)
	}

	// Create the block
	blockResult := types.NewBlockWithHeader(header).WithBody(types.Body{Transactions: body.Transactions, Uncles: body.Uncles})

	return blockResult, receiptsResult, nil
}

func BatchCallRpcClient(ctx context.Context, sugar *zap.SugaredLogger, calls []rpc.BatchElem, rpcClient *rpc.Client) error {
	limiter := rate.NewLimiter(5, 10) // 5 requests/sec, burst of 10
	batchSize := 25
	maxRetries := config.MaxRetriesAttempt
	baseDelay := 500 * time.Millisecond

	for i := 0; i < len(calls); i += batchSize {
		end := i + batchSize
		if end > len(calls) {
			end = len(calls) // Prevent out-of-bounds error
		}

		// Rate limit enforcement
		if err := limiter.Wait(ctx); err != nil {
			if !strings.Contains(err.Error(), "context deadline exceeded") {
				sugar.Errorw("Failed to wait for rate limit", "err", err)
			}
			return err
		}

		// Retry loop with exponential backoff
		for attempt := 0; attempt < maxRetries; attempt++ {
			err := rpcClient.BatchCallContext(ctx, calls[i:end])
			if err == nil {
				break // Success, move to the next batch
			}

			// Log error
			if attempt > 2 {
				sugar.Warnf("Batch call failed %s %d", "attempt", attempt+1)
			}

			// Check if it's a 429 rate limit error
			if strings.Contains(err.Error(), "too many requests") || strings.Contains(err.Error(), "429") {
				delay := time.Duration(math.Pow(2, float64(attempt))) * baseDelay // Exponential backoff
				if attempt > 2 {
					sugar.Warnw("Rate limit hit, retrying after delay", "delay", delay)
				}

				select {
				case <-ctx.Done():
					return fmt.Errorf("context canceled: %w", ctx.Err())
				case <-time.After(delay):
					continue
				}
			}
			if !strings.Contains(err.Error(), "context deadline exceeded") {
				sugar.Errorf("Batch call failed %s %s", "err", err)
			}

			// If it's not a 429 error, return immediately
			return err
		}
	}

	return nil
}

// RetryWithBackoff executes the provided function with exponential backoff on rate limit errors
// It returns the result of type T and any error encountered
func RetryWithBackoff[T any](
	ctx context.Context,
	sugar *zap.SugaredLogger,
	maxAttempts int,
	operation string,
	fn func() (T, error),
) (T, error) {
	var result T
	var err error

	for attempt := 0; attempt < maxAttempts; attempt++ {
		result, err = fn()
		if err == nil {
			return result, nil
		}

		if strings.Contains(err.Error(), "too many requests") || strings.Contains(err.Error(), "429") {
			delay := time.Duration(math.Pow(2, float64(attempt))) * time.Second
			if attempt > 2 {
				sugar.Warnw(fmt.Sprintf("Rate limit hit, retrying after delay for %s", operation), "delay", delay)
			}

			select {
			case <-ctx.Done():
				return result, ctx.Err()
			case <-time.After(delay):
				continue
			}
		}

		sugar.Errorw(fmt.Sprintf("Failed to %s", operation), "err", err)
		return result, err
	}

	return result, fmt.Errorf("max retry attempts reached")
}
