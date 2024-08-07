package logging

import (
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"os"
	"time"
)

// Configure configures the logger to print to console when cfg.Debug is true,
// otherwise use the zerolog structured logging.
func Configure(isDebug bool) {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	if isDebug {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
		log.Logger = zerolog.New(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339}).
			With().
			Timestamp().Caller().
			Logger()
	} else {
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
		log.Logger = zerolog.New(os.Stderr).With().Timestamp().Caller().Logger()
	}
}
