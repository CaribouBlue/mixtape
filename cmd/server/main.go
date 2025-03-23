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
		log.SetDefaultLogger(log.Logger.Output(zerolog.ConsoleWriter{Out: os.Stderr}))
		log.Logger.Warn().Msg("Using development mode")
	}

	log.Logger.Info().Str("address", s.Addr).Msg("Starting server")
	if err := s.ListenAndServe(); err != nil {
		log.Logger.Fatal().Err(err).Msg("Failed to start server")
	}
}
