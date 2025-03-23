package main

import (
	"os"

	"github.com/CaribouBlue/mixtape/internal/log"
	"github.com/rs/zerolog"

	"github.com/CaribouBlue/mixtape/internal/config"
	"github.com/CaribouBlue/mixtape/internal/server"
)

func main() {
	config.Load()
	s := server.NewServer()

	if config.GetConfigValue(config.ConfEnv) == config.EnvDevelopment {
		log.SetDefaultLogger(log.Logger().Output(zerolog.ConsoleWriter{Out: os.Stderr}))
		log.Logger().Warn().Msg("Using development mode")
	} else {
		logFile, _ := os.OpenFile(
			config.GetConfigValue(config.ConfLogFilePath),
			os.O_APPEND|os.O_CREATE|os.O_WRONLY,
			0664,
		)
		defer logFile.Close()

		multi := zerolog.MultiLevelWriter(os.Stdout, logFile)
		log.SetDefaultLogger(zerolog.New(multi).Level(zerolog.InfoLevel).With().Timestamp().Logger())
	}

	log.Logger().Info().Str("address", s.Addr).Msg("Starting server")
	if err := s.ListenAndServe(); err != nil {
		log.Logger().Fatal().Err(err).Msg("Failed to start server")
	}
}
