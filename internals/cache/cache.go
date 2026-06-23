package cache

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/redis/go-redis/v9"
)

const DefaultTTL = 10 * time.Minute

func GetFromCache[T any](ctx context.Context, rdb *redis.Client, rkey string, dst *T) error {
	if rdb == nil {
		return redis.Nil
	}

	data, err := rdb.Get(ctx, rkey).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return redis.Nil
		}
		return err
	}
	return json.Unmarshal([]byte(data), dst)
}

func SaveToCache(ctx context.Context, rdb *redis.Client, rkey string, data any, ttl ...time.Duration) error {
	if rdb == nil {
		return nil
	}

	expiry := DefaultTTL
	if len(ttl) > 0 {
		expiry = ttl[0]
	}
	b, err := json.Marshal(data)
	if err != nil {
		return err
	}
	return rdb.Set(ctx, rkey, b, expiry).Err()
}

func DelFromCache(ctx context.Context, rdb *redis.Client, rkeys ...string) error {
	if rdb == nil || len(rkeys) == 0 {
		return nil
	}

	if err := rdb.Del(ctx, rkeys...).Err(); err != nil {
		return err
	}
	return nil
}
