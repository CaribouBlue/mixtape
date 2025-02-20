package main

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	dbPath := os.Getenv("DB_PATH")
	e := os.Remove(dbPath)
	if e != nil && !os.IsNotExist(e) {
		log.Fatal(e)
	}
}
