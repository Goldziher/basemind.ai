package config

import (
	"context"
	"github.com/basemind-ai/monorepo/shared/go/exc"
	"github.com/sethvargo/go-envconfig"
	"sync"
)

// Config - the shared configuration object.
//
//goland:noinspection GoUnnecessarilyExportedIdentifiers
type Config struct {
	DatabaseURL      string `env:"DATABASE_URL,required"`
	Environment      string `env:"ENVIRONMENT,default=test"`
	FrontendBaseURL  string `env:"FRONTEND_BASE_URL,required"`
	GcpProjectID     string `env:"GCP_PROJECT_ID,required"`
	JWTSecret        string `env:"JWT_SECRET,required"`
	RedisURL         string `env:"REDIS_CONNECTION_STRING,required"`
	ServerHost       string `env:"SERVER_HOST,required"`
	URLSigningSecret string `env:"URL_SIGNING_SECRET,required"`
	CryptoPassKey    string `env:"CRYPTO_PASS_KEY,required"`
	ServerPort       int    `env:"SERVER_PORT,required"`
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
