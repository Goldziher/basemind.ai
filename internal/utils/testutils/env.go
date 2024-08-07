package testutils

import (
	"github.com/basemind-ai/monorepo/e2e/factories"
	"os"
	"testing"
	"time"
)

func SetTestEnv(t *testing.T) {
	t.Helper()
	t.Setenv("DATABASE_URL", "postgresql://basemind:basemind@db:5432/basemind")
	t.Setenv("ENVIRONMENT", "development")
	t.Setenv("FRONTEND_BASE_URL", "http://localhost:3000")
	t.Setenv("GCP_PROJECT_ID", "basemind-ai-development")
	t.Setenv("JWT_SECRET", factories.RandomString(24))
	t.Setenv("REDIS_CONNECTION_STRING", "redis://redis:6379")
	t.Setenv("SERVER_HOST", "localhost")
	t.Setenv("SERVER_PORT", "3000")
	t.Setenv("URL_SIGNING_SECRET", factories.RandomString(10))
	t.Setenv("CRYPTO_PASS_KEY", factories.RandomString(32))
}

func UnsetTestEnv(t *testing.T) {
	t.Helper()
	_ = os.Unsetenv("DATABASE_URL")
	_ = os.Unsetenv("ENVIRONMENT")
	_ = os.Unsetenv("FRONTEND_BASE_URL")
	_ = os.Unsetenv("GCP_PROJECT_ID")
	_ = os.Unsetenv("JWT_SECRET")
	_ = os.Unsetenv("REDIS_CONNECTION_STRING")
	_ = os.Unsetenv("SERVER_HOST")
	_ = os.Unsetenv("SERVER_PORT")
	_ = os.Unsetenv("URL_SIGNING_SECRET")
	_ = os.Unsetenv("CRYPTO_PASS_KEY")
}

func GetSleepTimeout() time.Duration {
	sleepTimeout, err := time.ParseDuration(os.Getenv("TEST_SLEEP_TIMEOUT"))
	if err != nil {
		return 100 * time.Millisecond
	}

	return sleepTimeout
}
