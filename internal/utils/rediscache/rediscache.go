package rediscache

import (
	"context"
	"fmt"
	"github.com/basemind-ai/monorepo/internal/utils/exc"
	"github.com/go-redis/cache/v9"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog/log"
	"os"
	"sync"
	"time"
)

var (
	once   sync.Once
	client *cache.Cache
)

// SetClient is a helper function that allows you to set the redis client. This is useful for testing.
func SetClient(c *cache.Cache) {
	client = c
}

// New is a helper function that will initialize the redis client.
// It will only initialize the client once, and subsequent calls will return the same client.
func New(redisURL string) *cache.Cache {
	once.Do(func() {
		opt := exc.MustResult(redis.ParseURL(redisURL))
		opt.ClientName = os.Getenv("SERVICE_NAME")
		opt.OnConnect = func(ctx context.Context, cn *redis.Conn) error {
			if err := cn.Ping(ctx).Err(); err != nil {
				log.Fatal().Err(err).Msg("failed to ping redis")
			}
			log.Info().Msg("connected to redis")
			return nil
		}
		SetClient(cache.New(&cache.Options{
			Redis:      redis.NewClient(opt),
			LocalCache: cache.NewTinyLFU(1000, time.Minute),
		}))
	})
	return client
}

func GetClient() *cache.Cache {
	return exc.ReturnNotNil(client, "redis client is not initialized")
}

// With is a helper function that will check if a key exists in redis, and if it does, it will return the value. If it
// does not exist, it will call the fallback function, set the value in redis, and return the value.
func With[T any](
	ctx context.Context,
	key string,
	target *T,
	ttl time.Duration,
	fallback func() (*T, error),
) (*T, error) {
	if getErr := client.Get(ctx, key, target); getErr == nil {
		return target, nil
	}

	retrieved, retrieveErr := fallback()
	if retrieveErr != nil {
		return nil, fmt.Errorf("failed to retrieve value from database: %w", retrieveErr)
	}

	if setErr := client.Set(&cache.Item{
		Ctx:   ctx,
		Key:   key,
		Value: *retrieved,
		TTL:   ttl,
	}); setErr != nil {
		return nil, fmt.Errorf("failed to set value in Redis: %w", setErr)
	}

	return retrieved, nil
}

// Invalidate - deletes all the passed in keys from cache.
func Invalidate(ctx context.Context, keys ...string) {
	for _, key := range keys {
		exc.LogIfErr(
			client.Delete(ctx, key),
			fmt.Sprintf("failed to delete key from cache: %s", key),
		)
	}
}
