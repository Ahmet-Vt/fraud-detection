package cache

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

const (
	LastLocationTTL            = 24 * time.Hour
	Last24HoursAverageSpendTTL = 24 * time.Hour
	RequestCountTTL            = 60 * time.Second
)

type TransactionStateCache struct {
	client redis.Cmdable
}

func NewTransactionStateCache(client redis.Cmdable) *TransactionStateCache {
	return &TransactionStateCache{client: client}
}

func (c *TransactionStateCache) GetLastLocationByUserID(ctx context.Context, userID uuid.UUID) (string, bool, error) {
	if c == nil || c.client == nil {
		return "", false, fmt.Errorf("transaction state cache is nil")
	}

	location, err := c.client.Get(ctx, lastLocationCacheKey(userID)).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return "", false, nil
		}
		return "", false, err
	}

	return location, true, nil
}

func (c *TransactionStateCache) SetLastLocationByUserID(ctx context.Context, userID uuid.UUID, location string) error {
	if c == nil || c.client == nil {
		return fmt.Errorf("transaction state cache is nil")
	}

	return c.client.Set(ctx, lastLocationCacheKey(userID), location, LastLocationTTL).Err()
}

func (c *TransactionStateCache) GetLastTransactionCreatedAtByUserID(ctx context.Context, userID uuid.UUID) (time.Time, bool, error) {
	if c == nil || c.client == nil {
		return time.Time{}, false, fmt.Errorf("transaction state cache is nil")
	}

	createdAtRaw, err := c.client.Get(ctx, lastTransactionCreatedAtCacheKey(userID)).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return time.Time{}, false, nil
		}
		return time.Time{}, false, err
	}

	createdAt, err := time.Parse(time.RFC3339Nano, createdAtRaw)
	if err != nil {
		return time.Time{}, false, fmt.Errorf("parse last transaction created at: %w", err)
	}

	return createdAt, true, nil
}

func (c *TransactionStateCache) SetLastTransactionCreatedAtByUserID(ctx context.Context, userID uuid.UUID, createdAt time.Time) error {
	if c == nil || c.client == nil {
		return fmt.Errorf("transaction state cache is nil")
	}

	return c.client.Set(
		ctx,
		lastTransactionCreatedAtCacheKey(userID),
		createdAt.UTC().Format(time.RFC3339Nano),
		LastLocationTTL,
	).Err()
}

func (c *TransactionStateCache) GetLast24HoursAverageSpendByUserID(ctx context.Context, userID uuid.UUID) (int64, bool, error) {
	if c == nil || c.client == nil {
		return 0, false, fmt.Errorf("transaction state cache is nil")
	}

	averageRaw, err := c.client.Get(ctx, last24HoursAverageSpendCacheKey(userID)).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return 0, false, nil
		}
		return 0, false, err
	}

	average, err := strconv.ParseInt(averageRaw, 10, 64)
	if err != nil {
		return 0, false, fmt.Errorf("parse last 24 hours average spend: %w", err)
	}

	return average, true, nil
}

func (c *TransactionStateCache) SetLast24HoursAverageSpendByUserID(ctx context.Context, userID uuid.UUID, average int64) error {
	if c == nil || c.client == nil {
		return fmt.Errorf("transaction state cache is nil")
	}

	return c.client.Set(
		ctx,
		last24HoursAverageSpendCacheKey(userID),
		strconv.FormatInt(average, 10),
		Last24HoursAverageSpendTTL,
	).Err()
}

func (c *TransactionStateCache) IncrementAndGetRequestCountByUserID(ctx context.Context, userID uuid.UUID, requestTime time.Time) (int, error) {
	if c == nil || c.client == nil {
		return 0, fmt.Errorf("transaction state cache is nil")
	}

	windowStartMs := requestTime.Add(-RequestCountTTL).UnixMilli()
	currentRequestMs := requestTime.UnixMilli()
	member := fmt.Sprintf("%d-%s", currentRequestMs, uuid.NewString())

	pipe := c.client.TxPipeline()
	pipe.ZRemRangeByScore(ctx, requestCountCacheKey(userID), "-inf", fmt.Sprintf("(%d", windowStartMs))
	pipe.ZAdd(ctx, requestCountCacheKey(userID), redis.Z{
		Score:  float64(currentRequestMs),
		Member: member,
	})
	countCmd := pipe.ZCard(ctx, requestCountCacheKey(userID))
	pipe.Expire(ctx, requestCountCacheKey(userID), RequestCountTTL+5*time.Second)

	if _, err := pipe.Exec(ctx); err != nil {
		return 0, err
	}

	return int(countCmd.Val()), nil
}

func (c *TransactionStateCache) GetRequestCountByUserID(ctx context.Context, userID uuid.UUID, referenceTime time.Time) (int, error) {
	if c == nil || c.client == nil {
		return 0, fmt.Errorf("transaction state cache is nil")
	}

	windowStartMs := referenceTime.Add(-RequestCountTTL).UnixMilli()

	pipe := c.client.TxPipeline()
	pipe.ZRemRangeByScore(ctx, requestCountCacheKey(userID), "-inf", fmt.Sprintf("(%d", windowStartMs))
	countCmd := pipe.ZCard(ctx, requestCountCacheKey(userID))

	if _, err := pipe.Exec(ctx); err != nil {
		return 0, err
	}

	return int(countCmd.Val()), nil
}

func lastLocationCacheKey(userID uuid.UUID) string {
	return fmt.Sprintf("last_location:%s", userID)
}

func requestCountCacheKey(userID uuid.UUID) string {
	return fmt.Sprintf("request_count:%s", userID)
}

func lastTransactionCreatedAtCacheKey(userID uuid.UUID) string {
	return fmt.Sprintf("last_transaction_created_at:%s", userID)
}

func last24HoursAverageSpendCacheKey(userID uuid.UUID) string {
	return fmt.Sprintf("last_24h_average_spend:%s", userID)
}
