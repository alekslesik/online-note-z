package logger

import (
	"os"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/pkgerrors"
)

var logger zerolog.Logger

// Create a global instance of the logger, and return it in a factory function
func New(level string) zerolog.Logger {
	// We are making sure that we can extract stack traces
	// from an error by setting the ErrorStackMarshaler.
	// This is important, as we want to trace the root of our errors
	zerolog.ErrorStackMarshaler = pkgerrors.MarshalStack

	// Setting the loglevel
	loglevel, err := zerolog.ParseLevel(level)
	if err != nil {
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	}

	zerolog.SetGlobalLevel(loglevel)

	logger = zerolog.New(os.Stdout).With().Timestamp().CallerWithSkipFrameCount(2).Logger()

	// Instance of the logger should be used in the Ctx() function of Zerolog
	zerolog.DefaultContextLogger = &logger

	return logger
}
