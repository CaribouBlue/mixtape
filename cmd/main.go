package main

import (
	"log"

	"github.com/CaribouBlue/top-spot/db"
	"github.com/CaribouBlue/top-spot/server"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	db.InitGlobal()
	server.StartServer()
}
