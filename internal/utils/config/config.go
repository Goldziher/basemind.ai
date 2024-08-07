package config

import (
	"context"
	"github.com/basemind-ai/monorepo/internal/utils/exc"
	"github.com/sethvargo/go-envconfig"
	"sync"
)

// Config - the shared configuration object.
//
//goland:noinspection GoUnnecessarilyExportedIdentifiers
type Config struct {
	Environment      string `env:"ENVIRONMENT,default=test"`
	JWTSecret        string `env:"JWT_SECRET,required"`
	RedisURL         string `env:"REDIS_CONNECTION_STRING,required"`
	ServerPort       int    `env:"SERVER_PORT,required"`
	URLSigningSecret string `env:"URL_SIGNING_SECRET,required"`
	CryptoPassKey    string `env:"CRYPTO_PASS_KEY,required"`
}

var (
	config *Config
	once   sync.Once
)

// Get - returns the config object.
// Panics if the config is not initialized.
// This function is idempotent.
func Get(ctx context.Context) *Config {
	once.Do(func() {
		config = &Config{}
		exc.Must(envconfig.Process(ctx, config))
	})
	return config
}
