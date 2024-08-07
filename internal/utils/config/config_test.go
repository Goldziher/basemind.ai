package config_test

import (
	"context"
	"github.com/basemind-ai/monorepo/internal/utils/config"
	"github.com/basemind-ai/monorepo/internal/utils/testutils"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConfig(t *testing.T) {
	t.Run("panics if any env variables are missing", func(t *testing.T) {
		testutils.UnsetTestEnv(t)
		assert.Panics(t, func() {
			_ = config.Get(context.Background())
		})
	})
	t.Run("does not panic when all env vars are present", func(t *testing.T) {
		testutils.SetTestEnv(t)
		assert.NotPanics(t, func() {
			_ = config.Get(context.Background())
		})
	})
}
