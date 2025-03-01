package config

import (
	"os"

	"github.com/joho/godotenv"
)

type ConfigProperty struct {
	Key          string
	DefaultValue string
}

var (
	ConfDockerContext ConfigProperty = ConfigProperty{"DOCKER_CONTEXT", "default"}
	ConfDbPath        ConfigProperty = ConfigProperty{"DB_PATH", ""}
)

var isLoaded bool = false

func Load() error {
	isLoaded = true
	return godotenv.Load(".env.cli")
}

func GetConfigValue(prop ConfigProperty) string {
	// TODO: Fix this
	if !isLoaded {
		Load()
		isLoaded = true
	}

	val := os.Getenv(string(prop.Key))
	if val == "" {
		val = prop.DefaultValue
	}
	return val
}
