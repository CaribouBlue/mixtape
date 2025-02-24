package main

import (
	"log"
	"os"
	"strings"

	"github.com/CaribouBlue/mixtape/internal/server"
	"github.com/joho/godotenv"
)

func main() {
	envFilesStr := os.Getenv("ENV_FILES")
	if envFilesStr != "" {
		envFiles := strings.Split(envFilesStr, ",")
		err := godotenv.Load(envFiles...)
		if err != nil {
			log.Fatal("Error loading .env files: ", err)
		}
	} else {
		err := godotenv.Load()
		if err != nil {
			log.Default().Println("WARN | Error loading .env file: ", err)
		}
	}

	server.StartServer()
}
