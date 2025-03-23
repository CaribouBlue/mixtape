package config

import (
	"os"
	"strings"

	"github.com/CaribouBlue/mixtape/internal/log"

	"github.com/joho/godotenv"
)

type Env = string

const (
	EnvProduction  Env = "production"
	EnvDevelopment Env = "development"
)

type ConfigProperty struct {
	key          string
	defaultValue string
	isRequired   bool
	isSecret     bool
	validate     func(string) bool
}

type ConfigPropertyOpt func(*ConfigProperty) *ConfigProperty

func withDefaultValue(defaultValue string) ConfigPropertyOpt {
	return func(prop *ConfigProperty) *ConfigProperty {
		prop.defaultValue = defaultValue
		prop.isRequired = false
		return prop
	}
}

func withIsRequired(isRequired bool) ConfigPropertyOpt {
	return func(prop *ConfigProperty) *ConfigProperty {
		prop.isRequired = isRequired
		requiredConfigProperties = append(requiredConfigProperties, prop)
		return prop
	}
}

func withValidation(validate func(string) bool) ConfigPropertyOpt {
	return func(prop *ConfigProperty) *ConfigProperty {
		prop.validate = validate
		unvalidatedConfigProperties = append(unvalidatedConfigProperties, prop)
		return prop
	}
}

func newConfigProperty(key string, isSecret bool, opts ...ConfigPropertyOpt) ConfigProperty {
	prop := ConfigProperty{key: key, isSecret: isSecret, validate: func(string) bool { return true }}
	for _, opt := range opts {
		opt(&prop)
	}
	return prop
}

var (
	ConfAccessCode          ConfigProperty = newConfigProperty("ACCESS_CODE", true)
	ConfDbPath              ConfigProperty = newConfigProperty("DB_PATH", false, withIsRequired(true))
	ConfGmailUsername       ConfigProperty = newConfigProperty("GMAIL_USERNAME", true)
	ConfGmailPassword       ConfigProperty = newConfigProperty("GMAIL_PASSWORD", true)
	ConfHost                ConfigProperty = newConfigProperty("HOST", false)
	ConfPort                ConfigProperty = newConfigProperty("PORT", false)
	ConfAppDataPath         ConfigProperty = newConfigProperty("APP_DATA_PATH", false, withDefaultValue("."))
	ConfSpotifyClientId     ConfigProperty = newConfigProperty("SPOTIFY_CLIENT_ID", true, withIsRequired(true))
	ConfSpotifyClientSecret ConfigProperty = newConfigProperty("SPOTIFY_CLIENT_SECRET", true, withIsRequired(true))
	ConfSpotifyRedirectUri  ConfigProperty = newConfigProperty("SPOTIFY_REDIRECT_URI", false, withIsRequired(true))
	ConfSpotifyScope        ConfigProperty = newConfigProperty("SPOTIFY_SCOPE", false, withIsRequired(true))
	ConfEnvFiles            ConfigProperty = newConfigProperty("ENV_FILES", false)
	ConfEnv                 ConfigProperty = newConfigProperty("ENV", false, withDefaultValue(EnvProduction), withValidation(func(value string) bool {
		return value == EnvProduction || value == EnvDevelopment
	}))
)

var requiredConfigProperties = []*ConfigProperty{}
var unvalidatedConfigProperties = []*ConfigProperty{}

func Load() error {
	err := godotenv.Load()
	if err != nil {
		log.Logger().Warn().Msg("Unable to load default .env file")
	} else {
		log.Logger().Debug().Msg("Loaded default .env")
	}

	envFilesConfig := GetConfigValue(ConfEnvFiles)
	if envFilesConfig != "" {
		envFiles := strings.Split(envFilesConfig, ",")
		err := godotenv.Load(envFiles...)
		if err != nil {
			log.Logger().Fatal().Err(err).Msg("Error loading additional .env files")
		} else {
			log.Logger().Debug().Str("files", envFilesConfig).Msg("Loaded additional .env files")
		}
	}

	for _, prop := range requiredConfigProperties {
		if GetConfigValue(*prop) == "" {
			log.Logger().Fatal().Str("property", prop.key).Msg("Missing required config property")
		}
	}

	for _, prop := range unvalidatedConfigProperties {
		isValid := prop.validate(GetConfigValue(*prop))
		if !isValid {
			log.Logger().Fatal().Str("property", prop.key).Msg("Invalid config property value")
		}
	}

	return nil
}

func GetConfigValue(prop ConfigProperty) string {
	val := os.Getenv(string(prop.key))
	if val == "" {
		val = prop.defaultValue
	}
	return val
}
