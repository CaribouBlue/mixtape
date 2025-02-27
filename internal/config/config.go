package config

import (
	"log"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

type ConfigProperty struct {
	key          string
	defaultValue string
	isRequired   bool
	isSecret     bool
}

type ConfigPropertyOpt func(*ConfigProperty) *ConfigProperty

func withDefaultValue(defaultValue string) ConfigPropertyOpt {
	return func(prop *ConfigProperty) *ConfigProperty {
		prop.defaultValue = defaultValue
		prop.isRequired = false
		return prop
	}
}

func isRequired(prop *ConfigProperty) *ConfigProperty {
	prop.isRequired = true
	requiredConfigProperties = append(requiredConfigProperties, prop)
	return prop
}

func newConfigProperty(key string, isSecret bool, opts ...ConfigPropertyOpt) ConfigProperty {
	prop := ConfigProperty{key: key, isSecret: isSecret}
	for _, opt := range opts {
		opt(&prop)
	}
	return prop
}

var (
	ConfAccessCode          ConfigProperty = newConfigProperty("ACCESS_CODE", true)
	ConfDbPath              ConfigProperty = newConfigProperty("DB_PATH", false, isRequired)
	ConfGmailUsername       ConfigProperty = newConfigProperty("GMAIL_USERNAME", true)
	ConfGmailPassword       ConfigProperty = newConfigProperty("GMAIL_PASSWORD", true)
	ConfHost                ConfigProperty = newConfigProperty("HOST", false)
	ConfPort                ConfigProperty = newConfigProperty("PORT", false)
	ConfAppDataPath         ConfigProperty = newConfigProperty("APP_DATA_PATH", false, withDefaultValue("."))
	ConfSpotifyClientId     ConfigProperty = newConfigProperty("SPOTIFY_CLIENT_ID", true, isRequired)
	ConfSpotifyClientSecret ConfigProperty = newConfigProperty("SPOTIFY_CLIENT_SECRET", true, isRequired)
	ConfSpotifyRedirectUri  ConfigProperty = newConfigProperty("SPOTIFY_REDIRECT_URI", false, isRequired)
	ConfSpotifyScope        ConfigProperty = newConfigProperty("SPOTIFY_SCOPE", false, isRequired)
	ConfEnvFiles            ConfigProperty = newConfigProperty("ENV_FILES", false)
)

var requiredConfigProperties = []*ConfigProperty{}

func Load() error {
	err := godotenv.Load()
	if err != nil {
		log.Default().Println("WARN | Unable to load default .env file: ", err)
	} else {
		log.Default().Println("INFO | Loaded default .env")
	}

	envFilesConfig := GetConfigValue(ConfEnvFiles)
	if envFilesConfig != "" {
		envFiles := strings.Split(envFilesConfig, ",")
		err := godotenv.Load(envFiles...)
		if err != nil {
			log.Fatalln("Error loading .env files: ", err)
		} else {
			log.Default().Println("INFO | Loaded additional .env files:", envFilesConfig)
		}
	}

	for _, prop := range requiredConfigProperties {
		if GetConfigValue(*prop) == "" {
			log.Fatalln("Missing required config property:", prop.key)
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
