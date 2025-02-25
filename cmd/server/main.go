package main

import (
	"log"
	"os"
	"strings"

	"github.com/CaribouBlue/mixtape/internal/server"
	"github.com/joho/godotenv"
)

func main() {
	loadEnvFiles()
	s := server.NewServer()

	log.Default().Printf("Starting server at %s", s.Addr)
	if err := s.ListenAndServe(); err != nil {
		log.Fatalln("Error starting server:", err)
	}
}

func loadEnvFiles() {
	err := godotenv.Load()
	if err != nil {
		log.Default().Println("WARN | Unable to load default .env file: ", err)
	} else {
		log.Default().Println("INFO | Loaded default .env")
	}

	envFilesStr := os.Getenv("ENV_FILES")
	if envFilesStr != "" {
		envFiles := strings.Split(envFilesStr, ",")
		err := godotenv.Load(envFiles...)
		if err != nil {
			log.Fatalln("Error loading .env files: ", err)
		} else {
			log.Default().Println("INFO | Loaded additional .env files:", envFilesStr)
		}
	}
}
