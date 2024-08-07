package testutils

import (
	"github.com/basemind-ai/monorepo/internal/utils/rediscache"
	"github.com/go-redis/cache/v9"
	"github.com/go-redis/redismock/v9"
	"github.com/redis/go-redis/v9"
	"testing"
)

func CreateMockRedisClient(t *testing.T) (*redis.Client, redismock.ClientMock) {
	t.Helper()

	db, mockRedis := redismock.NewClientMock()
	rediscache.SetClient(cache.New(&cache.Options{
		Redis: db,
	}))

	t.Cleanup(func() {
		_ = db.Close()
	})

	return db, mockRedis
}
