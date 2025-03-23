package log

import (
	"os"

	"github.com/rs/zerolog"
)

var Logger zerolog.Logger = zerolog.New(os.Stderr).With().Timestamp().Logger()

func SetDefaultLogger(logger zerolog.Logger) {
	Logger = logger
}
