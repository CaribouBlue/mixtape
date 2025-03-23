package log

import (
	"os"

	"github.com/rs/zerolog"
)

var defaultLogger zerolog.Logger = zerolog.New(os.Stderr).With().Timestamp().Logger()

func Logger() *zerolog.Logger {
	logger := defaultLogger
	return &logger
}

func SetDefaultLogger(logger zerolog.Logger) {
	defaultLogger = logger
}
